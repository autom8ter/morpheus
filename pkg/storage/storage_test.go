package storage

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestSetGet(t *testing.T) {
	d, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(d)
	s := NewStorage(context.Background(), d, 1024, true)
	defer s.Close()
	b := s.GetBucket("users")

	cases := []struct {
		key    string
		values map[string]interface{}
		check  func(map[string]interface{}) error
	}{
		{
			key: "usr_1",
			values: map[string]interface{}{
				"name": "coleman",
			},
			check: func(m map[string]interface{}) error {
				if m["name"] != "coleman" {
					return fmt.Errorf("expected name to be coleman")
				}
				return nil
			},
		},
		{
			key: "usr_2",
			values: map[string]interface{}{
				"name": "lacee",
			},
			check: func(m map[string]interface{}) error {
				if m["name"] != "lacee" {
					return fmt.Errorf("expected name to be lacee")
				}
				return nil
			},
		},
	}
	for _, cs := range cases {
		now := time.Now()
		defer func() {
			t.Logf("%s %v", cs.key, time.Since(now).Nanoseconds())
		}()
		if err := b.Set(cs.key, cs.values); err != nil {
			t.Fatal(err)
		}
		val, err := b.Get(cs.key)
		if err != nil {
			t.Fatal(err)
		}
		if err := cs.check(val); err != nil {
			t.Fatal(err)
		}
	}
}

/*
goos: darwin
goarch: amd64
pkg: github.com/autom8ter/morpheus/pkg/storage
cpu: Intel(R) Core(TM) i9-9880H CPU @ 2.30GHz
BenchmarkSetGetOneKey-16          504018              2090 ns/op            1388 B/op         12 allocs/op
*/
func BenchmarkSetGetOneKey(b *testing.B) {
	d, err := ioutil.TempDir("", "bucket-test")
	if err != nil {
		b.Fatal(err)
	}
	s := NewStorage(context.Background(), d, 1024, false)

	bucket := s.GetBucket("users")
	b.ResetTimer()
	b.ReportAllocs()
	key := "usr_1"
	for n := 0; n < b.N; n++ {
		if err := bucket.Set(key, map[string]interface{}{
			"name": "coleman",
		}); err != nil {
			b.Fatal(err)
		}
		_, err := bucket.Get(key)
		if err != nil {
			b.Fatal(err)
		}
	}
	b.Cleanup(func() {
		s.Close()
		os.RemoveAll(d)
	})
}


// BenchmarkSetGetManyKey-16                 720487              3854 ns/op            2224 B/op         22 allocs/op
func BenchmarkSetGetManyKey(b *testing.B) {
	d, err := ioutil.TempDir("", "")
	if err != nil {
		b.Fatal(err)
	}
	s := NewStorage(context.Background(), d, 1024, false)
	bucket := s.GetBucket("users")
	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		key := fmt.Sprintf("id-%v", n)
		if err := bucket.Set(key, map[string]interface{}{
			"name": "coleman",
		}); err != nil {
			b.Fatal(err)
		}
		_, err := bucket.Get(key)
		if err != nil {
			b.Fatal(err)
		}
	}
	b.Cleanup(func() {
		bucket.Close()
		os.RemoveAll(d)

	})
}

// BenchmarkGetOneKey-16                    9067416               208.0 ns/op             0 B/op          0 allocs/op
func BenchmarkGetOneKey(b *testing.B) {
	d, err := ioutil.TempDir("", "")
	if err != nil {
		b.Fatal(err)
	}
	s := NewStorage(context.Background(), d, 1024, false)
	bucket := s.GetBucket("users")
	key := "usr_1"
	if err := bucket.Set(key, map[string]interface{}{
		"name": "coleman",
	}); err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		_, err := bucket.Get(key)
		if err != nil {
			b.Fatal(err)
		}
	}
	b.Cleanup(func() {
		bucket.Close()
		os.RemoveAll(d)

	})
}
//
////     72301             18549 ns/op            2414 B/op         60 allocs/op
//func BenchmarkBadgerSetGetManyKey(b *testing.B) {
//	d, err := ioutil.TempDir("", "badger-test")
//	if err != nil {
//		b.Fatal(err)
//	}
//	// defer os.RemoveAll(d)
//	opts := badger.DefaultOptions(d)
//
//	db, err := badger.Open(opts)
//	if err != nil {
//		panic(err)
//	}
//	b.ResetTimer()
//	b.ReportAllocs()
//	for n := 0; n < b.N; n++ {
//		if err := db.Update(func(txn *badger.Txn) error {
//			bits, _ := json.Marshal(map[string]interface{}{
//				"name": "coleman",
//			})
//			return txn.Set([]byte(fmt.Sprint(n)), bits)
//		}); err != nil {
//			b.Fatal(err)
//		}
//		if err := db.View(func(txn *badger.Txn) error {
//			_, err := txn.Get([]byte(fmt.Sprint(n)))
//			if err != nil {
//				b.Fatal(err)
//			}
//			return nil
//		}); err != nil {
//			b.Fatal(err)
//		}
//	}
//	b.Cleanup(func() {
//		db.Close()
//		os.RemoveAll(d)
//	})
//}
//
////     72301             18549 ns/op            2414 B/op         60 allocs/op
//func BenchmarkBadgerGetOneKey(b *testing.B) {
//	d, err := ioutil.TempDir("", "badger-test")
//	if err != nil {
//		b.Fatal(err)
//	}
//	// defer os.RemoveAll(d)
//	opts := badger.DefaultOptions(d)
//
//	db, err := badger.Open(opts)
//	if err != nil {
//		panic(err)
//	}
//	if err := db.Update(func(txn *badger.Txn) error {
//		bits, _ := json.Marshal(map[string]interface{}{
//			"name": "coleman",
//		})
//		return txn.Set([]byte("usr_1"), bits)
//	}); err != nil {
//		b.Fatal(err)
//	}
//	b.ResetTimer()
//	b.ReportAllocs()
//	for n := 0; n < b.N; n++ {
//
//		if err := db.View(func(txn *badger.Txn) error {
//			_, err := txn.Get([]byte("usr_1"))
//			if err != nil {
//				b.Fatal(err)
//			}
//			return nil
//		}); err != nil {
//			b.Fatal(err)
//		}
//	}
//	b.Cleanup(func() {
//		db.Close()
//		os.RemoveAll(d)
//	})
//}
