package inmem

import (
	"fmt"
	"github.com/autom8ter/morpheus/pkg/api"
)

func NewGraph() api.Graph {
	nodes := map[string]map[string]api.Node{}
	nodeRelationships := map[string]map[string]map[api.Direction]map[string]map[string]api.Relationship{}
	return api.NewGraph(
		func(nodeType string, nodeID string) (api.Node, error) {
			if nodes[nodeType] == nil {
				return nil, fmt.Errorf("not found")
			}
			return nodes[nodeType][nodeID], nil
		},
		func(nodeType string, nodeID string, properties map[string]interface{}) (api.Node, error) {
			if nodes[nodeType] == nil {
				nodes[nodeType] = map[string]api.Node{}
			}
			if nodeRelationships[nodeType] == nil {
				nodeRelationships[nodeType] = map[string]map[api.Direction]map[string]map[string]api.Relationship{}
			}
			if nodeRelationships[nodeType][nodeID] == nil {
				nodeRelationships[nodeType][nodeID] = map[api.Direction]map[string]map[string]api.Relationship{}
			}
			if nodeRelationships[nodeType][nodeID][api.Outgoing] == nil {
				nodeRelationships[nodeType][nodeID][api.Outgoing] = map[string]map[string]api.Relationship{}
			}
			if nodeRelationships[nodeType][nodeID][api.Incoming] == nil {
				nodeRelationships[nodeType][nodeID][api.Incoming] = map[string]map[string]api.Relationship{}
			}
			if properties == nil {
				properties = map[string]interface{}{}
			}
			nodeEntity := NewEntity(nodeType, nodeID, properties)
			node := api.NewNode(
				nodeEntity,
				func(direction api.Direction, relation, relationshipID string, node api.Node) api.Relationship {
					relationship := api.NewRelationship(
						NewEntity(relation, relationshipID, map[string]interface{}{}),
						func() api.Node {
							if direction == api.Outgoing {
								return nodes[nodeType][nodeID]
							}
							return node
						},
						func() api.Node {
							if direction == api.Outgoing {
								return node
							}
							return nodes[nodeType][nodeID]
						},
					)
					if nodeRelationships[nodeType][nodeID][direction][relation] == nil {
						nodeRelationships[nodeType][nodeID][direction][relation] = map[string]api.Relationship{}
					}

					nodeRelationships[nodeType][nodeID][direction][relation][relationshipID] = relationship

					if direction == api.Outgoing {
						if nodeRelationships[node.Type()][node.ID()][api.Incoming][relation] == nil {
							nodeRelationships[node.Type()][node.ID()][api.Incoming][relation] = map[string]api.Relationship{}
						}
						nodeRelationships[node.Type()][node.ID()][api.Incoming][relation][relationshipID] = relationship.Reverse()
					} else {
						if nodeRelationships[node.Type()][node.ID()][api.Outgoing][relation] == nil {
							nodeRelationships[node.Type()][node.ID()][api.Outgoing][relation] = map[string]api.Relationship{}
						}
						nodeRelationships[node.Type()][node.ID()][api.Outgoing][relation][relationshipID] = relationship.Reverse()
					}
					return relationship
				},
				func(direction api.Direction, relationship, id string) {
					if nodeRelationships[nodeType] == nil {
						return
					}
					if nodeRelationships[nodeType][nodeID] == nil {
						return
					}
					if nodeRelationships[nodeType][nodeID][direction] == nil {
						return
					}
					if nodeRelationships[nodeType][nodeID][direction][relationship] == nil {
						return
					}
					delete(nodeRelationships[nodeType][nodeID][direction][relationship], id)
				},
				func(direction api.Direction, relation string, fn func(node api.Relationship) bool) {
					if nodeRelationships[nodeType] == nil {
						return
					}
					if nodeRelationships[nodeType][nodeID] == nil {
						return
					}
					if nodeRelationships[nodeType][nodeID][direction] == nil {
						return
					}
					if nodeRelationships[nodeType][nodeID][direction][relation] == nil {
						return
					}
					for _, rel := range nodeRelationships[nodeType][nodeID][direction][relation] {
						if !fn(rel) {
							break
						}
					}
				},
			)
			nodes[nodeType][nodeID] = node
			return node, nil
		},
		func(nodeType string, id string) error {
			delete(nodes[nodeType], id)
			return nil
		},
		func(nodeType string, fn func(node api.Node) bool) error {
			for _, n := range nodes[nodeType] {
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
