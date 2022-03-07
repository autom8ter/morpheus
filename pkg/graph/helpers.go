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
	case model.OperatorEq:
		return val == exp.Value
	case model.OperatorNeq:
		return val != exp.Value
	case model.OperatorGt:
		return cast.ToFloat64(val) > cast.ToFloat64(exp.Value)
	case model.OperatorLt:
		return cast.ToFloat64(val) < cast.ToFloat64(exp.Value)
	case model.OperatorGte:
		return cast.ToFloat64(val) >= cast.ToFloat64(exp.Value)
	case model.OperatorLte:
		return cast.ToFloat64(val) <= cast.ToFloat64(exp.Value)
	case model.OperatorContains:
		return strings.Contains(cast.ToString(val), cast.ToString(exp.Value))
	case model.OperatorHasPrefix:
		return strings.HasPrefix(cast.ToString(val), cast.ToString(exp.Value))
	case model.OperatorHasSuffix:
		return strings.HasSuffix(cast.ToString(val), cast.ToString(exp.Value))
	}
	return false
}

func toNode(n api.Node) *model.Node {
	if n == nil {
		return &model.Node{}
	}
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
