// Package gorialize is an embedded database that stores Go structs serialized to gobs
package gorialize

import (
	"fmt"
)

type Index map[string][]int

func (idx Index) getIDs(model string, field string, value interface{}) (ids []int, ok bool) {
	key := makeKey(model, field, value)
	ids, ok = idx[key]
	return
}

func (idx Index) appendID(model string, field string, value interface{}, id int) {
	key := makeKey(model, field, value)
	idx[key] = append(idx[key], id)
}

func (idx Index) appendIDbyKey(key string, id int) {
	idx[key] = append(idx[key], id)
}

func (idx Index) removeID(model string, field string, value interface{}, id int) {
	key := makeKey(model, field, value)
	for i := range idx[key] {
		if idx[key][i] == id {
			idx[key][i] = idx[key][len(idx[key])-1]
			idx[key] = idx[key][:len(idx[key])-1]
			break
		}
	}
}

func (idx Index) removeIDbyKey(key string, id int) {
	for i := range idx[key] {
		if idx[key][i] == id {
			last := len(idx[key])-1
			idx[key][i] = idx[key][last]
			idx[key] = idx[key][:last]
			break
		}
	}
}

func makeKey(model string, field string, value interface{}) (key string) {
	key = fmt.Sprintf("%s:%s:%v", model, field, value)
	return
}
