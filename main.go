package main

import (
	"fmt"
	"reflect"
)

type DB struct {
	Tables map[string]Table
}

type Table struct {
	Counter   int
	Resources map[int]interface{}
}

type Resource interface {
	GetID() int
	SetID(ID int)
}

func newDB() DB {
	return DB{
		Tables: make(map[string]Table),
	}
}

func (db *DB) Insert(resource Resource) {
	model := reflect.TypeOf(resource).String()
	table := db.Tables[model]
	defer func() { db.Tables[model] = table }()

	if table.Resources == nil {
		table.Resources = make(map[int]interface{})
	}

	table.Counter++
	id := table.Counter
	table.Resources[id] = resource
	resource.SetID(id)
}

func (db *DB) Get(id int, resource interface{}) interface{} {
	model := reflect.TypeOf(resource).String()
	return db.Tables[model].Resources[id]
}

// func (db *DB) GetAll(model interface{}) interface{} {
// 	return db.Tables[reflect.TypeOf(model).String()]
// }

// func (db *DB) Update(id uint, resource interface{}) {
// 	model := reflect.TypeOf(resource).String()
// 	db.Tables[model][id] = resource
// }

type User struct {
	ID   int
	Name string
	Age  uint
}

func (self *User) GetID() int {
	return self.ID
}

func (self *User) SetID(ID int) {
	self.ID = ID
}

func main() {
	db := newDB()

	user1 := User{
		Name: "John Doe",
		Age:  42,
	}
	fmt.Println(user1)

	db.Insert(&user1)

	fmt.Println(db.Tables)
	fmt.Println(user1)

	user1.Name = "Tom"

	userX1 := db.Get(1, &User{}).(*User)
	fmt.Println(userX1)
	userX1.Name = "Hans"
	fmt.Println(userX1)

	userX2 := db.Get(1, &User{}).(*User)
	fmt.Println(userX2)
}
