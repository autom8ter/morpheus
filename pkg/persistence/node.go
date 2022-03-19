package persistence

import (
	"github.com/autom8ter/morpheus/pkg/api"
	"github.com/autom8ter/morpheus/pkg/constants"
	"github.com/autom8ter/morpheus/pkg/encode"
	"github.com/autom8ter/morpheus/pkg/graph/model"
	"github.com/dgraph-io/badger/v3"
	"github.com/palantir/stacktrace"
	"github.com/spf13/cast"
	"sort"
	"strings"
)

type Node struct {
	nodeType string
	nodeID   string
	data     map[string]interface{}
	db       *DB
}

func (n Node) Type() string {
	return n.nodeType
}

func (n Node) ID() string {
	return n.nodeID
}

func (n Node) Properties() (map[string]interface{}, error) {
	if n.data != nil {
		return n.data, nil
	}
	if err := n.db.db.View(func(txn *badger.Txn) error {
		var key = getNodePath(n.nodeType, n.nodeID)
		data := map[string]interface{}{}
		item, err := txn.Get(key)
		if err != nil {
			return stacktrace.Propagate(err, "failed to get key %s", string(key))
		}
		if err := item.Value(func(val []byte) error {
			return encode.Unmarshal(val, &data)
		}); err != nil {
			return stacktrace.Propagate(err, "failed to get key %s", string(key))
		}
		n.data = data
		return nil
	}); err != nil {
		return nil, err
	}
	return n.data, nil
}

func (n Node) GetProperty(name string) (interface{}, error) {
	all, err := n.Properties()
	if err != nil {
		return nil, stacktrace.Propagate(err, "")
	}
	if all == nil {
		return nil, stacktrace.Propagate(constants.ErrNotFound, "")
	}
	return all[name], nil
}

func (n Node) SetProperties(properties map[string]interface{}) error {
	if properties == nil {
		properties = map[string]interface{}{}
	}
	_, err := n.db.AddNode(n.nodeType, n.nodeID, properties)
	if err != nil {
		return stacktrace.Propagate(err, "")
	}
	return nil
}

func (n Node) DelProperty(name string) error {
	all, err := n.Properties()
	if err != nil {
		return stacktrace.Propagate(err, "")
	}
	delete(all, name)
	if err := n.SetProperties(all); err != nil {
		return stacktrace.Propagate(err, "")
	}
	return nil
}

func (n Node) AddRelation(direction api.Direction, relation string, properties map[string]interface{}, node api.Node) (api.Relation, error) {
	if properties == nil {
		properties = map[string]interface{}{}
	}
	relID := getRelationID(n.Type(), n.ID(), relation, node.Type(), node.ID())
	n.db.relationTypes.Store(relation, struct{}{})
	rkey := getRelationPath(relation, relID)
	var sourceNode api.Node
	var targetNode api.Node
	if direction == api.Outgoing {
		targetNode = node
		sourceNode = n
	} else {
		targetNode = n
		sourceNode = node
	}
	source := getNodeRelationPath(sourceNode.Type(), sourceNode.ID(), direction, relation, targetNode.Type(), targetNode.ID(), relID)
	target := getNodeRelationPath(targetNode.Type(), targetNode.ID(), direction.Opposite(), relation, sourceNode.Type(), sourceNode.ID(), relID)

	properties[Internal_Direction] = direction
	properties[Internal_SourceType] = sourceNode.Type()
	properties[Internal_SourceID] = sourceNode.ID()
	properties[Internal_TargetType] = targetNode.Type()
	properties[Internal_TargetID] = targetNode.ID()
	properties[Internal_ID] = relID
	properties[Internal_Relation] = relation
	properties[Internal_Type] = relation

	bits, err := encode.Marshal(&properties)
	if err != nil {
		return nil, stacktrace.Propagate(err, "")
	}
	if err := n.db.db.Update(func(txn *badger.Txn) error {
		if err := txn.Set(rkey, bits); err != nil {
			return stacktrace.Propagate(err, "")
		}
		if err := txn.Set(source, bits); err != nil {
			return stacktrace.Propagate(err, "")
		}
		if err := txn.Set(target, bits); err != nil {
			return stacktrace.Propagate(err, "")
		}
		for k, v := range properties {
			n.db.relationFieldMap.Store(strings.Join([]string{relation, k}, ","), struct{}{})
			key := getRelationFieldPath(relation, k, v, relID)
			if err := txn.Set(key, bits); err != nil {
				return stacktrace.Propagate(err, "")
			}
		}
		return nil
	}); err != nil {
		return nil, stacktrace.Propagate(err, "")
	}
	r := &Relation{
		relationType: relation,
		relationID:   relID,
		item:         properties,
		db:           n.db,
	}
	n.db.cache.Set(string(rkey), r, 1)
	return r, nil
}

func (n Node) DelRelation(relation string, id string) error {
	rel, ok, err := n.GetRelation(relation, id)
	if err != nil {
		return stacktrace.Propagate(err, "")
	}
	if !ok {
		return stacktrace.Propagate(constants.ErrNotFound, "")
	}
	targetNode, err := rel.Target()
	if err != nil {
		return stacktrace.Propagate(err, "")
	}

	rkey := getRelationPath(rel.Type(), rel.ID())
	source := getNodeRelationPath(n.Type(), n.ID(), api.Outgoing, relation, targetNode.Type(), targetNode.ID(), rel.ID())
	target := getNodeRelationPath(targetNode.Type(), targetNode.ID(), api.Incoming, relation, n.Type(), n.ID(), rel.ID())
	if err := n.db.db.Update(func(txn *badger.Txn) error {
		if err := txn.Delete(rkey); err != nil {
			return stacktrace.Propagate(err, "")
		}
		if err := txn.Delete(source); err != nil {
			return stacktrace.Propagate(err, "")
		}
		if err := txn.Delete(target); err != nil {
			return stacktrace.Propagate(err, "")
		}
		props, err := rel.Properties()
		if err != nil {
			return stacktrace.Propagate(err, "")
		}
		if err := n.db.db.View(func(txn *badger.Txn) error {
			for k, v := range props {
				key := getRelationFieldPath(rel.Type(), k, v, rel.ID())
				if err := txn.Delete(key); err != nil {
					return stacktrace.Propagate(err, "")
				}
			}
			return nil
		}); err != nil {
			return stacktrace.Propagate(err, "")
		}
		return nil
	}); err != nil {
		return stacktrace.Propagate(err, "")
	}
	n.db.cache.Del(string(rkey))
	return nil
}

func (n Node) GetRelation(relation, id string) (api.Relation, bool, error) {
	rkey := getRelationPath(relation, id)
	if val, ok := n.db.cache.Get(string(rkey)); ok {
		return val.(api.Relation), true, nil
	}

	rel := &Relation{
		relationType: relation,
		relationID:   id,
		item:         map[string]interface{}{},
		db:           n.db,
	}
	if err := n.db.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(rkey)
		if err != nil {
			return stacktrace.Propagate(err, "")
		}
		if err := item.Value(func(val []byte) error {
			return encode.Unmarshal(val, &rel.item)
		}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, false, stacktrace.Propagate(err, "")
	}
	if len(rel.item) == 0 {
		return nil, false, stacktrace.Propagate(constants.ErrNotFound, "")
	}
	return rel, true, nil
}

func (n Node) Relations(where *model.RelationWhere) (string, []api.Relation, error) {
	source := getNodeRelationPath(n.Type(), n.ID(), api.Direction(where.Direction), where.Relation, where.TargetType, "", "")
	// Iterate over 1000 items
	var (
		skipped int
		skip    int
		rels    []api.Relation
		err     error
	)
	if where.Cursor != nil {
		skip, err = parseCursor(*where.Cursor)
		if err != nil {
			return "", nil, stacktrace.Propagate(err, "")
		}
	}
	if where.PageSize == nil {
		defaultSize := prefetchSize
		where.PageSize = &defaultSize
	}

	if err := n.db.db.View(func(txn *badger.Txn) error {
		opt := badger.DefaultIteratorOptions
		opt.PrefetchSize = prefetchSize
		it := txn.NewIterator(opt)
		defer it.Close()
		for it.Seek(source); it.ValidForPrefix(source); it.Next() {
			if len(rels) >= *where.PageSize {
				return nil
			}
			if skipped < skip {
				skipped++
				continue
			}
			item := it.Item()
			split := strings.Split(string(item.Key()), ",")
			var rel api.Relation
			cached, ok := n.db.cache.Get(string(getRelationPath(where.Relation, split[len(split)-1])))
			if ok {
				rel = cached.(api.Relation)
			} else {
				rel = &Relation{
					relationType: where.Relation,
					relationID:   split[len(split)-1],
					item:         nil,
					db:           n.db,
				}
			}
			passed := true
			if len(where.Expressions) > 0 {
				for _, exp := range where.Expressions {
					passed, err = eval(exp, rel)
					if err != nil {
						return stacktrace.Propagate(err, "")
					}
					if !passed {
						break
					}
				}
			}
			if passed {
				rels = append(rels, rel)
			}
		}
		return nil
	}); err != nil {
		return "", nil, stacktrace.Propagate(err, "")
	}
	if len(rels) > 0 {
		if where.OrderBy == nil {
			sort.Slice(rels, func(i, j int) bool {
				return rels[i].ID() < rels[j].ID()
			})
		} else {
			sort.Slice(rels, func(i, j int) bool {
				switch where.OrderBy.Field {
				case "type":
					if where != nil && where.OrderBy.Reverse != nil && *where.OrderBy.Reverse {
						return rels[i].Type() > rels[j].Type()
					}
					return rels[i].Type() < rels[j].Type()
				default:
					iprops, err := rels[i].Properties()
					if err != nil {
						return true
					}
					if iprops[where.OrderBy.Field] == nil {
						return true
					}
					jprops, err := rels[j].Properties()
					if err != nil {
						return false
					}
					if jprops[where.OrderBy.Field] == nil {
						return true
					}
					if where.OrderBy.Reverse != nil && *where.OrderBy.Reverse {
						if val, err := cast.ToFloat64E(iprops[where.OrderBy.Field]); err == nil {
							return val > cast.ToFloat64(jprops[where.OrderBy.Field])
						}
						return cast.ToString(iprops[where.OrderBy.Field]) < cast.ToString(jprops[where.OrderBy.Field])
					}
					if val, err := cast.ToFloat64E(iprops[where.OrderBy.Field]); err == nil {
						return val < cast.ToFloat64(jprops[where.OrderBy.Field])
					}
					return cast.ToString(iprops[where.OrderBy.Field]) < cast.ToString(jprops[where.OrderBy.Field])
				}
			})
		}
	}
	return createCursor(skipped + skip), rels, nil
}
