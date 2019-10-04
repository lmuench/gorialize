package main

import (
	"fmt"
	"log"

	"github.com/lmuench/gorialize/example/model"
	"github.com/lmuench/gorialize/gorialize"
)

// Example gorialize usage
func main() {
	dir := gorialize.NewDirectory(gorialize.DirectoryConfig{
		Path:       "/tmp/gorialize/example_dev",
		Encrypted:  true,
		Passphrase: "my secret passphrase",
		Log:        true,
	})

	u1 := model.User{
		Name: "John Doe",
		Age:  42,
	}
	err := dir.Create(&u1)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(u1)

	tdl1 := model.TodoList{
		UserID: u1.GetID(),
		Title:  "My Todo List",
	}
	err = dir.Create(&tdl1)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(tdl1)

	var tdlX1 model.TodoList
	_ = dir.Read(&tdlX1, tdl1.GetID())
	fmt.Println(tdlX1)

	uX1, err := tdlX1.GetUser(dir)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(uX1)

	todoListsX1, err := model.GetAllTodoLists(dir)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(todoListsX1)

	tdlX1.Title = "A different title"
	err = dir.Replace(&tdlX1)
	if err != nil {
		log.Fatal(err)
	}

	todoListsX2, err := model.GetAllTodoLists(dir)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(todoListsX2)

	// for _, tdl := range todoListsX2 {
	// 	err := dir.Delete(&tdl)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// }

	// err = dir.DeleteAll(&model.User{})
	// if err != nil {
	// 	log.Fatal(err)
	// }
}
