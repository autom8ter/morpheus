package storage

import (
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
	s := NewStorage(d, 1024)
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
BenchmarkSetGetOneKey-16          121875              9142 ns/op            3500 B/op         29 allocs/op
*/
func BenchmarkSetGetOneKey(b *testing.B) {
	d, err := ioutil.TempDir("", "bucket-test")
	if err != nil {
		b.Fatal(err)
	}
	s := NewStorage(d, 1024)
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
	//b.Cleanup(func() {
	//	bucket.Close()
	//	os.RemoveAll(d)
	//})
}

//
//// BenchmarkSetGetManyKey-16                  83254             13831 ns/op            3707 B/op         34 allocs/op
//func BenchmarkSetGetManyKey(b *testing.B) {
//	d, err := ioutil.TempDir("", "")
//	if err != nil {
//		b.Fatal(err)
//	}
//	s := NewStorage(d, 1024)
//	bucket := s.GetBucket("users")
//	b.ResetTimer()
//	b.ReportAllocs()
//	for n := 0; n < b.N; n++ {
//		key := fmt.Sprintf("id-%v", n)
//		if err := bucket.Set(key, map[string]interface{}{
//			"name": "coleman",
//		}); err != nil {
//			b.Fatal(err)
//		}
//		_, err := bucket.Get(key)
//		if err != nil {
//			b.Fatal(err)
//		}
//	}
//	b.Cleanup(func() {
//		bucket.Close()
//		os.RemoveAll(d)
//
//	})
//}
//
//func BenchmarkGetOneKey(b *testing.B) {
//	d, err := ioutil.TempDir("", "")
//	if err != nil {
//		b.Fatal(err)
//	}
//	s := NewStorage(d, 1024)
//	bucket := s.GetBucket("users")
//	key := "usr_1"
//	if err := bucket.Set(key, map[string]interface{}{
//		"name": "coleman",
//	}); err != nil {
//		b.Fatal(err)
//	}
//	b.ResetTimer()
//	b.ReportAllocs()
//	for n := 0; n < b.N; n++ {
//		_, err := bucket.Get(key)
//		if err != nil {
//			b.Fatal(err)
//		}
//	}
//	b.Cleanup(func() {
//		bucket.Close()
//		os.RemoveAll(d)
//
//	})
//}
//
////     9525            106777 ns/op            2412 B/op         60 allocs/op
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
