package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"
	"github.com/google/uuid"

	"github.com/autom8ter/morpheus/pkg/graph/generated"
	"github.com/autom8ter/morpheus/pkg/graph/model"
)

func (r *mutationResolver) AddNode(ctx context.Context, typeArg string, properties map[string]interface{}) (*model.Node, error) {
	res, err := r.graph.AddNode(typeArg, uuid.New().String(), properties)
	if err != nil {
		return nil, err
	}
	return &model.Node{
		ID:                 res.ID(),
		Type:               res.Type(),
		Properties:         res.Properties(),
		GetProperty:        nil,
		SetProperty:        false,
		DelProperty:        false,
		AddRelationship:    nil,
		RemoveRelationship: false,
		Relationships:      nil,
	}, nil
}

func (r *mutationResolver) DelNode(ctx context.Context, key model.Key) (bool, error) {
	if err := r.graph.DelNode(key.Type, key.ID); err != nil {
		return false, err
	}
	return true, nil
}

func (r *queryResolver) NodeTypes(ctx context.Context) ([]string, error) {
	return r.graph.NodeTypes(), nil
}

func (r *queryResolver) GetNode(ctx context.Context, key model.Key) (*model.Node, error) {
	res, err := r.graph.GetNode(key.Type, key.ID)
	if err != nil {
		return nil, err
	}
	return &model.Node{
		ID:                 res.ID(),
		Type:               res.Type(),
		Properties:         res.Properties(),
		GetProperty:        nil,
		SetProperty:        false,
		DelProperty:        false,
		AddRelationship:    nil,
		RemoveRelationship: false,
		Relationships:      nil,
	}, nil
}

func (r *queryResolver) GetNodes(ctx context.Context, filter model.Filter) ([]*model.Node, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Size(ctx context.Context) (int, error) {
	return r.graph.Size(), nil
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
