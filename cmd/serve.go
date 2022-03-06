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
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
	"net"
	"time"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "start server",
	Run: func(_ *cobra.Command, _ []string) {
		storage_path := viper.GetString("database.storage_path")
		g := badger.NewGraph(fmt.Sprintf("%s/storage", storage_path))
		defer func() {
			if err := g.Close(); err != nil {
				logger.L.Error("failed to close graph", map[string]interface{}{
					"error": err,
				})
			}
		}()
		rlis, err := net.Listen("tcp", fmt.Sprintf(":%v", viper.GetInt("server.raft_port")))
		if err != nil {
			logger.L.Error("failed to start raft listener", map[string]interface{}{
				"error": err,
			})
			return
		}
		defer rlis.Close()
		joinRaft := viper.GetString("server.raft_cluster")
		rft, err := raft.NewRaft(
			g,
			rlis,
			raft.WithRaftDir(fmt.Sprintf("%s/raft", storage_path)),
			raft.WithTimeout(3*time.Second),
			raft.WithIsLeader(joinRaft == ""),
			raft.WithClusterSecret(viper.GetString("server.raft_secret")),
		)
		if err != nil {
			logger.L.Error("failed to create raft", map[string]interface{}{
				"error": err,
			})
			return
		}
		resolver := graph.NewResolver(g, rft)
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
		gport := viper.GetInt("server.graphql_port")
		logger.L.Info("starting server", map[string]interface{}{
			"graphql_port": gport,
			"raft_addr":    rlis.Addr().String(),
		})
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		egp := errgroup.Group{}
		if joinRaft != "" {
			egp.Go(func() error {
				if !raft.JoinCluster(
					context.Background(),
					joinRaft,
					viper.GetString("server.raft_local_addr"),
					viper.GetString("server.raft_secret"),
				) {
					cancel()
					return errors.New("failed to join raft cluster")
				}
				return nil
			})
		}

		egp.Go(func() error {
			return server.Serve(ctx, &server.Opts{
				Tracing:       viper.GetBool("features.tracing"),
				Introspection: viper.GetBool("features.introspection"),
				LogQueries:    viper.GetBool("features.log_queries"),
				Port:          fmt.Sprintf(":%v", gport),
			}, rft, schema)
		})

		if err := egp.Wait(); err != nil {
			logger.L.Error("server failure", map[string]interface{}{
				"error": err,
			})
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
