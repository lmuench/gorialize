package gobdb

import (
	"bytes"
	"encoding/gob"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
)

type DB struct {
	Path string
}

type Resource interface {
	GetID() int
	SetID(ID int)
}

func (db DB) Insert(resource Resource) {
	tablePath := db.TablePath(resource)
	metadataPath := TableMetadataPath(tablePath)

	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		err = os.MkdirAll(metadataPath, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
	}

	counterPath := TableCounterPath(metadataPath)
	var counter int
	b, err := db.SafeRead(counterPath)
	if err == nil {
		counter, err = strconv.Atoi(string(b))
		if err != nil {
			log.Fatal(err)
		}
		counter++
	} else {
		counter = 1
	}
	id := counter
	resource.SetID(id)

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err = enc.Encode(resource)
	if err != nil {
		log.Fatal(err)
	}

	err = db.SafeWrite(ResourcePath(tablePath, id), buf.Bytes())
	if err != nil {
		log.Fatal(err)
	}

	err = db.SafeWrite(counterPath, []byte(strconv.Itoa(counter)))
	if err != nil {
		log.Fatal(err)
	}
}

func (db DB) Get(resource interface{}, id int) error {
	tablePath := db.TablePath(resource)
	if _, err := os.Stat(tablePath); os.IsNotExist(err) {
		return err
	}

	b, err := db.SafeRead(ResourcePath(tablePath, id))
	if err != nil {
		return err
	}

	buf := bytes.NewReader(b)
	dec := gob.NewDecoder(buf)
	err = dec.Decode(resource)
	return err
}

func (db DB) GetAll(resource interface{}, callback func(resource interface{})) error {
	tablePath := db.TablePath(resource)
	if _, err := os.Stat(tablePath); os.IsNotExist(err) {
		return err
	}

	files, err := ioutil.ReadDir(tablePath)
	if err != nil {
		return err
	}

	db.ThwartIOBasePathEscape(tablePath)
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
		dec := gob.NewDecoder(buf)
		err = dec.Decode(resource)
		if err != nil {
			return err
		}
		callback(resource)
	}
	return nil
}

func (db DB) Delete(resource Resource) error {
	tablePath := db.TablePath(resource)
	resourcePath := ResourcePath(tablePath, resource.GetID())
	db.ThwartIOBasePathEscape(resourcePath)
	err := os.Remove(resourcePath)
	return err
}

func (db DB) DeleteAll(resource Resource) error {
	tablePath := db.TablePath(resource)
	if _, err := os.Stat(tablePath); os.IsNotExist(err) {
		return nil
	}
	files, err := ioutil.ReadDir(tablePath)
	if err != nil {
		return err
	}

	db.ThwartIOBasePathEscape(tablePath)
	for _, f := range files {
		err = DeleteFileWithIntegerNameOnly(tablePath, f)
		if err != nil {
			log.Println(err)
		}
	}
	return nil
}

func DeleteFileWithIntegerNameOnly(path string, f os.FileInfo) error {
	if f.IsDir() {
		return nil
	}
	id, err := strconv.Atoi(f.Name())
	if err != nil {
		return nil
	}
	err = os.Remove(ResourcePath(path, id))
	if err != nil {
		return err
	}
	return nil
}

func (db DB) Update(resource Resource) error {
	tablePath := db.TablePath(resource)
	if _, err := os.Stat(tablePath); os.IsNotExist(err) {
		return err
	}

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(resource)
	if err != nil {
		log.Fatal(err)
	}

	id := resource.GetID()
	err = db.SafeWrite(ResourcePath(tablePath, id), buf.Bytes())
	return err
}

func ModelName(resource interface{}) string {
	return reflect.TypeOf(resource).String()[1:]
}

func (db DB) TablePath(resource interface{}) string {
	model := ModelName(resource)
	return db.Path + "/" + model
}

func TableMetadataPath(tablePath string) string {
	return tablePath + "/metadata"
}

func TableCounterPath(metadataPath string) string {
	return metadataPath + "/counter"
}

func ResourcePath(tablePath string, id int) string {
	return tablePath + "/" + strconv.Itoa(id)
}

func (db DB) ThwartIOBasePathEscape(ioOperationPath string) {
	if !strings.HasPrefix(ioOperationPath, db.Path) {
		log.Fatal("Thwarted attempted IO operation outside of", db.Path)
	}
	if strings.Contains(ioOperationPath, "..") {
		log.Fatal("Thwarted attempted IO operation with path containing '..'")
	}
}

func (db DB) SafeWrite(path string, b []byte) error {
	db.ThwartIOBasePathEscape(path)
	err := ioutil.WriteFile(path, b, 0644)
	return err
}

func (db DB) SafeRead(path string) ([]byte, error) {
	db.ThwartIOBasePathEscape(path)
	b, err := ioutil.ReadFile(path)
	return b, err
}
