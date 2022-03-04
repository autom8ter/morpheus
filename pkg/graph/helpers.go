package graph

import (
	"github.com/autom8ter/morpheus/pkg/api"
	"github.com/autom8ter/morpheus/pkg/graph/model"
	"github.com/spf13/cast"
	"strings"
)

func eval(exp *model.Expression, ent api.Entity) bool {
	val := ent.GetProperty(exp.Key)
	switch exp.Operator {
	case "==":
		return val == exp.Value
	case "!=":
		return val != exp.Value
	case "contains":
		return strings.Contains(cast.ToString(val), cast.ToString(exp.Value))
	}
	return false
}

func toNode(n api.Node) *model.Node {
	return &model.Node{
		ID:   n.ID(),
		Type: n.Type(),
	}
}

func toRelationship(rel api.Relationship) *model.Relationship {
	return &model.Relationship{
		ID:     rel.ID(),
		Type:   rel.Type(),
		Source: toNode(rel.Source()),
		Target: toNode(rel.Target()),
	}
}
