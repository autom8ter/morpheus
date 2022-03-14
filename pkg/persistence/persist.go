package persistence

import (
	"fmt"
	"github.com/autom8ter/morpheus/pkg/api"
	"github.com/autom8ter/morpheus/pkg/encode"
	"github.com/autom8ter/morpheus/pkg/logger"
	"github.com/dgraph-io/badger/v3"
	lru "github.com/hashicorp/golang-lru"
	"github.com/palantir/stacktrace"
)

func NewPersistantGraph(dir string, rcacheSize, wcacheSize int) (api.Graph, error) {
	db, err := badger.Open(badger.DefaultOptions(dir))
	if err != nil {
		return nil, stacktrace.Propagate(err, "")
	}
	rcache, err := lru.New(rcacheSize)
	if err != nil {
		return nil, stacktrace.Propagate(err, "")
	}
	return api.NewGraph(func(prefix, nodeType, nodeID string, properties map[string]interface{}) api.Entity {
		compoundKey := getKey(prefix, nodeType, nodeID)
		if properties == nil {
			properties = map[string]interface{}{}
		}
		e := api.NewEntity(
			nodeType,
			nodeID,
			get(db, rcache, compoundKey),
			set(db, rcache, compoundKey),
		)
		e.SetProperties(properties)
		return e
	}, func() error {
		return db.Close()
	}), nil
}

func set(db *badger.DB, cache *lru.Cache, compoundKey string) func(props map[string]interface{}) {
	return func(props map[string]interface{}) {
		if err := db.Update(func(txn *badger.Txn) error {
			bits, err := encode.Marshal(props)
			if err != nil {
				return stacktrace.Propagate(err, "")
			}
			if err := txn.Set([]byte(compoundKey), bits); err != nil {
				return stacktrace.Propagate(err, "failed to set key: %s", compoundKey)
			}
			return nil
		}); err != nil {
			logger.L.Error("failed to persist", map[string]interface{}{
				"error": err,
			})
		}
		if cache.Contains(compoundKey) {
			cache.Add(compoundKey, props)
		}
	}
}

func get(db *badger.DB, rcache *lru.Cache, compoundKey string) func() map[string]interface{} {
	return func() map[string]interface{} {
		if val, ok := rcache.Get(compoundKey); ok {
			return val.(map[string]interface{})
		}
		data := map[string]interface{}{}

		if err := db.View(func(txn *badger.Txn) error {
			val, err := txn.Get([]byte(compoundKey))
			if err != nil {
				return stacktrace.Propagate(err, "failed to get key")
			}
			if err := val.Value(func(val []byte) error {
				return encode.Unmarshal(val, &data)
			}); err != nil {
				return stacktrace.Propagate(err, "")
			}
			return nil
		}); err != nil {
			logger.L.Error("failed to get properties", map[string]interface{}{
				"error": err,
			})
		} else {
			rcache.Add(compoundKey, data)
			return data
		}
		return nil
	}
}

func getKey(prefix, nodeType, nodeID string) string {
	return fmt.Sprintf("%s_%s_%s", prefix, nodeType, nodeID)
}
