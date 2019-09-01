package main

import (
	"fmt"
	"os"
)

func main() {
	argCnt := len(os.Args) - 1
	if argCnt < 1 {
		PrintHelpText()
		return
	}
	args := os.Args[1:]

	var err error
	command := args[0]
	path := args[1]

	switch command {
	case "show", "s":
		if argCnt < 3 {
			err = ShowAll(path)
		} else {
			filename := args[2]
			err = ShowOne(path, filename)
		}
	case "generate", "g":
		if argCnt < 3 || argCnt == 4 {
			PrintHelpText()
		}
		model := args[2]
		if argCnt > 4 {
			switch args[3] {
			case "referencing", "references", "ref", "belongs_to":
				owner := args[4]
				err = GenerateWithOwner(path, model, owner)
			default:
				PrintHelpText()
			}
		} else {
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
    generate [model path] [model]                     Generate a model
    generate [model path] [model] referencing [owner] Generate a model belonging to another model
    show [table path]                                 Show a table
    show [table path] [resource ID]                   Show a resource
	`)
}
