package persistence

import (
	"github.com/autom8ter/morpheus/pkg/encode"
	"github.com/dgraph-io/badger/v3"
	"github.com/palantir/stacktrace"
	"strings"
)

type Entity struct {
	nodePath         string
	relationshipPath string
	db               DB
}

func (n Entity) Type() string {
	if n.relationshipPath != "" {
		split := strings.Split(n.relationshipPath, ",")
		return split[1]
	}
	split := strings.Split(n.nodePath, ",")
	return split[1]
}

func (n Entity) ID() string {
	if n.relationshipPath != "" {
		split := strings.Split(n.relationshipPath, ",")
		return split[2]
	}
	split := strings.Split(n.nodePath, ",")
	return split[2]
}

func (n Entity) Properties() map[string]interface{} {
	data := map[string]interface{}{}
	if err := n.db.db.View(func(txn *badger.Txn) error {
		var key []byte
		if n.relationshipPath != "" {
			key = []byte(n.relationshipPath)
		} else {
			key = []byte(n.nodePath)
		}
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
		panic(err)
	}

	return data
}

func (n Entity) GetProperty(name string) interface{} {
	all := n.Properties()
	if all == nil {
		return nil
	}
	return all[name]
}

func (n Entity) SetProperties(properties map[string]interface{}) {
	bits, err := encode.Marshal(properties)
	if err != nil {
		panic(stacktrace.Propagate(err, ""))
	}
	var key []byte
	if n.relationshipPath != "" {
		key = []byte(n.relationshipPath)
	} else {
		key = []byte(n.nodePath)
	}
	if err := n.db.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, bits)
	}); err != nil {
		panic(stacktrace.Propagate(err, ""))
	}
}

func (n Entity) Del(name string) {
	all := n.Properties()
	delete(all, name)
	n.SetProperties(all)
}
