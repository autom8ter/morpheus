package scripts

import (
	"context"
	"github.com/autom8ter/morpheus/pkg/client"
)

type Script func(ctx context.Context, client *client.Client) error
