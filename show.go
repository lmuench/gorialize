package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/drosseau/degob"
)

func ShowOne(tablePath string, filename string) error {
	id, err := strconv.Atoi(filename)
	if err != nil {
		return err
	}
	b, err := ioutil.ReadFile(ResourcePath(tablePath, id))
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
		b, err := ioutil.ReadFile(ResourcePath(tablePath, id))
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

func ResourcePath(tablePath string, id int) string {
	return tablePath + "/" + strconv.Itoa(id)
}
