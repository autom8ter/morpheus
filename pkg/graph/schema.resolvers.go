package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/autom8ter/morpheus/pkg/api"
	"github.com/autom8ter/morpheus/pkg/config"
	"github.com/autom8ter/morpheus/pkg/constants"
	"github.com/autom8ter/morpheus/pkg/graph/fsm"
	"github.com/autom8ter/morpheus/pkg/graph/generated"
	"github.com/autom8ter/morpheus/pkg/graph/model"
	"github.com/autom8ter/morpheus/pkg/helpers"
	"github.com/autom8ter/morpheus/pkg/logger"
	"github.com/google/uuid"
	"github.com/palantir/stacktrace"
)

func (r *nodeResolver) Properties(ctx context.Context, obj *model.Node) (map[string]interface{}, error) {
	op := graphql.GetOperationContext(ctx)
	_, err := r.mw.RequireRole(ctx, config.READER)
	if err != nil {
		return nil, stacktrace.RootCause(err)
	}
	n, err := r.graph.GetNode(obj.Type, obj.ID)
	if err != nil {
		logger.L.Error("graphql resolver error", stacktrace.Propagate(err, ""), map[string]interface{}{
			"operation.name": op.OperationName,
			"node.type":      obj.Type,
			"node.id":        obj.ID,
		})
		return nil, stacktrace.RootCause(err)
	}
	props, err := n.Properties()
	if err != nil {
		logger.L.Error("graphql resolver error", stacktrace.Propagate(err, ""), map[string]interface{}{
			"operation.name": op.OperationName,
			"node.type":      obj.Type,
			"node.id":        obj.ID,
		})
		return nil, stacktrace.RootCause(err)
	}
	return props, nil
}

func (r *nodeResolver) GetProperty(ctx context.Context, obj *model.Node, key string) (interface{}, error) {
	op := graphql.GetOperationContext(ctx)
	_, err := r.mw.RequireRole(ctx, config.READER)
	if err != nil {
		return nil, stacktrace.RootCause(err)
	}
	n, err := r.graph.GetNode(obj.Type, obj.ID)
	if err != nil {
		logger.L.Error("graphql resolver error", stacktrace.Propagate(err, ""), map[string]interface{}{
			"operation.name": op.OperationName,
			"node.type":      obj.Type,
			"node.id":        obj.ID,
		})
		return nil, stacktrace.RootCause(err)
	}
	val, err := n.GetProperty(key)
	if err != nil {
		logger.L.Error("graphql resolver error", stacktrace.Propagate(err, ""), map[string]interface{}{
			"operation.name": op.OperationName,
			"node.type":      obj.Type,
			"node.id":        obj.ID,
		})
		return nil, stacktrace.RootCause(err)
	}
	return val, nil
}

func (r *nodeResolver) SetProperties(ctx context.Context, obj *model.Node, properties map[string]interface{}) (bool, error) {
	op := graphql.GetOperationContext(ctx)
	_, err := r.mw.RequireRole(ctx, config.WRITER)
	if err != nil {
		return false, stacktrace.Propagate(err, "")
	}
	cmd := &fsm.CMD{
		Method:     fsm.MethodNodeSetProperties,
		Properties: properties,
		Timestamp:  time.Now(),
		Metadata: map[string]string{
			"id":   obj.ID,
			"type": obj.Type,
		},
	}
	_, err = r.applyCMD(cmd)
	if err != nil {
		logger.L.Error("graphql resolver error", stacktrace.Propagate(err, ""), map[string]interface{}{
			"operation.name": op.OperationName,
			"node.type":      obj.Type,
			"node.id":        obj.ID,
		})
		return false, stacktrace.RootCause(err)
	}
	return true, nil
}

func (r *nodeResolver) GetRelationship(ctx context.Context, obj *model.Node, relationship string, id string) (*model.Relationship, error) {
	op := graphql.GetOperationContext(ctx)
	_, err := r.mw.RequireRole(ctx, config.READER)
	if err != nil {
		return nil, stacktrace.RootCause(err)
	}
	n, err := r.graph.GetNode(obj.Type, obj.ID)
	if err != nil {
		return nil, stacktrace.RootCause(stacktrace.Propagate(err, ""))
	}
	rel, ok, err := n.GetRelationship(relationship, id)
	if err != nil {
		logger.L.Error("graphql resolver error", stacktrace.Propagate(err, ""), map[string]interface{}{
			"operation.name": op.OperationName,
			"node.type":      obj.Type,
			"node.id":        obj.ID,
		})
		return nil, stacktrace.RootCause(err)
	}
	if !ok {
		return nil, stacktrace.RootCause(constants.ErrNotFound)
	}
	resp, err := toRelationship(rel)
	if err != nil {
		logger.L.Error("graphql resolver error", stacktrace.Propagate(err, ""), map[string]interface{}{
			"operation.name": op.OperationName,
			"node.type":      obj.Type,
			"node.id":        obj.ID,
		})
		return nil, stacktrace.RootCause(stacktrace.Propagate(err, ""))
	}
	return resp, nil
}

func (r *nodeResolver) AddRelationship(ctx context.Context, obj *model.Node, relationship string, nodeKey model.Key) (*model.Relationship, error) {
	op := graphql.GetOperationContext(ctx)
	_, err := r.mw.RequireRole(ctx, config.WRITER)
	if err != nil {
		return nil, stacktrace.RootCause(err)
	}
	cmd := &fsm.CMD{
		Method:    fsm.MethodNodeAddRelation,
		Key:       nodeKey,
		Timestamp: time.Now(),
		Metadata: map[string]string{
			"source.type":  obj.Type,
			"source.id":    obj.ID,
			"relationship": relationship,
		},
	}
	val, err := r.applyCMD(cmd)
	if err != nil {
		logger.L.Error("graphql resolver error", stacktrace.Propagate(err, ""), map[string]interface{}{
			"operation.name": op.OperationName,
			"node.type":      obj.Type,
			"node.id":        obj.ID,
			"relationship":   relationship,
		})
		return nil, stacktrace.RootCause(err)
	}
	rel, err := toRelationship(val.(api.Relationship))
	if err != nil {
		logger.L.Error("graphql resolver error", stacktrace.Propagate(err, ""), map[string]interface{}{
			"operation.name": op.OperationName,
			"node.type":      obj.Type,
			"node.id":        obj.ID,
		})
		return nil, stacktrace.RootCause(err)
	}
	return rel, nil
}

func (r *nodeResolver) DelRelationship(ctx context.Context, obj *model.Node, key model.Key) (bool, error) {
	op := graphql.GetOperationContext(ctx)
	_, err := r.mw.RequireRole(ctx, config.WRITER)
	if err != nil {
		return false, stacktrace.Propagate(err, "")
	}
	cmd := &fsm.CMD{
		Method:    fsm.MethodNodeDelRelation,
		Key:       key,
		Timestamp: time.Now(),
		Metadata: map[string]string{
			"id":   obj.Type,
			"type": obj.ID,
		},
	}
	_, err = r.applyCMD(cmd)
	if err != nil {
		logger.L.Error("graphql resolver error", stacktrace.Propagate(err, ""), map[string]interface{}{
			"operation.name": op.OperationName,
			"node.type":      obj.Type,
			"node.id":        obj.ID,
		})
		return false, stacktrace.RootCause(err)
	}
	return true, nil
}

func (r *nodeResolver) Relationships(ctx context.Context, obj *model.Node, where model.RelationWhere) (*model.Relationships, error) {
	op := graphql.GetOperationContext(ctx)
	_, err := r.mw.RequireRole(ctx, config.READER)
	if err != nil {
		return nil, stacktrace.RootCause(err)
	}
	n, err := r.graph.GetNode(obj.Type, obj.ID)
	if err != nil {
		logger.L.Error("failed to list relationships", stacktrace.Propagate(err, ""), map[string]interface{}{
			"operation.name": op.OperationName,
			"node.type":      obj.Type,
			"node.id":        obj.ID,
		})
		return nil, stacktrace.RootCause(err)
	}
	cursor, rels, err := n.Relationships(&where)
	if err != nil {
		logger.L.Error("failed to list relationships", stacktrace.Propagate(err, ""), map[string]interface{}{
			"operation.name": op.OperationName,
			"node.type":      obj.Type,
			"node.id":        obj.ID,
		})
		return nil, stacktrace.RootCause(err)
	}
	var resp []*model.Relationship
	for _, rel := range rels {
		i, err := toRelationship(rel)
		if err != nil {
			logger.L.Error("graphql resolver error", stacktrace.Propagate(err, ""), map[string]interface{}{
				"node.type":     obj.Type,
				"node.id":       obj.ID,
				"relation.type": rel.Type(),
				"relation.id":   rel.ID(),
			})
			return nil, stacktrace.RootCause(err)
		}
		resp = append(resp, i)
	}
	return &model.Relationships{
		Cursor: cursor,
		Values: resp,
	}, nil
}

func (r *nodeResolver) AddIncomingNode(ctx context.Context, obj *model.Node, relation string, addNode model.AddNode) (*model.Node, error) {
	n, err := r.Query().Add(ctx, addNode)
	if err != nil {
		logger.L.Error("graphql resolver error", stacktrace.Propagate(err, ""), map[string]interface{}{
			"node.type":     obj.Type,
			"node.id":       obj.ID,
			"relation.type": relation,
		})
		return nil, stacktrace.RootCause(err)
	}
	if _, err := r.Node().AddRelationship(ctx, n, relation, model.Key{
		Type: obj.Type,
		ID:   obj.ID,
	}); err != nil {
		return nil, stacktrace.RootCause(err)
	}
	return n, nil
}

func (r *nodeResolver) AddOutboundNode(ctx context.Context, obj *model.Node, relation string, addNode model.AddNode) (*model.Node, error) {
	n, err := r.Query().Add(ctx, addNode)
	if err != nil {
		logger.L.Error("graphql resolver error", stacktrace.Propagate(err, ""), map[string]interface{}{
			"node.type":     obj.Type,
			"node.id":       obj.ID,
			"relation.type": relation,
		})
		return nil, stacktrace.RootCause(err)
	}
	if _, err := r.Node().AddRelationship(ctx, obj, relation, model.Key{
		Type: n.Type,
		ID:   n.ID,
	}); err != nil {
		return nil, stacktrace.RootCause(err)
	}
	return n, nil
}

func (r *queryResolver) Types(ctx context.Context) ([]string, error) {
	op := graphql.GetOperationContext(ctx)
	_, err := r.mw.RequireRole(ctx, config.READER)
	if err != nil {
		logger.L.Error("failed to list relationships", stacktrace.Propagate(err, ""), map[string]interface{}{
			"operation.name": op.OperationName,
		})
		return nil, stacktrace.RootCause(err)
	}
	return r.graph.NodeTypes(), nil
}

func (r *queryResolver) Get(ctx context.Context, key model.Key) (*model.Node, error) {
	op := graphql.GetOperationContext(ctx)
	_, err := r.mw.RequireRole(ctx, config.READER)
	if err != nil {
		return nil, stacktrace.RootCause(err)
	}
	n, err := r.graph.GetNode(key.Type, key.ID)
	if err != nil {
		logger.L.Error("graphql resolver error", stacktrace.Propagate(err, ""), map[string]interface{}{
			"operation.name": op.OperationName,
			"node.type":      key.Type,
			"node.id":        key.ID,
		})
		return nil, stacktrace.RootCause(err)
	}
	node, err := toNode(n)
	if err != nil {
		logger.L.Error("graphql resolver error", stacktrace.Propagate(err, ""), map[string]interface{}{
			"operation.name": op.OperationName,
			"node.type":      key.Type,
			"node.id":        key.ID,
		})
		return nil, stacktrace.RootCause(err)
	}
	return node, nil
}

func (r *queryResolver) List(ctx context.Context, where model.NodeWhere) (*model.Nodes, error) {
	op := graphql.GetOperationContext(ctx)
	_, err := r.mw.RequireRole(ctx, config.READER)
	if err != nil {
		return nil, stacktrace.RootCause(err)
	}
	cursor, nodes, err := r.graph.RangeNodes(&where)
	if err != nil {
		logger.L.Error("graphql resolver error", stacktrace.Propagate(err, ""), map[string]interface{}{
			"operation.name": op.OperationName,
			"where.type":     where.Type,
			"where.cursor":   where.Cursor,
		})
		return nil, stacktrace.RootCause(err)
	}
	var resp = &model.Nodes{Cursor: cursor}
	for _, node := range nodes {
		n, err := toNode(node)
		if err != nil {
			logger.L.Error("graphql resolver error", stacktrace.Propagate(err, ""), map[string]interface{}{
				"operation.name": op.OperationName,
				"where.type":     where.Type,
				"where.cursor":   where.Cursor,
			})
			return nil, stacktrace.RootCause(err)
		}
		resp.Values = append(resp.Values, n)
	}
	return resp, nil
}

func (r *queryResolver) Add(ctx context.Context, add model.AddNode) (*model.Node, error) {
	op := graphql.GetOperationContext(ctx)
	_, err := r.mw.RequireRole(ctx, config.WRITER)
	if err != nil {
		return nil, stacktrace.RootCause(err)
	}
	a := &add
	if a.ID == nil {
		id := uuid.New().String()
		a.ID = &id
	}
	cmd := &fsm.CMD{
		Method: fsm.MethodAdd,
		Node: model.Node{
			ID:         *a.ID,
			Type:       a.Type,
			Properties: a.Properties,
		},
		Timestamp: time.Now(),
		Metadata:  nil,
	}
	result, err := r.applyCMD(cmd)
	if err != nil {
		logger.L.Error("graphql resolver error", stacktrace.Propagate(err, ""), map[string]interface{}{
			"operation.name": op.OperationName,
			"node.id":        *add.ID,
			"node.type":      add.Type,
		})
		return nil, stacktrace.RootCause(err)
	}
	node, err := toNode(result.(api.Node))
	if err != nil {
		logger.L.Error("graphql resolver error", stacktrace.Propagate(err, ""), map[string]interface{}{
			"operation.name": op.OperationName,
			"node.id":        *add.ID,
			"node.type":      add.Type,
		})
		return nil, stacktrace.RootCause(err)
	}
	return node, nil
}

func (r *queryResolver) Set(ctx context.Context, set model.SetNode) (*model.Node, error) {
	op := graphql.GetOperationContext(ctx)
	_, err := r.mw.RequireRole(ctx, config.WRITER)
	if err != nil {
		return nil, stacktrace.RootCause(err)
	}
	cmd := &fsm.CMD{
		Method: fsm.MethodSet,
		Node: model.Node{
			ID:         set.ID,
			Type:       set.Type,
			Properties: set.Properties,
		},
		Timestamp: time.Now(),
	}
	result, err := r.applyCMD(cmd)
	if err != nil {
		logger.L.Error("graphql resolver error", stacktrace.Propagate(err, ""), map[string]interface{}{
			"operation.name": op.OperationName,
			"node.id":        set.ID,
			"node.type":      set.Type,
		})
		return nil, stacktrace.RootCause(err)
	}
	n, err := toNode(result.(api.Node))
	if err != nil {
		logger.L.Error("graphql resolver error", stacktrace.Propagate(err, ""), map[string]interface{}{
			"operation.name": op.OperationName,
			"node.id":        set.ID,
			"node.type":      set.Type,
		})
		return nil, stacktrace.RootCause(err)
	}
	return n, nil
}

func (r *queryResolver) Del(ctx context.Context, del model.Key) (bool, error) {
	op := graphql.GetOperationContext(ctx)
	_, err := r.mw.RequireRole(ctx, config.WRITER)
	if err != nil {
		return false, stacktrace.Propagate(err, "")
	}
	cmd := &fsm.CMD{
		Method:    fsm.MethodDel,
		Key:       del,
		Timestamp: time.Now(),
	}
	_, err = r.applyCMD(cmd)
	if err != nil {
		logger.L.Error("graphql resolver error", stacktrace.Propagate(err, ""), map[string]interface{}{
			"operation.name": op.OperationName,
			"node.id":        del.ID,
			"node.type":      del.Type,
		})
		return false, stacktrace.RootCause(err)
	}
	return true, nil
}

func (r *queryResolver) BulkAdd(ctx context.Context, add []*model.AddNode) (bool, error) {
	op := graphql.GetOperationContext(ctx)
	_, err := r.mw.RequireRole(ctx, config.WRITER)
	if err != nil {
		return false, stacktrace.Propagate(err, "")
	}

	for _, a := range add {
		if a.ID == nil {
			id := uuid.New().String()
			a.ID = &id
		}
	}
	cmd := &fsm.CMD{
		Method:    fsm.MethodBulkAdd,
		AddNodes:  add,
		Timestamp: time.Now(),
		Metadata:  nil,
	}
	_, err = r.applyCMD(cmd)
	if err != nil {
		logger.L.Error("graphql resolver error", stacktrace.Propagate(err, ""), map[string]interface{}{
			"operation.name": op.OperationName,
		})
		return false, stacktrace.RootCause(err)
	}
	return true, nil
}

func (r *queryResolver) BulkSet(ctx context.Context, set []*model.SetNode) (bool, error) {
	op := graphql.GetOperationContext(ctx)
	_, err := r.mw.RequireRole(ctx, config.WRITER)
	if err != nil {
		return false, stacktrace.Propagate(err, "")
	}
	cmd := &fsm.CMD{
		Method:    fsm.MethodBulkSet,
		SetNodes:  set,
		Timestamp: time.Now(),
	}
	_, err = r.applyCMD(cmd)
	if err != nil {
		logger.L.Error("graphql resolver error", stacktrace.Propagate(err, ""), map[string]interface{}{
			"operation.name": op.OperationName,
		})
		return false, stacktrace.RootCause(err)
	}
	return true, nil
}

func (r *queryResolver) BulkDel(ctx context.Context, del []*model.Key) (bool, error) {
	op := graphql.GetOperationContext(ctx)
	_, err := r.mw.RequireRole(ctx, config.WRITER)
	if err != nil {
		return false, stacktrace.Propagate(err, "")
	}
	cmd := &fsm.CMD{
		Method:    fsm.MethodBulkDel,
		Keys:      del,
		Timestamp: time.Now(),
		Metadata:  map[string]string{},
	}
	_, err = r.applyCMD(cmd)
	if err != nil {
		logger.L.Error("graphql resolver error", stacktrace.Propagate(err, ""), map[string]interface{}{
			"operation.name": op.OperationName,
		})
		return false, stacktrace.RootCause(err)
	}
	return true, nil
}

func (r *queryResolver) Login(ctx context.Context, username string, password string) (string, error) {
	op := graphql.GetOperationContext(ctx)
	token, err := r.mw.Login(username, password)
	if err != nil {
		logger.L.Error("graphql resolver error", stacktrace.Propagate(err, ""), map[string]interface{}{
			"operation.name": op.OperationName,
			"username":       username,
		})
		return "", stacktrace.RootCause(err)
	}
	expired, _, err := helpers.JWTExpired(token)
	if err != nil {
		logger.L.Error("graphql resolver error", stacktrace.Propagate(err, ""), map[string]interface{}{
			"operation.name": op.OperationName,
			"username":       username,
		})
		return "", stacktrace.RootCause(err)
	}
	if expired {
		return "", fmt.Errorf("expired jwt (internal): %s", token)
	}
	return token, nil
}

func (r *relationshipResolver) Properties(ctx context.Context, obj *model.Relationship) (map[string]interface{}, error) {
	op := graphql.GetOperationContext(ctx)
	_, err := r.mw.RequireRole(ctx, config.READER)
	if err != nil {
		return nil, stacktrace.RootCause(err)
	}
	n, err := r.graph.GetRelationship(obj.Type, obj.ID)
	if err != nil {
		logger.L.Error("graphql resolver error", stacktrace.Propagate(err, ""), map[string]interface{}{
			"operation.name": op.OperationName,
			"relation.id":    obj.ID,
			"relation.type":  obj.Type,
		})
		return nil, stacktrace.RootCause(err)
	}
	props, err := n.Properties()
	if err != nil {
		logger.L.Error("graphql resolver error", stacktrace.Propagate(err, ""), map[string]interface{}{
			"operation.name": op.OperationName,
			"relation.id":    obj.ID,
			"relation.type":  obj.Type,
		})
		return nil, stacktrace.RootCause(err)
	}
	return props, nil
}

func (r *relationshipResolver) GetProperty(ctx context.Context, obj *model.Relationship, key string) (interface{}, error) {
	op := graphql.GetOperationContext(ctx)
	_, err := r.mw.RequireRole(ctx, config.READER)
	if err != nil {
		return false, stacktrace.Propagate(err, "")
	}
	n, err := r.graph.GetRelationship(obj.Type, obj.ID)
	if err != nil {
		logger.L.Error("graphql resolver error", stacktrace.Propagate(err, ""), map[string]interface{}{
			"operation.name": op.OperationName,
			"relation.id":    obj.ID,
			"relation.type":  obj.Type,
		})
		return false, stacktrace.RootCause(err)
	}
	val, err := n.GetProperty(key)
	if err != nil {
		logger.L.Error("graphql resolver error", stacktrace.Propagate(err, ""), map[string]interface{}{
			"operation.name": op.OperationName,
			"relation.id":    obj.ID,
			"relation.type":  obj.Type,
		})
		return false, stacktrace.RootCause(err)
	}
	return val, nil
}

func (r *relationshipResolver) SetProperties(ctx context.Context, obj *model.Relationship, properties map[string]interface{}) (bool, error) {
	op := graphql.GetOperationContext(ctx)
	_, err := r.mw.RequireRole(ctx, config.WRITER)
	if err != nil {
		return false, stacktrace.Propagate(err, "")
	}
	cmd := &fsm.CMD{
		Method:     fsm.MethodRelationSetProperties,
		Properties: properties,
		Timestamp:  time.Now(),
		Metadata: map[string]string{
			"id":   obj.ID,
			"type": obj.Type,
		},
	}
	_, err = r.applyCMD(cmd)
	if err != nil {
		logger.L.Error("graphql resolver error", stacktrace.Propagate(err, ""), map[string]interface{}{
			"error":          stacktrace.Propagate(err, ""),
			"operation.name": op.OperationName,
			"relation.id":    obj.ID,
			"relation.type":  obj.Type,
		})
		return false, stacktrace.RootCause(err)
	}
	return true, nil
}

// Node returns generated.NodeResolver implementation.
func (r *Resolver) Node() generated.NodeResolver { return &nodeResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

// Relationship returns generated.RelationshipResolver implementation.
func (r *Resolver) Relationship() generated.RelationshipResolver { return &relationshipResolver{r} }

type nodeResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type relationshipResolver struct{ *Resolver }
