package cmd

import (
	"context"
	"fmt"
	"github.com/autom8ter/morpheus/pkg/backends/inmem"
	"github.com/autom8ter/morpheus/pkg/graph"
	"github.com/autom8ter/morpheus/pkg/graph/generated"
	"github.com/autom8ter/morpheus/pkg/logger"
	"github.com/autom8ter/morpheus/pkg/server"
	"github.com/spf13/cobra"
	"log"
)

var (
	port int
	introspection bool
	logQueries bool
	tracing bool
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "start server",
	Run: func(_ *cobra.Command, _ []string) {
		g := inmem.NewGraph()
		schema := generated.NewExecutableSchema(generated.Config{
			Resolvers:  graph.NewResolver(g),
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
			log.Printf("server failure: %s", err)
		}
	},
}

func init() {
	serveCmd.Flags().IntVarP(&port, "port", "p", 8080, "port to serve on")
	serveCmd.Flags().BoolVarP(&logQueries, "log-queries", "l", false, "log all graphql requests")
	serveCmd.Flags().BoolVarP(&introspection, "introspection", "i", false, "enable introspection")
	serveCmd.Flags().BoolVarP(&tracing, "tracing", "t", false, "enable apollo tracing")
	rootCmd.AddCommand(serveCmd)
}
