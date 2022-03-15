package persistence

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/autom8ter/morpheus/pkg/api"
	"github.com/autom8ter/morpheus/pkg/graph/model"
	"github.com/spf13/cast"
	"strconv"
	"strings"
)

var (
	nodesPrefix          InternalProperty = "1"
	relationshipPrefix   InternalProperty = "2"
	nodeRelationPrefix   InternalProperty = "3"
	nodeFieldsPrefix     InternalProperty = "4"
	relationFieldsPrefix InternalProperty = "5"
)

const (
	prefetchSize = 25
)

type InternalProperty string

const (
	ID         = "_id"
	Type       = "_type"
	Direction  = "_direction"
	Relation   = "_relation"
	SourceType = "_source_type"
	SourceID   = "_source_id"
	TargetType = "_target_type"
	TargetID   = "_target_id"
)

func getNodePath(typee, id string) []byte {
	key := append([]string{string(nodesPrefix)})
	if typee != "" {
		key = append(key, typee)
	}
	if id != "" {
		key = append(key, id)
	}
	return []byte(strings.Join(key, ","))
}

func getNodeRelationshipPath(sourceType, sourceID string, direction api.Direction, relation, targetType, targetID, relationID string) []byte {
	key := append([]string{string(nodeRelationPrefix)}, sourceType, sourceID, string(direction), relation, targetType)
	if targetID != "" {
		key = append(key, targetID)
	}
	if relationID != "" {
		key = append(key, relationID)
	}
	return []byte(strings.Join(key, ","))
}

func getRelationID(sourceType, sourceID string, relation, targetType, targetID string) string {
	key := append([]string{string(nodeRelationPrefix)}, sourceType, sourceID, relation, targetType, targetID)
	s := sha1.New()
	s.Write([]byte(strings.Join(key, ",")))
	id := s.Sum(nil)
	return hex.EncodeToString(id)
}

func getRelationshipPath(relation, id string) []byte {
	key := append([]string{string(relationshipPrefix)}, relation, id)
	return []byte(strings.Join(key, ","))
}

func getNodeTypeFieldPath(nodeType, field string, fieldValue interface{}, nodeID string) []byte {
	key := append([]string{string(nodeFieldsPrefix)}, nodeType, field, fmt.Sprint(fieldValue), nodeID)
	return []byte(strings.Join(key, ","))
}

func getRelationshipFieldPath(relation, field string, fieldValue interface{}, relationID string) []byte {
	key := append([]string{string(relationFieldsPrefix)}, relation, field, fmt.Sprint(fieldValue), relationID)
	return []byte(strings.Join(key, ","))
}

func parseCursor(cursor string) (int, error) {
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
func createCursor(skip int) string {
	cursor := fmt.Sprintf("cursor-%v", skip)
	str := base64.StdEncoding.EncodeToString([]byte(cursor))
	return str
}

func eval(exp *model.Expression, ent api.Entity) (bool, error) {

	val, err := ent.GetProperty(exp.Key)
	if err != nil {
		return false, err
	}
	switch v := ent.(type) {
	case api.Relationship:
		if strings.HasPrefix(exp.Key, "source.") {
			source, err := v.Source()
			if err != nil {
				return false, err
			}
			return eval(&model.Expression{
				Key:      strings.TrimPrefix(exp.Key, "source."),
				Operator: exp.Operator,
				Value:    exp.Value,
			}, source)
		}
		if strings.HasPrefix(exp.Key, "target.") {
			target, err := v.Target()
			if err != nil {
				return false, err
			}
			return eval(&model.Expression{
				Key:      strings.TrimPrefix(exp.Key, "target."),
				Operator: exp.Operator,
				Value:    exp.Value,
			}, target)
		}
	}

	switch exp.Operator {
	case model.OperatorEq:
		return val == exp.Value, nil
	case model.OperatorNeq:
		return val != exp.Value, nil
	case model.OperatorGt:
		return cast.ToFloat64(val) > cast.ToFloat64(exp.Value), nil
	case model.OperatorLt:
		return cast.ToFloat64(val) < cast.ToFloat64(exp.Value), nil
	case model.OperatorGte:
		return cast.ToFloat64(val) >= cast.ToFloat64(exp.Value), nil
	case model.OperatorLte:
		return cast.ToFloat64(val) <= cast.ToFloat64(exp.Value), nil
	case model.OperatorContains:
		return strings.Contains(cast.ToString(val), cast.ToString(exp.Value)), nil
	case model.OperatorHasPrefix:
		return strings.HasPrefix(cast.ToString(val), cast.ToString(exp.Value)), nil
	case model.OperatorHasSuffix:
		return strings.HasSuffix(cast.ToString(val), cast.ToString(exp.Value)), nil
	}
	return false, nil
}
