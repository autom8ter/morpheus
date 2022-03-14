package graph

import (
	"encoding/base64"
	"fmt"
	"github.com/autom8ter/morpheus/pkg/api"
	"github.com/autom8ter/morpheus/pkg/graph/model"
	"github.com/spf13/cast"
	"strconv"
	"strings"
)

func eval(exp *model.Expression, ent api.Entity) bool {

	val := ent.GetProperty(exp.Key)
	if val == nil && exp.Key == "type" {
		val = ent.Type()
	}
	if val == nil && exp.Key == "id" {
		val = ent.ID()
	}
	switch v := ent.(type) {
	case api.Relationship:
		if strings.HasPrefix(exp.Key, "source.") {
			return eval(&model.Expression{
				Key:      strings.TrimPrefix(exp.Key, "source."),
				Operator: exp.Operator,
				Value:    exp.Value,
			}, v.Source())
		}
		if strings.HasPrefix(exp.Key, "target.") {
			return eval(&model.Expression{
				Key:      strings.TrimPrefix(exp.Key, "target."),
				Operator: exp.Operator,
				Value:    exp.Value,
			}, v.Target())
		}
	}

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
func (r *Resolver) parseCursor(cursor string) (int, error) {
	bits, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return 0, err
	}
	split := strings.Split(string(bits), "cursor-")
	if len(split) < 2 {
		return 0, fmt.Errorf("bad cursor")
	}
	return strconv.Atoi(split[1])
}
func (r *Resolver) createCursor(skip int) string {
	cursor := fmt.Sprintf("cursor-%v", skip)
	str := base64.StdEncoding.EncodeToString([]byte(cursor))
	return str
}
