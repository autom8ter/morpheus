package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	client2 "github.com/autom8ter/morpheus/pkg/client"
	"github.com/palantir/stacktrace"
	"github.com/spf13/cobra"
	"io/ioutil"
	"time"
)

func getQueryCmd() *cobra.Command {
	var (
		endpoint string
		user     string
		password string
		query    string
		file     string
		timeout  time.Duration
		vars     = map[string]string{}
	)
	queryCmd := &cobra.Command{
		Use:   "query",
		Short: "run a graphql query",
		Run: func(_ *cobra.Command, _ []string) {
			if file != "" {
				bits, err := ioutil.ReadFile(file)
				if err != nil {
					fmt.Println(stacktrace.Propagate(err, "failed to read file: %s", file))
					return
				}
				query = string(bits)
			}
			if query == "" {
				fmt.Println("empty query")
				return
			}
			client := client2.NewClient(user, password, endpoint, timeout)
			resp, err := client.Query(context.Background(), query, vars)
			if err != nil {
				fmt.Println(stacktrace.Propagate(err, ""))
				return
			}
			bits, err := json.MarshalIndent(resp, "", "    ")
			if err != nil {
				fmt.Println(stacktrace.Propagate(err, ""))
				return
			}
			fmt.Println(string(bits))
		},
	}
	queryCmd.Flags().StringVarP(&endpoint, "", "e", "http://localhost:8080/query", "server endpoint")
	queryCmd.Flags().StringVarP(&user, "username", "u", "", "basic auth username")
	queryCmd.Flags().StringVarP(&password, "password", "p", "", "basic auth password")
	queryCmd.Flags().StringVarP(&file, "file", "f", "", "load query from graphql file path")
	queryCmd.Flags().DurationVarP(&timeout, "timeout", "t", 1*time.Minute, "query timeout")
	queryCmd.Flags().StringToStringVarP(&vars, "vars", "v", nil, "query variables")
	return queryCmd
}

func init() {
	rootCmd.AddCommand(getQueryCmd())
}
