package persistence

import (
	"crypto/sha1"
	"encoding/hex"
	"github.com/autom8ter/morpheus/pkg/api"
	"strings"
)

var (
	nodesPrefix        = "1"
	relationshipPrefix = "2"
	nodeRelationPrefix = "3"
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
	key := append([]string{nodesPrefix}, typee, id)
	return []byte(strings.Join(key, ","))
}

func getNodeRelationshipPath(sourceType, sourceID string, direction api.Direction, relation, targetType, targetID, relationID string) []byte {
	key := append([]string{nodeRelationPrefix}, sourceType, sourceID, string(direction), relation, targetType, targetID, relationID)
	return []byte(strings.Join(key, ","))
}

func getRelationID(sourceType, sourceID string, relation, targetType, targetID string) string {
	key := append([]string{nodeRelationPrefix}, sourceType, sourceID, relation, targetType, targetID)
	s := sha1.New()
	s.Write([]byte(strings.Join(key, ",")))
	id := s.Sum(nil)
	return hex.EncodeToString(id)
}

func getNodeRelationshipPrefix(sourceType, sourceID string, direction api.Direction, relation, targetType string) []byte {
	key := append([]string{nodeRelationPrefix}, sourceType, sourceID, string(direction), relation, targetType)
	return []byte(strings.Join(key, ","))
}

func getRelationshipPath(relation, id string) []byte {
	key := append([]string{relationshipPrefix}, relation, id)
	return []byte(strings.Join(key, ","))
}
