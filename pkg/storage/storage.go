package storage

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"os"
	"sync"
)

type Storage struct {
	buckets    map[string]Bucket
	mu         sync.RWMutex
	recordSize int
}

func NewStorage(recordSize int) *Storage {
	return &Storage{mu: sync.RWMutex{}, recordSize: recordSize, buckets: map[string]Bucket{}}
}

func (s *Storage) GetBucket(bucket string) Bucket {
	b, ok := s.buckets[bucket]
	if !ok {
		os.Mkdir(bucket, 0700)
		b = Bucket{
			dir:        bucket,
			filekeys:   map[string]struct{}{},
			pointers:   map[string]int64{},
			files:      map[string]*os.File{},
			recordSize: s.recordSize,
		}
		s.buckets[bucket] = b
	}
	return b
}

type Bucket struct {
	dir        string
	filekeys   map[string]struct{}
	pointers   map[string]int64
	files      map[string]*os.File
	recordSize int
}

func (s Bucket) Get(key string) (map[string]interface{}, error) {
	f, ok := s.getFileIfExists(key)
	if !ok {
		return nil, errors.New("not found")
	}
	offset, ok := s.pointers[key]
	if !ok {
		return nil, errors.New("not found")
	}
	buf := make([]byte, s.recordSize)
	count, err := f.ReadAt(buf, offset)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read value")
	}
	if count != s.recordSize {
		return nil, fmt.Errorf("incorrect file size")
	}
	buf = bytes.Trim(buf, "\x00")
	data := map[string]interface{}{}
	if err := json.Unmarshal(buf, &data); err != nil {
		return nil, errors.Wrap(err, "failed to decode value")
	}
	return data, nil
}

func (s Bucket) Set(key string, value map[string]interface{}) error {
	buf := make([]byte, s.recordSize)
	bits, err := json.Marshal(value)
	if err != nil {
		return errors.Wrap(err, "failed to encode record")
	}
	for i, b := range bits {
		buf[i] = b
	}
	f, err := s.getFile(key)
	if err != nil {
		return errors.Wrap(err, "failed to get file for record")
	}
	_, err = f.Write(buf)
	if err != nil {
		return errors.Wrap(err, "failed to write record")
	}
	//if count != s.recordSize {
	//	return fmt.Errorf("incorrect file size: %v", count)
	//}
	s.filekeys[key] = struct{}{}
	s.pointers[key] = int64((len(s.filekeys) - 1) * s.recordSize)
	return nil
}

func hash(key string) string {
	h := sha1.New()
	h.Write([]byte(key))

	return hex.EncodeToString(h.Sum(nil))
}

func (s Bucket) getFile(key string) (*os.File, error) {
	var err error
	h := hash(key)
	path := fmt.Sprintf("%s/%s", s.dir, h[:3])

	f, ok := s.files[string(h[:3])]
	if !ok {
		f, err = os.Create(path)
		if err != nil {
			return nil, err
		}
		s.files[string(h[:3])] = f
	}
	return f, nil
}

func (s Bucket) getFileIfExists(key string) (*os.File, bool) {
	h := hash(key)
	f, ok := s.files[h[:3]]
	if !ok {
		return nil, false
	}
	return f, true
}

func (b Bucket) Close() error {
	for _, file := range b.files {
		file.Close()
	}
	return nil
}
