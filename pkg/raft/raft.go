package raft

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/autom8ter/morpheus/pkg/api"
	"github.com/autom8ter/morpheus/pkg/logger"
	fsm2 "github.com/autom8ter/morpheus/pkg/raft/fsm"
	"github.com/autom8ter/morpheus/pkg/raft/storage"
	transport2 "github.com/autom8ter/morpheus/pkg/raft/transport"
	"github.com/hashicorp/raft"
	"github.com/pkg/errors"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"time"
)

type Raft struct {
	raft *raft.Raft
	opts *Options
}

func NewRaft(g api.Graph, lis net.Listener, opts ...Opt) (*Raft, error) {
	fsm := fsm2.NewGraphFSM(g)

	options := &Options{}
	for _, o := range opts {
		o(options)
	}
	options.setDefaults()
	config := raft.DefaultConfig()
	config.LogLevel = "INFO"
	if options.debug {
		config.LogLevel = "DEBUG"
	}
	config.NoSnapshotRestoreOnStart = !options.restoreSnapshotOnRestart
	config.LocalID = raft.ServerID(options.peerID)
	if options.leaseTimeout != 0 {
		config.LeaderLeaseTimeout = options.leaseTimeout
	}
	if options.heartbeatTimeout != 0 {
		config.HeartbeatTimeout = options.heartbeatTimeout
	}
	if options.electionTimeout != 0 {
		config.ElectionTimeout = options.electionTimeout
	}
	if options.commitTimeout != 0 {
		config.CommitTimeout = options.commitTimeout
	}
	host, _ := os.Hostname()
	path := fmt.Sprintf("%s/%s", options.raftDir, host)
	snapshotPath := fmt.Sprintf("%s/snapshots", path)
	os.MkdirAll(snapshotPath, 0700)
	transport := transport2.NewNetworkTransport(lis, options.advertise, options.maxPool, options.timeout, os.Stderr)
	snapshots, err := raft.NewFileSnapshotStore(snapshotPath, options.retainSnapshots, os.Stderr)
	if err != nil {
		return nil, err
	}
	storagePath := fmt.Sprintf("%s/storage", path)
	os.MkdirAll(storagePath, 0700)

	strg, err := storage.NewStorage(storagePath)
	if err != nil {
		return nil, err
	}

	ra, err := raft.NewRaft(config, fsm, strg, strg, snapshots, transport)
	if err != nil {
		return nil, err
	}

	if options.isLeader {
		configuration := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      config.LocalID,
					Address: transport.LocalAddr(),
				},
			},
		}
		ra.BootstrapCluster(configuration)
	}
	return &Raft{
		opts: options,
		raft: ra,
	}, nil
}

func (r *Raft) State() raft.RaftState {
	return r.raft.State()
}

func (s *Raft) Servers() ([]raft.Server, error) {
	configFuture := s.raft.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		return nil, err
	}
	return configFuture.Configuration().Servers, nil
}

func (s *Raft) Join(nodeID, addr string) error {
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return errors.Wrapf(err, "failed to parse voter %s address", nodeID)
	}
	configFuture := s.raft.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		return errors.Wrap(err, "failed to get raft configuration")
	}

	for _, srv := range configFuture.Configuration().Servers {
		if srv.ID == raft.ServerID(nodeID) || srv.Address == raft.ServerAddress(addr) {
			// However if *both* the ID and the address are the same, then nothing -- not even
			// a join operation -- is needed.
			if srv.Address == raft.ServerAddress(addr) && srv.ID == raft.ServerID(nodeID) {
				// already a member
				return nil
			}

			future := s.raft.RemoveServer(srv.ID, 0, 0)
			if err := future.Error(); err != nil {
				return fmt.Errorf("error removing existing node %s at %s: %s", nodeID, addr, err)
			}
		}
	}
	errs := 0
	for {
		f := s.raft.AddVoter(raft.ServerID(nodeID), raft.ServerAddress(tcpAddr.String()), 0, 0)
		if err := f.Error(); err != nil {
			errs++
			if errs >= 5 {
				return errors.Wrap(err, "failed to add raft voter")
			}
			time.Sleep(1 * time.Second)
			continue
		} else {
			break
		}
	}
	return nil
}

func (s *Raft) LeaderAddr() string {
	return string(s.raft.Leader())
}

func (s *Raft) Stats() map[string]string {
	return s.raft.Stats()
}

func (s *Raft) Apply(bits []byte) (interface{}, error) {
	f := s.raft.Apply(bits, s.opts.timeout)
	if err := f.Error(); err != nil {
		return nil, err
	}
	resp := f.Response()
	if err, ok := resp.(error); ok {
		return nil, err
	}
	return resp, nil
}

func (r *Raft) Close() error {
	return r.raft.Shutdown().Error()
}

func (r *Raft) PeerID() string {
	return r.opts.peerID
}

func hash(val []byte) string {
	h := sha1.New()
	h.Write(val)
	bs := h.Sum(nil)
	return hex.EncodeToString(bs)
}

const raftUser = "morpheus-raft"

type JoinRaftRequest struct {
	NodeID   string `json:"nodeId"`
	NodeAddr string `json:"nodeAddr"`
}

func (rft *Raft) JoinHTTPHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		usr, pw, ok := r.BasicAuth()
		if !ok {
			http.Error(w, "raft: missing username/password", http.StatusUnauthorized)
			return
		}
		if usr != raftUser {
			http.Error(w, "raft: incorrect username", http.StatusUnauthorized)
			return
		}
		if pw != rft.opts.raftSecret {
			http.Error(w, "raft: incorrect password", http.StatusUnauthorized)
			return
		}
		req := &JoinRaftRequest{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "raft: failed to encode request", http.StatusBadRequest)
			return
		}
		if err := rft.Join(req.NodeID, req.NodeAddr); err != nil {
			logger.L.Error("raft: failed to join cluster", map[string]interface{}{
				"error":     err,
				"node_id":   req.NodeID,
				"node_addr": req.NodeAddr,
			})
			http.Error(w, "raft: failed to join cluster", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func JoinCluster(ctx context.Context, clusterAddr, localAddr, raftSecret string) bool {
	for x := 0; x < 30; x++ {
		if ctx.Err() != nil {
			return false
		}
		if err := pingTCP(localAddr); err != nil {
			logger.L.Error("failed to join cluster - retrying",
				map[string]interface{}{
					"error": err,
				},
			)
			time.Sleep(1 * time.Second)
			continue
		}
		req, err := http.NewRequest(http.MethodPost, clusterAddr, nil)
		if err != nil {
			logger.L.Error("failed to join cluster", map[string]interface{}{
				"error": err,
			})
			time.Sleep(1 * time.Second)
			continue
		}
		req.SetBasicAuth(raftUser, raftSecret)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			logger.L.Error("failed to join cluster", map[string]interface{}{
				"error": err,
			})
			time.Sleep(1 * time.Second)
			continue
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			return true
		}
		if resp.StatusCode != http.StatusOK {
			bits, _ := ioutil.ReadAll(resp.Body)
			logger.L.Error("failed to join cluster", map[string]interface{}{
				"response": string(bits),
			})
			time.Sleep(1 * time.Second)
			continue
		}
	}
	return false
}

func pingTCP(address string) error {
	timeout := time.Second
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return errors.Wrap(err, "tcp ping failure")
	}
	defer conn.Close()
	return nil
}
