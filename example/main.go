package main

import (
	"fmt"
	"log"

	"github.com/lmuench/gobdb/example/model"
	"github.com/lmuench/gobdb/gobdb"
)

// Example gobdb usage
func main() {
	db := &gobdb.DB{Path: "/tmp/gobdb/example_dev"}

	u1 := model.User{
		Name: "John Doe",
		Age:  42,
	}
	db.Insert(&u1)

	tdl1 := model.TodoList{
		UserID: u1.GetID(),
		Title:  "My Todo List",
	}
	db.Insert(&tdl1)

	var tdlX1 model.TodoList
	_ = db.Get(&tdlX1, tdl1.GetID())
	fmt.Println(tdlX1)

	uX1, err := tdlX1.GetUser(db)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(uX1)

	todoListsX1, err := model.GetAllTodoLists(db)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(todoListsX1)

	tdlX1.Title = "A different title"
	err = db.Update(&tdlX1)
	if err != nil {
		log.Fatal(err)
	}

	todoListsX2, err := model.GetAllTodoLists(db)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(todoListsX2)
}
