package cmd

import (
	"context"
	"fmt"
	"github.com/autom8ter/morpheus/pkg/backends/badger"
	"github.com/autom8ter/morpheus/pkg/graph"
	"github.com/autom8ter/morpheus/pkg/graph/generated"
	"github.com/autom8ter/morpheus/pkg/logger"
	"github.com/autom8ter/morpheus/pkg/raft"
	"github.com/autom8ter/morpheus/pkg/server"
	"github.com/spf13/cobra"
	"net"
)

var (
	port          int
	raftPort int
	introspection bool
	logQueries    bool
	tracing       bool
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "start server",
	Run: func(_ *cobra.Command, _ []string) {
		g := badger.NewGraph("./db/storage")
		rlis, err := net.Listen("tcp", fmt.Sprintf(":%v", raftPort))
		if err != nil {
			logger.L.Error("failed to start raft listener", map[string]interface{}{
				"error": err,
			})
			return
		}
		resolver, err := graph.NewResolver(g, rlis, raft.WithRaftDir("./db/raft"))
		if err != nil {
			logger.L.Error("failed to create graphql resolver", map[string]interface{}{
				"error": err,
			})
			return
		}
		schema := generated.NewExecutableSchema(generated.Config{
			Resolvers:  resolver,
			Directives: generated.DirectiveRoot{},
			Complexity: generated.ComplexityRoot{},
		})
		logger.L.Info("starting server", map[string]interface{}{
			"port": port,
		})
		if err := server.Serve(context.Background(), &server.Opts{
			Tracing:       tracing,
			Introspection: introspection,
			LogQueries:    logQueries,
			Port:          fmt.Sprintf(":%v", port),
		}, schema); err != nil {
			logger.L.Error("server failure", map[string]interface{}{
				"error": err,
			})
		}
	},
}

func init() {
	serveCmd.Flags().IntVarP(&port, "port", "p", 8080, "port to serve on")
	serveCmd.Flags().IntVarP(&raftPort, "raft-port", "r", 5283, "port to serve raft on")
	serveCmd.Flags().BoolVarP(&logQueries, "log-queries", "l", false, "log all graphql requests")
	serveCmd.Flags().BoolVarP(&introspection, "introspection", "i", false, "enable introspection")
	serveCmd.Flags().BoolVarP(&tracing, "tracing", "t", false, "enable apollo tracing")
	rootCmd.AddCommand(serveCmd)
}
