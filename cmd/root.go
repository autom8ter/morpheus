package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

func init() {
	viper.AutomaticEnv()
	viper.SetDefault("graphql_port", 8080)
	viper.SetDefault("raft_port", 7598)
	viper.SetDefault("storage_path", "~/.morpheus")
	viper.SetDefault("features.log_queries", false)
	viper.SetDefault("features.introspection", false)
	viper.SetDefault("features.apollo_tracing", false)
}
