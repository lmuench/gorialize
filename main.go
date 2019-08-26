package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"reflect"
)

type DB struct {
	Tables map[string]Table
}

type Table struct {
	Counter   int
	Resources map[int]bytes.Buffer
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

func (db *DB) Insert(resource Resource) error {
	model := reflect.TypeOf(resource).String()
	table := db.Tables[model]
	defer func() { db.Tables[model] = table }()

	if table.Resources == nil {
		table.Resources = make(map[int]bytes.Buffer)
	}

	table.Counter++
	id := table.Counter
	resource.SetID(id)

	var resourceBuffer bytes.Buffer
	defer func() { table.Resources[id] = resourceBuffer }()

	enc := gob.NewEncoder(&resourceBuffer)
	err := enc.Encode(resource)
	fmt.Println(resourceBuffer)
	return err
}

func (db *DB) Get(id int, resource interface{}) error {
	model := reflect.TypeOf(resource).String()
	resourceBuffer := db.Tables[model].Resources[id]
	dec := gob.NewDecoder(&resourceBuffer)
	err := dec.Decode(resource)
	return err
}

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

	user1 := &User{
		Name: "John Doe",
		Age:  42,
	}
	fmt.Println(user1)

	_ = db.Insert(user1)

	user1.Name = "Tom"
	fmt.Println(user1)

	userX1 := &User{}
	err := db.Get(1, userX1)
	fmt.Println(userX1, err)

	userX1.Name = "Hans"
	fmt.Println(userX1)

	userX2 := &User{}
	err = db.Get(1, userX2)
	fmt.Println(userX2, err)

	userX2.Name = "Alice"
	_ = db.Insert(userX2)

	userX3 := &User{}
	err = db.Get(2, userX3)
	fmt.Println(userX3, err)
}
