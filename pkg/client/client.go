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
		Transport: &basicAuthTransport{
			username:  username,
			password:  password,
			transport: http.DefaultTransport,
		},
		Timeout: timeout,
	}
	gclient := graphql.NewClient(endpoint, graphql.WithHTTPClient(client))
	return &Client{
		client: gclient,
	}
}

func (c *Client) Query(ctx context.Context, query string, vars map[string]string) (map[string]interface{}, error) {
	req := graphql.NewRequest(query)
	if vars != nil {
		for k, v := range vars {
			req.Var(k, v)
		}
	}
	resp := map[string]interface{}{}
	if err := c.client.Run(ctx, req, &resp); err != nil {
		return nil, stacktrace.Propagate(err, "failed to do request")
	}
	return resp, nil
}

type basicAuthTransport struct {
	username  string
	password  string
	transport http.RoundTripper
}

func (b *basicAuthTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	request.SetBasicAuth(b.username, b.password)
	return b.transport.RoundTrip(request)
}
