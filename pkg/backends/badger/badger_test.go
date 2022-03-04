package badger

import (
	"github.com/autom8ter/morpheus/pkg/api"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

func Test(t *testing.T) {
	dir, err := ioutil.TempDir("", "badger-test")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)
	g := NewGraph(dir)
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
	coleman.AddRelationship(api.Outgoing, "works_at", "1", choozle)

	coleman.Relationships(api.Outgoing, "works_at", func(relationship api.Relationship) bool {
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

goos: darwin
goarch: amd64
pkg: github.com/autom8ter/morpheus/pkg/backends/inmem
cpu: Intel(R) Core(TM) i9-9880H CPU @ 2.30GHz
Benchmark-16             3159924               373.7 ns/op            32 B/op          2 allocs/op
*/
func Benchmark(b *testing.B) {
	dir, err := ioutil.TempDir("", "badger-test")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)
	g := NewGraph(dir)
	coleman, err := g.AddNode("user", "colemanword@gmail.com", map[string]interface{}{
		"name": "Coleman Word",
	})
	if err != nil {
		b.Fatal(err)
	}
	choozle, err := g.AddNode("business", "www.choozle.com", map[string]interface{}{
		"name": "Choozle",
	})
	if err != nil {
		b.Fatal(err)
	}
	coleman.AddRelationship(api.Outgoing, "works_at", "1", choozle)
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		coleman.Relationships(api.Outgoing, "works_at", func(relationship api.Relationship) bool {
			relationship.Target().Relationships(api.Incoming, "works_at", func(relationship2 api.Relationship) bool {
				if relationship2.Source().GetProperty("name") != "Coleman Word" {
					b.Fatal("fail - ", relationship2.Target().ID(), relationship2.Source().ID())
				}
				return false
			})
			return true
		})

		{
			c, err := g.GetNode("user", "colemanword@gmail.com")
			if err != nil {
				b.Fatal(err)
			}
			if !reflect.DeepEqual(coleman, c) {
				b.Fatal("not equal")
			}
		}
	}
}
