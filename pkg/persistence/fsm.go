package persistence

import (
	"github.com/autom8ter/morpheus/pkg/encode"
	"github.com/autom8ter/morpheus/pkg/graph/fsm"
	"github.com/autom8ter/morpheus/pkg/graph/model"
	"github.com/hashicorp/raft"
	"github.com/palantir/stacktrace"
	"io"
)

func (d *DB) FSM() raft.FSM {
	return &fsm.FSM{
		ApplyFunc: func(log *raft.Log) interface{} {
			var cmd fsm.CMD
			if err := encode.Unmarshal(log.Data, &cmd); err != nil {
				return stacktrace.Propagate(err, "")
			}
			switch cmd.Method {
			case fsm.MethodAdd:
				addNode := cmd.Value.(model.AddNode)
				n, err := d.AddNode(addNode.Type, *addNode.ID, addNode.Properties)
				if err != nil {
					return stacktrace.Propagate(err, "")
				}
				return n
			case fsm.MethodSet:
				addNode := cmd.Value.(model.SetNode)
				n, err := d.AddNode(addNode.Type, addNode.ID, addNode.Properties)
				if err != nil {
					return stacktrace.Propagate(err, "")
				}
				return n
			case fsm.MethodDel:
				key := cmd.Value.(model.Key)
				err := d.DelNode(key.Type, key.ID)
				if err != nil {
					return stacktrace.Propagate(err, "")
				}
				return true
			case fsm.MethodBulkDel:
				keys := cmd.Value.([]*model.Key)
				for _, key := range keys {
					err := d.DelNode(key.Type, key.ID)
					if err != nil {
						return stacktrace.Propagate(err, "")
					}
				}
				return true
			case fsm.MethodBulkSet:
				sets := cmd.Value.([]*model.SetNode)
				for _, set := range sets {
					_, err := d.AddNode(set.Type, set.ID, set.Properties)
					if err != nil {
						return stacktrace.Propagate(err, "")
					}
				}
				return true
			case fsm.MethodBulkAdd:
				adds := cmd.Value.([]*model.AddNode)
				for _, add := range adds {
					_, err := d.AddNode(add.Type, *add.ID, add.Properties)
					if err != nil {
						return stacktrace.Propagate(err, "")
					}
				}
				return true
			case fsm.MethodNodeSetProperties:
				props := cmd.Value.(map[string]interface{})
				n, err := d.GetNode(cmd.Metadata["type"], cmd.Metadata["id"])
				if err != nil {
					return stacktrace.Propagate(err, "")
				}
				if err := n.SetProperties(props); err != nil {
					return stacktrace.Propagate(err, "")
				}
				return n
			case fsm.MethodNodeAddRelation:
				key := cmd.Value.(model.Key)
				source, err := d.GetNode(cmd.Metadata["type"], cmd.Metadata["id"])
				if err != nil {
					return stacktrace.Propagate(err, "")
				}
				target, err := d.GetNode(key.Type, key.ID)
				if err != nil {
					return stacktrace.Propagate(err, "")
				}
				rel, err := source.AddRelationship(cmd.Metadata["relation"], target)
				if err != nil {
					return stacktrace.Propagate(err, "")
				}
				return rel
			case fsm.MethodNodeDelRelation:
				key := cmd.Value.(model.Key)
				source, err := d.GetNode(cmd.Metadata["type"], cmd.Metadata["id"])
				if err != nil {
					return stacktrace.Propagate(err, "")
				}
				err = source.DelRelationship(key.Type, key.ID)
				if err != nil {
					return stacktrace.Propagate(err, "")
				}
				return true
			case fsm.MethodRelationSetProperties:
				props := cmd.Value.(map[string]interface{})
				rel, err := d.GetRelationship(cmd.Metadata["type"], cmd.Metadata["id"])
				if err != nil {
					return stacktrace.Propagate(err, "")
				}
				if err := rel.SetProperties(props); err != nil {
					return stacktrace.Propagate(err, "")
				}
				return rel
			default:
				return stacktrace.NewError("unknown method: %s", cmd.Method)
			}
		},
		SnapshotFunc: func() (*fsm.Snapshot, error) {
			return nil, stacktrace.NewError("unimplemented")
		},
		RestoreFunc: func(closer io.ReadCloser) error {
			return stacktrace.NewError("unimplemented")
		},
	}
}
