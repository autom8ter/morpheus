package storage

import (
	"io/ioutil"
	"os"
	"testing"
)

func Test(t *testing.T) {
	d, err := ioutil.TempDir("temp", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(d)
	s := NewStorage(1024)
	b := s.GetBucket("users")
	if err := b.Set("usr_1", map[string]interface{}{
		"name": "coleman",
	}); err != nil {
		t.Fatal(err)
	}
	if err := b.Set("usr_2", map[string]interface{}{
		"name": "lacee",
	}); err != nil {
		t.Fatal(err)
	}
	val, err := b.Get("usr_2")
	if err != nil {
		t.Fatal(err)
	}
	if val == nil {
		t.Fatal("empty value")
	}
	if val["name"] != "lacee" {
		t.Fatal("bad record")
	}
}
