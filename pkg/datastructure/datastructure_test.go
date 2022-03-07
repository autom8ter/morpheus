package datastructure

import "testing"

func Test(t *testing.T) {
	o := NewOrderedMap()
	o.Add("1", 1)
	val, ok := o.Get("1")
	if !ok {
		t.Fatal("fail")
	}
	if val != 1 {
		t.Fatal("fail")
	}
}
