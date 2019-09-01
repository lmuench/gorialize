package model

import "github.com/lmuench/gobdb/gobdb"

type Item struct {
	ID int
	// Add attributes here
}

func (self *Item) GetID() int {
	return self.ID
}

func (self *Item) SetID(ID int) {
	self.ID = ID
}

func GetAllItems(db *gobdb.DB) ([]Item, error) {
	items := []Item{}

	err := db.GetAll(&Item{}, func(resource interface{}) {
		item := *resource.(*Item)
		items = append(items, item)
	})
	return items, err
}

func GetAllItemsMap(db *gobdb.DB) (map[int]Item, error) {
	items := make(map[int]Item)

	err := db.GetAll(&Item{}, func(resource interface{}) {
		item := *resource.(*Item)
		items[item.GetID()] = item
	})
	return items, err
}
