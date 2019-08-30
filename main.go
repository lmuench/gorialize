package main

import (
	"fmt"
	"log"

	"github.com/jinzhu/copier"
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
	ID    int
	Owner *User
	Title string
}

func (self *TodoList) GetID() int {
	return self.ID
}

func (self *TodoList) SetID(ID int) {
	self.ID = ID
}

// Example gobdb usage
func main() {
	db := &gobdb.DB{Path: "/tmp/gobdb"}

	u1 := User{
		Name: "John Doe",
		Age:  42,
	}
	tdl1 := TodoList{
		Owner: &u1,
		Title: "My Todo List",
	}
	u1.Age++
	db.Insert(&tdl1)

	u2 := User{
		Name: "Steve Miller",
		Age:  26,
	}
	tdl2 := TodoList{
		Owner: &u2,
		Title: "Stuff I need to do",
	}
	u2.Age++
	db.Insert(&tdl2)

	var tdlX1 TodoList
	_ = db.Get(&tdlX1, 1)
	fmt.Println(tdlX1.Owner)
	var tdlX2 TodoList
	_ = db.Get(&tdlX2, 2)
	fmt.Println(tdlX2.Owner)

	todoLists, err := GetAllTodoLists(db)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(todoLists)
	for _, tdl := range todoLists {
		fmt.Println(tdl.Owner)
	}
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

func GetAllTodoLists(db *gobdb.DB) ([]TodoList, error) {
	todoLists := []TodoList{}

	err := db.GetAll(&TodoList{}, func(resource interface{}) {
		var todoList TodoList
		err := copier.Copy(&todoList, resource)
		if err != nil {
			log.Fatal(err)
		}
		todoLists = append(todoLists, todoList)
	})
	return todoLists, err
}

// func GetAllTodoLists(db *gobdb.DB) ([]TodoList, error) {
// 	todoLists := []TodoList{}

// 	err := db.GetAll(&TodoList{}, func(resource interface{}) {
// 		todoList := *resource.(*TodoList)
// 		owner := *todoList.Owner
// 		todoList.Owner = &owner
// 		todoLists = append(todoLists, todoList)
// 	})
// 	return todoLists, err
// }
