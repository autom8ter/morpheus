package graph

import (
	"github.com/autom8ter/morpheus/pkg/api"
	"github.com/autom8ter/morpheus/pkg/middleware"
	"github.com/autom8ter/morpheus/pkg/raft"
	"sync"
)

type Resolver struct {
	graph api.Graph
	raft  *raft.Raft
	mu    *sync.RWMutex
	mw    *middleware.Middleware
}

func NewResolver(graph api.Graph, r *raft.Raft, mw *middleware.Middleware) *Resolver {
	return &Resolver{graph: graph, raft: r, mu: &sync.RWMutex{}, mw: mw}
}
