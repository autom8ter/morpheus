package iterator

type Iterator struct {
	index int
	items []interface{}
}

func NewIterator(index int, items []interface{}) *Iterator {
	return &Iterator{index: index, items: items}
}

func (i *Iterator) HasNext() bool {
	if i.index < len(i.items) {
		return true
	}
	return false
}

func (i *Iterator) GetNext() interface{} {
	if i.HasNext() {
		user := i.items[i.index]
		i.index++
		return user
	}
	return nil
}

func (i *Iterator) Do(fn func(interface{}) bool) {
	if i.HasNext() {
		val := i.items[i.index]
		i.index++
		if !fn((val)) {
			return
		}
	}
}
