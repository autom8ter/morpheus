package badger

import (
	"fmt"
	"github.com/autom8ter/morpheus/pkg/api"
	"github.com/autom8ter/morpheus/pkg/encode"
	"github.com/dgraph-io/badger/v3"
	lru "github.com/hashicorp/golang-lru"
)

func NewGraph(dir string, cacheSize int) (api.Graph, error) {
	db, err := badger.Open(badger.DefaultOptions(dir))
	if err != nil {
		return nil, err
	}
	cache, err := lru.NewWithEvict(cacheSize, func(key interface{}, value interface{}) {
		if err := db.Update(func(txn *badger.Txn) error {
			bits, _ := encode.Marshal(value)
			if err := txn.Set([]byte(key.(string)), bits); err != nil {
				return err
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
		if properties == nil {
			properties = map[string]interface{}{}
		}
		cache.Add(getKey(prefix, nodeType, nodeID), properties)
		return api.NewEntity(
			nodeType,
			nodeID,
			func() map[string]interface{} {
				if val, ok := cache.Get(getKey(prefix, nodeType, nodeID)); ok {
					return val.(map[string]interface{})
				}

				data := map[string]interface{}{}
				if err := db.View(func(txn *badger.Txn) error {
					val, err := txn.Get([]byte(getKey(prefix, nodeType, nodeID)))
					if err != nil {
						return err
					}
					if err := val.Value(func(val []byte) error {
						return encode.Unmarshal(val, &data)
					}); err != nil {
						return err
					}
					cache.Add(getKey(prefix, nodeType, nodeID), data)
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
				cache.Add(getKey(prefix, nodeType, nodeID), properties)
			})
	}, func() error {
		return db.Close()
	}), nil
}

func getKey(prefix, nodeType, nodeID string) string {
	return fmt.Sprintf("%s_%s_%s", prefix, nodeType, nodeID)
}
