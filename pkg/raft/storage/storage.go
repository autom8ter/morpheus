package storage

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"github.com/autom8ter/morpheus/pkg/helpers"
	"github.com/dgraph-io/badger/v3"
	"github.com/hashicorp/raft"
	"math"
	"strconv"
)

var (
	dbLogsPrefix = []byte("logs")
	dbConfPrefix = []byte("conf")

	// ErrKeyNotFound is an error indicating a given key does not exist
	ErrKeyNotFound = errors.New("not found")
)

type Storage struct {
	db *badger.DB
}

func NewStorage(path string) (*Storage, error) {
	db, err := badger.Open(badger.DefaultOptions(path))
	if err != nil {
		return nil, err
	}
	return &Storage{db: db}, nil
}
func (b *Storage) FirstIndex() (uint64, error) {
	first := uint64(0)
	err := b.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		it.Seek(dbLogsPrefix)
		if it.ValidForPrefix(dbLogsPrefix) {
			item := it.Item()
			k := string(item.Key()[len(dbLogsPrefix):])
			idx, err := strconv.ParseUint(k, 10, 64)
			if err != nil {
				return err
			}
			first = idx
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return first, nil
}

func (b *Storage) LastIndex() (uint64, error) {
	last := uint64(0)
	if err := b.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Reverse = true
		it := txn.NewIterator(opts)
		defer it.Close()
		// ensure reverse seeking will include the
		// see https://github.com/dgraph-io/badger/issues/436 and
		// https://github.com/dgraph-io/badger/issues/347
		seekKey := append(dbLogsPrefix, 0xFF)
		it.Seek(seekKey)
		if it.ValidForPrefix(dbLogsPrefix) {
			item := it.Item()
			k := string(item.Key()[len(dbLogsPrefix):])
			idx, err := strconv.ParseUint(k, 10, 64)
			if err != nil {
				return err
			}
			last = idx
		}
		return nil
	}); err != nil {
		return 0, err
	}
	return last, nil
}

func (b *Storage) GetLog(idx uint64, log *raft.Log) error {
	return b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(fmt.Sprintf("%s%d", dbLogsPrefix, idx)))
		if err != nil {
			return err
		}
		if item == nil {
			return raft.ErrLogNotFound
		}
		return item.Value(func(val []byte) error {
			buf := bytes.NewBuffer(val)
			dec := gob.NewDecoder(buf)
			return dec.Decode(&log)
		})
	})
}

func (b *Storage) StoreLog(log *raft.Log) error {
	return b.StoreLogs([]*raft.Log{log})
}

// StoreLogs is used to store a set of raft logs
func (b *Storage) StoreLogs(logs []*raft.Log) error {
	maxBatchSize := b.db.MaxBatchSize()
	min := uint64(0)
	max := uint64(len(logs))
	ranges := b.generateRanges(min, max, maxBatchSize)
	for _, r := range ranges {
		txn := b.db.NewTransaction(true)
		defer txn.Discard()
		for index := r.from; index < r.to; index++ {
			log := logs[index]
			key := []byte(fmt.Sprintf("%s%d", dbLogsPrefix, log.Index))
			var out bytes.Buffer
			enc := gob.NewEncoder(&out)
			enc.Encode(log)
			if err := txn.Set(key, out.Bytes()); err != nil {
				return err
			}
		}
		if err := txn.Commit(); err != nil {
			return err
		}
	}
	return nil
}

type iteratorRange struct{ from, to uint64 }

func (b *Storage) generateRanges(min, max uint64, batchSize int64) []iteratorRange {
	nSegments := int(math.Round(float64((max - min) / uint64(batchSize))))
	segments := []iteratorRange{}
	if (max - min) <= uint64(batchSize) {
		segments = append(segments, iteratorRange{from: min, to: max})
		return segments
	}
	for len(segments) < nSegments {
		nextMin := min + uint64(batchSize)
		segments = append(segments, iteratorRange{from: min, to: nextMin})
		min = nextMin + 1
	}
	segments = append(segments, iteratorRange{from: min, to: max})
	return segments
}

func (b *Storage) DeleteRange(min, max uint64) error {
	maxBatchSize := b.db.MaxBatchSize()
	ranges := b.generateRanges(min, max, maxBatchSize)
	for _, r := range ranges {
		txn := b.db.NewTransaction(true)
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer txn.Discard()

		it.Rewind()
		// Get the key to start at
		minKey := []byte(fmt.Sprintf("%s%d", dbLogsPrefix, r.from))
		for it.Seek(minKey); it.ValidForPrefix(dbLogsPrefix); it.Next() {
			item := it.Item()
			// get the index as a string to convert to uint64
			k := string(item.Key()[len(dbLogsPrefix):])
			idx, err := strconv.ParseUint(k, 10, 64)
			if err != nil {
				it.Close()
				return err
			}
			// Handle out-of-range index
			if idx > r.to {
				break
			}
			// Delete in-range index
			delKey := []byte(fmt.Sprintf("%s%d", dbLogsPrefix, idx))
			if err := txn.Delete(delKey); err != nil {
				it.Close()
				return err
			}
		}
		it.Close()
		if err := txn.Commit(); err != nil {
			return err
		}
	}
	return nil
}

func (b *Storage) Set(k, v []byte) error {
	return b.db.Update(func(txn *badger.Txn) error {
		key := []byte(fmt.Sprintf("%s%d", dbConfPrefix, k))
		return txn.Set(key, v)
	})
}

func (b *Storage) Get(k []byte) ([]byte, error) {
	txn := b.db.NewTransaction(false)
	defer txn.Discard()
	key := []byte(fmt.Sprintf("%s%d", dbConfPrefix, k))
	item, err := txn.Get(key)
	if item == nil {
		return nil, ErrKeyNotFound
	}
	if err != nil {
		return nil, err
	}
	v, err := item.ValueCopy(nil)
	if err != nil {
		return nil, err
	}
	if err := txn.Commit(); err != nil {
		return nil, err
	}
	return append([]byte(nil), v...), nil
}

func (b *Storage) SetUint64(key []byte, val uint64) error {
	return b.Set(key, helpers.Uint64ToBytes(val))
}

func (b *Storage) GetUint64(key []byte) (uint64, error) {
	val, err := b.Get(key)
	if err != nil {
		return 0, err
	}
	return helpers.BytesToUint64(val), nil
}
