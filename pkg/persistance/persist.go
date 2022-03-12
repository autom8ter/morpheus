package persistance

import (
	"fmt"
	"github.com/autom8ter/morpheus/pkg/api"
	"github.com/autom8ter/morpheus/pkg/encode"
	"github.com/dgraph-io/badger/v3"
	lru "github.com/hashicorp/golang-lru"
	"github.com/palantir/stacktrace"
)

type cachedItem struct {
	properties map[string]interface{}
	commited   bool
}

func NewPersistantGraph(dir string, cacheSize int) (api.Graph, error) {
	db, err := badger.Open(badger.DefaultOptions(dir))
	if err != nil {
		return nil, err
	}
	cache, err := lru.NewWithEvict(cacheSize, func(key interface{}, value interface{}) {
		item := value.(*cachedItem)
		if item.commited {
			return
		}
		if err := db.Update(func(txn *badger.Txn) error {
			bits, _ := encode.Marshal(item)
			if err := txn.Set([]byte(key.(string)), bits); err != nil {
				return stacktrace.Propagate(err, "failed to set key: %s", key)
			}
			return nil
		}); err != nil {
			panic(err)
		}
	})
	if err != nil {
		return nil, err
	}
	return api.NewGraph(func(prefix, nodeType, nodeID string, properties map[string]interface{}) api.Entity {
		compoundKey := getKey(prefix, nodeType, nodeID)
		if properties == nil {
			properties = map[string]interface{}{}
		}
		e := api.NewEntity(
			nodeType,
			nodeID,
			get(db, cache, compoundKey),
			set(cache, compoundKey),
		)
		e.SetProperties(properties)
		return e
	}, func() error {
		cache.Purge()
		return db.Close()
	}), nil
}

func set(cache *lru.Cache, compoundKey string) func(props map[string]interface{}) {
	return func(props map[string]interface{}) {
		cache.Add(compoundKey, &cachedItem{
			properties: props,
			commited:   false,
		})
	}
}

func get(db *badger.DB, cache *lru.Cache, compoundKey string) func() map[string]interface{} {
	return func() map[string]interface{} {
		if val, ok := cache.Get(compoundKey); ok {
			return val.(*cachedItem).properties
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
				return err
			}
			cache.ContainsOrAdd(compoundKey, &cachedItem{
				properties: data,
				commited:   true,
			})
			return nil
		}); err != nil {
			panic(err)
		}
		return data
	}
}

func getKey(prefix, nodeType, nodeID string) string {
	return fmt.Sprintf("%s_%s_%s", prefix, nodeType, nodeID)
}
