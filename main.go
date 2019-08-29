package main

import (
	"fmt"
	"log"

	"github.com/lmuench/gobdb/gobdb"
)

// User implements gobdb.Resource interface
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

// Example gobdb usage
func main() {
	db := &gobdb.DB{Path: "/tmp/gobdb"}

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

// Helper functions you can define
// SELECT * FROM USERS
func GetAllUsersMap(db *gobdb.DB) (map[int]User, error) {
	users := make(map[int]User)

	err := db.GetAll(&User{}, func(resource interface{}) {
		user := *resource.(*User)
		users[user.GetID()] = user
	})
	return users, err
}

// SELECT * FROM USERS
func GetAllUsersSlice(db *gobdb.DB) ([]User, error) {
	users := []User{}

	err := db.GetAll(&User{}, func(resource interface{}) {
		user := *resource.(*User)
		users = append(users, user)
	})
	return users, err
}

// SELECT * FROM USERS WHERE AGE > 42
func GetAllUsersOver42Slice(db *gobdb.DB) ([]User, error) {
	users := []User{}

	err := db.GetAll(&User{}, func(resource interface{}) {
		user := *resource.(*User)
		if user.Age > 42 {
			users = append(users, user)
		}
	})
	return users, err
}
