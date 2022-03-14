package persistence

import (
	"encoding/json"
	"github.com/autom8ter/morpheus/pkg/api"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

func Test(t *testing.T) {
	dir, err := ioutil.TempDir("", "badger-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	g, err := New(dir, 100)
	if err != nil {
		t.Fatal(err)
	}
	coleman, err := g.AddNode("user", "colemanword@gmail.com", map[string]interface{}{
		"name": "Coleman Word",
	})
	if err != nil {
		t.Fatal(err)
	}
	choozle, err := g.AddNode("business", "www.choozle.com", nil)
	if err != nil {
		t.Fatal(err)
	}
	coleman.AddRelationship("works_at", choozle)

	coleman.Relationships(0, "works_at", "business", func(relationship api.Relationship) bool {
		t.Logf("relationships - %s %s %s", coleman.GetProperty("name"), relationship.Type(), relationship.Target().ID())
		return true
	})

	{
		c, err := g.GetNode("user", "colemanword@gmail.com")
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(coleman, c) {
			t.Fatal("not equal")
		}
	}
	t.Log(g.NodeTypes())
}

/*
go test  -bench=Benchmark . -benchmem -run=^$

Benchmark: 170589              7558 ns/op            3840 B/op         64 allocs/op
//*/
//func Benchmark(b *testing.B) {
//	dir, err := ioutil.TempDir("", "badger-test")
//	if err != nil {
//		b.Fatal(err)
//	}
//	defer os.RemoveAll(dir)
//	g, err := New(dir, 100)
//	if err != nil {
//		b.Fatal(err)
//	}
//	b.ReportAllocs()
//	b.ResetTimer()
//
//	for n := 0; n < b.N; n++ {
//		id := fmt.Sprintf("colemanword%v@gmail.com", n)
//		coleman, err := g.AddNode("user", id, map[string]interface{}{
//			"name": "Coleman Word",
//		})
//		if err != nil {
//			b.Fatal(err)
//		}
//		choozle, err := g.AddNode("business", "www.choozle.com", map[string]interface{}{
//			"name": "Choozle",
//		})
//		if err != nil {
//			b.Fatal(err)
//		}
//		coleman.AddRelationship(api.Outgoing, "works_at", "1", choozle)
//		coleman.Relationships(0, api.Outgoing, "works_at", func(relationship api.Relationship) bool {
//			relationship.Target().Relationships(0, api.Incoming, "works_at", func(relationship2 api.Relationship) bool {
//				if relationship2.Source().GetProperty("name") != "Coleman Word" {
//					b.Fatal("fail - ", relationship2.Target().ID(), relationship2.Source().ID())
//				}
//				return false
//			})
//			return true
//		})
//	}
//}

func jsonString(value interface{}) string {
	bits, _ := json.MarshalIndent(value, "", "    ")
	return string(bits)
}
