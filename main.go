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

type TodoList struct {
	ID     int
	UserID int
	Title  string
}

func (self *TodoList) GetID() int {
	return self.ID
}

func (self *TodoList) SetID(ID int) {
	self.ID = ID
}

// Example gobdb usage
func main() {
	db := &gobdb.DB{Path: "/tmp/gobdb/dev"}

	u1 := User{
		Name: "John Doe",
		Age:  42,
	}

	db.Insert(&u1)

	tdl1 := TodoList{
		UserID: u1.GetID(),
		Title:  "My Todo List",
	}
	db.Insert(&tdl1)

	var tdlX1 TodoList
	_ = db.Get(&tdlX1, tdl1.GetID())
	fmt.Println(tdlX1)

	var uX1 User
	_ = db.Get(&uX1, tdlX1.UserID)
	fmt.Println(uX1)

	todoLists, err := GetAllTodoLists(db)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(todoLists)
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
func GetAllUsers(db *gobdb.DB) ([]User, error) {
	users := []User{}

	err := db.GetAll(&User{}, func(resource interface{}) {
		user := *resource.(*User)
		users = append(users, user)
	})
	return users, err
}

// SELECT * FROM USERS WHERE AGE > 42
func GetAllUsersOver42(db *gobdb.DB) ([]User, error) {
	users := []User{}

	err := db.GetAll(&User{}, func(resource interface{}) {
		user := *resource.(*User)
		if user.Age > 42 {
			users = append(users, user)
		}
	})
	return users, err
}

// SELECT * FROM TODOLISTS
func GetAllTodoLists(db *gobdb.DB) ([]TodoList, error) {
	todoLists := []TodoList{}

	err := db.GetAll(&TodoList{}, func(resource interface{}) {
		todoList := *resource.(*TodoList)
		todoLists = append(todoLists, todoList)
	})
	return todoLists, err
}
