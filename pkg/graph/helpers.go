package graph

import (
	"encoding/base64"
	"fmt"
	"github.com/autom8ter/morpheus/pkg/api"
	"github.com/autom8ter/morpheus/pkg/graph/model"
	"strconv"
	"strings"
)

func toNode(n api.Node) *model.Node {
	props, _ := n.Properties()
	if n == nil {
		return &model.Node{}
	}
	return &model.Node{
		ID:         n.ID(),
		Type:       n.Type(),
		Properties: props,
	}
}

func toRelationship(rel api.Relationship) (*model.Relationship, error) {
	source, err := rel.Source()
	if err != nil {
		return nil, err
	}
	target, err := rel.Target()
	if err != nil {
		return nil, err
	}
	return &model.Relationship{
		ID:     rel.ID(),
		Type:   rel.Type(),
		Source: toNode(source),
		Target: toNode(target),
	}, nil
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
