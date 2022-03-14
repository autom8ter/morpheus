package cmd

import (
	"github.com/autom8ter/morpheus/cmd/client"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(serveCmd, client.RootCmd)

}

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
