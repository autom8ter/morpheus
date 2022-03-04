package api

type Direction string

const (
	Outgoing Direction = "OUTGOING"
	Incoming Direction = "INCOMING"
)

type Entity interface {
	Hash() string
	ID() string
	Type() string
	Properties() map[string]interface{}
	GetProperty(name string) interface{}
	SetProperties(properties map[string]interface{})
	DelProperty(name string)
}

type Node interface {
	Entity
	AddRelationship(direction Direction, relationship string, id string, node Node) Relationship
	RemoveRelationship(direction Direction, relationship string, id string)
	Relationships(direction Direction, typee string, fn func(relationship Relationship) bool)
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
	RangeNodes(typee string, fn func(node Node) bool) error
	NodeTypes() []string
	Size() int
}

type Operation func(graph Graph, input map[string]string, output chan interface{}) error
