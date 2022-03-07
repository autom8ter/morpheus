package datastructure

type OrderedMap interface {
	Get(key string) (interface{}, bool)
	Range(skip int, fn func(val interface{}) bool)
	Del(key string)
	Add(key string, val interface{})
	Len() int
	Keys() []string
}

func NewOrderedMap() OrderedMap {
	return &orderedMap{
		keys:   nil,
		values: map[string]interface{}{},
	}
}

type orderedMap struct {
	keys   []string
	values map[string]interface{}
}

func (o *orderedMap) Get(key string) (interface{}, bool) {
	if val, ok := o.values[key]; ok {
		return val, true
	}
	return nil, false
}

func (o *orderedMap) Del(key string) {
	if _, ok := o.values[key]; ok {
		remove(o.keys, key)
		delete(o.values, key)
	}
}

func (o *orderedMap) Add(key string, val interface{}) {
	if val, ok := o.values[key]; ok {
		o.values[key] = val
		return
	}
	o.values[key] = val
	o.keys = append(o.keys, key)
}

func (o *orderedMap) Range(skip int, fn func(val interface{}) bool) {
	for _, key := range o.keys[skip:] {
		if !fn(o.values[key]) {
			break
		}
	}
}

func (o *orderedMap) Len() int {
	return len(o.values)
}

func (o *orderedMap) Keys() []string {
	return o.keys
}

func remove(slice []string, key string) []string {
	for i, val := range slice {
		if val == key {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}
