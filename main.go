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
)

const basePath string = "/tmp/gobdb"

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
	// Restore(db)
	// DumpOnShutdown(db)
	return db
}

func Restore(db DB) {
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		log.Fatal(err)
	}

	fmt.Println("Restoring DB from disk:")
	filenames, err := ioutil.ReadDir(basePath)
	if err != nil {
		log.Fatal(err)
	}

	for _, filename := range filenames {
		b, err := ioutil.ReadFile(basePath + filename.Name())
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
		fmt.Println(resource)
	}
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
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		err := os.Mkdir(basePath, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
	}

	fmt.Println("Dumping DB to disk...")

	for model, table := range db.Tables {
		err := ioutil.WriteFile(basePath+model+"."+"counter", []byte(strconv.Itoa(table.Counter)), 0644)
		if err != nil {
			panic(err)
		}
		for id, resource := range table.Resources {
			err := ioutil.WriteFile(basePath+model+"."+strconv.Itoa(id)+".gob", resource.Bytes(), 0644)
			if err != nil {
				panic(err)
			}
		}
	}
}

func TablePath(model string) string {
	return basePath + "/" + model
}

func TableMetadataPath(tablePath string) string {
	return tablePath + "/metadata"
}

func ResourcePath(tablePath string, id int) string {
	return tablePath + "/" + strconv.Itoa(id)
}

func ModelName(resource interface{}) string {
	return reflect.TypeOf(resource).String()[1:]
}

func (db *DB) Insert(resource Resource) {
	model := ModelName(resource)
	tablePath := TablePath(model)
	metadataPath := TableMetadataPath(tablePath)

	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		err = os.MkdirAll(metadataPath, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
	}

	// table := db.Tables[model]
	// defer func() { db.Tables[model] = table }()

	// if table.Resources == nil {
	// table.Resources = make(map[int]bytes.Buffer)
	// }

	// table.Counter++

	// id := table.Counter

	var counter int
	b, err := ioutil.ReadFile(metadataPath + "/counter")
	if err == nil {
		counter, err = strconv.Atoi(string(b))
		if err != nil {
			panic(err)
		}
		counter++
	} else {
		counter = 1
	}
	id := counter
	resource.SetID(id)

	var buf bytes.Buffer
	// defer func() { table.Resources[id] = buf }()

	enc := gob.NewEncoder(&buf)
	err = enc.Encode(resource)
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile(tablePath+"/"+strconv.Itoa(id), buf.Bytes(), 0644)
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile(metadataPath+"/"+"counter", []byte(strconv.Itoa(counter)), 0644)
	if err != nil {
		panic(err)
	}
}

func (db *DB) Get(resource interface{}, id int) error {
	model := ModelName(resource)
	tablePath := TablePath(model)
	metadataPath := TableMetadataPath(tablePath)

	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		return err
	}

	b, err := ioutil.ReadFile(ResourcePath(tablePath, id))
	if err != nil {
		panic(err)
	}

	buf := bytes.NewReader(b)
	dec := gob.NewDecoder(buf)
	err = dec.Decode(resource)
	if err != nil {
		return err
	}
	return nil
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

	user1 := User{
		Name: "John Doe",
		Age:  42,
	}
	db.Insert(&user1)

	var userX1 User
	err := db.Get(&userX1, 1)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(userX1)
	userX1.Age++
	db.Insert(&userX1)

	// time.Sleep(time.Second * 3)
}
