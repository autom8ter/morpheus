package persistence

import (
	"github.com/autom8ter/morpheus/pkg/api"
	"github.com/autom8ter/morpheus/pkg/constants"
	"github.com/autom8ter/morpheus/pkg/encode"
	"github.com/dgraph-io/badger/v3"
	"github.com/palantir/stacktrace"
	"sort"
	"strings"
	"sync"
)

type DB struct {
	dir               string
	db                *badger.DB
	nodeTypes         sync.Map
	relationshipTypes sync.Map
}

func New(dir string) (*DB, error) {
	db, err := badger.Open(badger.DefaultOptions(dir).WithLogger(bLogger{}))
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to create database storage")
	}
	return &DB{
		dir:               dir,
		db:                db,
		nodeTypes:         sync.Map{},
		relationshipTypes: sync.Map{},
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
	d.nodeTypes.Store(nodeType, struct{}{})
	key := getNodePath(nodeType, nodeID)
	bits, err := encode.Marshal(properties)
	if err != nil {
		return nil, stacktrace.Propagate(err, "")
	}
	if err := d.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, bits)
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

func (d *DB) RangeNodes(skip int, nodeType string, fn func(node api.Node) bool) error {
	var skipped int
	key := getNodePath(nodeType, "")
	if err := d.db.View(func(txn *badger.Txn) error {
		opt := badger.DefaultIteratorOptions
		opt.PrefetchSize = 10
		it := txn.NewIterator(opt)
		defer it.Close()
		for it.Rewind(); it.ValidForPrefix(key); it.Next() {
			if skipped < skip {
				skip++
				continue
			}
			item := it.Item()
			key := item.KeyCopy(nil)
			split := strings.Split(string(key), ",")
			if !fn(&Node{
				nodeType: nodeType,
				nodeID:   split[len(split)-1],
				db:       d,
			}) {
				break
			}
		}
		return nil
	}); err != nil {
		panic(err)
	}
	return nil
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

func (d *DB) RangeRelationships(skip int, relation string, fn func(node api.Relationship) bool) error {
	source := getRelationshipPath(relation, "")
	// Iterate over 1000 items
	var skipped int
	if err := d.db.View(func(txn *badger.Txn) error {
		opt := badger.DefaultIteratorOptions
		opt.PrefetchSize = 10
		it := txn.NewIterator(opt)
		defer it.Close()
		for it.Rewind(); it.ValidForPrefix(source); it.Next() {
			if skipped < skip {
				skip++
				continue
			}
			item := it.Item()
			split := strings.Split(string(item.Key()), ",")
			if !fn(&Relationship{
				relationshipType: relation,
				relationshipID:   split[len(split)-1],
				db:               d,
			}) {
				break
			}
		}
		return nil
	}); err != nil {
		panic(err)
	}
	return nil
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

func (d *DB) Size() int {
	count := 0
	for _, t := range d.NodeTypes() {
		if err := d.RangeNodes(0, t, func(node api.Node) bool {
			count++
			return true
		}); err != nil {
			panic(stacktrace.Propagate(err, ""))
		}
	}
	return count
}

func (d *DB) Close() error {
	return d.db.Close()
}
