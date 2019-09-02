package gobdb

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

var mutex sync.Mutex

type Resource interface {
	GetID() int
	SetID(ID int)
}

type DB struct {
	Path string
	Log  bool
}

type Query struct {
	FatalError   error
	DB           DB
	Writer       bytes.Buffer
	ReadBuffer   []byte
	Model        string
	Resource     Resource
	ID           int
	Counter      int
	CounterPath  string
	MetadataPath string
	ResourcePath string
	TablePath    string
}

func (q Query) Log(operation string) {
	if q.DB.Log {
		fmt.Println("Operation    :", operation)
		fmt.Println("Model        :", q.Model)
		fmt.Println("ID           :", q.ID)
		fmt.Println("Resource     :", q.Resource)
		fmt.Println("Table Path   :", q.TablePath)
		fmt.Println("Fatal Error  :", q.FatalError)
	}
}

func (db DB) NewQueryWithoutID(resource Resource) *Query {
	return &Query{
		DB:       db,
		Resource: resource,
	}
}

func (db DB) NewQueryWithID(resource Resource, id int) *Query {
	return &Query{
		DB:       db,
		Resource: resource,
		ID:       id,
	}
}

func (db DB) Insert(resource Resource) {
	mutex.Lock()
	defer mutex.Unlock()

	q := db.NewQueryWithoutID(resource)
	q.ReflectModelNameFromResource()
	q.BuildTablePath()
	q.BuildMetadataPath()
	q.CreateMetadataDirectoryIfNotExist()
	q.BuildCounterPath()
	q.ReadCounterFromDisk()
	q.IncrementCounterAndSetID()
	q.EncodeResource()
	q.BuildResourcePath()
	q.WriteResourceToDisk()
	q.WriteCounterToDisk()
	q.Log("insert")
}

func (db DB) Get(resource Resource, id int) error {
	mutex.Lock()
	defer mutex.Unlock()

	q := db.NewQueryWithID(resource, id)
	q.ReflectModelNameFromResource()
	q.BuildTablePath()
	q.ExitIfTableNotExist()
	q.BuildResourcePath()
	q.ReadFromDiskIntoBuffer()
	q.DecodeBufferIntoResource()
	q.Log("Get")
	return q.FatalError
}

// func (db DB) GetAll(resource interface{}, callback func(resource interface{})) error {
// 	mutex.Lock()
// 	defer mutex.Unlock()

// 	tablePath := db.TablePath(resource)
// 	if _, err := os.Stat(tablePath); os.IsNotExist(err) {
// 		return err
// 	}

// 	files, err := ioutil.ReadDir(tablePath)
// 	if err != nil {
// 		return err
// 	}

// 	db.ThwartIOBasePathEscape(tablePath)
// 	for _, f := range files {
// 		if f.IsDir() {
// 			continue
// 		}

// 		id, err := strconv.Atoi(f.Name())
// 		if err != nil {
// 			continue
// 		}
// 		b, err := ioutil.ReadFile(ResourcePath(tablePath, id))
// 		if err != nil {
// 			return err
// 		}

// 		buf := bytes.NewReader(b)
// 		dec := gob.NewDecoder(buf)
// 		err = dec.Decode(resource)
// 		if err != nil {
// 			return err
// 		}
// 		callback(resource)
// 	}
// 	return nil
// }

// func (db DB) Update(resource Resource) error {
// 	mutex.Lock()
// 	defer mutex.Unlock()

// 	tablePath := db.TablePath(resource)
// 	if _, err := os.Stat(tablePath); os.IsNotExist(err) {
// 		return err
// 	}

// 	var buf bytes.Buffer
// 	enc := gob.NewEncoder(&buf)
// 	err := enc.Encode(resource)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	id := resource.GetID()
// 	err = db.SafeWrite(ResourcePath(tablePath, id), buf.Bytes())
// 	return err
// }

// func (db DB) Delete(resource Resource) error {
// 	mutex.Lock()
// 	defer mutex.Unlock()

// 	tablePath := db.TablePath(resource)
// 	resourcePath := ResourcePath(tablePath, resource.GetID())
// 	db.ThwartIOBasePathEscape(resourcePath)
// 	err := os.Remove(resourcePath)
// 	return err
// }

// func (db DB) DeleteAll(resource Resource) error {
// 	mutex.Lock()
// 	defer mutex.Unlock()

// 	tablePath := db.TablePath(resource)
// 	if _, err := os.Stat(tablePath); os.IsNotExist(err) {
// 		return nil
// 	}
// 	files, err := ioutil.ReadDir(tablePath)
// 	if err != nil {
// 		return err
// 	}

// 	db.ThwartIOBasePathEscape(tablePath)
// 	for _, f := range files {
// 		err = DeleteFileWithIntegerNameOnly(tablePath, f)
// 		if err != nil {
// 			log.Println(err)
// 		}
// 	}
// 	return nil
// }

// func DeleteFileWithIntegerNameOnly(path string, f os.FileInfo) error {
// 	if f.IsDir() {
// 		return nil
// 	}
// 	id, err := strconv.Atoi(f.Name())
// 	if err != nil {
// 		return nil
// 	}
// 	err = os.Remove(ResourcePath(path, id))
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

func (q *Query) ReflectModelNameFromResource() {
	if q.FatalError != nil {
		return
	}
	if q.Resource == nil {
		q.FatalError = errors.New("Resource missing")
		return
	}
	q.Model = reflect.TypeOf(q.Resource).String()[1:]
}

func (q *Query) BuildTablePath() {
	if q.FatalError != nil {
		return
	}
	if q.DB.Path == "" {
		q.FatalError = errors.New("DB path missing")
		return
	}
	if q.Model == "" {
		q.FatalError = errors.New("Model name missing")
		return
	}
	q.TablePath = q.DB.Path + "/" + q.Model
}

func (q *Query) BuildMetadataPath() {
	if q.FatalError != nil {
		return
	}
	if q.TablePath == "" {
		q.FatalError = errors.New("Table path missing")
		return
	}
	q.MetadataPath = q.TablePath + "/metadata"
}

func (q *Query) CreateMetadataDirectoryIfNotExist() {
	if q.FatalError != nil {
		return
	}
	if q.MetadataPath == "" {
		q.FatalError = errors.New("Metadata path missing")
		return
	}
	if _, err := os.Stat(q.MetadataPath); os.IsNotExist(err) {
		q.FatalError = os.MkdirAll(q.MetadataPath, os.ModePerm)
	}
}

func (q *Query) BuildCounterPath() {
	if q.FatalError != nil {
		return
	}
	if q.MetadataPath == "" {
		q.FatalError = errors.New("Metadata path missing")
		return
	}
	q.CounterPath = q.MetadataPath + "/counter"
}

func (q *Query) BuildResourcePath() {
	if q.FatalError != nil {
		return
	}
	if q.TablePath == "" {
		q.FatalError = errors.New("Table path missing")
		return
	}
	if q.ID < 1 {
		q.FatalError = errors.New("ID smaller than 1")
		return
	}
	q.ResourcePath = q.TablePath + "/" + strconv.Itoa(q.ID)
}

func (q *Query) ReadCounterFromDisk() {
	if q.FatalError != nil {
		return
	}
	if q.CounterPath == "" {
		q.FatalError = errors.New("Counter path missing")
		return
	}
	b, err := q.DB.SafeRead(q.CounterPath)
	if err == nil {
		q.Counter, q.FatalError = strconv.Atoi(string(b))
	} else {
		q.Counter = 0
	}
}

func (q *Query) IncrementCounterAndSetID() {
	if q.FatalError != nil {
		return
	}
	q.Counter++
	q.ID = q.Counter
	q.Resource.SetID(q.ID)
}

func (q *Query) EncodeResource() {
	if q.FatalError != nil {
		return
	}
	enc := gob.NewEncoder(&q.Writer)
	q.FatalError = enc.Encode(q.Resource)
}

func (q *Query) WriteResourceToDisk() {
	if q.FatalError != nil {
		return
	}
	if q.ResourcePath == "" {
		q.FatalError = errors.New("Resource path missing")
		return
	}
	q.FatalError = q.DB.SafeWrite(q.ResourcePath, q.Writer.Bytes())
}

func (q *Query) WriteCounterToDisk() {
	if q.FatalError != nil {
		return
	}
	if q.CounterPath == "" {
		q.FatalError = errors.New("Counter path missing")
		return
	}
	q.FatalError = q.DB.SafeWrite(q.CounterPath, []byte(strconv.Itoa(q.Counter)))
}

func (q *Query) ExitIfTableNotExist() {
	if q.FatalError != nil {
		return
	}
	if q.TablePath == "" {
		q.FatalError = errors.New("Table path missing")
		return
	}
	if _, err := os.Stat(q.TablePath); os.IsNotExist(err) {
		q.FatalError = errors.New("Table does not exist")
	}
}

func (q *Query) ReadFromDiskIntoBuffer() {
	if q.FatalError != nil {
		return
	}
	if q.ResourcePath == "" {
		q.FatalError = errors.New("Resource path missing")
		return
	}
	q.ReadBuffer, q.FatalError = q.DB.SafeRead(q.ResourcePath)
}

func (q *Query) DecodeBufferIntoResource() {
	if q.FatalError != nil {
		return
	}
	if len(q.ReadBuffer) == 0 {
		q.FatalError = errors.New("Read buffer empty")
		return
	}
	reader := bytes.NewReader(q.ReadBuffer)
	dec := gob.NewDecoder(reader)
	q.FatalError = dec.Decode(q.Resource)
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
