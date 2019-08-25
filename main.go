package main

import (
	"fmt"
	"reflect"
)

type DB struct {
	Tables map[string]Table
}

type Table struct {
	Counter int
	Rows    map[int]Resource
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

func (db *DB) Insert(resource Resource) Resource {
	model := reflect.TypeOf(resource).String()
	table := db.Tables[model]
	if table.Rows == nil {
		table.Rows = make(map[int]Resource)
	}
	table.Counter++
	table.Rows[table.Counter] = resource
	resource.SetID(table.Counter)
	db.Tables[model] = table
	return resource
}

// func (db *DB) Get(id uint, model interface{}) interface{} {
// 	return db.Tables[reflect.TypeOf(model).String()][id]
// }

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
	db.Insert(&user1)

	fmt.Println(db.Tables)

	// fmt.Println(db.GetAll(User{}))

	// userX := db.Get(0, User{}).(User)

	// userX.Age = 43
	// fmt.Println(user1)
	// fmt.Println(userX)

	// db.Update(0, userX)
	// fmt.Println(db.GetAll(User{}))
	// fmt.Println(db.Get(0, User{}))
}
