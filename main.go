package main

import (
	"fmt"
	"log"

	"github.com/lmuench/gobdb/example"
	"github.com/lmuench/gobdb/gobdb"
)

// Example gobdb usage
func main() {
	db := &gobdb.DB{Path: "/tmp/gobdb/dev"}

	u1 := example.User{
		Name: "John Doe",
		Age:  42,
	}
	db.Insert(&u1)

	tdl1 := example.TodoList{
		OwnerID: u1.GetID(),
		Title:   "My Todo List",
	}
	db.Insert(&tdl1)

	var tdlX1 example.TodoList
	_ = db.Get(&tdlX1, tdl1.GetID())
	fmt.Println(tdlX1)

	uX1, err := tdlX1.GetOwner(db)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(uX1)

	todoLists, err := example.GetAllTodoLists(db)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(todoLists)
}
