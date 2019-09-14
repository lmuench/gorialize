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

func (self TodoList) GetUser(dir *gorialize.Directory) (User, error) {
	var user User
	err := dir.Read(&user, self.UserID)
	return user, err
}

// Helpers

// SELECT * FROM TODOLISTS
func GetAllTodoLists(dir *gorialize.Directory) ([]TodoList, error) {
	todoLists := []TodoList{}

	err := dir.ReadAll(&TodoList{}, func(resource interface{}) {
		todoList := *resource.(*TodoList)
		todoLists = append(todoLists, todoList)
	})
	return todoLists, err
}
