package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/lmuench/gobdb/gobdb"

	"github.com/drosseau/degob"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println(`
  Commands:
    show [table path]                 Show a table
    show [table path] [resource ID]   Show a resource
		`)
		return
	}

	// exe, err := os.Executable()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// cwd := filepath.Dir(exe)

	var err error
	path := os.Args[1]
	if len(os.Args) < 3 {
		err = ShowAll(path)
	} else {
		filename := os.Args[2]
		err = ShowOne(path, filename)
	}
	if err != nil {
		fmt.Println(err)
	}
}

func ShowOne(path string, filename string) error {
	id, err := strconv.Atoi(filename)
	if err != nil {
		return err
	}
	b, err := ioutil.ReadFile(gobdb.ResourcePath(path, id))
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

func ShowAll(path string) error {
	return nil
}
