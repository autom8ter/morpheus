package cmd

import (
	"context"
	"fmt"
	"github.com/autom8ter/morpheus/pkg/logger"
	"github.com/autom8ter/morpheus/pkg/persistance"
	"github.com/autom8ter/morpheus/pkg/server"
	"github.com/spf13/cobra"
)

const defaultCacheSize = 100000

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "start server",
	Run: func(_ *cobra.Command, _ []string) {
		g, err := persistance.NewPersistantGraph(fmt.Sprintf("%s/storage", cfg.Database.StoragePath), defaultCacheSize)
		if err != nil {
			panic(err)
		}
		logger.L.Info("starting server", map[string]interface{}{
			"port": cfg.Server.Port,
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
