package persistence

import (
	"github.com/autom8ter/morpheus/pkg/api"
	"github.com/autom8ter/morpheus/pkg/constants"
	"github.com/autom8ter/morpheus/pkg/encode"
	"github.com/autom8ter/morpheus/pkg/graph/model"
	"github.com/autom8ter/morpheus/pkg/logger"
	"github.com/dgraph-io/badger/v3"
	"github.com/dgraph-io/ristretto"
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
	cache                *ristretto.Cache
}

func New(dir string) (api.Graph, error) {
	db, err := badger.Open(badger.DefaultOptions(dir).WithLogger(logger.BadgerLogger()))
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to create database storage")
	}
	d := &DB{
		dir:                  dir,
		db:                   db,
		nodeTypes:            sync.Map{},
		nodeFieldMap:         sync.Map{},
		relationshipTypes:    sync.Map{},
		relationshipFieldMap: sync.Map{},
	}
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,     // number of keys to track frequency of (10M).
		MaxCost:     1 << 30, // maximum cost of cache (1GB).
		BufferItems: 64,      // number of keys per Get buffer.
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to create database cache")
	}
	d.cache = cache

	return d, nil
}

func (d *DB) GetNode(nodeType, nodeID string) (api.Node, error) {
	if nodeType == "" {
		return nil, stacktrace.NewError("empty node type")
	}
	if nodeID == "" {
		return nil, stacktrace.NewError("empty node id")
	}
	val, ok := d.cache.Get(string(getNodePath(nodeType, nodeID)))
	if ok {
		return val.(api.Node), nil
	}
	n := &Node{
		nodeType: nodeType,
		nodeID:   nodeID,
		db:       d,
	}

	if err := d.db.View(func(txn *badger.Txn) error {
		data := map[string]interface{}{}
		key := getNodePath(nodeType, nodeID)
		item, err := txn.Get(key)
		if err != nil {
			return stacktrace.Propagate(err, "key=%s", string(key))
		}
		if err := item.Value(func(val []byte) error {
			return encode.Unmarshal(val, &data)
		}); err != nil {
			return stacktrace.Propagate(err, "key=%s", string(key))
		}
		n.data = data
		return nil

	}); err != nil {
		return nil, stacktrace.Propagate(err, "")
	}

	if len(n.data) == 0 {
		return nil, stacktrace.Propagate(constants.ErrNotFound, "")
	}
	d.cache.Set(string(getNodePath(nodeType, nodeID)), n, 1)
	return n, nil
}

func (d *DB) AddNode(nodeType, nodeID string, properties map[string]interface{}) (api.Node, error) {
	if properties == nil {
		properties = map[string]interface{}{}
	}
	existing, _ := d.GetNode(nodeType, nodeID)
	if existing != nil && existing.ID() != "" {
		if err := d.db.Update(func(txn *badger.Txn) error {
			for k, v := range properties {
				key := getNodeTypeFieldPath(nodeType, k, v, nodeID)
				if err := txn.Delete(key); err != nil {
					return stacktrace.Propagate(err, "")
				}
			}
			return nil
		}); err != nil {
			logger.L.Error("failed to delete existing value index", stacktrace.Propagate(err, ""), map[string]interface{}{})
		}
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
		return nil, stacktrace.Propagate(err, "")
	}
	n := &Node{
		nodeType: nodeType,
		nodeID:   nodeID,
		data:     properties,
		db:       d,
	}
	d.cache.Set(string(key), n, 1)
	return n, nil
}

func (d *DB) DelNode(nodeType, nodeID string) error {
	key := getNodePath(nodeType, nodeID)
	if err := d.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	}); err != nil {
		return stacktrace.Propagate(err, "")
	}
	d.cache.Del(key)
	return nil
}

func (d *DB) RangeNodes(where *model.NodeWhere) (string, []api.Node, error) {
	if where.PageSize == nil {
		pageSize := 25
		where.PageSize = &pageSize
	}
	if len(where.Expressions) > 0 {
		switch where.Expressions[0].Operator {
		case model.OperatorEq:
			return d.rangeEQNodes(where)
		case model.OperatorHasPrefix:
			return d.rangeHasPrefixNodes(where)
		}
	}

	var (
		skipped int
		skip    int
		err     error
	)
	key := getNodePath(where.Type, "")
	if where.Cursor != nil {
		skip, err = parseCursor(*where.Cursor)
		if err != nil {
			return "", nil, stacktrace.Propagate(err, "")
		}
	}
	evalFunc := func(n api.Node) (bool, error) {
		passed := true
		for _, exp := range where.Expressions {
			passed, err = eval(exp, n)
			if err != nil {
				return false, stacktrace.Propagate(err, "")
			}
			if !passed {
				break
			}
		}
		return passed, nil
	}
	var nodes []api.Node
	if err := d.db.View(func(txn *badger.Txn) error {
		opt := badger.DefaultIteratorOptions
		opt.PrefetchSize = prefetchSize
		it := txn.NewIterator(opt)
		defer it.Close()
		for it.Seek(key); it.ValidForPrefix(key); it.Next() {
			if len(nodes) >= *where.PageSize {
				return nil
			}
			if skipped < skip {
				skipped++
				continue
			}
			item := it.Item()
			val, ok := d.cache.Get(string(item.Key()))
			if ok {
				n := val.(api.Node)
				passed, err := evalFunc(n)
				if err != nil {
					return stacktrace.Propagate(err, "")
				}
				if passed {
					nodes = append(nodes, n)
				}
				continue
			}
			split := strings.Split(string(item.Key()), ",")
			nodeID := split[len(split)-1]
			n := &Node{
				nodeType: where.Type,
				nodeID:   nodeID,
				db:       d,
			}
			data := map[string]interface{}{}
			if err := item.Value(func(val []byte) error {
				return encode.Unmarshal(val, &data)
			}); err != nil {
				return stacktrace.Propagate(err, "")
			}
			n.data = data
			passed, err := evalFunc(n)
			if err != nil {
				return err
			}
			if passed {
				nodes = append(nodes, n)
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
		db:               d,
	}
	data := map[string]interface{}{}
	key := getRelationshipPath(relation, id)
	if err := d.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return stacktrace.Propagate(err, "")
		}
		if err := item.Value(func(val []byte) error {
			return encode.Unmarshal(val, &data)
		}); err != nil {
			return stacktrace.Propagate(err, "")
		}
		return nil

	}); err != nil {
		return nil, err
	}
	r.item = data
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
				db:               d,
			}
			data := map[string]interface{}{}
			if err := item.Value(func(val []byte) error {
				return encode.Unmarshal(val, &data)
			}); err != nil {
				return stacktrace.Propagate(err, "")
			}
			rel.item = data
			if len(rel.item) > 0 {
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
