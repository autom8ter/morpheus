package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/autom8ter/morpheus/pkg/api"
	"github.com/autom8ter/morpheus/pkg/graph/generated"
	"github.com/autom8ter/morpheus/pkg/graph/model"
	"github.com/google/uuid"
)

func (r *mutationResolver) AddNode(ctx context.Context, typeArg string, properties map[string]interface{}) (*model.Node, error) {
	res, err := r.graph.AddNode(typeArg, uuid.New().String(), properties)
	if err != nil {
		return nil, err
	}
	return toNode(res), nil
}

func (r *mutationResolver) DelNode(ctx context.Context, key model.Key) (bool, error) {
	if err := r.graph.DelNode(key.Type, key.ID); err != nil {
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
	n, err := r.graph.GetNode(obj.Type, obj.ID)
	if err != nil {
		return false, err
	}
	n.SetProperties(properties)
	return true, nil
}

func (r *nodeResolver) AddRelationship(ctx context.Context, obj *model.Node, direction model.Direction, relationship string, nodeKey model.Key) (*model.Relationship, error) {
	n, err := r.graph.GetNode(obj.Type, obj.ID)
	if err != nil {
		return nil, err
	}
	target, err := r.graph.GetNode(nodeKey.Type, nodeKey.ID)
	if err != nil {
		return nil, err
	}
	rel := n.AddRelationship(api.Direction(direction), relationship, uuid.NewString(), target)
	return toRelationship(rel), nil
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
	n, err := r.graph.GetNode(obj.Type, obj.ID)
	if err != nil {
		return false, err
	}
	n.SetProperties(properties)
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
