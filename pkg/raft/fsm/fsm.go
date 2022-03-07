package fsm

import (
	"encoding/gob"
	"fmt"
	"github.com/autom8ter/morpheus/pkg/api"
	"github.com/autom8ter/morpheus/pkg/encode"
	"github.com/hashicorp/raft"
	"io"
)

func init() {
	gob.Register(&raft.Log{})
}

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

type AddNode struct {
	Type       string
	ID         string
	Properties map[string]interface{}
}

type DelNode struct {
	Type string
	ID   string
}

type SetNodeProperty struct {
	Type       string
	ID         string
	Properties map[string]interface{}
}

type SetRelationshipProperty struct {
	NodeType       string
	NodeID         string
	Relationship   string
	RelationshipID string
	Properties     map[string]interface{}
}

type AddRelationship struct {
	NodeType       string
	NodeID         string
	Direction      string
	Relationship   string
	RelationshipID string
	Node2Type      string
	Node2ID        string
}

type DelRelationship struct {
	NodeType       string
	NodeID         string
	Direction      string
	Relationship   string
	RelationshipID string
}

type Method string

const (
	AddNodes                  = "ADD NODES"
	DelNodes                  = "DEL NODES"
	SetNodeProperties         = "SET NODE PROPERTIES"
	AddRelationships          = "ADD RELATIONSHIPS"
	DelRelationships          = "DEL RELATIONSHIPS"
	SetRelationshipProperties = "SET RELATIONSHIP PROPERTIES"
)

type CMD struct {
	Method                    Method                    `json:"method"`
	AddNodes                  []AddNode                 `json:"addNodes"`
	DelNodes                  []DelNode                 `json:"delNodes"`
	SetNodeProperties         []SetNodeProperty         `json:"setNodeProperties"`
	SetRelationshipProperties []SetRelationshipProperty `json:"setRelationshipProperties"`
	AddRelationships          []AddRelationship         `json:"addRelationships"`
	DelRelationships          []DelRelationship         `json:"delRelationships"`
}

func NewGraphFSM(g api.Graph) raft.FSM {
	return &FSM{
		ApplyFunc: func(log *raft.Log) interface{} {
			cmd := &CMD{}
			if err := encode.Unmarshal(log.Data, cmd); err != nil {
				return err
			}
			switch cmd.Method {
			case AddNodes:
				var nodes []api.Node
				for _, n := range cmd.AddNodes {
					n, err := g.AddNode(n.Type, n.ID, n.Properties)
					if err != nil {
						return err
					}
					nodes = append(nodes, n)
				}
				return nodes
			case DelNodes:
				for _, n := range cmd.DelNodes {
					err := g.DelNode(n.Type, n.ID)
					if err != nil {
						return err
					}
				}
				return nil
			case SetNodeProperties:
				var nodes []api.Node
				for _, node := range cmd.SetNodeProperties {
					n, err := g.GetNode(node.Type, node.ID)
					if err != nil {
						return err
					}
					n.SetProperties(node.Properties)
					nodes = append(nodes, n)
				}
				return nodes
			case SetRelationshipProperties:
				var rels []api.Relationship
				for _, relationship := range cmd.SetRelationshipProperties {
					r, err := g.GetRelationship(relationship.Relationship, relationship.RelationshipID)
					if err != nil {
						return err
					}
					r.SetProperties(relationship.Properties)
					rels = append(rels, r)
				}
				return rels
			case AddRelationships:
				var rels []api.Relationship
				for _, relation := range cmd.AddRelationships {
					n, err := g.GetNode(relation.NodeType, relation.NodeID)
					if err != nil {
						return err
					}
					n2, err := g.GetNode(relation.Node2Type, relation.Node2ID)
					if err != nil {
						return err
					}
					rels = append(rels, n.AddRelationship(api.Direction(relation.Direction), relation.Relationship, relation.RelationshipID, n2))
				}
				return rels
			case DelRelationships:
				for _, relation := range cmd.DelRelationships {
					n, err := g.GetNode(relation.NodeType, relation.NodeID)
					if err != nil {
						return err
					}
					n.DelRelationship(api.Direction(relation.Direction), relation.Relationship, relation.RelationshipID)
				}
				return nil
			default:
				return fmt.Errorf("unsupported method: %s", cmd.Method)
			}
		},
		SnapshotFunc: func() (*Snapshot, error) {
			return nil, fmt.Errorf("raft: snapshot unimplemented")

		},
		RestoreFunc: func(closer io.ReadCloser) error {
			return fmt.Errorf("raft: restore unimplemented")
		},
	}
}
