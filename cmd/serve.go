package cmd

import (
	"context"
	"fmt"
	"github.com/autom8ter/morpheus/pkg/server"
	"github.com/spf13/cobra"
	"log"
)

var port int

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "start server",
	Run: func(_ *cobra.Command, _ []string) {
		log.Printf("starting server on port: %v", port)
		if err := server.Serve(context.Background(), fmt.Sprintf(":%v", port)); err != nil {
			log.Printf("server failure: %s", err)
		}
	},
}

func init() {
	serveCmd.Flags().IntVarP(&port, "port", "p", 8080, "port to serve on")
	rootCmd.AddCommand(serveCmd)
}
