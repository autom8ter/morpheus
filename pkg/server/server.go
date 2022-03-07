package server

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/apollotracing"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/autom8ter/morpheus/pkg/api"
	"github.com/autom8ter/morpheus/pkg/auth"
	"github.com/autom8ter/morpheus/pkg/config"
	"github.com/autom8ter/morpheus/pkg/dataloader"
	"github.com/autom8ter/morpheus/pkg/graph"
	"github.com/autom8ter/morpheus/pkg/graph/generated"
	"github.com/autom8ter/morpheus/pkg/logger"
	"github.com/autom8ter/morpheus/pkg/raft"
	"golang.org/x/sync/errgroup"
	"net"
	"net/http"
	"time"
)

func Serve(ctx context.Context, g api.Graph, cfg *config.Config) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	if cfg.Server.GraphQLPort == 0 {
		cfg.Server.GraphQLPort = 8080
	}

	rlis, err := net.Listen("tcp", fmt.Sprintf(":%v", cfg.Server.RaftPort))
	if err != nil {
		return err
	}
	defer rlis.Close()
	glis, err := net.Listen("tcp", fmt.Sprintf(":%v", cfg.Server.GraphQLPort))
	if err != nil {
		return err
	}
	defer glis.Close()
	if cfg.Server.TLSCert != "" {
		cer, err := tls.X509KeyPair([]byte(cfg.Server.TLSCert), []byte(cfg.Server.TLSKey))
		if err != nil {
			return err
		}
		tlsConfig := &tls.Config{Certificates: []tls.Certificate{cer}}
		rlis = tls.NewListener(rlis, tlsConfig)
		glis = tls.NewListener(glis, tlsConfig)
	}
	joinRaft := cfg.Server.RaftCluster
	rft, err := raft.NewRaft(
		g,
		rlis,
		raft.WithRaftDir(fmt.Sprintf("%s/raft", cfg.Database.StoragePath)),
		raft.WithIsLeader(joinRaft == ""),
		raft.WithClusterSecret(cfg.Server.RaftSecret),
	)
	if err != nil {
		return err
	}
	resolver := graph.NewResolver(g, rft)
	defer func() {
		if err := resolver.Close(); err != nil {
			logger.L.Error("failed to close graphql resolver", map[string]interface{}{
				"error": err,
			})
		}
	}()
	schema := generated.NewExecutableSchema(generated.Config{
		Resolvers:  resolver,
		Directives: generated.DirectiveRoot{},
		Complexity: generated.ComplexityRoot{},
	})

	srv := handler.NewDefaultServer(schema)
	srv.SetQueryCache(lru.New(1000))
	mux := http.NewServeMux()

	if cfg.Features != nil {
		if cfg.Features.Introspection {
			srv.Use(extension.Introspection{})
		}
		if cfg.Features.ApolloTracing {
			srv.Use(apollotracing.Tracer{})
		}
		if cfg.Features.LogQueries {
			srv.AroundOperations(func(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
				oc := graphql.GetOperationContext(ctx)
				logger.L.Info("executing operation", map[string]interface{}{
					"operation_name": oc.OperationName,
					"raw_query":      oc.RawQuery,
				})
				return next(ctx)
			})
		}
	}

	mux.Handle("/", playground.Handler("GraphQL playground", "/query"))

	if cfg.Auth != nil && len(cfg.Auth.Users) > 0 {
		mux.Handle("/query", dataloader.Middleware(g, auth.Middleware(auth.BasicAuth(cfg.Auth.Users), srv)))
	} else {
		mux.Handle("/query", dataloader.Middleware(g, srv))
	}

	mux.Handle("/raft/join", rft.JoinHTTPHandler())

	server := &http.Server{Handler: logger.Middleware(logger.L, mux)}
	wg := errgroup.Group{}
	wg.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
				defer cancel()
				return server.Shutdown(ctx)
			}
		}
	})
	wg.Go(func() error {
		return server.Serve(glis)
	})
	if cfg.Server.RaftCluster != "" && cfg.Server.RaftBroadcast != "" {
		wg.Go(func() error {
			if !raft.JoinCluster(
				context.Background(),
				cfg.Server.RaftCluster,
				cfg.Server.RaftBroadcast,
				cfg.Server.RaftSecret,
			) {
				cancel()
				return errors.New("failed to join raft cluster")
			}
			return nil
		})
	}
	return wg.Wait()
}
