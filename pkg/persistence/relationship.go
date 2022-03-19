package persistence

import (
	"github.com/autom8ter/morpheus/pkg/api"
	"github.com/autom8ter/morpheus/pkg/constants"
	"github.com/autom8ter/morpheus/pkg/encode"
	"github.com/dgraph-io/badger/v3"
	"github.com/palantir/stacktrace"
	"github.com/spf13/cast"
)

type Relation struct {
	relationType string
	relationID   string
	item         map[string]interface{}
	db           *DB
}

func (n *Relation) Type() string {
	return n.relationType
}

func (n *Relation) ID() string {
	return n.relationID
}

func (n *Relation) Properties() (map[string]interface{}, error) {
	if len(n.item) > 0 {
		return n.item, nil
	}
	data := map[string]interface{}{}
	if err := n.db.db.View(func(txn *badger.Txn) error {
		var key = getRelationPath(n.relationType, n.relationID)
		item, err := txn.Get(key)
		if err != nil {
			return stacktrace.Propagate(err, "failed to get relation properties")
		}
		if err := item.Value(func(val []byte) error {
			return encode.Unmarshal(val, &data)
		}); err != nil {
			return stacktrace.Propagate(err, "failed to get relation properties")
		}
		return nil
	}); err != nil {
		return nil, stacktrace.Propagate(err, "failed to get relation properties")
	}
	n.item = data
	return data, nil
}

func (n *Relation) GetProperty(name string) (interface{}, error) {
	all, err := n.Properties()
	if err != nil {
		return nil, stacktrace.Propagate(err, "")
	}
	if all == nil {
		return nil, stacktrace.Propagate(constants.ErrNotFound, "")
	}
	return all[name], nil
}

func (n *Relation) SetProperties(properties map[string]interface{}) error {

	bits, err := encode.Marshal(properties)
	if err != nil {
		return stacktrace.Propagate(err, "")
	}
	var key = getRelationPath(n.relationType, n.relationID)
	if err := n.db.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, bits)
	}); err != nil {
		return stacktrace.Propagate(err, "")
	}
	n.item = properties
	return nil
}

func (n *Relation) DelProperty(name string) error {
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

func (r Relation) Direction() (api.Direction, error) {
	all, err := r.Properties()
	if err != nil {
		return api.Direction(""), stacktrace.Propagate(err, "")
	}
	if err := r.SetProperties(all); err != nil {
		return api.Direction(""), stacktrace.Propagate(err, "")
	}
	return api.Direction(cast.ToString(all[Internal_Direction])), nil
}

func (r Relation) Relation() (string, error) {
	all, err := r.Properties()
	if err != nil {
		return "", stacktrace.Propagate(err, "")
	}
	return cast.ToString(all[Internal_Relation]), nil
}

func (r Relation) Source() (api.Node, error) {
	all, err := r.Properties()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to load source node")
	}
	n, err := r.db.GetNode(cast.ToString(all[Internal_SourceType]), cast.ToString(all[Internal_SourceID]))
	if err != nil {
		return nil, stacktrace.Propagate(err, "")
	}
	return n, nil
}

func (r Relation) Target() (api.Node, error) {
	all, err := r.Properties()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to load target node")
	}
	n, err := r.db.GetNode(cast.ToString(all[Internal_TargetType]), cast.ToString(all[Internal_TargetID]))
	if err != nil {
		return nil, stacktrace.Propagate(err, "")
	}
	return n, nil
}
