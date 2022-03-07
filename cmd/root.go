package cmd

import (
	"github.com/autom8ter/morpheus/pkg/config"
	"github.com/spf13/cobra"
	"log"
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
