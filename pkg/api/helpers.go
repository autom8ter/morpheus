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
	closer     func() error
}

func newGraph(
	getNode func(typee string, id string) (Node, error),
	addNode func(typee string, id string, properties map[string]interface{}) (Node, error),
	delNode func(typee string, id string) error,
	rangeNodes func(typee string, fn func(node Node) bool) error,
	nodeTypes func() []string,
	size func() int,
	closer func() error,
) Graph {
	return &graph{getNode: getNode, addNode: addNode, delNode: delNode, rangeNodes: rangeNodes, nodeTypes: nodeTypes, size: size, closer: closer}
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

func (g graph) Close() error {
	return g.closer()
}

func NewGraph(entityFunc EntityCreationFunc, closer func() error) Graph {
	nodes := map[string][]Node{}
	nodeRelationships := map[string]map[string]map[Direction]map[string][]Relationship{}
	return newGraph(
		func(nodeType string, nodeID string) (Node, error) {
			for _, node := range nodes[nodeType] {
				if node.ID() == nodeID {
					return node, nil
				}
			}
			return nil, fmt.Errorf("not found")
		},
		func(nodeType string, nodeID string, properties map[string]interface{}) (Node, error) {
			if nodeRelationships[nodeType] == nil {
				nodeRelationships[nodeType] = map[string]map[Direction]map[string][]Relationship{}
			}
			if nodeRelationships[nodeType][nodeID] == nil {
				nodeRelationships[nodeType][nodeID] = map[Direction]map[string][]Relationship{}
			}
			if nodeRelationships[nodeType][nodeID][Outgoing] == nil {
				nodeRelationships[nodeType][nodeID][Outgoing] = map[string][]Relationship{}
			}
			if nodeRelationships[nodeType][nodeID][Incoming] == nil {
				nodeRelationships[nodeType][nodeID][Incoming] = map[string][]Relationship{}
			}
			if properties == nil {
				properties = map[string]interface{}{}
			}
			nodeEntity := entityFunc("1_", nodeType, nodeID, properties)
			node := NewNode(
				nodeEntity,
				func(direction Direction, relation, relationID string) (Relationship, bool) {
					if nodeRelationships[nodeType][nodeID][direction] == nil {
						return nil, false
					}
					if nodeRelationships[nodeType][nodeID][direction][relation] == nil {
						return nil, false
					}
					for _, rel := range nodeRelationships[nodeType][nodeID][direction][relation] {
						if rel.ID() == relationID {
							return rel, true
						}
					}
					return nil, false
				},
				func(direction Direction, relation, relationshipID string, node Node) Relationship {
					relationship := NewRelationship(
						entityFunc("2_", relation, relationshipID, map[string]interface{}{}),
						func() Node {
							if direction == Outgoing {
								for _, node := range nodes[nodeType] {
									if node.ID() == nodeID {
										return node
									}
								}
								return nil
							}
							return node
						},
						func() Node {
							if direction == Outgoing {
								return node
							}
							for _, node := range nodes[nodeType] {
								if node.ID() == nodeID {
									return node
								}
							}
							return nil
						},
					)
					nodeRelationships[nodeType][nodeID][direction][relation] = append(nodeRelationships[nodeType][nodeID][direction][relation], relationship)

					if direction == Outgoing {
						nodeRelationships[node.Type()][node.ID()][Incoming][relation] = append(nodeRelationships[node.Type()][node.ID()][Incoming][relation], relationship)
					} else {
						nodeRelationships[node.Type()][node.ID()][Outgoing][relation] = append(nodeRelationships[node.Type()][node.ID()][Outgoing][relation], relationship)
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

					for _, rel := range nodeRelationships[nodeType][nodeID][direction][relationship] {
						if rel.ID() == id {
							nodeRelationships[rel.Source().Type()][rel.Source().ID()][Outgoing][relationship] = removeRelationshipID(nodeRelationships[rel.Source().Type()][rel.Source().ID()][Outgoing][relationship], id)
							nodeRelationships[rel.Source().Type()][rel.Source().ID()][Incoming][relationship] = removeRelationshipID(nodeRelationships[rel.Source().Type()][rel.Source().ID()][Incoming][relationship], id)

							nodeRelationships[rel.Target().Type()][rel.Target().ID()][Outgoing][relationship] = removeRelationshipID(nodeRelationships[rel.Target().Type()][rel.Target().ID()][Outgoing][relationship], id)
							nodeRelationships[rel.Target().Type()][rel.Target().ID()][Incoming][relationship] = removeRelationshipID(nodeRelationships[rel.Target().Type()][rel.Target().ID()][Incoming][relationship], id)
						}
					}
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
			nodes[nodeType] = append(nodes[nodeType], node)
			return node, nil
		},
		func(nodeType string, nodeID string) error {
			nodes[nodeType] = removeNodeID(nodes[nodeType], nodeID)
			delete(nodeRelationships[nodeType][nodeID], Outgoing)
			delete(nodeRelationships[nodeType][nodeID], Incoming)
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
		closer,
	)
}

func removeRelationshipID(slice []Relationship, id string) []Relationship {
	for i, element := range slice {
		if element.ID() == id {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

func removeNodeID(slice []Node, id string) []Node {
	for i, element := range slice {
		if element.ID() == id {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}
