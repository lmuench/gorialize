package gobdb

import (
	"log"
	"testing"
)

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

var db *DB

func before() {
	db = &DB{
		Path: "/tmp/gobdb/gobdb_test",
	}

	err := db.DeleteAll(&User{})
	if err != nil {
		log.Fatal(err)
	}
}

func TestInsertAndGet(t *testing.T) {
	before()

	newUser := &User{
		Name: "John Doe",
		Age:  42,
	}
	db.Insert(newUser)

	storedUser := &User{}
	err := db.Get(storedUser, newUser.GetID())
	if err != nil {
		t.Error(err)
	}

	if storedUser.ID != newUser.ID {
		t.Error("IDs don't equal")
	}
	if storedUser.Name != newUser.Name {
		t.Error("Names don't equal")
	}
	if storedUser.Age != newUser.Age {
		t.Error("Ages don't equal")
	}
}
