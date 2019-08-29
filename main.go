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
}

type Resource interface {
	GetID() int
	SetID(ID int)
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

	var counter int
	b, err := ioutil.ReadFile(metadataPath + "/counter")
	if err == nil {
		counter, err = strconv.Atoi(string(b))
		if err != nil {
			log.Fatal(err)
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
		log.Fatal(err)
	}

	err = ioutil.WriteFile(tablePath+"/"+strconv.Itoa(id), buf.Bytes(), 0644)
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile(metadataPath+"/"+"counter", []byte(strconv.Itoa(counter)), 0644)
	if err != nil {
		log.Fatal(err)
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
		return err
	}

	buf := bytes.NewReader(b)
	dec := gob.NewDecoder(buf)
	err = dec.Decode(resource)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) GetAll(resource interface{}, callback func(resource interface{})) error {
	model := ModelName(resource)
	tablePath := TablePath(model)

	if _, err := os.Stat(tablePath); os.IsNotExist(err) {
		return err
	}

	files, err := ioutil.ReadDir(tablePath)
	if err != nil {
		return err
	}

	for _, f := range files {
		if f.IsDir() {
			continue
		}

		b, err := ioutil.ReadFile(tablePath + "/" + f.Name())
		if err != nil {
			return err
		}

		buf := bytes.NewReader(b)
		dec := gob.NewDecoder(buf)
		err = dec.Decode(resource)
		if err != nil {
			return err
		}
		callback(resource)
	}

	return nil
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
	userX1.Age++
	db.Insert(&userX1)

	usersMap, err := GetAllUsersMap(db)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(usersMap)

	usersSlice, err := GetAllUsersSlice(db)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(usersSlice)

	usersOver42Slice, err := GetAllUsersOver42Slice(db)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(usersOver42Slice)
}

func GetAllUsersMap(db *DB) (map[int]User, error) {
	users := make(map[int]User)

	err := db.GetAll(&User{}, func(resource interface{}) {
		user := *resource.(*User)
		users[user.GetID()] = user
	})
	if err != nil {
		return nil, err
	}
	return users, err
}

func GetAllUsersSlice(db *DB) ([]User, error) {
	users := []User{}

	err := db.GetAll(&User{}, func(resource interface{}) {
		user := *resource.(*User)
		users = append(users, user)
	})
	if err != nil {
		return nil, err
	}
	return users, err
}

func GetAllUsersOver42Slice(db *DB) ([]User, error) {
	users := []User{}

	err := db.GetAll(&User{}, func(resource interface{}) {
		user := *resource.(*User)
		if user.Age > 42 {
			users = append(users, user)
		}
	})
	return users, err
}
