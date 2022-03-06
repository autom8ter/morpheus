package api

import (
	"fmt"
)

type iEntity struct {
	id            string
	typee         string
	getProperties func() map[string]interface{}
	setProperties func(properties map[string]interface{})
}

func NewEntity(
	typee string,
	id string,
	getProperties func() map[string]interface{},
	setProperties func(properties map[string]interface{})) Entity {
	return &iEntity{id: id, typee: typee, getProperties: getProperties, setProperties: setProperties}
}

func (i iEntity) Hash() string {
	return fmt.Sprintf("%s/%s", i.typee, i.id)
}

func (i iEntity) ID() string {
	return i.id
}

func (i iEntity) Type() string {
	return i.typee
}

func (i iEntity) Properties() map[string]interface{} {
	return i.getProperties()
}

func (i iEntity) GetProperty(name string) interface{} {
	return i.getProperties()[name]
}

func (i iEntity) SetProperties(properties map[string]interface{}) {
	i.setProperties(properties)
}

func (i iEntity) DelProperty(name string) {
	props := i.getProperties()
	delete(props, name)
	i.SetProperties(props)
}

type iRelationship struct {
	Entity
	getSource func() Node
	getTarget func() Node
}

func NewRelationship(entity Entity, getSource func() Node, getTarget func() Node) Relationship {
	return &iRelationship{Entity: entity, getSource: getSource, getTarget: getTarget}
}

func (i iRelationship) Source() Node {
	return i.getSource()
}

func (i iRelationship) Target() Node {
	return i.getTarget()
}

func (i iRelationship) Reverse() Relationship {
	return iRelationship{
		Entity:    i.Entity,
		getSource: i.getTarget,
		getTarget: i.getSource,
	}
}

type iNode struct {
	Entity
	getRel        func(direction Direction, relation, id string) (Relationship, bool)
	addRel        func(direction Direction, relationship string, id string, node Node) Relationship
	delRel        func(direction Direction, relationship string, id string)
	relationships func(direction Direction, typee string, fn func(relationship Relationship) bool)
}

func NewNode(entity Entity,
	getRel func(direction Direction, relation, id string) (Relationship, bool),
	addRel func(direction Direction, relationship string, id string, node Node) Relationship,
	delRel func(direction Direction, relationship string, id string),
	relationships func(direction Direction, typee string, fn func(relationship Relationship) bool)) Node {
	return &iNode{Entity: entity, getRel: getRel, addRel: addRel, delRel: delRel, relationships: relationships}
}

func (i iNode) GetRelationship(direction Direction, relationship, id string) (Relationship, bool) {
	return i.getRel(direction, relationship, id)
}

func (i iNode) AddRelationship(direction Direction, relationship, id string, node Node) Relationship {
	return i.addRel(direction, relationship, id, node)
}

func (i iNode) DelRelationship(direction Direction, relationship, id string) {
	i.delRel(direction, relationship, id)
}

func (i iNode) Relationships(direction Direction, typee string, fn func(relationship Relationship) bool) {
	i.relationships(direction, typee, fn)
}

type graph struct {
	getNode    func(typee string, id string) (Node, error)
	addNode    func(typee string, id string, properties map[string]interface{}) (Node, error)
	delNode    func(typee string, id string) error
	rangeNodes func(typee string, fn func(node Node) bool) error
	nodeTypes  func() []string
	size       func() int
}

func newGraph(
	getNode func(typee string, id string) (Node, error),
	addNode func(typee string, id string, properties map[string]interface{}) (Node, error),
	delNode func(typee string, id string) error,
	rangeNodes func(typee string, fn func(node Node) bool) error,
	nodeTypes func() []string,
	size func() int) Graph {
	return &graph{getNode: getNode, addNode: addNode, delNode: delNode, rangeNodes: rangeNodes, nodeTypes: nodeTypes, size: size}
}

func (g graph) GetNode(typee string, id string) (Node, error) {
	return g.getNode(typee, id)
}

func (g graph) AddNode(typee string, id string, properties map[string]interface{}) (Node, error) {
	return g.addNode(typee, id, properties)
}

func (g graph) DelNode(typee string, id string) error {
	return g.delNode(typee, id)
}

func (g graph) RangeNodes(typee string, fn func(node Node) bool) error {
	return g.rangeNodes(typee, fn)
}

func (g graph) NodeTypes() []string {
	return g.nodeTypes()
}

func (g graph) Size() int {
	return g.size()
}

func NewGraph(entityFunc EntityCreationFunc) Graph {
	nodes := map[string]map[string]Node{}
	nodeRelationships := map[string]map[string]map[Direction]map[string]map[string]Relationship{}
	return newGraph(
		func(nodeType string, nodeID string) (Node, error) {
			if nodes[nodeType] == nil {
				return nil, fmt.Errorf("not found")
			}
			return nodes[nodeType][nodeID], nil
		},
		func(nodeType string, nodeID string, properties map[string]interface{}) (Node, error) {
			if nodes[nodeType] == nil {
				nodes[nodeType] = map[string]Node{}
			}
			if nodeRelationships[nodeType] == nil {
				nodeRelationships[nodeType] = map[string]map[Direction]map[string]map[string]Relationship{}
			}
			if nodeRelationships[nodeType][nodeID] == nil {
				nodeRelationships[nodeType][nodeID] = map[Direction]map[string]map[string]Relationship{}
			}
			if nodeRelationships[nodeType][nodeID][Outgoing] == nil {
				nodeRelationships[nodeType][nodeID][Outgoing] = map[string]map[string]Relationship{}
			}
			if nodeRelationships[nodeType][nodeID][Incoming] == nil {
				nodeRelationships[nodeType][nodeID][Incoming] = map[string]map[string]Relationship{}
			}
			if properties == nil {
				properties = map[string]interface{}{}
			}
			nodeEntity := entityFunc("1_", nodeType, nodeID, properties)
			node := NewNode(
				nodeEntity,
				func(direction Direction, relation, id string) (Relationship, bool) {
					if nodeRelationships[nodeType][nodeID][direction] == nil {
						return nil, false
					}
					if nodeRelationships[nodeType][nodeID][direction][relation] == nil {
						return nil, false
					}
					if rel, ok := nodeRelationships[nodeType][nodeID][direction][relation][id]; ok {
						return rel, true
					}
					return nil, false
				},
				func(direction Direction, relation, relationshipID string, node Node) Relationship {
					relationship := NewRelationship(
						entityFunc("2_", relation, relationshipID, map[string]interface{}{}),
						func() Node {
							if direction == Outgoing {
								return nodes[nodeType][nodeID]
							}
							return node
						},
						func() Node {
							if direction == Outgoing {
								return node
							}
							return nodes[nodeType][nodeID]
						},
					)
					if nodeRelationships[nodeType][nodeID][direction][relation] == nil {
						nodeRelationships[nodeType][nodeID][direction][relation] = map[string]Relationship{}
					}

					nodeRelationships[nodeType][nodeID][direction][relation][relationshipID] = relationship

					if direction == Outgoing {
						if nodeRelationships[node.Type()][node.ID()][Incoming][relation] == nil {
							nodeRelationships[node.Type()][node.ID()][Incoming][relation] = map[string]Relationship{}
						}
						nodeRelationships[node.Type()][node.ID()][Incoming][relation][relationshipID] = relationship
					} else {
						if nodeRelationships[node.Type()][node.ID()][Outgoing][relation] == nil {
							nodeRelationships[node.Type()][node.ID()][Outgoing][relation] = map[string]Relationship{}
						}
						nodeRelationships[node.Type()][node.ID()][Outgoing][relation][relationshipID] = relationship
					}
					return relationship
				},
				func(direction Direction, relationship, id string) {
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
				func(direction Direction, relation string, fn func(node Relationship) bool) {
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
		func(nodeType string, fn func(node Node) bool) error {
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
