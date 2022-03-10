package custom

import (
	"github.com/autom8ter/morpheus/pkg/api"
	"io/ioutil"
	"os"
	"testing"
)

/*
go test  -bench=Benchmark . -benchmem -run=^$

goos: darwin
goarch: amd64
pkg: github.com/autom8ter/morpheus/pkg/backends/inmem
cpu: Intel(R) Core(TM) i9-9880H CPU @ 2.30GHz
Benchmark-16              520010              2241 ns/op             104 B/op          7 allocs/op
*/
func Benchmark(b *testing.B) {
	dir, err := ioutil.TempDir("", "badger-test")
	if err != nil {
		panic(err)
	}
	g := NewGraph(dir)

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
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
		coleman.Relationships(0, api.Outgoing, "works_at", func(relationship api.Relationship) bool {
			relationship.Target().Relationships(0, api.Incoming, "works_at", func(relationship2 api.Relationship) bool {
				if relationship2.Source().GetProperty("name") != "Coleman Word" {
					b.Fatal("fail - ", relationship2.Target().ID(), relationship2.Source().ID())
				}
				return false
			})
			return true
		})

		{
			_, err := g.GetNode("user", "colemanword@gmail.com")
			if err != nil {
				b.Fatal(err)
			}
			//if !reflect.DeepEqual(coleman, c) {
			//	b.Fatal("not equal")
			//}
		}
	}
	b.Cleanup(func() {
		g.Close()
		os.RemoveAll(dir)
	})
}

