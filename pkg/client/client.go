package client

import (
	"context"
	"fmt"
	"github.com/autom8ter/morpheus/pkg/helpers"
	"github.com/machinebox/graphql"
	"github.com/palantir/stacktrace"
	"github.com/pkg/errors"
	"github.com/spf13/cast"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Client struct {
	next       int64
	httpClient *http.Client
	endpoint   string
	client     *graphql.Client
	username   string
	password   string
	token      string
	mu         sync.RWMutex
}

func NewClient(username, password, endpoint string, timeout time.Duration) *Client {
	client := &http.Client{
		Transport: http.DefaultTransport,
		Timeout:   timeout,
	}
	gclient := graphql.NewClient(endpoint, graphql.WithHTTPClient(client))
	c := &Client{
		next:       0,
		httpClient: client,
		endpoint:   endpoint,
		client:     gclient,
		username:   username,
		password:   password,
		token:      "",
		mu:         sync.RWMutex{},
	}
	if err := c.checkToken(); err != nil {
		panic(err)
	}
	return c
}

func (c *Client) checkToken() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.token != "" {
		if c.next > time.Now().Unix() {
			return nil
		}
		exp, unix, err := helpers.JWTExpired(c.token)
		if err != nil {
			return err
		}
		c.next = unix
		if !exp {
			return nil
		}
	}
	var loginQuery = fmt.Sprintf(`query { login(username: "%s", password: "%s") }`, c.username, c.password)
	req := graphql.NewRequest(loginQuery)
	//req.Var("user", c.username)
	//req.Var("password", c.password)

	resp := map[string]interface{}{}
	client := graphql.NewClient(c.endpoint)
	if err := client.Run(context.Background(), req, &resp); err != nil {
		return stacktrace.Propagate(err, "failed to do request %s %s  %s", c.username, c.password, loginQuery)
	}

	if resp["login"] == nil {
		return errors.New("failed to login")
	}
	token := cast.ToString(resp["login"])
	_, unix, err := helpers.JWTExpired(token)
	if err != nil {
		return err
	}
	c.token = token
	c.next = unix
	return nil
}

func (c *Client) Query(ctx context.Context, query string, vars map[string]string) (map[string]interface{}, error) {
	if err := c.checkToken(); err != nil {
		return nil, err
	}
	req := graphql.NewRequest(query)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.token))
	if vars != nil {
		for k, v := range vars {
			req.Var(k, v)
		}
	}
	resp := map[string]interface{}{}
	bearToken := req.Header.Get("Authorization")
	token := strings.TrimSpace(strings.TrimPrefix(bearToken, "Bearer "))
	if token == "" {
		panic(bearToken)
	}
	if err := c.client.Run(ctx, req, &resp); err != nil {
		return nil, stacktrace.Propagate(err, "failed to do request")
	}
	return resp, nil
}

func (c *Client) Queryx(ctx context.Context, query string, vars map[string]interface{}) (map[string]interface{}, error) {
	if err := c.checkToken(); err != nil {
		return nil, err
	}
	req := graphql.NewRequest(query)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.token))
	if vars != nil {
		for k, v := range vars {
			req.Var(k, v)
		}
	}
	resp := map[string]interface{}{}
	bearToken := req.Header.Get("Authorization")
	token := strings.TrimSpace(strings.TrimPrefix(bearToken, "Bearer "))
	if token == "" {
		panic(bearToken)
	}
	if err := c.client.Run(ctx, req, &resp); err != nil {
		return nil, stacktrace.Propagate(err, "failed to do request")
	}
	return resp, nil
}
