package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/palantir/stacktrace"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

const prefixSize = 2

type keyMeta struct {
	File   *os.File
	Offset int64
}

type Bucket struct {
	dir        string
	filekeys   map[string]map[string]struct{}
	keymeta    map[string]*keyMeta
	files      map[string]*os.File
	cache      map[string]map[string]interface{}
	lock       *sync.RWMutex
	recordSize int
	closed     *int64
}

type cachedItem struct {
	timestamp time.Time
	values    map[string]interface{}
}

func (s *Bucket) Get(key string) (map[string]interface{}, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	if val, ok := s.cache[key]; ok {
		return val, nil
	}

	pointer, ok := s.keymeta[key]
	if !ok {
		return nil, stacktrace.NewError("not found")
	}
	buf := make([]byte, s.recordSize)
	count, err := pointer.File.ReadAt(buf, pointer.Offset)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to read value")
	}
	if count != s.recordSize {
		return nil, stacktrace.NewError("incorrect file size")
	}
	buf = bytes.Trim(buf, "\x00")
	data := map[string]interface{}{}
	if err := json.Unmarshal(buf, &data); err != nil {
		return nil, stacktrace.Propagate(err, "failed to decode value")
	}
	s.cache[key] = data
	return data, nil
}

func (s *Bucket) Set(key string, value map[string]interface{}) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.cache[key] = value
	return nil
}

func (s *Bucket) getFile(key string) (*os.File, string, error) {
	var err error
	h := hash(key)
	path := fmt.Sprintf("%s/%s", s.dir, h[:prefixSize])

	f, ok := s.files[h[:prefixSize]]
	if !ok {
		f, err = os.Create(path)
		if err != nil {
			return nil, path, stacktrace.Propagate(err, "failed to create file: %s", path)
		}
		s.files[h[:prefixSize]] = f
	}
	return f, path, nil
}

func (s *Bucket) getFileIfExists(key string) (*os.File, bool) {
	h := hash(key)
	f, ok := s.files[h[:prefixSize]]
	if !ok {
		return nil, false
	}
	return f, true
}

func (b *Bucket) Close() error {
	atomic.StoreInt64(b.closed, 1)
	for {
		b.lock.RLock()
		defer b.lock.RUnlock()
		if len(b.cache) == 0 {
			break
		}
	}
	time.Sleep(1 * time.Second)
	for _, file := range b.files {
		file.Close()
	}
	return nil
}

func (b *Bucket) gc(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	tick := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-tick.C:
			b.lock.Lock()
			defer b.lock.Unlock()
			for k, val := range b.cache {
				buf := make([]byte, b.recordSize)
				bits, err := json.Marshal(val)
				if err != nil {
					panic(stacktrace.Propagate(err, "failed to encode record"))
				}
				fmt.Println("DEBUG", k, string(bits))
				for i, b := range bits {
					buf[i] = b
				}
				f, path, err := b.getFile(k)
				if err != nil {
					panic(stacktrace.Propagate(err, "failed to get file for record: %v", k))
				}
				if pointer, ok := b.keymeta[k]; ok {
					_, err := f.WriteAt(buf, pointer.Offset)
					if err != nil {
						panic(stacktrace.Propagate(err, "failed to write record"))
					}
				} else {
					_, err = f.Write(buf)
					if err != nil {
						panic(stacktrace.Propagate(err, "failed to write record"))
					}
				}
				if b.filekeys[path] == nil {
					b.filekeys[path] = map[string]struct{}{}
				}
				b.filekeys[path][k] = struct{}{}
				b.keymeta[k] = &keyMeta{
					File:   f,
					Offset: int64((len(b.filekeys[path]) - 1) * b.recordSize),
				}
				delete(b.cache, k)
			}
		}
	}
}
