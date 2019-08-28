package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"reflect"
	"strconv"
	"time"
)

const path string = "/tmp/gobdb/"

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
	db := DB{
		Tables: make(map[string]Table),
	}
	Restore(db)
	DumpOnShutdown(db)
	return db
}

func Restore(db DB) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Fatal(err)
	}

	fmt.Println("Restoring DB from disk")
	filenames, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}

	for _, filename := range filenames {
		b, err := ioutil.ReadFile(path + filename.Name())
		if err != nil {
			panic(err)
		}
		var resource User
		buffer := bytes.NewReader(b)
		dec := gob.NewDecoder(buffer)
		err = dec.Decode(&resource)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("---")
		fmt.Println(resource)
		fmt.Println("---")
	}

	fmt.Println("###")
	fmt.Println(db.Tables)
	fmt.Println("###")
}

func DumpOnShutdown(db DB) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		Dump(db)
	}()
}

func Dump(db DB) {
	// TODO dump table.Counter to file

	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.Mkdir(path, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
	}

	fmt.Println("Dumping DB to disk:")
	fmt.Println(db.Tables)

	for model, table := range db.Tables {
		for id, resource := range table.Resources {
			err := ioutil.WriteFile(path+model+"."+strconv.Itoa(id)+".gob", resource.Bytes(), 0644)
			if err != nil {
				panic(err)
			}
		}
	}
}

func (db *DB) Insert(resource Resource) error {
	model := reflect.TypeOf(resource).String()[1:]
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
	userX2.Age = 34
	_ = db.Insert(userX2)
	_ = db.Insert(userX2)

	fmt.Println(db.Tables)
	time.Sleep(time.Second * 3)

	userX3 := &User{}
	err = db.Get(2, userX3)
	fmt.Println(userX3, err)
}
