package persistence

import (
	"github.com/autom8ter/morpheus/pkg/api"
	"github.com/autom8ter/morpheus/pkg/encode"
	"github.com/dgraph-io/badger/v3"
	"github.com/palantir/stacktrace"
	"github.com/spf13/cast"
)

type Relationship struct {
	relationshipType string
	relationshipID   string
	item             []byte
	db               *DB
}

func (n Relationship) Type() string {
	return n.relationshipType
}

func (n Relationship) ID() string {
	return n.relationshipID
}

func (n Relationship) Properties() map[string]interface{} {
	data := map[string]interface{}{}
	if n.item != nil {
		if err := encode.Unmarshal(n.item, &data); err != nil {
			panic(stacktrace.Propagate(err, ""))
		}
		n.item = nil
		return data
	}
	if err := n.db.db.View(func(txn *badger.Txn) error {
		var key = getRelationshipPath(n.relationshipType, n.relationshipID)
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

func (n Relationship) GetProperty(name string) interface{} {
	all := n.Properties()
	if all == nil {
		return nil
	}
	return all[name]
}

func (n Relationship) SetProperties(properties map[string]interface{}) {
	bits, err := encode.Marshal(properties)
	if err != nil {
		panic(stacktrace.Propagate(err, ""))
	}
	var key = getRelationshipPath(n.relationshipType, n.relationshipID)
	if err := n.db.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, bits)
	}); err != nil {
		panic(stacktrace.Propagate(err, ""))
	}
}

func (n Relationship) DelProperty(name string) {
	all := n.Properties()
	delete(all, name)
	n.SetProperties(all)
}

func (r Relationship) Direction() api.Direction {
	all := r.Properties()
	return api.Direction(cast.ToString(all[Direction]))
}

func (r Relationship) Relation() string {
	all := r.Properties()
	return cast.ToString(all[Relation])
}

func (r Relationship) Source() api.Node {
	all := r.Properties()
	return &Node{
		nodeType: cast.ToString(all[SourceType]),
		nodeID:   cast.ToString(all[SourceID]),
		db:       r.db,
	}
}

func (r Relationship) Target() api.Node {
	all := r.Properties()
	return &Node{
		nodeType: cast.ToString(all[TargetType]),
		nodeID:   cast.ToString(all[TargetID]),
		db:       r.db,
	}
}
