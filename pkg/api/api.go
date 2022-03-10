package api

type Direction string

const (
	Outgoing Direction = "OUTGOING"
	Incoming Direction = "INCOMING"
)

type Entity interface {
	ID() string
	Type() string
	Properties() map[string]interface{}
	GetProperty(name string) interface{}
	SetProperties(properties map[string]interface{})
	DelProperty(name string)
}

type EntityCreationFunc func(prefix, nodeType, nodeID string, properties map[string]interface{}) Entity

type Node interface {
	Entity
	AddRelationship(direction Direction, relationship string, id string, node Node) Relationship
	DelRelationship(direction Direction, relationship string, id string)
	GetRelationship(direction Direction, relation, id string) (Relationship, bool)
	Relationships(skip int, direction Direction, typee string, fn func(relationship Relationship) bool)
}

type Relationship interface {
	Entity
	Source() Node
	Target() Node
}

type Graph interface {
	GetNode(typee string, id string) (Node, error)
	AddNode(typee string, id string, properties map[string]interface{}) (Node, error)
	DelNode(typee string, id string) error
	RangeNodes(skip int, typee string, fn func(node Node) bool) error
	NodeTypes() []string

	GetRelationship(relation string, id string) (Relationship, error)
	RangeRelationships(skip int, relation string, fn func(node Relationship) bool) error
	RelationshipTypes() []string

	Size() int
	Close() error
}

type Operation func(graph Graph, input map[string]string, output chan interface{}) error
