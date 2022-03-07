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
		values: []interface{}{},
		index:  map[string]int{},
	}
}

type orderedMap struct {
	values []interface{}
	index  map[string]int
}

func (o *orderedMap) Get(key string) (interface{}, bool) {
	if _, ok := o.index[key]; !ok {
		return nil, false
	}
	return o.values[o.index[key]], true
}

func (o *orderedMap) Del(key string) {
	if index, ok := o.index[key]; ok {
		remove(o.values, index)
		delete(o.index, key)
		return
	}
}

func (o *orderedMap) Add(key string, val interface{}) {
	if index, ok := o.index[key]; ok {
		o.values[index] = val
		return
	}
	index := len(o.values)
	o.index[key] = index
	o.values = append(o.values, val)
}

func (o *orderedMap) Range(skip int, fn func(val interface{}) bool) {
	for _, val := range o.values[skip:] {
		if !fn(val) {
			break
		}
	}
}

func (o *orderedMap) Len() int {
	return len(o.values)
}

func (o *orderedMap) Keys() []string {
	var keys []string
	for k, _ := range o.index {
		keys = append(keys, k)
	}
	return keys
}

func remove(slice []interface{}, i int) []interface{} {
	return append(slice[:i], slice[i+1:]...)
}
