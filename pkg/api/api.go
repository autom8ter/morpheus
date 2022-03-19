package api

import (
	"github.com/autom8ter/morpheus/pkg/graph/model"
	"github.com/hashicorp/raft"
)

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
	Properties() (map[string]interface{}, error)
	GetProperty(name string) (interface{}, error)
	SetProperties(properties map[string]interface{}) error
	DelProperty(name string) error
}

type Node interface {
	Entity
	AddRelation(direction Direction, relation string, properties map[string]interface{}, node Node) (Relation, error)
	DelRelation(relation string, id string) error
	GetRelation(relation, id string) (Relation, bool, error)
	Relations(where *model.RelationWhere) (string, []Relation, error)
}

type Relation interface {
	Entity
	Source() (Node, error)
	Target() (Node, error)
}

type Graph interface {
	GetNode(typee string, id string) (Node, error)
	AddNode(typee string, id string, properties map[string]interface{}) (Node, error)
	DelNode(typee string, id string) error
	RangeNodes(where *model.NodeWhere) (string, []Node, error)
	NodeTypes() []string

	GetRelation(relation string, id string) (Relation, error)
	RangeRelations(where *model.RelationWhere) (string, []Relation, error)
	RelationTypes() []string

	Close() error
	FSM() raft.FSM
}
