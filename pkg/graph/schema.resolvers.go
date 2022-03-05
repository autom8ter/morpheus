package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/autom8ter/morpheus/pkg/raft/fsm"

	"github.com/autom8ter/morpheus/pkg/api"
	"github.com/autom8ter/morpheus/pkg/graph/generated"
	"github.com/autom8ter/morpheus/pkg/graph/model"
	"github.com/google/uuid"
)

func (r *mutationResolver) AddNode(ctx context.Context, typeArg string, properties map[string]interface{}) (*model.Node, error) {
	id := uuid.New().String()
	cmd := &fsm.CMD{
		Method: fsm.AddNodes,
		AddNodes: []fsm.AddNode{
			{
				Type:       typeArg,
				ID:         id,
				Properties: properties,
			},
		},
	}
	bits, err := json.Marshal(cmd)
	if err != nil {
		return nil, err
	}
	val, err := r.raft.Apply(bits)
	if err != nil {
		return nil, err
	}
	if err, ok := val.(error); ok {
		return nil, err
	}

	return toNode(val.([]api.Node)[0]), nil
}

func (r *mutationResolver) DelNode(ctx context.Context, key model.Key) (bool, error) {
	cmd := &fsm.CMD{
		Method: fsm.DelNodes,
		DelNodes: []fsm.DelNode{
			{
				Type: key.Type,
				ID:   key.ID,
			},
		},
	}
	bits, err := json.Marshal(cmd)
	if err != nil {
		return false, err
	}
	val, err := r.raft.Apply(bits)
	if err != nil {
		return false, err
	}
	if err, ok := val.(error); ok {
		return false, err
	}
	return true, nil
}

func (r *nodeResolver) Properties(ctx context.Context, obj *model.Node, input *string) (map[string]interface{}, error) {
	n, err := r.graph.GetNode(obj.Type, obj.ID)
	if err != nil {
		return nil, err
	}
	return n.Properties(), nil
}

func (r *nodeResolver) GetProperty(ctx context.Context, obj *model.Node, key string) (interface{}, error) {
	n, err := r.graph.GetNode(obj.Type, obj.ID)
	if err != nil {
		return nil, err
	}
	return n.GetProperty(key), nil
}

func (r *nodeResolver) SetProperties(ctx context.Context, obj *model.Node, properties map[string]interface{}) (bool, error) {
	cmd := &fsm.CMD{
		Method: fsm.SetNodeProperties,
		SetNodeProperties: []fsm.SetNodeProperty{
			{
				Type:       obj.Type,
				ID:         obj.ID,
				Properties: properties,
			},
		},
	}
	bits, err := json.Marshal(cmd)
	if err != nil {
		return false, err
	}
	val, err := r.raft.Apply(bits)
	if err != nil {
		return false, err
	}
	if err, ok := val.(error); ok {
		return false, err
	}
	return true, nil
}

func (r *nodeResolver) GetRelationship(ctx context.Context, obj *model.Node, direction model.Direction, relationship string, id string) (*model.Relationship, error) {
	n, err := r.graph.GetNode(obj.Type, obj.ID)
	if err != nil {
		return nil, err
	}
	rel, ok := n.GetRelationship(api.Direction(direction), relationship, id)
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return toRelationship(rel), nil
}

func (r *nodeResolver) AddRelationship(ctx context.Context, obj *model.Node, direction model.Direction, relationship string, nodeKey model.Key) (*model.Relationship, error) {
	cmd := &fsm.CMD{
		Method: fsm.AddRelationships,
		AddRelationships: []fsm.AddRelationship{
			{
				NodeType:       obj.Type,
				NodeID:         obj.ID,
				Direction:      string(direction),
				Relationship:   relationship,
				RelationshipID: uuid.NewString(),
				Node2Type:      nodeKey.Type,
				Node2ID:        nodeKey.ID,
			},
		},
	}
	bits, err := json.Marshal(cmd)
	if err != nil {
		return nil, err
	}
	val, err := r.raft.Apply(bits)
	if err != nil {
		return nil, err
	}
	if err, ok := val.(error); ok {
		return nil, err
	}
	return toRelationship(val.([]api.Relationship)[0]), nil
}

func (r *nodeResolver) Relationships(ctx context.Context, obj *model.Node, direction model.Direction, typeArg string, filter *model.Filter) ([]*model.Relationship, error) {
	n, err := r.graph.GetNode(obj.Type, obj.ID)
	if err != nil {
		return nil, err
	}
	var rels []*model.Relationship
	n.Relationships(api.Direction(direction), typeArg, func(relationship api.Relationship) bool {

		for _, exp := range filter.Expressions {
			if !eval(exp, relationship) {
				return true
			}
		}
		rels = append(rels, toRelationship(relationship))
		return true
	})
	return rels, nil
}

func (r *queryResolver) NodeTypes(ctx context.Context) ([]string, error) {
	return r.graph.NodeTypes(), nil
}

func (r *queryResolver) GetNode(ctx context.Context, key model.Key) (*model.Node, error) {
	res, err := r.graph.GetNode(key.Type, key.ID)
	if err != nil {
		return nil, err
	}
	return toNode(res), nil
}

func (r *queryResolver) GetNodes(ctx context.Context, typeArg string, filter model.Filter) ([]*model.Node, error) {
	var nodes []*model.Node
	if err := r.graph.RangeNodes(typeArg, func(node api.Node) bool {
		for _, exp := range filter.Expressions {
			if !eval(exp, node) {
				return true
			}
		}
		nodes = append(nodes, toNode(node))
		return true
	}); err != nil {
		return nil, err
	}
	return nodes, nil
}

func (r *queryResolver) Size(ctx context.Context) (int, error) {
	return r.graph.Size(), nil
}

func (r *relationshipResolver) Properties(ctx context.Context, obj *model.Relationship, input *string) (map[string]interface{}, error) {
	n, err := r.graph.GetNode(obj.Type, obj.ID)
	if err != nil {
		return nil, err
	}
	return n.Properties(), nil
}

func (r *relationshipResolver) GetProperty(ctx context.Context, obj *model.Relationship, key string) (interface{}, error) {
	n, err := r.graph.GetNode(obj.Type, obj.ID)
	if err != nil {
		return false, err
	}
	return n.GetProperty(key), nil
}

func (r *relationshipResolver) SetProperties(ctx context.Context, obj *model.Relationship, properties map[string]interface{}) (bool, error) {
	cmd := &fsm.CMD{
		Method: fsm.SetRelationshipProperties,
		SetRelationshipProperties: []fsm.SetRelationshipProperty{
			{
				NodeType:       obj.Source.Type,
				NodeID:         obj.Source.ID,
				Relationship:   obj.Type,
				RelationshipID: obj.ID,
				Properties:     properties,
			},
		},
	}
	bits, err := json.Marshal(cmd)
	if err != nil {
		return false, err
	}
	val, err := r.raft.Apply(bits)
	if err != nil {
		return false, err
	}
	if err, ok := val.(error); ok {
		return false, err
	}
	return true, nil
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Node returns generated.NodeResolver implementation.
func (r *Resolver) Node() generated.NodeResolver { return &nodeResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

// Relationship returns generated.RelationshipResolver implementation.
func (r *Resolver) Relationship() generated.RelationshipResolver { return &relationshipResolver{r} }

type mutationResolver struct{ *Resolver }
type nodeResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type relationshipResolver struct{ *Resolver }
