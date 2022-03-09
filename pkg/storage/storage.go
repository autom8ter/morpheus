package storage

import (
	"context"
	"fmt"
	"github.com/palantir/stacktrace"
	"os"
	"sync"
)

type Storage struct {
	rootDir    string
	buckets    map[string]*Bucket
	mu         sync.RWMutex
	recordSize int
	ctx        context.Context
}

func NewStorage(rootDir string, recordSize int) *Storage {
	return &Storage{mu: sync.RWMutex{}, recordSize: recordSize, buckets: map[string]*Bucket{}, rootDir: rootDir}
}

func (s *Storage) GetBucket(bucket string) *Bucket {
	b, ok := s.buckets[bucket]
	if !ok {
		dir := fmt.Sprintf("%s/%s", s.rootDir, bucket)
		if err := os.MkdirAll(dir, 0700); err != nil {
			panic(stacktrace.Propagate(err, ""))
		}
		closed := int64(0)
		b = &Bucket{
			dir:        dir,
			filekeys:   map[string]map[string]struct{}{},
			keymeta:    map[string]*keyMeta{},
			files:      map[string]*os.File{},
			cache:      map[string]map[string]interface{}{},
			lock:       &sync.RWMutex{},
			recordSize: s.recordSize,
			closed:     &closed,
		}
		s.buckets[bucket] = b
		go func() {
			b.gc(context.Background())
		}()
	}
	return b
}
