package persistence

import (
	"github.com/autom8ter/morpheus/pkg/api"
	"github.com/autom8ter/morpheus/pkg/encode"
	"github.com/dgraph-io/badger/v3"
	"github.com/google/uuid"
	"github.com/palantir/stacktrace"
	"strings"
)

type Node struct {
	nodeType string
	nodeID   string
	item     []byte
	db       *badger.DB
}

func (n Node) Type() string {
	return n.nodeType
}

func (n Node) ID() string {
	return n.nodeID
}

func (n Node) Properties() map[string]interface{} {
	data := map[string]interface{}{}
	if n.item != nil {
		if err := encode.Unmarshal(n.item, &data); err != nil {
			panic(stacktrace.Propagate(err, ""))
		}
		n.item = nil
		return data
	}
	if err := n.db.View(func(txn *badger.Txn) error {
		var key = getNodePath(n.nodeType, n.nodeID)
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

func (n Node) GetProperty(name string) interface{} {
	all := n.Properties()
	if all == nil {
		return nil
	}
	return all[name]
}

func (n Node) SetProperties(properties map[string]interface{}) {
	bits, err := encode.Marshal(properties)
	if err != nil {
		panic(stacktrace.Propagate(err, ""))
	}
	var key = getNodePath(n.nodeType, n.nodeID)
	if err := n.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, bits)
	}); err != nil {
		panic(stacktrace.Propagate(err, ""))
	}
}

func (n Node) DelProperty(name string) {
	all := n.Properties()
	delete(all, name)
	n.SetProperties(all)
}

func (n Node) AddRelationship(relationship string, node api.Node) api.Relationship {

	rkey := getRelationshipPath(relationship, uuid.NewString())
	source := getNodeRelationshipPath(n.Type(), n.ID(), api.Outgoing, relationship, node.Type(), node.ID())
	target := getNodeRelationshipPath(node.Type(), node.ID(), api.Incoming, relationship, node.Type(), node.ID())

	values := map[string]interface{}{
		Direction: api.Outgoing,
	}
	values[Direction] = api.Outgoing
	values[SourceType] = n.Type()
	values[SourceID] = n.ID()
	values[TargetID] = node.ID()
	values[TargetType] = node.Type()
	values[ID] = getRelationID(n.Type(), n.ID(), relationship, node.Type(), node.ID())
	values[Relation] = relationship
	values[Type] = relationship

	bits, err := encode.Marshal(values)
	if err != nil {
		panic(err)
	}
	if err := n.db.Update(func(txn *badger.Txn) error {
		if err := txn.Set(rkey, bits); err != nil {
			return stacktrace.Propagate(err, "")
		}
		if err := txn.Set(source, bits); err != nil {
			return stacktrace.Propagate(err, "")
		}
		if err := txn.Set(target, bits); err != nil {
			return stacktrace.Propagate(err, "")
		}
		return nil
	}); err != nil {
		panic(err)
	}
	return nil
}

func (n Node) DelRelationship(relationship string, id string) {
	rel, ok := n.GetRelationship(relationship, id)
	if !ok {
		return
	}
	rkey := getRelationshipPath(rel.Type(), rel.ID())
	source := getNodeRelationshipPath(n.Type(), n.ID(), api.Outgoing, relationship, rel.Target().Type(), rel.Target().ID())
	target := getNodeRelationshipPath(rel.Target().Type(), rel.Target().ID(), api.Incoming, relationship, n.Type(), n.ID())

	if err := n.db.Update(func(txn *badger.Txn) error {
		if err := txn.Delete(rkey); err != nil {
			return stacktrace.Propagate(err, "")
		}
		if err := txn.Delete(source); err != nil {
			return stacktrace.Propagate(err, "")
		}
		if err := txn.Delete(target); err != nil {
			return stacktrace.Propagate(err, "")
		}
		return nil
	}); err != nil {
		panic(err)
	}
}

func (n Node) GetRelationship(relation, id string) (api.Relationship, bool) {
	rkey := getRelationshipPath(relation, id)
	rel := &Relationship{
		relationshipType: relation,
		relationshipID:   id,
		item:             nil,
		db:               n.db,
	}
	if err := n.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(rkey)
		if err != nil {
			return stacktrace.Propagate(err, "")
		}
		if err := item.Value(func(val []byte) error {
			rel.item = val
			return nil
		}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		panic(err)
	}
	if rel.item == nil {
		return nil, false
	}
	return rel, true
}

func (n Node) Relationships(skip int, relation string, targetType string, fn func(relationship api.Relationship) bool) {
	source := getNodeRelationshipPrefix(n.Type(), n.ID(), api.Outgoing, relation, targetType)
	// Iterate over 1000 items
	var skipped int
	if err := n.db.View(func(txn *badger.Txn) error {
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
			rel := &Relationship{
				relationshipType: relation,
				relationshipID:   split[len(split)-1],
				db:               n.db,
			}
			if err := item.Value(func(val []byte) error {
				rel.item = val
				return nil
			}); err != nil {
				return stacktrace.Propagate(err, "")
			}
			if rel.item != nil {
				if !fn(rel) {
					break
				}
			}
		}
		return nil
	}); err != nil {
		panic(err)
	}
}
