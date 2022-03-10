package cmd

import (
	"context"
	"fmt"
	"github.com/autom8ter/morpheus/pkg/backends/badger"
	"github.com/autom8ter/morpheus/pkg/logger"
	"github.com/autom8ter/morpheus/pkg/server"
	"github.com/spf13/cobra"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "start server",
	Run: func(_ *cobra.Command, _ []string) {
		g, err := badger.NewGraph(fmt.Sprintf("%s/storage", cfg.Database.StoragePath), 1000000)
		if err != nil {
			panic(err)
		}
		defer func() {
			if err := g.Close(); err != nil {
				logger.L.Error("failed to close graph", map[string]interface{}{
					"error": err,
				})
			}
		}()
		logger.L.Info("starting server", map[string]interface{}{
			"graphql_port": cfg.Server.GraphQLPort,
			"raft_port":    cfg.Server.RaftPort,
		})
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		if err := server.Serve(ctx, g, cfg); err != nil {
			logger.L.Error("server failure", map[string]interface{}{
				"error": err,
			})
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
