package api

type Direction string

const (
	Outgoing Direction = "OUTGOING"
	Incoming Direction = "INCOMING"
)

func (d Direction) Opposite() Direction {
	switch d {
	case Outgoing:
		return Incoming
	default:
		return Outgoing
	}
}

type Entity interface {
	ID() string
	Type() string
	Properties() map[string]interface{}
	GetProperty(name string) interface{}
	SetProperties(properties map[string]interface{})
	DelProperty(name string)
}

type Node interface {
	Entity
	AddRelationship(relationship string, node Node) Relationship
	DelRelationship(relationship string, id string)
	GetRelationship(relation, id string) (Relationship, bool)
	Relationships(skip int, relation string, targetType string, fn func(relationship Relationship) bool)
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
