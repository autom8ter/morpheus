package persistence

import (
	"fmt"
	"github.com/autom8ter/morpheus/pkg/api"
	"github.com/autom8ter/morpheus/pkg/constants"
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

func (n *Relationship) Type() string {
	return n.relationshipType
}

func (n *Relationship) ID() string {
	return n.relationshipID
}

func (n *Relationship) Properties() (map[string]interface{}, error) {
	data := map[string]interface{}{}
	//if len(n.item) > 0 {
	//	if err := encode.Unmarshal(n.item, &data); err != nil {
	//		return nil, stacktrace.Propagate(err, "failed to get relationship properties")
	//	}
	//	n.item = nil
	//	return data, nil
	//}
	if err := n.db.db.View(func(txn *badger.Txn) error {
		var key = getRelationshipPath(n.relationshipType, n.relationshipID)
		item, err := txn.Get(key)
		if err != nil {
			return stacktrace.Propagate(err, "failed to get relationship properties")
		}
		if err := item.Value(func(val []byte) error {
			n.item = val
			return encode.Unmarshal(val, &data)
		}); err != nil {
			return stacktrace.Propagate(err, "failed to get relationship properties")
		}
		return nil
	}); err != nil {
		return nil, stacktrace.Propagate(err, "failed to get relationship properties")
	}
	return data, nil
}

func (n *Relationship) GetProperty(name string) (interface{}, error) {
	all, err := n.Properties()
	if err != nil {
		return nil, stacktrace.Propagate(err, "")
	}
	if all == nil {
		return nil, stacktrace.Propagate(constants.ErrNotFound, "")
	}
	return all[name], nil
}

func (n *Relationship) SetProperties(properties map[string]interface{}) error {
	bits, err := encode.Marshal(properties)
	if err != nil {
		return stacktrace.Propagate(err, "")
	}
	var key = getRelationshipPath(n.relationshipType, n.relationshipID)
	if err := n.db.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, bits)
	}); err != nil {
		return stacktrace.Propagate(err, "")
	}
	return nil
}

func (n *Relationship) DelProperty(name string) error {
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

func (r Relationship) Direction() (api.Direction, error) {
	all, err := r.Properties()
	if err != nil {
		return api.Direction(""), stacktrace.Propagate(err, "")
	}
	if err := r.SetProperties(all); err != nil {
		return api.Direction(""), stacktrace.Propagate(err, "")
	}
	return api.Direction(cast.ToString(all[Direction])), nil
}

func (r Relationship) Relation() (string, error) {
	all, err := r.Properties()
	if err != nil {
		return "", stacktrace.Propagate(err, "")
	}
	return cast.ToString(all[Relation]), nil
}

func (r Relationship) Source() (api.Node, error) {
	all, err := r.Properties()
	if err != nil {
		return nil, stacktrace.Propagate(err, fmt.Sprintf("%s %s", all[SourceType], all[SourceID]))
	}
	return &Node{
		nodeType: cast.ToString(all[SourceType]),
		nodeID:   cast.ToString(all[SourceID]),
		db:       r.db,
	}, nil
}

func (r Relationship) Target() (api.Node, error) {
	all, err := r.Properties()
	if err != nil {
		return nil, stacktrace.Propagate(err, fmt.Sprintf("%s %s", r.Type(), r.ID()))
	}
	return &Node{
		nodeType: cast.ToString(all[TargetType]),
		nodeID:   cast.ToString(all[TargetID]),
		db:       r.db,
	}, nil
}
