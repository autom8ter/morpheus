package persistence

import (
	"fmt"
	"github.com/autom8ter/morpheus/pkg/encode"
	"github.com/autom8ter/morpheus/pkg/graph/fsm"
	"github.com/autom8ter/morpheus/pkg/logger"
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
				addNode := cmd.Node
				logger.L.Info(fmt.Sprintf("%s/%s", addNode.Type, addNode.ID), addNode.Properties)
				n, err := d.AddNode(addNode.Type, addNode.ID, addNode.Properties)
				if err != nil {
					return stacktrace.Propagate(err, "")
				}
				return n
			case fsm.MethodSet:
				addNode := cmd.Node
				n, err := d.AddNode(addNode.Type, addNode.ID, addNode.Properties)
				if err != nil {
					return stacktrace.Propagate(err, "")
				}
				return n
			case fsm.MethodDel:
				key := cmd.Key
				err := d.DelNode(key.Type, key.ID)
				if err != nil {
					return stacktrace.Propagate(err, "")
				}
				return true
			case fsm.MethodBulkDel:
				keys := cmd.Keys
				for _, key := range keys {
					err := d.DelNode(key.Type, key.ID)
					if err != nil {
						return stacktrace.Propagate(err, "")
					}
				}
				return true
			case fsm.MethodBulkSet:
				sets := cmd.SetNodes
				for _, set := range sets {
					_, err := d.AddNode(set.Type, set.ID, set.Properties)
					if err != nil {
						return stacktrace.Propagate(err, "")
					}
				}
				return true
			case fsm.MethodBulkAdd:
				adds := cmd.AddNodes
				for _, add := range adds {
					_, err := d.AddNode(add.Type, *add.ID, add.Properties)
					if err != nil {
						return stacktrace.Propagate(err, "")
					}
				}
				return true
			case fsm.MethodNodeSetProperties:
				props := cmd.Properties
				n, err := d.GetNode(cmd.Metadata["type"], cmd.Metadata["id"])
				if err != nil {
					return stacktrace.Propagate(err, "")
				}
				if err := n.SetProperties(props); err != nil {
					return stacktrace.Propagate(err, "")
				}
				return n
			case fsm.MethodNodeAddRelation:
				key := cmd.Key
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
				key := cmd.Key
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
				props := cmd.Properties
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
