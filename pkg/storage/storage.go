package storage

import (
	"context"
	"fmt"
	"github.com/palantir/stacktrace"
	"os"
	"sync"
	"time"
)

type Storage struct {
	debug      bool
	rootDir    string
	buckets    map[string]*Bucket
	mu         sync.RWMutex
	recordSize int
	ctx        context.Context
	cancel     func()
}

func NewStorage(ctx context.Context, rootDir string, recordSize int, debug bool) *Storage {
	ctx, cancel := context.WithCancel(ctx)
	return &Storage{
		debug:      debug,
		rootDir:    rootDir,
		buckets:    map[string]*Bucket{},
		mu:         sync.RWMutex{},
		recordSize: recordSize,
		ctx:        ctx,
		cancel:     cancel,
	}
}

func (s *Storage) GetBucket(bucket string) *Bucket {
	b, ok := s.buckets[bucket]
	if !ok {
		dir := fmt.Sprintf("%s/%s", s.rootDir, bucket)
		if err := os.MkdirAll(dir, 0700); err != nil {
			panic(stacktrace.Propagate(err, ""))
		}
		b = NewBucket(s.ctx, dir, s.recordSize, 1000000, 1*time.Minute, s.debug)
		s.buckets[bucket] = b
	}
	return b
}

func (s *Storage) Close() {
	s.cancel()
	for _, b := range s.buckets {
		b.Close()
	}
}
