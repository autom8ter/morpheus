package persistence

import (
	"encoding/json"
)

//
//func Test(t *testing.T) {
//	dir, err := ioutil.TempDir("", "badger-test")
//	if err != nil {
//		t.Fatal(err)
//	}
//	defer os.RemoveAll(dir)
//	g, err := New(dir)
//	if err != nil {
//		t.Fatal(err)
//	}
//	coleman, err := g.AddNode("user", "colemanword@gmail.com", map[string]interface{}{
//		"name": "Coleman Word",
//	})
//	if err != nil {
//		t.Fatal(err)
//	}
//	choozle, err := g.AddNode("business", "www.choozle.com", nil)
//	if err != nil {
//		t.Fatal(err)
//	}
//	coleman.AddRelation("works_at", choozle)
//
//	found := false
//	coleman.Relations(0, "works_at", "business", func(relation api.Relation) bool {
//		found = true
//		t.Logf("relations - %s %s %s", coleman.GetProperty("name"), relation.Type(), relation.Target().ID())
//		return true
//	})
//	if !found {
//		t.Fatal("failed to find relation")
//	}
//
//	{
//		c, err := g.GetNode("user", "colemanword@gmail.com")
//		if err != nil {
//			t.Fatal(err)
//		}
//		if !reflect.DeepEqual(coleman.Properties(), c.Properties()) {
//			t.Fatal("not equal", jsonString(coleman.Properties()), jsonString(c.Properties()))
//		}
//	}
//	t.Log(g.NodeTypes())
//}

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
//		coleman.AddRelation(api.Outgoing, "works_at", "1", choozle)
//		coleman.Relations(0, api.Outgoing, "works_at", func(relation api.Relation) bool {
//			relation.Target().Relations(0, api.Incoming, "works_at", func(relation2 api.Relation) bool {
//				if relation2.Source().GetProperty("name") != "Coleman Word" {
//					b.Fatal("fail - ", relation2.Target().ID(), relation2.Source().ID())
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
