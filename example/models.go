// Example modeles
package example

import "github.com/lmuench/gobdb/gobdb"

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

// TodoList implements gobdb.Resource interface
type TodoList struct {
	ID      int
	OwnerID int
	Title   string
}

func (self *TodoList) GetID() int {
	return self.ID
}

func (self *TodoList) SetID(ID int) {
	self.ID = ID
}

// Helpers

func (self TodoList) GetOwner(db *gobdb.DB) (User, error) {
	var user User
	err := db.Get(&user, self.OwnerID)
	return user, err
}

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
