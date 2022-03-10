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
		if value.(*cachedItem).commited {
			return
		}
		if err := db.Update(func(txn *badger.Txn) error {
			bits, _ := encode.Marshal(value)
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
		cache.Add(compoundKey, &cachedItem{
			properties: properties,
			commited:   false,
		})
		return api.NewEntity(
			nodeType,
			nodeID,
			func() map[string]interface{} {
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
					cache.Add(compoundKey, &cachedItem{
						properties: data,
						commited:   true,
					})
					return nil
				}); err != nil {
					panic(err)
				}
				return data
			},
			func(properties map[string]interface{}) {
				if properties == nil {
					properties = map[string]interface{}{}
				}
				cache.Add(compoundKey, properties)
			})
	}, func() error {
		cache.Purge()
		return db.Close()
	}), nil
}

func getKey(prefix, nodeType, nodeID string) string {
	return fmt.Sprintf("%s_%s_%s", prefix, nodeType, nodeID)
}
