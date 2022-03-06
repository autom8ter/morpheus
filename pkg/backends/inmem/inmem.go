package inmem

import (
	"github.com/autom8ter/morpheus/pkg/api"
)

func NewGraph() api.Graph {
	return api.NewGraph(func(prefix, nodeType, nodeID string, properties map[string]interface{}) api.Entity {
		return api.NewEntity(nodeType, nodeID,
			func() map[string]interface{} {
				return properties
			}, func(properties map[string]interface{}) {
				for k, v := range properties {
					properties[k] = v
				}
			})
	}, func() error {
		return nil
	})
}
