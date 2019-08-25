package main

import (
	"fmt"
	"reflect"
)

type DB struct {
	Tables map[string]Table
}

func newDB() DB {
	return DB{
		Tables: make(map[string]Table),
	}
}

type Table []interface{}

func (db *DB) Insert(resource interface{}) uint {
	model := reflect.TypeOf(resource).String()
	db.Tables[model] = append(db.Tables[model], resource)
	return uint(len(db.Tables[model]))
}

func (db *DB) Get(id uint, model interface{}) interface{} {
	return db.Tables[reflect.TypeOf(model).String()][id]
}

func (db *DB) Update(id uint, resource interface{}) {
	model := reflect.TypeOf(resource).String()
	db.Tables[model][id] = resource
}

type User struct {
	Name string
	Age  uint
}

func main() {
	db := newDB()

	user1 := User{
		Name: "John Doe",
		Age:  42,
	}
	db.Insert(user1)

	fmt.Println(db.Tables)

	userX := db.Get(0, User{}).(User)

	userX.Age = 43
	fmt.Println(user1)
	fmt.Println(userX)

	db.Update(0, userX)
	fmt.Println(db.Tables)
	fmt.Println(db.Get(0, User{}))
}
