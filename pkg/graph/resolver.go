package graph

import (
	"github.com/autom8ter/morpheus/pkg/api"
	"github.com/autom8ter/morpheus/pkg/raft"
	"net"
)

type Resolver struct {
	graph api.Graph
	raft  *raft.Raft
}

func NewResolver(graph api.Graph, raftLis net.Listener) (*Resolver, error) {
	r, err := raft.NewRaft(graph, raftLis)
	if err != nil {
		return nil, err
	}
	return &Resolver{graph: graph, raft: r}, nil
}

func (r *Resolver) Close() error {
	return r.raft.Close()
}
