package storage

import (
	"fmt"
	"os"
	"sync"
)

type Storage struct {
	rootDir    string
	buckets    map[string]Bucket
	mu         sync.RWMutex
	recordSize int
}

func NewStorage(rootDir string, recordSize int) *Storage {
	os.MkdirAll(rootDir, 0700)
	return &Storage{mu: sync.RWMutex{}, recordSize: recordSize, buckets: map[string]Bucket{}, rootDir: rootDir}
}

func (s *Storage) GetBucket(bucket string) Bucket {
	b, ok := s.buckets[bucket]
	if !ok {
		dir := fmt.Sprintf("%s/%s", s.rootDir, bucket)
		os.MkdirAll(dir, 0700)
		b = Bucket{
			dir:        dir,
			filekeys:   map[string]map[string]struct{}{},
			keymeta:    map[string]*keyMeta{},
			files:      map[string]*os.File{},
			recordSize: s.recordSize,
		}
		s.buckets[bucket] = b
	}
	return b
}
