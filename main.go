package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"strconv"
)

const basePath string = "/tmp/gobdb"

type DB struct {
	// Name string
	// Tables map[string]Table
}

// type Table struct {
// 	// Counter   int
// 	Resources map[int]bytes.Buffer
// }

type Resource interface {
	GetID() int
	SetID(ID int)
}

// type Resources map[int]Resource

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

	if _, err := os.Stat(tablePath); os.IsNotExist(err) {
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

func (db *DB) GetAll(resource interface{}) (map[int]interface{}, error) {
	model := ModelName(resource)
	tablePath := TablePath(model)

	if _, err := os.Stat(tablePath); os.IsNotExist(err) {
		return nil, err
	}

	files, err := ioutil.ReadDir(tablePath)
	if err != nil {
		return nil, err
	}

	resources := make(map[int]interface{})

	for _, f := range files {
		if f.IsDir() {
			continue
		}

		b, err := ioutil.ReadFile(tablePath + "/" + f.Name())
		if err != nil {
			return nil, err
		}
		id, err := strconv.Atoi(f.Name())
		if err != nil {
			return nil, err
		}
		buf := bytes.NewReader(b)
		dec := gob.NewDecoder(buf)
		err = dec.Decode(resource)
		resources[id] = resource
		fmt.Println(id)
		if err != nil {
			return nil, err
		}
	}
	return resources, nil
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
	db := &DB{}

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

	users, err := db.GetAll(&User{})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(users)

	for _, user := range users {
		fmt.Println(*user.(*User))
	}
}
