// Package gorialize is an embedded database that stores Go structs serialized to gobs
package gorialize

import (
	"fmt"
)

type Index struct {
	IDs  map[string][]int
	Keys map[int][]string
}

func NewIndex() Index {
	return Index{
		IDs: map[string][]int{},
		Keys: map[int][]string{},
	}
}

func (idx Index) getIDs(model string, field string, value interface{}) (ids []int, ok bool) {
	key := makeKey(model, field, value)
	ids, ok = idx.IDs[key]
	return
}

func (idx Index) add(model string, field string, value interface{}, id int) {
	key := makeKey(model, field, value)
	idx.IDs[key] = append(idx.IDs[key], id)
	idx.Keys[id] = append(idx.Keys[id], key)
}

func (idx Index) setDirectly(key string, id int) {
	idx.IDs[key] = append(idx.IDs[key], id)
	idx.Keys[id] = append(idx.Keys[id], key)
}

func (idx Index) remove(id int) {
	keys := idx.Keys[id]
	for _, key := range keys {
		for i := range idx.IDs[key] {
			last := len(idx.IDs[key])-1
			if idx.IDs[key][i] == id {
				idx.IDs[key][i] = idx.IDs[key][last]
				idx.IDs[key] = idx.IDs[key][:last]
				break
			}
		}
	}
	delete(idx.Keys, id)
}

func makeKey(model string, field string, value interface{}) (key string) {
	key = fmt.Sprintf("%s:%s:%v", model, field, value)
	return
}
