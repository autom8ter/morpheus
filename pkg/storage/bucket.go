package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/autom8ter/machine/v3"
	"github.com/autom8ter/morpheus/pkg/storage/cache"
	"github.com/palantir/stacktrace"
	"os"
	"sync"
	"time"
)

const prefixSize = 3

type keyMeta struct {
	File   *os.File
	Offset int64
}

type Bucket struct {
	debug      bool
	dir        string
	filekeys   map[string]map[string]struct{}
	keymeta    map[string]*keyMeta
	files      map[string]*os.File
	cache      *cache.Cache
	queue      chan map[string]map[string]interface{}
	lock       *sync.RWMutex
	recordSize int
	gcInterval time.Duration
	ctx        context.Context
	cancel     func()
	wg         sync.WaitGroup
	machine    machine.Machine
}

func NewBucket(ctx context.Context, dir string, recordSize, cacheSize int, gcInterval time.Duration, debug bool) *Bucket {
	ctx, cancel := context.WithCancel(ctx)
	b := &Bucket{
		debug:      debug,
		dir:        dir,
		filekeys:   map[string]map[string]struct{}{},
		keymeta:    map[string]*keyMeta{},
		files:      map[string]*os.File{},
		cache:      cache.NewCache(cacheSize, gcInterval),
		queue:      make(chan map[string]map[string]interface{}, 100000),
		lock:       &sync.RWMutex{},
		recordSize: recordSize,
		gcInterval: gcInterval,
		ctx:        ctx,
		cancel:     cancel,
		wg:         sync.WaitGroup{},
		machine:    machine.New(machine.WithThrottledRoutines(25)),
	}
	go b.start()
	return b
}

func (s *Bucket) Get(key string) (map[string]interface{}, error) {
	if err := s.ctx.Err(); err != nil {
		return nil, stacktrace.Propagate(err, "")
	}
	if val, ok := s.cache.Get(key); ok {
		return val, nil
	}
	if s.debug {
		fmt.Println(stacktrace.NewError("rlock"))
	}
	s.lock.RLock()
	defer func() {
		s.lock.RUnlock()
		if s.debug {
			fmt.Println(stacktrace.NewError("runlock"))
		}
	}()
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
		return nil, stacktrace.Propagate(err, "incorrect file size")
	}
	buf = bytes.Trim(buf, "\x00")
	data := map[string]interface{}{}
	if err := json.Unmarshal(buf, &data); err != nil {
		return nil, stacktrace.Propagate(err, "failed to decode value")
	}
	s.cache.Set(key, data, true)
	return data, nil
}

func (s *Bucket) Set(key string, value map[string]interface{}) error {
	if err := s.ctx.Err(); err != nil {
		return stacktrace.Propagate(err, "")
	}
	s.queue <- map[string]map[string]interface{}{key: value}
	s.cache.Set(key, value, false)
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
	b.cancel()
	for {
		if len(b.queue) == 0 {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	b.machine.Wait()
	b.wg.Wait()
	for _, file := range b.files {
		file.Close()
	}
	return nil
}

func (b *Bucket) start() {
	b.wg.Add(1)
	go func() {
		defer b.wg.Done()
		b.cache.GC(b.ctx)
	}()
	b.wg.Add(1)
	go func() {
		defer b.wg.Done()
		for {
			select {
			default:
				if len(b.queue) == 0 && b.ctx.Err() != nil {
					return
				}
			case values := <-b.queue:
				b.machine.Go(b.ctx, func(ctx context.Context) error {
					for k, val := range values {
						if err := b.store(k, val); err != nil {
							fmt.Println(stacktrace.Propagate(err, ""))
						}
					}
					return nil
				})
			}
		}
	}()
}

func (b *Bucket) store(k string, val map[string]interface{}) error {
	buf := make([]byte, b.recordSize)
	bits, err := json.Marshal(val)
	if err != nil {
		panic(stacktrace.Propagate(err, "failed to encode record"))
	}
	//fmt.Println("writing record"))
	for i, b := range bits {
		buf[i] = b
	}
	b.lock.Lock()
	defer b.lock.Unlock()
	f, path, err := b.getFile(k)
	if err != nil {
		return stacktrace.Propagate(err, "failed to get file for record: %v", k)
	}

	if pointer, ok := b.keymeta[k]; ok {
		_, err := f.WriteAt(buf, pointer.Offset)
		if err != nil {
			return stacktrace.Propagate(err, "failed to write record")
		}
	} else {
		_, err = f.Write(buf)
		if err != nil {
			return stacktrace.Propagate(err, "failed to write record")
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
	b.cache.Set(k, val, true)
	return nil
}
