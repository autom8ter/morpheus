package server

import (
	"context"
	"crypto/tls"
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
	"github.com/autom8ter/morpheus/pkg/graph"
	"github.com/autom8ter/morpheus/pkg/graph/generated"
	"github.com/autom8ter/morpheus/pkg/logger"
	"github.com/autom8ter/morpheus/pkg/raft"
	"github.com/palantir/stacktrace"
	"github.com/soheilhy/cmux"
	"golang.org/x/sync/errgroup"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func Serve(ctx context.Context, g api.Graph, cfg *config.Config) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%v", cfg.Server.Port))
	if err != nil {
		return err
	}
	defer lis.Close()
	if cfg.Server.TLSCert != "" {
		cer, err := tls.X509KeyPair([]byte(cfg.Server.TLSCert), []byte(cfg.Server.TLSKey))
		if err != nil {
			return err
		}
		tlsConfig := &tls.Config{Certificates: []tls.Certificate{cer}}
		lis = tls.NewListener(lis, tlsConfig)
	}
	clis := cmux.New(lis)

	glis := clis.Match(cmux.HTTP1(), cmux.HTTP2())
	defer glis.Close()
	tcplis := clis.Match(cmux.Any())
	defer tcplis.Close()

	joinRaft := cfg.Server.RaftCluster
	rft, err := raft.NewRaft(
		g,
		tcplis,
		raft.WithRaftDir(fmt.Sprintf("%s/raft", cfg.Database.StoragePath)),
		raft.WithIsLeader(joinRaft == ""),
		raft.WithClusterSecret(cfg.Server.RaftSecret),
	)
	if err != nil {
		return err
	}
	ath := auth.NewAuth(cfg.Auth)
	resolver := graph.NewResolver(g, rft, ath)
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

	mux.Handle("/", playground.Handler("GraphQL Console", "/query"))

	mux.Handle("/query", ath.JwtClaimsParser(srv, false))

	server := &http.Server{Handler: logger.Middleware(logger.L, mux)}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	wg, ctx := errgroup.WithContext(ctx)
	wg.Go(func() error {
		clis.Serve()
		return nil
	})
	wg.Go(func() error {
		if err := server.Serve(glis); err != nil && stacktrace.RootCause(err) != http.ErrServerClosed {
			return stacktrace.Propagate(server.Serve(glis), "")
		}
		return nil
	})
	select {
	case <-interrupt:
		logger.L.Info("shutdown signal received", map[string]interface{}{})
		cancel()
		break
	case <-ctx.Done():
		break
	}
	logger.L.Info("shutting down...", map[string]interface{}{})
	ctx, cancel = context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.L.Error("failed to shutdown server", map[string]interface{}{
			"error": err,
		})
	}
	g.Close()
	rft.Close()
	tcplis.Close()
	clis.Close()
	lis.Close()
	return wg.Wait()
}
