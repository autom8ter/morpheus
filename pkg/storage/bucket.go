package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/palantir/stacktrace"
	"os"
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
	recordSize int
}

func (s Bucket) Get(key string) (map[string]interface{}, error) {
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
	return data, nil
}

func (s Bucket) Set(key string, value map[string]interface{}) error {
	buf := make([]byte, s.recordSize)
	bits, err := json.Marshal(value)
	if err != nil {
		return stacktrace.Propagate(err, "failed to encode record")
	}
	for i, b := range bits {
		buf[i] = b
	}
	f, path, err := s.getFile(key)
	if err != nil {
		return stacktrace.Propagate(err, "failed to get file for record")
	}
	if pointer, ok := s.keymeta[key]; ok {
		_, err := f.WriteAt(buf, pointer.Offset)
		if err != nil {
			return stacktrace.Propagate(err, "failed to write record")
		}
		return nil
	}
	_, err = f.Write(buf)
	if err != nil {
		return stacktrace.Propagate(err, "failed to write record")
	}
	if s.filekeys[path] == nil {
		s.filekeys[path] = map[string]struct{}{}
	}
	s.filekeys[path][key] = struct{}{}
	s.keymeta[key] = &keyMeta{
		File:   f,
		Offset: int64((len(s.filekeys[path]) - 1) * s.recordSize),
	}
	return nil
}

func (s Bucket) getFile(key string) (*os.File, string, error) {
	var err error
	h := hash(key)
	path := fmt.Sprintf("%s/%s", s.dir, h[:prefixSize])

	f, ok := s.files[h[:prefixSize]]
	if !ok {
		f, err = os.Create(path)
		if err != nil {
			return nil, path, err
		}
		s.files[h[:prefixSize]] = f
	}
	return f, path, nil
}

func (s Bucket) getFileIfExists(key string) (*os.File, bool) {
	h := hash(key)
	f, ok := s.files[h[:prefixSize]]
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
