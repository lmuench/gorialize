package main

import (
	"os"
	"strings"
	"text/template"
)

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
	f, err := os.Create(path + "/" + d.Var + ".go")
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
