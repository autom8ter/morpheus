package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"
	"github.com/palantir/stacktrace"
	"sort"

	"github.com/autom8ter/morpheus/pkg/api"
	"github.com/autom8ter/morpheus/pkg/auth"
	"github.com/autom8ter/morpheus/pkg/encode"
	"github.com/autom8ter/morpheus/pkg/graph/generated"
	"github.com/autom8ter/morpheus/pkg/graph/model"
	"github.com/autom8ter/morpheus/pkg/raft/fsm"
	"github.com/google/uuid"
	"github.com/spf13/cast"
)

func (r *mutationResolver) Add(ctx context.Context, add model.AddNode) (*model.Node, error) {
	if usr, ok := auth.GetUser(ctx); ok && usr.ReadOnly {
		return nil, fmt.Errorf("authorization failed: readonly user")
	}
	if add.ID == nil {
		id := uuid.New().String()
		add.ID = &id
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	cmd := &fsm.CMD{
		Method: fsm.AddNodes,
		AddNodes: []fsm.AddNode{
			{
				Type:       add.Type,
				ID:         *add.ID,
				Properties: add.Properties,
			},
		},
	}
	bits, err := encode.Marshal(cmd)
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

func (r *mutationResolver) Set(ctx context.Context, set model.SetNode) (*model.Node, error) {
	if usr, ok := auth.GetUser(ctx); ok && usr.ReadOnly {
		return nil, fmt.Errorf("authorization failed: readonly user")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	cmd := &fsm.CMD{
		Method: fsm.SetNodeProperties,
		SetNodeProperties: []fsm.SetNodeProperty{
			{
				Type:       set.Type,
				ID:         set.ID,
				Properties: set.Properties,
			},
		},
	}
	bits, err := encode.Marshal(cmd)
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

func (r *mutationResolver) Del(ctx context.Context, del model.Key) (bool, error) {
	if usr, ok := auth.GetUser(ctx); ok && usr.ReadOnly {
		return false, fmt.Errorf("authorization failed: readonly user")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	cmd := &fsm.CMD{
		Method: fsm.DelNodes,
		DelNodes: []fsm.DelNode{
			{
				Type: del.Type,
				ID:   del.ID,
			},
		},
	}
	bits, err := encode.Marshal(cmd)
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

func (r *mutationResolver) BulkAdd(ctx context.Context, add []*model.AddNode) (bool, error) {
	if usr, ok := auth.GetUser(ctx); ok && usr.ReadOnly {
		return false, fmt.Errorf("authorization failed: readonly user")
	}
	var toAdd []fsm.AddNode
	for _, a := range add {
		if a.ID == nil {
			id := uuid.New().String()
			a.ID = &id
		}
		toAdd = append(toAdd, fsm.AddNode{
			Type:       a.Type,
			ID:         *a.ID,
			Properties: a.Properties,
		})
	}
	cmd := &fsm.CMD{
		Method:   fsm.AddNodes,
		AddNodes: toAdd,
	}
	bits, err := encode.Marshal(cmd)
	if err != nil {
		return false, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	val, err := r.raft.Apply(bits)
	if err != nil {
		return false, err
	}
	if err, ok := val.(error); ok {
		return false, err
	}
	return true, nil
}

func (r *mutationResolver) BulkSet(ctx context.Context, set []*model.SetNode) (bool, error) {
	if usr, ok := auth.GetUser(ctx); ok && usr.ReadOnly {
		return false, fmt.Errorf("authorization failed: readonly user")
	}
	var setProperties []fsm.SetNodeProperty
	for _, s := range set {
		setProperties = append(setProperties, fsm.SetNodeProperty{
			Type:       s.Type,
			ID:         s.ID,
			Properties: s.Properties,
		})
	}
	cmd := &fsm.CMD{
		Method:            fsm.SetNodeProperties,
		SetNodeProperties: setProperties,
	}
	bits, err := encode.Marshal(cmd)
	if err != nil {
		return false, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	val, err := r.raft.Apply(bits)
	if err != nil {
		return false, err
	}
	if err, ok := val.(error); ok {
		return false, err
	}
	return true, nil
}

func (r *mutationResolver) BulkDel(ctx context.Context, del []*model.Key) (bool, error) {
	if usr, ok := auth.GetUser(ctx); ok && usr.ReadOnly {
		return false, fmt.Errorf("authorization failed: readonly user")
	}
	var delNodes []fsm.DelNode
	for _, d := range del {
		delNodes = append(delNodes, fsm.DelNode{
			Type: d.Type,
			ID:   d.ID,
		})
	}
	cmd := &fsm.CMD{
		Method:   fsm.DelNodes,
		DelNodes: delNodes,
	}
	bits, err := encode.Marshal(cmd)
	if err != nil {
		return false, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	val, err := r.raft.Apply(bits)
	if err != nil {
		return false, err
	}
	if err, ok := val.(error); ok {
		return false, err
	}
	return true, nil
}

func (r *nodeResolver) Properties(ctx context.Context, obj *model.Node) (map[string]interface{}, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	n, err := r.graph.GetNode(obj.Type, obj.ID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "")
	}
	return n.Properties(), nil
}

func (r *nodeResolver) GetProperty(ctx context.Context, obj *model.Node, key string) (interface{}, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	n, err := r.graph.GetNode(obj.Type, obj.ID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "")
	}
	return n.GetProperty(key), nil
}

func (r *nodeResolver) SetProperties(ctx context.Context, obj *model.Node, properties map[string]interface{}) (bool, error) {
	if usr, ok := auth.GetUser(ctx); ok && usr.ReadOnly {
		return false, fmt.Errorf("authorization failed: readonly user")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
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
	bits, err := encode.Marshal(cmd)
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
	r.mu.RLock()
	defer r.mu.RUnlock()
	n, err := r.graph.GetNode(obj.Type, obj.ID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "")
	}
	rel, ok := n.GetRelationship(api.Direction(direction), relationship, id)
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return toRelationship(rel), nil
}

func (r *nodeResolver) AddRelationship(ctx context.Context, obj *model.Node, direction model.Direction, relationship string, nodeKey model.Key) (*model.Relationship, error) {
	if usr, ok := auth.GetUser(ctx); ok && usr.ReadOnly {
		return nil, fmt.Errorf("authorization failed: readonly user")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
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
	bits, err := encode.Marshal(cmd)
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

func (r *nodeResolver) DelRelationship(ctx context.Context, obj *model.Node, direction model.Direction, key model.Key) (bool, error) {
	if usr, ok := auth.GetUser(ctx); ok && usr.ReadOnly {
		return false, fmt.Errorf("authorization failed: readonly user")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	cmd := &fsm.CMD{
		Method: fsm.DelRelationships,
		DelRelationships: []fsm.DelRelationship{
			{
				NodeType:       obj.Type,
				NodeID:         obj.ID,
				Relationship:   obj.Type,
				RelationshipID: obj.ID,
				Direction:      string(direction),
			},
		},
	}
	bits, err := encode.Marshal(cmd)
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

func (r *nodeResolver) Relationships(ctx context.Context, obj *model.Node, direction model.Direction, filter *model.Filter) (*model.Relationships, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	n, err := r.graph.GetNode(obj.Type, obj.ID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "")
	}
	skip := 0
	if filter.Cursor != nil {
		skp, err := r.parseCursor(*filter.Cursor)
		if err != nil {
			return nil, err
		}
		skip = skp
	}

	var rels []*model.Relationship
	pageSize := 25
	if filter.PageSize != nil {
		pageSize = *filter.PageSize
	}
	n.Relationships(skip, api.Direction(direction), filter.Type, func(relationship api.Relationship) bool {
		if len(rels) >= pageSize {
			return false
		}
		skip++
		for _, exp := range filter.Expressions {
			if !eval(exp, relationship) {
				return true
			}
		}
		rels = append(rels, toRelationship(relationship))
		return true
	})
	if len(rels) > 0 {
		if filter.OrderBy != nil && rels[0].Properties[filter.OrderBy.Field] != nil {
			sort.Slice(rels, func(i, j int) bool {
				if rels[i].Properties[filter.OrderBy.Field] == nil {
					return false
				}
				if rels[j].Properties[filter.OrderBy.Field] == nil {
					return true
				}
				if filter.OrderBy.Reverse != nil && *filter.OrderBy.Reverse {
					if val, err := cast.ToFloat64E(rels[i].Properties[filter.OrderBy.Field]); err == nil {
						return val < cast.ToFloat64(rels[j].Properties[filter.OrderBy.Field])
					}
					return cast.ToString(rels[i].Properties[filter.OrderBy.Field]) < cast.ToString(rels[j].Properties[filter.OrderBy.Field])
				}
				if val, err := cast.ToFloat64E(rels[i].Properties[filter.OrderBy.Field]); err == nil {
					return val > cast.ToFloat64(rels[j].Properties[filter.OrderBy.Field])
				}
				return cast.ToString(rels[i].Properties[filter.OrderBy.Field]) > cast.ToString(rels[j].Properties[filter.OrderBy.Field])
			})
		} else {
			if filter.OrderBy != nil && filter.OrderBy.Reverse != nil && *filter.OrderBy.Reverse {
				sort.Slice(rels, func(i, j int) bool {
					return rels[i].ID < rels[j].ID
				})
			} else {
				sort.Slice(rels, func(i, j int) bool {
					return rels[i].ID > rels[j].ID
				})
			}
		}
	}
	return &model.Relationships{
		Cursor:        r.createCursor(skip),
		Relationships: rels,
	}, nil
}

func (r *queryResolver) Types(ctx context.Context) ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.graph.NodeTypes(), nil
}

func (r *queryResolver) Get(ctx context.Context, key model.Key) (*model.Node, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	n, err := r.graph.GetNode(key.Type, key.ID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "")
	}
	return toNode(n), nil
}

func (r *queryResolver) List(ctx context.Context, filter model.Filter) (*model.Nodes, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var nodes []*model.Node
	limit := 25
	skip := 0
	if filter.Cursor != nil {
		skp, err := r.parseCursor(*filter.Cursor)
		if err != nil {
			return nil, err
		}
		skip = skp
	}
	if filter.PageSize != nil {
		limit = *filter.PageSize
	}
	if err := r.graph.RangeNodes(skip, filter.Type, func(node api.Node) bool {
		if len(nodes) >= limit {
			return false
		}
		skip++
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
	if len(nodes) > 0 {
		if filter.OrderBy != nil && nodes[0].Properties[filter.OrderBy.Field] != nil {
			sort.Slice(nodes, func(i, j int) bool {
				if nodes[i].Properties[filter.OrderBy.Field] == nil {
					return false
				}
				if nodes[j].Properties[filter.OrderBy.Field] == nil {
					return true
				}
				if filter.OrderBy.Reverse != nil && *filter.OrderBy.Reverse {
					if val, err := cast.ToFloat64E(nodes[i].Properties[filter.OrderBy.Field]); err == nil {
						return val < cast.ToFloat64(nodes[j].Properties[filter.OrderBy.Field])
					}
					return cast.ToString(nodes[i].Properties[filter.OrderBy.Field]) < cast.ToString(nodes[j].Properties[filter.OrderBy.Field])
				}
				if val, err := cast.ToFloat64E(nodes[i].Properties[filter.OrderBy.Field]); err == nil {
					return val > cast.ToFloat64(nodes[j].Properties[filter.OrderBy.Field])
				}
				return cast.ToString(nodes[i].Properties[filter.OrderBy.Field]) > cast.ToString(nodes[j].Properties[filter.OrderBy.Field])
			})
		} else {
			if filter.OrderBy != nil && filter.OrderBy.Reverse != nil && *filter.OrderBy.Reverse {
				sort.Slice(nodes, func(i, j int) bool {
					return nodes[i].ID < nodes[j].ID
				})
			} else {
				sort.Slice(nodes, func(i, j int) bool {
					return nodes[i].ID > nodes[j].ID
				})
			}
		}
	}
	return &model.Nodes{
		Cursor: r.createCursor(skip),
		Nodes:  nodes,
	}, nil
}

func (r *queryResolver) Size(ctx context.Context) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.graph.Size(), nil
}

func (r *relationshipResolver) Properties(ctx context.Context, obj *model.Relationship) (map[string]interface{}, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	n, err := r.graph.GetRelationship(obj.Type, obj.ID)
	if err != nil {
		return nil, err
	}
	return n.Properties(), nil
}

func (r *relationshipResolver) GetProperty(ctx context.Context, obj *model.Relationship, key string) (interface{}, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	n, err := r.graph.GetRelationship(obj.Type, obj.ID)
	if err != nil {
		return false, err
	}
	return n.GetProperty(key), nil
}

func (r *relationshipResolver) SetProperties(ctx context.Context, obj *model.Relationship, properties map[string]interface{}) (bool, error) {
	if usr, ok := auth.GetUser(ctx); ok && usr.ReadOnly {
		return false, fmt.Errorf("authorization failed: readonly user")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
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
	bits, err := encode.Marshal(cmd)
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
