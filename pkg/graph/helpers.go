package graph

import (
	"encoding/base64"
	"fmt"
	"github.com/autom8ter/morpheus/pkg/api"
	"github.com/autom8ter/morpheus/pkg/graph/model"
	"github.com/palantir/stacktrace"
	"strconv"
	"strings"
)

func toNode(n api.Node) (*model.Node, error) {
	props, err := n.Properties()
	if err != nil {
		return nil, stacktrace.Propagate(err, "")
	}
	return &model.Node{
		ID:         n.ID(),
		Type:       n.Type(),
		Properties: props,
	}, nil
}

func toRelation(rel api.Relation) (*model.Relation, error) {
	source, err := rel.Source()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to load source")
	}
	target, err := rel.Target()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to load target")
	}
	sourceNode, err := toNode(source)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to load source node %s %s", source.Type(), source.ID())
	}
	targetNode, err := toNode(target)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to load target nod %s %s", target.Type(), target.ID())
	}
	return &model.Relation{
		ID:     rel.ID(),
		Type:   rel.Type(),
		Source: sourceNode,
		Target: targetNode,
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
