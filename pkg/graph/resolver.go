package graph

import (
	"github.com/autom8ter/morpheus/pkg/api"
	"github.com/autom8ter/morpheus/pkg/raft"
	"net"
	"sync"
)

type Resolver struct {
	graph api.Graph
	raft  *raft.Raft
	mu    *sync.RWMutex
}

func NewResolver(graph api.Graph, raftLis net.Listener, ropts ...raft.Opt) (*Resolver, error) {
	r, err := raft.NewRaft(graph, raftLis, ropts...)
	if err != nil {
		return nil, err
	}
	return &Resolver{graph: graph, raft: r, mu: &sync.RWMutex{}}, nil
}

func (r *Resolver) Close() error {
	return r.raft.Close()
}
