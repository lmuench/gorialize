package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"text/template"

	"github.com/lmuench/gobdb/gobdb"

	"github.com/drosseau/degob"
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

type modelTemplateData struct {
	Package string
	Model   string
	Var     string
}

func Generate(path string, model string) error {
	var d modelTemplateData
	substrings := strings.Split(path, "/")
	d.Package = strings.ToLower(substrings[len(substrings)-1])
	d.Var = strings.ToLower(model)
	d.Model = strings.Title(d.Var)

	t, err := template.New("model").Parse(modelTemplate)
	if err != nil {
		return err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return err
		}
	}
	f, err := os.Create(path + "/" + d.Model + ".go")
	if err != nil {
		return err
	}

	err = t.Execute(f, d)
	return err
}

var modelTemplate = `package {{.Package}}

import "github.com/lmuench/gobdb/gobdb"

type {{.Model}} struct {
	ID int
	// Add attributes here
}

func (self *{{.Model}}) GetID() int {
	return self.ID
}

func (self *{{.Model}}) SetID(ID int) {
	self.ID = ID
}

func GetAll{{.Model}}s(db *gobdb.DB) ([]{{.Model}}, error) {
	{{.Var}}s := []{{.Model}}{}

	err := db.GetAll(&{{.Model}}{}, func(resource interface{}) {
		{{.Var}} := *resource.(*{{.Model}})
		{{.Var}}s = append({{.Var}}s, {{.Var}})
	})
	return {{.Var}}s, err
}

func GetAll{{.Model}}sMap(db *gobdb.DB) (map[int]{{.Model}}, error) {
	{{.Var}}s := make(map[int]{{.Model}})

	err := db.GetAll(&{{.Model}}{}, func(resource interface{}) {
		{{.Var}} := *resource.(*{{.Model}})
		{{.Var}}s[{{.Var}}.GetID()] = {{.Var}}
	})
	return {{.Var}}s, err
}
`

func ShowOne(tablePath string, filename string) error {
	id, err := strconv.Atoi(filename)
	if err != nil {
		return err
	}
	b, err := ioutil.ReadFile(gobdb.ResourcePath(tablePath, id))
	if err != nil {
		return err
	}

	buf := bytes.NewReader(b)
	dec := degob.NewDecoder(buf)
	gobs, err := dec.Decode()
	if err != nil {
		return err
	}

	for _, g := range gobs {
		err = g.WriteValue(os.Stdout, degob.SingleLine)
		if err != nil {
			return err
		}
	}
	return nil
}

func ShowAll(tablePath string) error {
	files, err := ioutil.ReadDir(tablePath)
	if err != nil {
		return err
	}

	for _, f := range files {
		if f.IsDir() {
			continue
		}

		id, err := strconv.Atoi(f.Name())
		if err != nil {
			continue
		}
		b, err := ioutil.ReadFile(gobdb.ResourcePath(tablePath, id))
		if err != nil {
			return err
		}

		buf := bytes.NewReader(b)
		dec := degob.NewDecoder(buf)
		gobs, err := dec.Decode()
		if err != nil {
			return err
		}

		for _, g := range gobs {
			err = g.WriteValue(os.Stdout, degob.SingleLine)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
