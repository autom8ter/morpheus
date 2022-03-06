package server

import (
	"context"
	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/apollotracing"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/autom8ter/morpheus/pkg/auth"
	"github.com/autom8ter/morpheus/pkg/logger"
	"github.com/autom8ter/morpheus/pkg/raft"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
	"net"
	"net/http"
	"time"
)

const defaultPort = "8080"

type Opts struct {
	Tracing       bool
	Introspection bool
	LogQueries    bool
	Port          string
}

func Serve(ctx context.Context, opts *Opts, r *raft.Raft, schema graphql.ExecutableSchema) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	if opts.Port == "" {
		opts.Port = defaultPort
	}
	srv := handler.NewDefaultServer(schema)
	srv.SetQueryCache(lru.New(1000))
	if opts.Introspection {
		srv.Use(extension.Introspection{})
	}
	if opts.Tracing {
		srv.Use(apollotracing.Tracer{})
	}
	if opts.LogQueries {
		srv.AroundOperations(func(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
			oc := graphql.GetOperationContext(ctx)
			logger.L.Info("executing operation", map[string]interface{}{
				"operation_name": oc.OperationName,
				"raw_query":      oc.RawQuery,
			})
			return next(ctx)
		})
	}

	mux := http.NewServeMux()
	mux.Handle("/", playground.Handler("GraphQL playground", "/query"))

	usrs := viper.GetStringMapString("auth.users")
	if len(usrs) > 0 {
		mux.Handle("/query", auth.Middleware(auth.BasicAuth(usrs), srv))
	} else {
		mux.Handle("/query", srv)
	}

	mux.Handle("/raft/join", r.JoinHTTPHandler())

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
		lis, err := net.Listen("tcp", opts.Port)
		if err != nil {
			return err
		}
		defer lis.Close()
		return server.Serve(lis)
	})
	return wg.Wait()
}
