package api

import "fmt"

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

type iNode struct {
	Entity
	addRel        func(direction Direction, relationship string, id string, node Node) Relationship
	removeRel     func(direction Direction, relationship string, id string)
	relationships func(direction Direction, typee string, fn func(relationship Relationship) bool)
}

func NewNode(entity Entity,
	addRel func(direction Direction, relationship string, id string, node Node) Relationship,
	removeRel func(direction Direction, relationship string, id string),
	relationships func(direction Direction, typee string, fn func(relationship Relationship) bool)) Node {
	return &iNode{Entity: entity, addRel: addRel, removeRel: removeRel, relationships: relationships}
}

func (i iNode) AddRelationship(direction Direction, relationship, id string, node Node) Relationship {
	return i.addRel(direction, relationship, id, node)
}

func (i iNode) RemoveRelationship(direction Direction, relationship, id string) {
	i.removeRel(direction, relationship, id)
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

func NewGraph(
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
