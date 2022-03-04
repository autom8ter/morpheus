package server

import (
	"context"
	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"golang.org/x/sync/errgroup"
	"net"
	"net/http"
	"time"
)

const defaultPort = "8080"

func Serve(ctx context.Context, port string, schema graphql.ExecutableSchema) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	if port == "" {
		port = defaultPort
	}
	srv := handler.NewDefaultServer(schema)
	mux := http.NewServeMux()
	mux.Handle("/", playground.Handler("GraphQL playground", "/query"))
	mux.Handle("/query", srv)
	server := &http.Server{Handler: mux}
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
		lis, err := net.Listen("tcp", port)
		if err != nil {
			return err
		}
		defer lis.Close()
		return server.Serve(lis)
	})
	return wg.Wait()
}
