package persistence

import (
	"github.com/autom8ter/morpheus/pkg/api"
	"github.com/autom8ter/morpheus/pkg/constants"
	"github.com/autom8ter/morpheus/pkg/encode"
	"github.com/autom8ter/morpheus/pkg/graph/model"
	"github.com/dgraph-io/badger/v3"
	"github.com/palantir/stacktrace"
	"sort"
	"strings"
	"sync"
)

type DB struct {
	dir                  string
	db                   *badger.DB
	nodeTypes            sync.Map
	nodeFieldMap         sync.Map
	relationshipTypes    sync.Map
	relationshipFieldMap sync.Map
}

func New(dir string) (api.Graph, error) {
	db, err := badger.Open(badger.DefaultOptions(dir).WithLogger(bLogger{}))
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to create database storage")
	}
	return &DB{
		dir:                  dir,
		db:                   db,
		nodeTypes:            sync.Map{},
		nodeFieldMap:         sync.Map{},
		relationshipTypes:    sync.Map{},
		relationshipFieldMap: sync.Map{},
	}, nil
}

func (d *DB) GetNode(nodeType, nodeID string) (api.Node, error) {
	n := &Node{
		nodeType: nodeType,
		nodeID:   nodeID,
		db:       d,
	}
	key := getNodePath(nodeType, nodeID)
	if err := d.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return stacktrace.Propagate(err, "")
		}
		if err := item.Value(func(val []byte) error {
			n.item = val
			return nil
		}); err != nil {
			return stacktrace.Propagate(err, "")
		}
		return nil

	}); err != nil {
		return nil, err
	}
	if n.item == nil {
		return nil, stacktrace.Propagate(constants.ErrNotFound, "")
	}
	return n, nil
}

func (d *DB) AddNode(nodeType, nodeID string, properties map[string]interface{}) (api.Node, error) {
	if properties == nil {
		properties = map[string]interface{}{}
	}
	d.nodeTypes.Store(nodeType, struct{}{})
	key := getNodePath(nodeType, nodeID)
	properties[ID] = nodeID
	properties[Type] = nodeType
	bits, err := encode.Marshal(properties)
	if err != nil {
		return nil, stacktrace.Propagate(err, "")
	}

	if err := d.db.Update(func(txn *badger.Txn) error {
		if err := txn.Set(key, bits); err != nil {
			return stacktrace.Propagate(err, "")
		}
		for k, v := range properties {
			d.nodeTypes.Store(strings.Join([]string{nodeType, k}, ","), struct{}{})
			key := getNodeTypeFieldPath(nodeType, k, v, nodeID)
			if err := txn.Set(key, bits); err != nil {
				return stacktrace.Propagate(err, "")
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return &Node{
		nodeType: nodeType,
		nodeID:   nodeID,
		db:       d,
	}, nil
}

func (d *DB) DelNode(nodeType, nodeID string) error {
	key := getNodePath(nodeType, nodeID)
	if err := d.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	}); err != nil {
		return stacktrace.Propagate(err, "")
	}
	return nil
}

func (d *DB) RangeNodes(where *model.NodeWhere) (string, []api.Node, error) {
	if where.PageSize == nil {
		pageSize := 25
		where.PageSize = &pageSize
	}
	var (
		skipped int
		skip    int
		nodes   []api.Node
		err     error
	)
	key := getNodePath(where.Type, "")
	if where.Cursor != nil {
		skip, err = parseCursor(*where.Cursor)
		if err != nil {
			return "", nil, stacktrace.Propagate(err, "")
		}
	}

	if err := d.db.View(func(txn *badger.Txn) error {
		opt := badger.DefaultIteratorOptions
		opt.PrefetchSize = prefetchSize
		it := txn.NewIterator(opt)
		defer it.Close()
		for it.Seek(key); it.ValidForPrefix(key); it.Next() {
			if len(nodes) >= *where.PageSize {
				return nil
			}
			if skipped <= skip {
				skipped++
				continue
			}
			item := it.Item()
			split := strings.Split(string(item.Key()), ",")
			node := &Node{
				nodeType: where.Type,
				nodeID:   split[len(split)-1],
				db:       d,
			}
			if err := item.Value(func(val []byte) error {
				node.item = val
				return nil
			}); err != nil {
				return stacktrace.Propagate(err, "")
			}
			if node.item != nil {
				passed := true
				for _, exp := range where.Expressions {
					passed, err = eval(exp, node)
					if err != nil {
						return stacktrace.Propagate(err, "")
					}
					if !passed {
						break
					}
				}
				if passed {
					nodes = append(nodes, node)
				}
			}
		}
		return nil
	}); err != nil {
		return "", nil, err
	}
	return createCursor(skip + skipped), nodes, nil
}

func (d *DB) NodeTypes() []string {
	var types []string
	d.nodeTypes.Range(func(key, value interface{}) bool {
		types = append(types, key.(string))
		return true
	})
	sort.Strings(types)
	return types
}

func (d *DB) GetRelationship(relation string, id string) (api.Relationship, error) {
	r := &Relationship{
		relationshipType: relation,
		relationshipID:   id,
		item:             nil,
		db:               d,
	}
	key := getRelationshipPath(relation, id)
	if err := d.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return stacktrace.Propagate(err, "")
		}
		if err := item.Value(func(val []byte) error {
			r.item = val
			return nil
		}); err != nil {
			return stacktrace.Propagate(err, "")
		}
		return nil

	}); err != nil {
		return nil, err
	}
	if r.item == nil {
		return nil, stacktrace.Propagate(constants.ErrNotFound, "")
	}
	return r, nil
}

func (d *DB) RangeRelationships(where *model.RelationWhere) (string, []api.Relationship, error) {
	if where.PageSize == nil {
		pageSize := prefetchSize
		where.PageSize = &pageSize
	}
	var (
		err     error
		skipped int
		skip    int
		rels    []api.Relationship
	)

	skip, err = parseCursor(*where.Cursor)
	if err != nil {
		return "", nil, stacktrace.Propagate(err, "")
	}

	if err := d.db.View(func(txn *badger.Txn) error {

		key := getRelationshipPath(where.Relation, "")

		opt := badger.DefaultIteratorOptions
		opt.PrefetchSize = prefetchSize
		it := txn.NewIterator(opt)

		defer it.Close()
		for it.Seek(key); it.ValidForPrefix(key); it.Next() {
			if len(rels) >= *where.PageSize {
				return nil
			}
			if skipped <= skip {
				skipped++
				continue
			}
			item := it.Item()
			split := strings.Split(string(item.Key()), ",")
			rel := &Relationship{
				relationshipType: where.Relation,
				relationshipID:   split[len(split)-1],
				item:             nil,
				db:               d,
			}
			if err := item.Value(func(val []byte) error {
				rel.item = val
				return nil
			}); err != nil {
				return stacktrace.Propagate(err, "")
			}
			if rel.item != nil {
				passed := true
				for _, exp := range where.Expressions {
					passed, err = eval(exp, rel)
					if err != nil {
						return stacktrace.Propagate(err, "")
					}
					if !passed {
						break
					}
				}
				if passed {
					rels = append(rels, rel)
				}
			}
		}
		return nil
	}); err != nil {
		return "", rels, err
	}
	return createCursor(skip + skipped), rels, nil
}

func (d *DB) RelationshipTypes() []string {
	var types []string
	d.relationshipTypes.Range(func(key, value interface{}) bool {
		types = append(types, key.(string))
		return true
	})
	sort.Strings(types)
	return types
}

func (d *DB) Close() error {
	return d.db.Close()
}
