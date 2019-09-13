package main

import (
	"fmt"
	"os"

	"github.com/lmuench/gorialize/gorialize"
)

func main() {
	argCnt := len(os.Args) - 1
	if argCnt < 2 {
		PrintHelpText()
		return
	}
	args := os.Args[1:]

	var err error
	command := args[0]
	path := args[1]

	switch command {
	case "generate", "g":
		err = HandleGenerateCommand(command, path, args, argCnt)
	case "show", "s":
		err = HandleShowCommand(command, path, args, argCnt)
	default:
		PrintHelpText()
	}
	if err != nil {
		fmt.Println(err)
	}
}

func HandleGenerateCommand(command string, path string, args []string, argCnt int) error {
	if argCnt < 3 {
		PrintHelpText()
	}
	model := args[2]

	if argCnt == 3 {
		return Generate(path, model)
	}

	if argCnt == 4 {
		PrintHelpText()
	}

	switch args[3] {
	case "referencing", "references", "ref", "belongs_to":
		owner := args[4]
		return GenerateWithOwner(path, model, owner)
	default:
		PrintHelpText()
	}
	return nil
}

func HandleShowCommand(command string, path string, args []string, argCnt int) error {
	if argCnt < 3 {
		return gorialize.ShowAll(path)
	}
	filename := args[2]
	return gorialize.ShowOne(path, filename)
}

func PrintHelpText() {
	fmt.Println(`
  Commands:
    generate [model path] [model]                     Generate a model
    generate [model path] [model] referencing [owner] Generate a model belonging to another model
    show [directory path]                             Show a directory's resources
    show [directory path] [resource ID]               Show a single resource
	`)
	os.Exit(1)
}
