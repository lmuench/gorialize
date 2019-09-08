package model

import "github.com/lmuench/gorialize/gorialize"

// TodoList implements gorialize.Resource interface
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

// Helpers

func (self TodoList) GetUser(db *gorialize.Directory) (User, error) {
	var user User
	err := db.Read(&user, self.UserID)
	return user, err
}

// Helpers

// SELECT * FROM TODOLISTS
func GetAllTodoLists(db *gorialize.Directory) ([]TodoList, error) {
	todoLists := []TodoList{}

	err := db.ReadAll(&TodoList{}, func(resource interface{}) {
		todoList := *resource.(*TodoList)
		todoLists = append(todoLists, todoList)
	})
	return todoLists, err
}
