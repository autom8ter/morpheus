package persistence

import (
	"github.com/autom8ter/morpheus/pkg/encode"
	"github.com/autom8ter/morpheus/pkg/graph/fsm"
	"github.com/autom8ter/morpheus/pkg/helpers"
	"github.com/hashicorp/raft"
	"github.com/palantir/stacktrace"
	"io"
)

func (d *DB) FSM() raft.FSM {
	return &fsm.FSM{
		ApplyFunc: func(log *raft.Log) interface{} {
			var cmd fsm.CMD
			if err := encode.Unmarshal(log.Data, &cmd); err != nil {
				return stacktrace.Propagate(err, "command = %s", helpers.JSONString(cmd))
			}
			switch cmd.Method {
			case fsm.MethodAdd:
				addNode := cmd.Node
				n, err := d.AddNode(addNode.Type, addNode.ID, addNode.Properties)
				if err != nil {
					return stacktrace.Propagate(err, "command = %s", helpers.JSONString(cmd))
				}
				return n
			case fsm.MethodSet:
				addNode := cmd.Node
				n, err := d.AddNode(addNode.Type, addNode.ID, addNode.Properties)
				if err != nil {
					return stacktrace.Propagate(err, "command = %s", helpers.JSONString(cmd))
				}
				return n
			case fsm.MethodDel:
				key := cmd.Key
				err := d.DelNode(key.Type, key.ID)
				if err != nil {
					return stacktrace.Propagate(err, "command = %s", helpers.JSONString(cmd))
				}
				return true
			case fsm.MethodBulkDel:
				keys := cmd.Keys
				for _, key := range keys {
					err := d.DelNode(key.Type, key.ID)
					if err != nil {
						return stacktrace.Propagate(err, "command = %s", helpers.JSONString(cmd))
					}
				}
				return true
			case fsm.MethodBulkSet:
				sets := cmd.SetNodes
				for _, set := range sets {
					_, err := d.AddNode(set.Type, set.ID, set.Properties)
					if err != nil {
						return stacktrace.Propagate(err, "command = %s", helpers.JSONString(cmd))
					}
				}
				return true
			case fsm.MethodBulkAdd:
				adds := cmd.AddNodes
				for _, add := range adds {
					_, err := d.AddNode(add.Type, *add.ID, add.Properties)
					if err != nil {
						return stacktrace.Propagate(err, "command = %s", helpers.JSONString(cmd))
					}
				}
				return true
			case fsm.MethodNodeSetProperties:
				var (
					sourceType = cmd.Metadata["type"]
					sourceID   = cmd.Metadata["id"]
				)
				if sourceType == "" || sourceID == "" {
					return stacktrace.NewError("bad raft cmd")
				}
				props := cmd.Properties
				n, err := d.GetNode(sourceType, sourceID)
				if err != nil {
					return stacktrace.Propagate(err, "command = %s", helpers.JSONString(cmd))
				}
				if err := n.SetProperties(props); err != nil {
					return stacktrace.Propagate(err, "command = %s", helpers.JSONString(cmd))
				}
				return n
			case fsm.MethodNodeAddRelation:
				key := cmd.Key
				var (
					sourceType = cmd.Metadata["source.type"]
					sourceID   = cmd.Metadata["source.id"]
					relation   = cmd.Metadata["relationship"]
				)
				if sourceType == "" || sourceID == "" || relation == "" {
					return stacktrace.NewError("bad raft cmd")
				}
				source, err := d.GetNode(sourceType, sourceID)
				if err != nil {
					return stacktrace.Propagate(err, "command = %s", helpers.JSONString(cmd))
				}
				target, err := d.GetNode(key.Type, key.ID)
				if err != nil {
					return stacktrace.Propagate(err, "command = %s", helpers.JSONString(cmd))
				}
				rel, err := source.AddRelationship(relation, target)
				if err != nil {
					return stacktrace.Propagate(err, "command = %s", helpers.JSONString(cmd))
				}
				return rel
			case fsm.MethodNodeDelRelation:
				var (
					sourceType = cmd.Metadata["type"]
					sourceID   = cmd.Metadata["id"]
				)
				key := cmd.Key
				source, err := d.GetNode(sourceType, sourceID)
				if err != nil {
					return stacktrace.Propagate(err, "command = %s", helpers.JSONString(cmd))
				}
				err = source.DelRelationship(key.Type, key.ID)
				if err != nil {
					return stacktrace.Propagate(err, "command = %s", helpers.JSONString(cmd))
				}
				return true
			case fsm.MethodRelationSetProperties:
				var (
					sourceType = cmd.Metadata["type"]
					sourceID   = cmd.Metadata["id"]
				)
				if sourceType == "" || sourceID == "" {
					return stacktrace.NewError("bad raft cmd")
				}
				props := cmd.Properties
				rel, err := d.GetRelationship(sourceType, sourceID)
				if err != nil {
					return stacktrace.Propagate(err, "command = %s", helpers.JSONString(cmd))
				}
				if err := rel.SetProperties(props); err != nil {
					return stacktrace.Propagate(err, "command = %s", helpers.JSONString(cmd))
				}
				return rel
			default:
				return stacktrace.NewError("unknown method: %s", cmd.Method)
			}
		},
		SnapshotFunc: func() (*fsm.Snapshot, error) {
			return nil, nil
		},
		RestoreFunc: func(closer io.ReadCloser) error {
			return stacktrace.NewError("unimplemented")
		},
	}
}
