package persistence

import (
	"fmt"
	"github.com/autom8ter/morpheus/pkg/api"
	"github.com/autom8ter/morpheus/pkg/encode"
	"github.com/autom8ter/morpheus/pkg/graph/model"
	"github.com/dgraph-io/badger/v3"
	"github.com/palantir/stacktrace"
	"strings"
)

func (d *DB) rangeEQNodes(where *model.NodeWhere) (string, []api.Node, error) {
	if where.PageSize == nil {
		pageSize := 25
		where.PageSize = &pageSize
	}
	var (
		skipped int
		skip    int
		err     error
	)

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
	for _, exp := range where.Expressions {
		key := getNodeTypeFieldPath(where.Type, exp.Key, exp.Value, "")
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
				split := strings.Split(string(item.Key()), ",")
				nodeID := split[len(split)-1]
				val, ok := d.cache.Get(string(getNodePath(where.Type, nodeID)))
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
	}

	return createCursor(skip + skipped), nodes, nil
}

func (d *DB) rangeContainsNodes(where *model.NodeWhere) (string, []api.Node, error) {
	if where.PageSize == nil {
		pageSize := 25
		where.PageSize = &pageSize
	}
	var (
		skipped int
		skip    int
		err     error
	)

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
	for _, exp := range where.Expressions {
		key := getNodeTypeFieldPath(where.Type, exp.Key, "", "")
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
				split := strings.Split(string(item.Key()), ",")
				nodeID := split[len(split)-1]
				val, ok := d.cache.Get(string(getNodePath(where.Type, nodeID)))
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
	}

	return createCursor(skip + skipped), nodes, nil
}

func (d *DB) rangeHasPrefixNodes(where *model.NodeWhere) (string, []api.Node, error) {
	if where.PageSize == nil {
		pageSize := 25
		where.PageSize = &pageSize
	}
	var (
		skipped int
		skip    int
		err     error
	)

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
	for _, exp := range where.Expressions {
		key := getNodeTypeFieldPath(where.Type, exp.Key, fmt.Sprint(exp.Value), "")
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
				split := strings.Split(string(item.Key()), ",")
				nodeID := split[len(split)-1]
				val, ok := d.cache.Get(string(getNodePath(where.Type, nodeID)))
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
	}

	return createCursor(skip + skipped), nodes, nil
}
