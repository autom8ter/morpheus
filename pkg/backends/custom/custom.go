package custom

import (
	"context"
	"fmt"
	"github.com/autom8ter/morpheus/pkg/api"
	"github.com/autom8ter/morpheus/pkg/storage"
)

func NewGraph(dir string) api.Graph {
	s := storage.NewStorage(context.Background(), dir, 1024, false)
	return api.NewGraph(func(prefix, nodeType, nodeID string, properties map[string]interface{}) api.Entity {
		setProperties := func(props map[string]interface{}) {
			if props == nil {
				props = map[string]interface{}{}
			}
			bucket := s.GetBucket(fmt.Sprintf("%s_%s", prefix, nodeType))
			if err := bucket.Set(nodeID, properties); err != nil {
				panic(err)
			}
		}
		setProperties(properties)
		return api.NewEntity(
			nodeType,
			nodeID,
			func() map[string]interface{} {
				bucket := s.GetBucket(fmt.Sprintf("%s_%s", prefix, nodeType))
				val, err := bucket.Get(nodeID)
				if err != nil {
					panic(err)
				}
				return val
			},
			setProperties)
	}, func() error {
		s.Close()
		return nil
	})
}
