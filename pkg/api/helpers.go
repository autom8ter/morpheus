package api

import (
	"github.com/autom8ter/morpheus/pkg/constants"
	"github.com/autom8ter/morpheus/pkg/datastructure"
	"github.com/palantir/stacktrace"
	"sort"
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

type iNode struct {
	Entity
	getRel        func(direction Direction, relation, id string) (Relationship, bool)
	addRel        func(direction Direction, relationship string, id string, node Node) Relationship
	delRel        func(direction Direction, relationship string, id string)
	relationships func(skip int, direction Direction, typee string, fn func(relationship Relationship) bool)
}

func NewNode(entity Entity,
	getRel func(direction Direction, relation, id string) (Relationship, bool),
	addRel func(direction Direction, relationship string, id string, node Node) Relationship,
	delRel func(direction Direction, relationship string, id string),
	relationships func(skip int, direction Direction, typee string, fn func(relationship Relationship) bool)) Node {
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

func (i iNode) Relationships(skip int, direction Direction, typee string, fn func(relationship Relationship) bool) {
	i.relationships(skip, direction, typee, fn)
}

type graph struct {
	getNode    func(typee string, id string) (Node, error)
	addNode    func(typee string, id string, properties map[string]interface{}) (Node, error)
	delNode    func(typee string, id string) error
	rangeNodes func(skip int, typee string, fn func(node Node) bool) error
	nodeTypes  func() []string
	getRel     func(typee string, id string) (Relationship, error)
	rangeRels  func(skip int, typee string, fn func(relation Relationship) bool) error
	relTypes   func() []string
	size       func() int
	closer     func() error
}

func newGraph(
	getNode func(typee string, id string) (Node, error),
	addNode func(typee string, id string, properties map[string]interface{}) (Node, error),
	delNode func(typee string, id string) error,
	rangeNodes func(skip int, typee string, fn func(node Node) bool) error,
	nodeTypes func() []string,
	getRel func(typee string, id string) (Relationship, error),
	rangeRels func(skip int, typee string, fn func(relation Relationship) bool) error,
	relTypes func() []string,
	size func() int,
	closer func() error,
) Graph {
	return &graph{
		getNode:    getNode,
		addNode:    addNode,
		delNode:    delNode,
		rangeNodes: rangeNodes,
		nodeTypes:  nodeTypes,
		getRel:     getRel,
		rangeRels:  rangeRels,
		relTypes:   relTypes,
		size:       size,
		closer:     closer,
	}
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

func (g graph) RangeNodes(skip int, typee string, fn func(node Node) bool) error {
	return g.rangeNodes(skip, typee, fn)
}

func (g graph) NodeTypes() []string {
	return g.nodeTypes()
}

func (g graph) GetRelationship(relation string, id string) (Relationship, error) {
	return g.getRel(relation, id)
}

func (g graph) RangeRelationships(skip int, relation string, fn func(node Relationship) bool) error {
	return g.rangeRels(skip, relation, fn)
}

func (g graph) RelationshipTypes() []string {
	return g.relTypes()
}

func (g graph) Size() int {
	return g.size()
}

func (g graph) Close() error {
	return g.closer()
}

func NewGraph(entityFunc EntityCreationFunc, closer func() error) Graph {
	nodes := map[string]datastructure.OrderedMap{}
	relationships := map[string]datastructure.OrderedMap{}
	nodeRelationships := map[string]map[string]map[Direction]map[string]datastructure.OrderedMap{}
	return newGraph(
		func(nodeType string, nodeID string) (Node, error) {
			if nodes[nodeType] == nil {
				return nil, stacktrace.Propagate(constants.ErrNotFound, "")
			}
			val, ok := nodes[nodeType].Get(nodeID)
			if !ok {
				return nil, stacktrace.Propagate(constants.ErrNotFound, "")
			}
			return val.(Node), nil
		},
		func(nodeType string, nodeID string, properties map[string]interface{}) (Node, error) {
			if nodes[nodeType] == nil {
				nodes[nodeType] = datastructure.NewOrderedMap()
			}
			if nodeRelationships[nodeType] == nil {
				nodeRelationships[nodeType] = map[string]map[Direction]map[string]datastructure.OrderedMap{}
			}
			if nodeRelationships[nodeType][nodeID] == nil {
				nodeRelationships[nodeType][nodeID] = map[Direction]map[string]datastructure.OrderedMap{}
			}
			if nodeRelationships[nodeType][nodeID][Outgoing] == nil {
				nodeRelationships[nodeType][nodeID][Outgoing] = map[string]datastructure.OrderedMap{}
			}
			if nodeRelationships[nodeType][nodeID][Incoming] == nil {
				nodeRelationships[nodeType][nodeID][Incoming] = map[string]datastructure.OrderedMap{}
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
					val, ok := nodeRelationships[nodeType][nodeID][direction][relation].Get(relationID)
					if !ok {
						return nil, false
					}
					return val.(Relationship), true
				},
				func(direction Direction, relation, relationshipID string, node Node) Relationship {
					if relationships[relation] == nil {
						relationships[relation] = datastructure.NewOrderedMap()
					}
					relationship := NewRelationship(
						entityFunc("2_", relation, relationshipID, map[string]interface{}{}),
						func() Node {
							if direction == Outgoing {
								val, _ := nodes[nodeType].Get(nodeID)
								return val.(Node)
							}
							return node
						},
						func() Node {
							if direction == Outgoing {
								return node
							}
							val, _ := nodes[nodeType].Get(nodeID)
							return val.(Node)
						},
					)
					if nodeRelationships[nodeType][nodeID][direction][relation] == nil {
						nodeRelationships[nodeType][nodeID][direction][relation] = datastructure.NewOrderedMap()
					}
					if nodeRelationships[node.Type()][node.ID()][Incoming][relation] == nil {
						nodeRelationships[node.Type()][node.ID()][Incoming][relation] = datastructure.NewOrderedMap()
					}

					nodeRelationships[nodeType][nodeID][direction][relation].Add(relationship.ID(), relationship)

					if direction == Outgoing {
						nodeRelationships[node.Type()][node.ID()][Incoming][relation].Add(relationship.ID(), relationship)
					} else {
						nodeRelationships[node.Type()][node.ID()][Outgoing][relation].Add(relationship.ID(), relationship)
					}
					relationships[relation].Add(relationshipID, relation)

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

					val, ok := nodeRelationships[nodeType][nodeID][direction][relationship].Get(id)
					rel := val.(Relationship)
					if ok {
						nodeRelationships[rel.Source().Type()][rel.Source().ID()][Outgoing][relationship].Del(id)
						nodeRelationships[rel.Source().Type()][rel.Source().ID()][Incoming][relationship].Del(id)

						nodeRelationships[rel.Target().Type()][rel.Target().ID()][Outgoing][relationship].Del(id)
						nodeRelationships[rel.Target().Type()][rel.Target().ID()][Incoming][relationship].Del(id)

						relationships[relationship].Del(id)
					}
				},
				func(skip int, direction Direction, relation string, fn func(node Relationship) bool) {
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
					nodeRelationships[nodeType][nodeID][direction][relation].Range(skip, func(val interface{}) bool {
						return fn(val.(Relationship))
					})
				},
			)
			nodes[nodeType].Add(node.ID(), node)
			return node, nil
		},
		func(nodeType string, nodeID string) error {
			for _, types := range nodeRelationships[nodeType][nodeID] {
				for relationship, values := range types {
					values.Range(0, func(val interface{}) bool {
						rel := val.(Relationship)
						relationships[relationship].Del(rel.ID())
						nodeRelationships[rel.Source().Type()][rel.Source().ID()][Outgoing][relationship].Del(rel.ID())
						nodeRelationships[rel.Source().Type()][rel.Source().ID()][Incoming][relationship].Del(rel.ID())

						nodeRelationships[rel.Target().Type()][rel.Target().ID()][Outgoing][relationship].Del(rel.ID())
						nodeRelationships[rel.Target().Type()][rel.Target().ID()][Incoming][relationship].Del(rel.ID())
						return true
					})
				}
			}
			delete(nodeRelationships[nodeType][nodeID], Outgoing)
			delete(nodeRelationships[nodeType][nodeID], Incoming)
			nodes[nodeType].Del(nodeID)
			return nil
		},
		func(skip int, nodeType string, fn func(node Node) bool) error {
			if nodes[nodeType] == nil {
				return nil
			}
			nodes[nodeType].Range(skip, func(val interface{}) bool {
				return fn(val.(Node))
			})
			return nil
		},
		func() []string {
			var types []string
			for k, _ := range nodes {
				types = append(types, k)
			}
			sort.Strings(types)
			return types
		},
		func(typee string, id string) (Relationship, error) {
			if relationships[typee] == nil {
				return nil, stacktrace.Propagate(constants.ErrNotFound, "")
			}
			val, ok := relationships[typee].Get(id)
			if !ok {
				return nil, stacktrace.Propagate(constants.ErrNotFound, "")
			}
			return val.(Relationship), nil
		},
		func(skip int, typee string, fn func(relation Relationship) bool) error {
			if nodes[typee] == nil {
				return nil
			}
			relationships[typee].Range(skip, func(val interface{}) bool {
				return fn(val.(Relationship))
			})
			return nil
		},
		func() []string {
			var types []string
			for k, _ := range relationships {
				types = append(types, k)
			}
			sort.Strings(types)
			return types
		},
		func() int {
			size := 0
			for _, n := range nodes {
				size = size + n.Len()
			}
			return size
		},
		closer,
	)
}
