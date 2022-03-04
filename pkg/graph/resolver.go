package graph

import "github.com/autom8ter/morpheus/pkg/api"

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	graph api.Graph
}

func NewResolver(graph api.Graph) *Resolver {
	return &Resolver{graph: graph}
}
