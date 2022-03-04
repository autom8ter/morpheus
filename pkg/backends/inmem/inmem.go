package inmem

import (
	"fmt"
	"github.com/autom8ter/morpheus/pkg/api"
)

func NewGraph() api.Graph {
	nodes := map[string]map[string]api.Node{}
	return api.NewGraph(
		func(typee string, id string) (api.Node, error) {
			if nodes[typee] == nil {
				return nil, fmt.Errorf("not found")
			}
			return nodes[typee][id], nil
		},
		func(typee string, id string, properties map[string]interface{}) (api.Node, error) {
			if nodes[typee] == nil {
				nodes[typee] = map[string]api.Node{}
			}
			nodeRelationships := map[api.Direction]map[string]map[string]api.Relationship{}
			if properties == nil {
				properties = map[string]interface{}{}
			}
			nodeEntity := NewEntity(typee, id, properties)
			node := api.NewNode(
				nodeEntity,
				func(direction api.Direction, relation, id string, node api.Node) api.Relationship {
					relationship := api.NewRelationship(
						NewEntity(relation, id, map[string]interface{}{}),
						func() api.Node {
							if direction == api.Outgoing {
								return nodes[typee][id]
							}
							return node
						},
						func() api.Node {
							if direction == api.Outgoing {
								return node
							}
							return nodes[typee][id]
						},
					)
					if nodeRelationships[direction] == nil {
						nodeRelationships[direction] = map[string]map[string]api.Relationship{}
					}
					if nodeRelationships[direction][relation] == nil {
						nodeRelationships[direction][relation] = map[string]api.Relationship{}
					}
					nodeRelationships[direction][relation][id] = relationship
					return relationship
				},
				func(direction api.Direction, relationship, id string) {
					if nodeRelationships[direction] == nil {
						return
					}
					if nodeRelationships[direction][relationship] == nil {
						return
					}
					delete(nodeRelationships[direction][relationship], id)
				},
				func(direction api.Direction, typee string, fn func(node api.Relationship) bool) {
					if nodeRelationships[direction] == nil {
						return
					}
					if nodeRelationships[direction][typee] == nil {
						return
					}
					for _, rel := range nodeRelationships[direction][typee] {
						if !fn(rel) {
							break
						}
					}
				},
			)
			nodes[typee][id] = node
			return node, nil
		},
		func(typee string, id string) error {
			delete(nodes[typee], id)
			return nil
		},
		func(typee string, fn func(node api.Node) bool) error {
			for _, n := range nodes[typee] {
				if !fn(n) {
					return nil
				}
			}
			return nil
		},
		func() []string {
			var types []string
			for k, _ := range nodes {
				types = append(types, k)
			}
			return types
		},
		func() int {
			size := 0
			for _, n := range nodes {
				size = size + len(n)
			}
			return size
		},
	)
}

func NewEntity(typee, id string, props map[string]interface{}) api.Entity {
	return api.NewEntity(typee, id,
		func() map[string]interface{} {
			return props
		}, func(properties map[string]interface{}) {
			for k, v := range properties {
				props[k] = v
			}
		})
}
