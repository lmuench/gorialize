package model

import "github.com/lmuench/gobdb/gobdb"

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

// SELECT * FROM TODOLISTS
func GetAllTodoLists(db *gobdb.DB) ([]TodoList, error) {
	todoLists := []TodoList{}

	err := db.GetAll(&TodoList{}, func(resource interface{}) {
		todoList := *resource.(*TodoList)
		todoLists = append(todoLists, todoList)
	})
	return todoLists, err
}
