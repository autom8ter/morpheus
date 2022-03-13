package graph

import (
	"github.com/autom8ter/morpheus/pkg/api"
	"github.com/autom8ter/morpheus/pkg/auth"
	"github.com/autom8ter/morpheus/pkg/raft"
	"sync"
)

type Resolver struct {
	graph api.Graph
	raft  *raft.Raft
	mu    *sync.RWMutex
	auth  *auth.Auth
}

func NewResolver(graph api.Graph, r *raft.Raft, ath *auth.Auth) *Resolver {
	return &Resolver{graph: graph, raft: r, mu: &sync.RWMutex{}, auth: ath}
}
