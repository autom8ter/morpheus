package cmd

import (
	"context"
	"fmt"
	"github.com/autom8ter/morpheus/pkg/config"
	"github.com/autom8ter/morpheus/pkg/constants"
	"github.com/autom8ter/morpheus/pkg/logger"
	"github.com/autom8ter/morpheus/pkg/persistence"
	"github.com/autom8ter/morpheus/pkg/server"
	"github.com/spf13/cobra"
	"log"
)

const defaultCacheSize = 100000

var (
	configFile = "./config.yaml"
	cfg        *config.Config
)

func init() {
	var err error
	cfg, err = config.LoadConfig(configFile)
	if err != nil {
		log.Fatal(err)
	}
}

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "start server",
	Run: func(_ *cobra.Command, _ []string) {
		if len(cfg.Auth.Users) == 0 {
			cfg.Auth.Users = []config.User{config.User{
				Username: constants.ProjectName,
				Password: constants.ProjectName,
				Roles: []config.Role{
					config.ADMIN,
				},
			}}
		}
		g, err := persistence.NewPersistantGraph(fmt.Sprintf("%s/storage", cfg.Database.StoragePath), defaultCacheSize, 100)
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

