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

func (n Node) AddRelationship(relationship string, node api.Node) (api.Relationship, error) {
	relID := getRelationID(n.Type(), n.ID(), relationship, node.Type(), node.ID())
	n.db.relationshipTypes.Store(relationship, struct{}{})
	rkey := getRelationshipPath(relationship, relID)
	source := getNodeRelationshipPath(n.Type(), n.ID(), api.Outgoing, relationship, node.Type(), node.ID(), relID)
	target := getNodeRelationshipPath(node.Type(), node.ID(), api.Incoming, relationship, n.Type(), n.ID(), relID)

	values := map[string]interface{}{
		Direction:  api.Outgoing,
		SourceType: n.Type(),
		SourceID:   n.ID(),
		TargetID:   node.ID(),
		TargetType: node.Type(),
		ID:         relID,
		Relation:   relationship,
		Type:       relationship,
	}

	bits, err := encode.Marshal(&values)
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
		for k, v := range values {
			n.db.relationshipFieldMap.Store(strings.Join([]string{relationship, k}, ","), struct{}{})
			key := getRelationshipFieldPath(relationship, k, v, relID)
			if err := txn.Set(key, bits); err != nil {
				return stacktrace.Propagate(err, "")
			}
		}
		return nil
	}); err != nil {
		return nil, stacktrace.Propagate(err, "")
	}
	r := &Relationship{
		relationshipType: relationship,
		relationshipID:   relID,
		item:             values,
		db:               n.db,
	}
	n.db.cache.Set(string(rkey), r, 1)
	return r, nil
}

func (n Node) DelRelationship(relationship string, id string) error {
	rel, ok, err := n.GetRelationship(relationship, id)
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

	rkey := getRelationshipPath(rel.Type(), rel.ID())
	source := getNodeRelationshipPath(n.Type(), n.ID(), api.Outgoing, relationship, targetNode.Type(), targetNode.ID(), rel.ID())
	target := getNodeRelationshipPath(targetNode.Type(), targetNode.ID(), api.Incoming, relationship, n.Type(), n.ID(), rel.ID())
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
				key := getRelationshipFieldPath(rel.Type(), k, v, rel.ID())
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

func (n Node) GetRelationship(relation, id string) (api.Relationship, bool, error) {
	rkey := getRelationshipPath(relation, id)
	if val, ok := n.db.cache.Get(string(rkey)); ok {
		return val.(api.Relationship), true, nil
	}

	rel := &Relationship{
		relationshipType: relation,
		relationshipID:   id,
		item:             map[string]interface{}{},
		db:               n.db,
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

func (n Node) Relationships(where *model.RelationWhere) (string, []api.Relationship, error) {
	source := getNodeRelationshipPath(n.Type(), n.ID(), api.Direction(where.Direction), where.Relation, where.TargetType, "", "")
	// Iterate over 1000 items
	var (
		skipped int
		skip    int
		rels    []api.Relationship
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
			var rel api.Relationship
			cached, ok := n.db.cache.Get(string(getRelationshipPath(where.Relation, split[len(split)-1])))
			if ok {
				rel = cached.(api.Relationship)
			} else {
				rel = &Relationship{
					relationshipType: where.Relation,
					relationshipID:   split[len(split)-1],
					item:             nil,
					db:               n.db,
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
