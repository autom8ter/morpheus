package cmd

import (
	"fmt"
	"github.com/autom8ter/morpheus/pkg/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "morpheus",
	Short: "A brief description of your application",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

var configFile = "./config.yaml"

func init() {
	homedir, _ := os.UserHomeDir()
	viper.AutomaticEnv()
	viper.SetDefault("server.graphql_port", 8080)
	viper.SetDefault("server.raft_port", 7598)
	viper.SetDefault("database.storage_path", fmt.Sprintf("%s/.morpheus", homedir))
	viper.SetDefault("features.log_queries", false)
	viper.SetDefault("features.introspection", false)
	viper.SetDefault("features.apollo_tracing", false)
	viper.SetDefault("server.raft_cluster", "")
	viper.SetConfigFile(configFile)
	if err := viper.ReadInConfig(); err != nil {
		logger.L.Info("missing config file", map[string]interface{}{
			"expected_path": configFile,
		})
	}
}
