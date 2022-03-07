// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

import (
	"fmt"
	"io"
	"strconv"
)

type Entity interface {
	IsEntity()
}

type Expression struct {
	Key      string      `json:"key"`
	Operator Operator    `json:"operator"`
	Value    interface{} `json:"value"`
}

type Filter struct {
	Cursor      *string       `json:"cursor"`
	Expressions []*Expression `json:"expressions"`
	PageSize    *int          `json:"page_size"`
	OrderBy     *OrderBy      `json:"order_by"`
}

type Key struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type Node struct {
	ID              string                 `json:"id"`
	Type            string                 `json:"type"`
	Properties      map[string]interface{} `json:"properties"`
	GetProperty     interface{}            `json:"getProperty"`
	SetProperties   bool                   `json:"setProperties"`
	DelProperty     bool                   `json:"delProperty"`
	GetRelationship *Relationship          `json:"getRelationship"`
	AddRelationship *Relationship          `json:"addRelationship"`
	DelRelationship bool                   `json:"delRelationship"`
	Relationships   *Relationships         `json:"relationships"`
}

func (Node) IsEntity() {}

type Nodes struct {
	Cursor string  `json:"cursor"`
	Nodes  []*Node `json:"nodes"`
}

type OrderBy struct {
	Field   string `json:"field"`
	Reverse *bool  `json:"reverse"`
}

type Relationship struct {
	ID            string                 `json:"id"`
	Type          string                 `json:"type"`
	Properties    map[string]interface{} `json:"properties"`
	GetProperty   interface{}            `json:"getProperty"`
	SetProperties bool                   `json:"setProperties"`
	DelProperty   bool                   `json:"delProperty"`
	Source        *Node                  `json:"source"`
	Target        *Node                  `json:"target"`
}

func (Relationship) IsEntity() {}

type Relationships struct {
	Cursor        string          `json:"cursor"`
	Relationships []*Relationship `json:"relationships"`
}

type Direction string

const (
	DirectionOutgoing Direction = "OUTGOING"
	DirectionIncoming Direction = "INCOMING"
)

var AllDirection = []Direction{
	DirectionOutgoing,
	DirectionIncoming,
}

func (e Direction) IsValid() bool {
	switch e {
	case DirectionOutgoing, DirectionIncoming:
		return true
	}
	return false
}

func (e Direction) String() string {
	return string(e)
}

func (e *Direction) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = Direction(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid Direction", str)
	}
	return nil
}

func (e Direction) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type Operator string

const (
	OperatorEq        Operator = "EQ"
	OperatorNeq       Operator = "NEQ"
	OperatorGt        Operator = "GT"
	OperatorLt        Operator = "LT"
	OperatorGte       Operator = "GTE"
	OperatorLte       Operator = "LTE"
	OperatorContains  Operator = "CONTAINS"
	OperatorHasPrefix Operator = "HAS_PREFIX"
	OperatorHasSuffix Operator = "HAS_SUFFIX"
)

var AllOperator = []Operator{
	OperatorEq,
	OperatorNeq,
	OperatorGt,
	OperatorLt,
	OperatorGte,
	OperatorLte,
	OperatorContains,
	OperatorHasPrefix,
	OperatorHasSuffix,
}

func (e Operator) IsValid() bool {
	switch e {
	case OperatorEq, OperatorNeq, OperatorGt, OperatorLt, OperatorGte, OperatorLte, OperatorContains, OperatorHasPrefix, OperatorHasSuffix:
		return true
	}
	return false
}

func (e Operator) String() string {
	return string(e)
}

func (e *Operator) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = Operator(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid Operator", str)
	}
	return nil
}

func (e Operator) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}
