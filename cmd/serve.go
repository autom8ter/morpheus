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
	"github.com/spf13/viper"
	"net"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "start server",
	Run: func(_ *cobra.Command, _ []string) {
		storage_path := viper.GetString("storage_path")
		g := badger.NewGraph(fmt.Sprintf("%s/storage", storage_path))
		rlis, err := net.Listen("tcp", fmt.Sprintf(":%v", viper.GetInt("raft_port")))
		if err != nil {
			logger.L.Error("failed to start raft listener", map[string]interface{}{
				"error": err,
			})
			return
		}
		defer rlis.Close()
		resolver, err := graph.NewResolver(g, rlis, raft.WithRaftDir(fmt.Sprintf("%s/raft", storage_path)))
		if err != nil {
			logger.L.Error("failed to create graphql resolver", map[string]interface{}{
				"error": err,
			})
			return
		}
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
		gport := viper.GetInt("graphql_port")
		logger.L.Info("starting server", map[string]interface{}{
			"graphql_port": gport,
			"raft_addr":    rlis.Addr().String(),
		})
		if err := server.Serve(context.Background(), &server.Opts{
			Tracing:       viper.GetBool("features.tracing"),
			Introspection: viper.GetBool("features.introspection"),
			LogQueries:    viper.GetBool("features.log_queries"),
			Port:          fmt.Sprintf(":%v", gport),
		}, schema); err != nil {
			logger.L.Error("server failure", map[string]interface{}{
				"error": err,
			})
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
