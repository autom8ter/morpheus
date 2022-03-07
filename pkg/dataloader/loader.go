package dataloader

import (
	"context"
	"fmt"
	"github.com/autom8ter/morpheus/pkg/api"
	"net/http"
	"strings"
	"time"
)

const loadersKey = "dataloaders"

type Loaders struct {
	NodeByID NodeLoader
}

func ID(typee string, id string) string {
	return fmt.Sprintf("%s///%s", typee, id)
}

func getTypeID(id string) (string, string, bool) {
	split := strings.Split(id, "///")
	if len(split) < 2 {
		return "", "", false
	}
	return split[0], split[1], true
}

func Middleware(g api.Graph, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), loadersKey, &Loaders{
			NodeByID: NodeLoader{
				maxBatch: 100,
				wait:     1 * time.Millisecond,
				fetch: func(ids []string) ([]*api.Node, []error) {
					var errs []error
					var nodes []*api.Node
					for _, id := range ids {
						typee, id, ok := getTypeID(id)
						if !ok {
							continue
						}
						n, err := g.GetNode(typee, id)
						if err != nil {
							errs = append(errs, err)
						}
						nodes = append(nodes, &n)
					}
					return nodes, errs
				},
			},
		})
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func For(ctx context.Context) *Loaders {
	return ctx.Value(loadersKey).(*Loaders)
}
