package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/lmuench/gobdb/gobdb"

	"github.com/drosseau/degob"
)

func ShowOne(tablePath string, filename string) error {
	passphrase := os.Getenv("GOBDB_PASS")
	var db *gobdb.DB
	if passphrase == "" {
		db = gobdb.NewDB("", false)
	} else {
		db = gobdb.NewEncryptedDB("", false, passphrase)
	}
	id, err := strconv.Atoi(filename)
	if err != nil {
		return errors.New("Resource ID parameter must be a number")
	}
	q := db.NewQueryWithID("show", nil, id)
	q.TablePath = tablePath
	q.ThwartIOBasePathEscape()
	q.ExitIfTableNotExist()
	q.BuildResourcePath()
	q.ReadGobFromDisk()
	q.DecryptGobBuffer()
	if q.FatalError != nil {
		if q.FatalError.Error()[:6] == "cipher" {
			fmt.Println("Failed to decrypt with GOBDB_PASS environment variable.")
		}
		return q.FatalError
	}
	reader := bytes.NewReader(q.GobBuffer)
	dec := degob.NewDecoder(reader)
	gobs, err := dec.Decode()
	if err != nil {
		fmt.Println("Failed to decode gob. If DB is encrypted set GOBDB_PASS environment variable.")
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

	passphrase := os.Getenv("GOBDB_PASS")
	var db *gobdb.DB
	if passphrase == "" {
		db = gobdb.NewDB("", false)
	} else {
		db = gobdb.NewEncryptedDB("", false, passphrase)
	}

	for _, f := range files {
		if f.IsDir() {
			continue
		}

		id, err := strconv.Atoi(f.Name())
		if err != nil {
			continue
		}
		q := db.NewQueryWithID("show", nil, id)
		q.TablePath = tablePath
		q.ThwartIOBasePathEscape()
		q.ExitIfTableNotExist()
		q.BuildResourcePath()
		q.ReadGobFromDisk()
		q.DecryptGobBuffer()
		if q.FatalError != nil {
			if q.FatalError.Error()[:6] == "cipher" {
				fmt.Println("Failed to decrypt with GOBDB_PASS environment variable.")
			}
			return q.FatalError
		}
		reader := bytes.NewReader(q.GobBuffer)
		dec := degob.NewDecoder(reader)
		gobs, err := dec.Decode()
		if err != nil {
			fmt.Println("Failed to decode gob. If DB is encrypted set GOBDB_PASS environment variable.")
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
