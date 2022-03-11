package client

import (
	"context"
	"github.com/machinebox/graphql"
	"github.com/palantir/stacktrace"
	"net/http"
	"time"
)

type Client struct {
	client *graphql.Client
}

func NewClient(username, password, endpoint string, timeout time.Duration) *Client {
	client := &http.Client{
		Transport: basicAuthTransport{
			username: username,
			password: password,
		},
		Timeout: timeout,
	}
	client.Transport = basicAuthTransport{
		username: username,
		password: password,
	}
	return &Client{client: graphql.NewClient(endpoint, graphql.WithHTTPClient(client))}
}

func (c *Client) Query(ctx context.Context, query string, vars map[string]string, resp interface{}) error {
	req := graphql.NewRequest(query)
	if vars != nil {
		for k, v := range vars {
			req.Var(k, v)
		}
	}
	if err := c.client.Run(ctx, req, resp); err != nil {
		return stacktrace.Propagate(err, "failed to do request")
	}
	return nil
}

type basicAuthTransport struct {
	username string
	password string
}

func (b basicAuthTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	request.SetBasicAuth(b.username, b.password)
	return http.DefaultTransport.RoundTrip(request)
}
