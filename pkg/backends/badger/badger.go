package badger

import (
	"fmt"
	"github.com/autom8ter/morpheus/pkg/api"
	"github.com/autom8ter/morpheus/pkg/encode"
	"github.com/dgraph-io/badger/v3"
)

func NewGraph(dir string) api.Graph {
	db, err := badger.Open(badger.DefaultOptions(dir))
	if err != nil {
		panic(err)
	}
	return api.NewGraph(func(prefix, nodeType, nodeID string, properties map[string]interface{}) api.Entity {
		setProperties := func(props map[string]interface{}) {
			if props == nil {
				props = map[string]interface{}{}
			}
			if err := db.Update(func(txn *badger.Txn) error {
				bits, _ := encode.Marshal(props)
				if err := txn.Set(getKey(prefix, nodeType, nodeID), bits); err != nil {
					return err
				}
				return nil
			}); err != nil {
				panic(err)
			}
		}
		setProperties(properties)
		return api.NewEntity(
			nodeType,
			nodeID,
			func() map[string]interface{} {
				data := map[string]interface{}{}
				if err := db.View(func(txn *badger.Txn) error {
					val, err := txn.Get(getKey(prefix, nodeType, nodeID))
					if err != nil {
						return err
					}
					if err := val.Value(func(val []byte) error {
						return encode.Unmarshal(val, &data)
					}); err != nil {
						return err
					}
					return nil
				}); err != nil {
					panic(err)
				}
				return data
			},
			setProperties)
	}, func() error {
		return db.Close()
	})
}

func getKey(prefix, nodeType, nodeID string) []byte {
	return []byte(fmt.Sprintf("%s_%s_%s", prefix, nodeType, nodeID))
}
