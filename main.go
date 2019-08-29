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

// TODO will registering an interface help?
func (db *DB) GetAll(resource interface{}, callback func(resource interface{}, id int)) error {
	model := ModelName(resource)
	tablePath := TablePath(model)

	if _, err := os.Stat(tablePath); os.IsNotExist(err) {
		return err
	}

	files, err := ioutil.ReadDir(tablePath)
	if err != nil {
		return err
	}

	resources := make(map[int]interface{})

	for _, f := range files {
		if f.IsDir() {
			continue
		}

		b, err := ioutil.ReadFile(tablePath + "/" + f.Name())
		if err != nil {
			return err
		}
		id, err := strconv.Atoi(f.Name())
		if err != nil {
			return err
		}
		buf := bytes.NewReader(b)
		dec := gob.NewDecoder(buf)
		err = dec.Decode(resource)
		fmt.Println(id, resource)
		callback(resource, id)
		resources[id] = resource // TODO won't work since this is a pointer and will change with every iteration
		// resources = append(resources.([]interface{}), resource)
		if err != nil {
			return err
		}
	}

	// var buf bytes.Buffer
	// enc := gob.NewEncoder(&buf)
	// err = enc.Encode(resources)
	// if err != nil {
	// 	return err
	// }
	// dec := gob.NewDecoder(&buf)
	// err = dec.Decode(result)
	// if err != nil {
	// 	return err
	// }
	return nil
}

func (db *DB) Where(resource interface{}, callback func(resource interface{}) bool) error {
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
		// id, err := strconv.Atoi(f.Name())
		// if err != nil {
		// 	return err
		// }
		buf := bytes.NewReader(b)
		dec := gob.NewDecoder(buf)
		err = dec.Decode(resource)
		if err != nil {
			return err
		}
		if callback(resource) {
			return nil
		}
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
	fmt.Println(userX1)
	userX1.Age++
	db.Insert(&userX1)
	fmt.Println(userX1)

	// users := make(map[int]*User)
	// users := make([]User, 1)
	// err = db.GetAll(users)
	// if err != nil {
	// log.Fatal(err)
	// }

	// fmt.Println(users)

	// for _, user := range users {
	// 	fmt.Println(*user.(*User))
	// }
	// var userWhere User
	// err = db.Where(&userWhere, func(user interface{}) bool {
	// 	return user.(*User).Age == 43
	// })
	// if err != nil {
	// 	log.Fatal(err)
	// }

	users := make(map[int]User)

	err = db.GetAll(&User{}, func(resource interface{}, id int) {
		users[id] = *resource.(*User)
	})
	if err != nil {
		log.Fatal(err)
	}
	// users := result.ToUserMap()

	fmt.Println(users)
	fmt.Println(users[1].Age)
	fmt.Println(users[2].Age)
	fmt.Println(users[3].Age)
}

// type IMap map[int]interface{}

// func (imap IMap) ToUserMap() map[int]User {
// 	users := make(map[int]User)
// 	for k, v := range imap {
// 		// fmt.Println(k, *v.(*User))
// 		users[k] = *v.(*User)
// 	}
// 	return users
// }
