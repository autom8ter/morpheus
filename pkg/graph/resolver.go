package graph

import (
	"github.com/autom8ter/morpheus/pkg/api"
	"github.com/autom8ter/morpheus/pkg/encode"
	"github.com/autom8ter/morpheus/pkg/graph/fsm"
	"github.com/autom8ter/morpheus/pkg/middleware"
	"github.com/autom8ter/morpheus/pkg/raft"
	lru "github.com/hashicorp/golang-lru"
	"github.com/palantir/stacktrace"
	"sync"
)

type Resolver struct {
	graph api.Graph
	raft  *raft.Raft
	mu    *sync.RWMutex
	mw    *middleware.Middleware
	cache *lru.Cache
}

func NewResolver(graph api.Graph, r *raft.Raft, mw *middleware.Middleware) *Resolver {
	c, err := lru.New(10000)
	if err != nil {
		panic(err)
	}
	return &Resolver{graph: graph, raft: r, mu: &sync.RWMutex{}, mw: mw, cache: c}
}

func (r *Resolver) applyCMD(cmd *fsm.CMD) (interface{}, error) {
	bits, err := encode.Marshal(cmd)
	if err != nil {
		return nil, stacktrace.Propagate(err, "")
	}
	val, err := r.raft.Apply(bits)
	if err != nil {
		return nil, stacktrace.Propagate(err, "")
	}
	if err, ok := val.(error); ok {
		return nil, stacktrace.Propagate(err, "")
	}
	return val, nil
}
