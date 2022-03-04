package inmem

import (
	"github.com/autom8ter/morpheus/pkg/api"
	"reflect"
	"testing"
)

func Test(t *testing.T) {
	g := NewGraph()
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
		t.Logf("%s %s %s", coleman.GetProperty("name"), relationship.Type(), relationship.Target().ID())
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

// go test -cpuprofile benchmark.prof -bench=Benchmark . -benchmem -run=^$
func Benchmark(b *testing.B) {
	g := NewGraph()
	coleman, err := g.AddNode("user", "colemanword@gmail.com", map[string]interface{}{
		"name": "Coleman Word",
	})
	if err != nil {
		b.Fatal(err)
	}
	choozle, err := g.AddNode("business", "www.choozle.com", nil)
	if err != nil {
		b.Fatal(err)
	}
	coleman.AddRelationship(api.Outgoing, "works_at", "1", choozle)
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		coleman.Relationships(api.Outgoing, "works_at", func(relationship api.Relationship) bool {
			relationship.Target().Relationships(api.Incoming, "works_at", func(relationship api.Relationship) bool {
				if relationship.Source().GetProperty("name") != "Coleman Word" {
					b.Fatal("fail")
				}
				return false
			})
			//b.Logf("%s %s %s", coleman.GetProperty("name"), relationship.Type(), relationship.Target().ID())
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
