package fsm

import (
	"github.com/autom8ter/morpheus/pkg/encode"
	"github.com/autom8ter/morpheus/pkg/graph/model"
	"github.com/hashicorp/raft"
	"github.com/palantir/stacktrace"
	"io"
	"time"
)

type Snapshot struct {
	PersistFunc func(sink raft.SnapshotSink) error
	ReleaseFunc func()
}

func (s *Snapshot) Persist(sink raft.SnapshotSink) error {
	return s.PersistFunc(sink)
}

func (s *Snapshot) Release() {
	s.ReleaseFunc()
}

type FSM struct {
	ApplyFunc    func(log *raft.Log) interface{}
	SnapshotFunc func() (*Snapshot, error)
	RestoreFunc  func(closer io.ReadCloser) error
}

func (f *FSM) Apply(log *raft.Log) interface{} {
	return f.ApplyFunc(log)
}

func (f *FSM) Snapshot() (raft.FSMSnapshot, error) {
	return f.SnapshotFunc()
}

func (f *FSM) Restore(closer io.ReadCloser) error {
	return f.RestoreFunc(closer)
}

type Method string

const (
	MethodRelationSetProperties Method = "relation.set_properties"

	MethodSet               Method = "set"
	MethodAdd               Method = "add"
	MethodDel               Method = "del"
	MethodNodeDelRelation   Method = "node.del_relation"
	MethodNodeAddRelation   Method = "node.add_relation"
	MethodNodeSetProperties Method = "node.set_properties"
	MethodBulkAdd           Method = "bulk_add"
	MethodBulkSet           Method = "bulk_set"
	MethodBulkDel           Method = "bulk_del"
)

type CMD struct {
	Method       Method `json:"method"`
	Node         model.Node
	Relationship model.Relationship
	AddNodes     []*model.AddNode
	SetNodes     []*model.SetNode
	Key          model.Key
	Keys         []*model.Key
	Properties   map[string]interface{}
	Timestamp    time.Time         `json:"timestamp"`
	Metadata     map[string]string `json:"metadata"`
}

type CMDHandlerFunc func(c CMD) ([]interface{}, error)

func NewFSM(handlers ...CMDHandlerFunc) raft.FSM {
	return &FSM{
		ApplyFunc: func(log *raft.Log) interface{} {
			cmd := CMD{}
			if err := encode.Unmarshal(log.Data, &cmd); err != nil {
				return stacktrace.Propagate(err, "")
			}
			for _, handler := range handlers {
				values, err := handler(cmd)
				if err != nil {
					return stacktrace.Propagate(err, "")
				}
				if len(values) > 0 {
					return values
				}
			}
			return stacktrace.NewError("failed to apply command")
		},
		SnapshotFunc: func() (*Snapshot, error) {
			return nil, stacktrace.NewError("unimplemented")
		},
		RestoreFunc: func(closer io.ReadCloser) error {
			return stacktrace.NewError("unimplemented")
		},
	}
}
