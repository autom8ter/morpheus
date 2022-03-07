package graph

import (
	"github.com/autom8ter/morpheus/pkg/api"
	"github.com/autom8ter/morpheus/pkg/raft"
	"sync"
)

type Resolver struct {
	graph api.Graph
	raft  *raft.Raft
	mu    *sync.RWMutex
}

func NewResolver(graph api.Graph, r *raft.Raft) *Resolver {
	return &Resolver{graph: graph, raft: r, mu: &sync.RWMutex{}}
}

func (r *Resolver) Close() error {
	return r.raft.Close()
}
