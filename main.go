package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 3 {
		PrintHelpText()
		return
	}

	var err error
	command := os.Args[1]
	path := os.Args[2]

	switch command {
	case "show", "s":
		if len(os.Args) < 4 {
			err = ShowAll(path)
		} else {
			filename := os.Args[3]
			err = ShowOne(path, filename)
		}
	case "generate", "g":
		if len(os.Args) < 4 {
			PrintHelpText()
		} else {
			model := os.Args[3]
			err = Generate(path, model)
		}
	default:
		PrintHelpText()
	}
	if err != nil {
		fmt.Println(err)
	}
}

func PrintHelpText() {
	fmt.Println(`
	Commands:
		generate [model path] [model name]    Generate a model
    show [table path]                     Show a table
    show [table path] [resource ID]       Show a resource
	`)
}
