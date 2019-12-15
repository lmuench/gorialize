// Package gorialize is an embedded database that stores Go structs serialized to gobs
package gorialize

import (
	"fmt"
	"errors"
)

type Index struct {
	KV map[string][]int
	VK map[string][]string
}

func NewIndex() Index {
	return Index{
		KV: map[string][]int{},
		VK: map[string][]string{},
	}
}

func (idx Index) getIDs(model string, field string, value interface{}) []int {
	key := makeKey(model, field, value)
	return idx.KV[key]
}

func (idx Index) getMatchingIDs(model string, clauses ...Where) (ids []int) {
	idMap := map[int]bool{}
	for _, clause := range clauses {
		var tmpIDs []int
		switch true {
		case len(clause.Range) > 0:
			for _, value := range clause.Range {
				tmpIDs = append(tmpIDs, idx.getIDs(model, clause.Field, value)...)
			}
		case len(clause.In) > 0:
			for _, value := range clause.In {
				tmpIDs = append(tmpIDs, idx.getIDs(model, clause.Field, value)...)
			}
		default:
			tmpIDs = idx.getIDs(model, clause.Field, clause.Equals)
		}

		if clause.And == nil {
			for _, id := range tmpIDs {
				idMap[id] = true
			}
		} else {
			tmpIDmap := map[int]int{}
			for _, id := range tmpIDs {
				tmpIDmap[id]++
			}
			idsToAnd := idx.getMatchingIDs(model, *clause.And)
			for _, id := range idsToAnd {
				tmpIDmap[id]++
			}
			for id, cnt := range tmpIDmap {
				if cnt > 1 {
					idMap[id] = true
				}
			}
		}
	}
	for id := range idMap {
		ids = append(ids, id)
	}
	return
}

func (idx Index) add(model string, field string, value interface{}, id int) {
	key := makeKey(model, field, value)
	val := makeVal(model, field, id)
	idx.KV[key] = append(idx.KV[key], id)
	idx.VK[val] = append(idx.VK[val], key)
}

func (idx Index) addDirectly(key string, id int) error {
	val, err := makeValFromKey(key, id)
	if err != nil {
		return err
	}
	idx.KV[key] = append(idx.KV[key], id)
	idx.VK[val] = append(idx.VK[val], key)
	return nil
}

func (idx Index) remove(model string, field string, id int) {
	val := makeVal(model, field, id)
	idx.removeDirectly(val, id)
}

func (idx Index) removeDirectly(val string, id int) {
	keys := idx.VK[val]
	for _, key := range keys {
		last := len(idx.KV[key])-1
		if last == 0 {
			delete(idx.KV, key)
		} else {
			for i := range idx.KV[key] {
				if idx.KV[key][i] == id {
					idx.KV[key][i] = idx.KV[key][last]
					idx.KV[key] = idx.KV[key][:last]
					break
				}
			}
		}
	}
	delete(idx.VK, val)
}

func makeKey(model string, field string, value interface{}) (key string) {
	key = fmt.Sprintf("%s:%s:%v", model, field, value)
	return
}

func makeVal(model string, field string, id int) (val string) {
	val = fmt.Sprintf("%s:%s:%d", model, field, id)
	return
}

func makeValFromKey(key string, id int) (string, error) {
	cnt := 0
	for i, c := range key {
		if c == ':' {
			cnt++
			if cnt == 2 {
				val := fmt.Sprintf("%s:%d", key[:i], id)
				return val, nil
			}
		}
	}
	return "", errors.New("Invalid key")
}
