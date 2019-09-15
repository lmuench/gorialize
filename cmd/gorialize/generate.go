package main

import (
	"fmt"
	"os"
	"strings"
	"text/template"
)

type modelTemplateData struct {
	Package  string
	Model    string
	ModelVar string
	Owner    string
	OwnerVar string
}

func Generate(path string, model string) error {
	var d modelTemplateData
	substrings := strings.Split(path, "/")
	d.Package = strings.ToLower(substrings[len(substrings)-1])
	d.ModelVar = unsnake(model)
	d.Model = camelize(model)

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

	filepath := path + "/" + d.ModelVar + ".go"
	if _, err := os.Stat(filepath); err == nil {
		fmt.Printf("%s already exists. Do you want to overwrite it? (y/N) ", filepath)
		var answer string
		fmt.Scanln(&answer)
		if answer != "y" && answer != "Y" {
			fmt.Println("Aborted.")
			return nil
		}
	}

	f, err := os.Create(filepath)
	if err != nil {
		return err
	}

	err = t.Execute(f, d)
	if err == nil {
		fmt.Printf("Generated %s in %s\n", d.Model, filepath)
	}
	return err
}

func GenerateWithOwner(path string, model string, owner string) error {
	var d modelTemplateData
	substrings := strings.Split(path, "/")
	d.Package = strings.ToLower(substrings[len(substrings)-1])
	d.ModelVar = unsnake(model)
	d.Model = camelize(model)
	d.OwnerVar = unsnake(owner)
	d.Owner = camelize(owner)

	t, err := template.New("model with owner").Parse(modelWithOwnerTemplate)
	if err != nil {
		return err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return err
		}
	}

	filepath := path + "/" + d.ModelVar + ".go"
	if _, err := os.Stat(filepath); err == nil {
		fmt.Printf("%s already exists. Do you want to overwrite it? (y/N) ", filepath)
		var answer string
		fmt.Scanln(&answer)
		if answer != "y" && answer != "Y" {
			fmt.Println("Aborted.")
			return nil
		}
	}

	f, err := os.Create(filepath)
	if err != nil {
		return err
	}

	err = t.Execute(f, d)
	if err == nil {
		fmt.Printf("Generated %s in %s\n--------- %s belongs to %s\n", d.Model, filepath, d.Model, d.Owner)
	}
	return err
}

func camelize(model string) string {
	subs := strings.Split(model, "_")
	if len(subs) > 1 {
		for i := 0; i < len(subs); i++ {
			subs[i] = strings.ToLower(subs[i])
			subs[i] = strings.Title(subs[i])
		}
		return strings.Join(subs, "")
	}
	return strings.Title(model)
}

func unsnake(model string) string {
	subs := strings.Split(model, "_")
	joined := strings.Join(subs, "")
	return strings.ToLower(joined)
}

var modelTemplate = `package {{.Package}}

import "github.com/lmuench/gorialize/gorialize"

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

func GetAll{{.Model}}s(dir *gorialize.Directory) ([]{{.Model}}, error) {
	{{.ModelVar}}s := []{{.Model}}{}

	err := dir.GetAll(&{{.Model}}{}, func(resource interface{}) {
		{{.ModelVar}} := *resource.(*{{.Model}})
		{{.ModelVar}}s = append({{.ModelVar}}s, {{.ModelVar}})
	})
	return {{.ModelVar}}s, err
}

func GetAll{{.Model}}sMap(dir *gorialize.Directory) (map[int]{{.Model}}, error) {
	{{.ModelVar}}s := make(map[int]{{.Model}})

	err := dir.GetAll(&{{.Model}}{}, func(resource interface{}) {
		{{.ModelVar}} := *resource.(*{{.Model}})
		{{.ModelVar}}s[{{.ModelVar}}.GetID()] = {{.ModelVar}}
	})
	return {{.ModelVar}}s, err
}
`

var modelWithOwnerTemplate = `package {{.Package}}

import "github.com/lmuench/gorialize/gorialize"

type {{.Model}} struct {
	ID int
	{{.Owner}}ID int
	// Add attributes here
}

func (self *{{.Model}}) GetID() int {
	return self.ID
}

func (self *{{.Model}}) SetID(ID int) {
	self.ID = ID
}

func (self {{.Model}}) Get{{.Owner}}(dir *gorialize.Directory) ({{.Owner}}, error) {
	var {{.OwnerVar}} {{.Owner}}
	err := dir.Get(&{{.OwnerVar}}, self.{{.Owner}}ID)
	return {{.OwnerVar}}, err
}

func GetAll{{.Model}}s(dir *gorialize.Directory) ([]{{.Model}}, error) {
	{{.ModelVar}}s := []{{.Model}}{}

	err := dir.GetAll(&{{.Model}}{}, func(resource interface{}) {
		{{.ModelVar}} := *resource.(*{{.Model}})
		{{.ModelVar}}s = append({{.ModelVar}}s, {{.ModelVar}})
	})
	return {{.ModelVar}}s, err
}

func GetAll{{.Model}}sMap(dir *gorialize.Directory) (map[int]{{.Model}}, error) {
	{{.ModelVar}}s := make(map[int]{{.Model}})

	err := dir.GetAll(&{{.Model}}{}, func(resource interface{}) {
		{{.ModelVar}} := *resource.(*{{.Model}})
		{{.ModelVar}}s[{{.ModelVar}}.GetID()] = {{.ModelVar}}
	})
	return {{.ModelVar}}s, err
}
`
