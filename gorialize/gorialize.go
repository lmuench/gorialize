// Package gorialize is a serialization framework for Go. It aims to provide an embedded persistence layer
// for applications that do not require all the features of a database. Gorialize lets you serialize
// your structs and other data types to gobs while organizing the serialized data like a database.
// It provides a CRUD API that accepts any type that implements the Gorialize Resource interface
package gorialize

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

var mutex sync.Mutex

type Index map[Field]map[Value][]ID
type Model string
type Field string
type Value interface{}
type ID int

// Directory exposes methods to read and write serialized data inside a base directory.
type Directory struct {
	Path      string
	Encrypted bool
	Key       *[32]byte
	Log       bool
	Indices   map[Model]Index
}

// DirectoryConfig holds parameters to be passed to NewDirectory().
type DirectoryConfig struct {
	Path       string
	Encrypted  bool
	Passphrase string
	Log        bool
}

// NewDirectory returns a new Directory struct for the given configuration.
func NewDirectory(config DirectoryConfig) *Directory {
	dir := &Directory{
		Path: config.Path,
		Log:  config.Log,
		Indices: map[Model]Index{},
	}

	if config.Encrypted {
		var key [32]byte
		h := hashPassphrase([]byte(config.Passphrase))
		copy(key[:], h[:32])
		dir.Key = &key
		dir.Encrypted = true
	}
	return dir
}

type Query struct {
	FatalError   error
	Dir          Directory
	Operation    string
	Writer       bytes.Buffer
	GobBuffer    []byte
	ResourceType reflect.Type
	Model        string
	Resource     interface{}
	ID           int
	Counter      int
	CounterPath  string
	MetadataPath string
	ResourcePath string
	DirPath      string
	SafeIOPath   bool
	DirFileInfo  []os.FileInfo
	WhereClauses []Where
	Matches      []ID
	IndexUpdateLog []string
}

type Where struct {
	Field Field
	Value Value
}

func (q Query) Log() {
	if q.Dir.Log {
		fmt.Println()
		fmt.Println("Operation       :", q.Operation)
		fmt.Println("Model           :", q.Model)
		fmt.Println("ID              :", q.ID)
		fmt.Println("Resource        :", q.Resource)
		fmt.Println("Directory       :", q.DirPath)
		fmt.Println("Fatal Error     :", q.FatalError)
		if len(q.IndexUpdateLog) > 0 {
			fmt.Println("Updated Indices :", q.IndexUpdateLog)
		}
		fmt.Println()
	}
}

func (dir Directory) newQueryWithoutID(operation string, resource interface{}) *Query {
	return &Query{
		Dir:       dir,
		Operation: operation,
		Resource:  resource,
	}
}

func (dir Directory) newQueryWithID(operation string, resource interface{}, id int) *Query {
	return &Query{
		Dir:       dir,
		Operation: operation,
		Resource:  resource,
		ID:        id,
	}
}

// Create creates a new serialized resource and sets its ID.
// TODO: update and Dir.Upsert indices
func (dir Directory) Create(resource interface{}) error {
	mutex.Lock()
	defer mutex.Unlock()

	q := dir.newQueryWithoutID("create", resource)
	q.ReflectTypeOfResource()
	q.ReflectModelNameFromType()
	q.BuildDirPath()
	q.ThwartIOBasePathEscape()
	q.BuildMetadataPath()
	q.CreateMetadataDirectoryIfNotExist()
	q.BuildCounterPath()
	q.ReadCounterFromDisk()
	q.IncrementCounterAndSetID()
	q.EncodeResourceToGob()
	q.EncryptGobBuffer()
	q.BuildResourcePath()
	q.WriteGobToDisk()
	q.WriteCounterToDisk()
	q.UpdateIndices()
	q.Log()
	return q.FatalError
}

// Read reads the serialized resource with the given ID.
func (dir Directory) Read(resource interface{}, id int) error {
	mutex.Lock()
	defer mutex.Unlock()

	q := dir.newQueryWithID("read", resource, id)
	q.ReflectTypeOfResource()
	q.ReflectModelNameFromType()
	q.BuildDirPath()
	q.ThwartIOBasePathEscape()
	q.ExitIfDirNotExist()
	q.BuildResourcePath()
	q.ReadGobFromDisk()
	q.DecryptGobBuffer()
	q.DecodeGobToResource()
	q.Log()
	return q.FatalError
}

// GetOwner reads the serialized resource which owns the given resource.
// The resource needs to have an addressable owner ID int field which
// follows a 'FooID' naming convention where 'Foo' is the owner type.
func (dir Directory) GetOwner(resource interface{}, owner interface{}) error {
	ownerID, err := getOwnerID(resource, owner)
	if err != nil {
		return err
	}

	err = dir.Read(owner, ownerID)
	return err
}

// readFromCustomSubdirectory reads the serialized resource with the given ID from a custom subdirectory.
// This method is intended for testing purposes.
func (dir Directory) readFromCustomSubdirectory(resource interface{}, id int, subdir string) error {
	mutex.Lock()
	defer mutex.Unlock()

	q := dir.newQueryWithID("read", resource, id)
	q.BuildCustomDirPath(subdir)
	q.ThwartIOBasePathEscape()
	q.ExitIfDirNotExist()
	q.BuildResourcePath()
	q.ReadGobFromDisk()
	q.DecryptGobBuffer()
	q.DecodeGobToResource()
	q.Log()
	return q.FatalError
}

// ReadAllIntoSlice reads all serialized resources of the given slice's elements's type and writes them into the slice.
func (dir Directory) ReadAllIntoSlice(slice interface{}) error {
	slicePtr := reflect.ValueOf(slice)
	sliceVal := reflect.Indirect(slicePtr)
	resourceTyp := reflect.TypeOf(slice).Elem().Elem()
	resourceVal := reflect.New(resourceTyp)
	resource := resourceVal.Interface()

	err := dir.ReadAll(resource, func(resource interface{}) {
		resourcePtr := reflect.ValueOf(resource)
		resourceVal := reflect.Indirect(resourcePtr)
		sliceVal.Set(reflect.Append(sliceVal, resourceVal))
	})
	return err
}

// ReadAll reads all serialized resource of the given type and calls the provided callback function on each.
func (dir Directory) ReadAll(resource interface{}, callback func(resource interface{})) error {
	mutex.Lock()
	defer mutex.Unlock()
	var err error

	q := dir.newQueryWithoutID("read", resource)
	q.ReflectTypeOfResource()
	q.ReflectModelNameFromType()
	q.BuildDirPath()
	q.ThwartIOBasePathEscape()
	q.ExitIfDirNotExist()
	q.ReadDirFileinfo()
	for _, f := range q.DirFileInfo {
		if f.IsDir() {
			continue
		}
		q.ID, err = strconv.Atoi(f.Name())
		if err != nil {
			continue
		}
		q.BuildResourcePath()
		q.ReadGobFromDisk()
		q.DecryptGobBuffer()
		q.DecodeGobToResource()
		q.PassResourceToCallback(callback)
		q.Log()
	}
	return q.FatalError
}

// Find reads the first serialized resources matching the given WHERE clauses
// Note: WHERE clauses can only be used with indexed fields.
func (dir Directory) Find(resource interface{}, whereClauses ...Where) error {
	mutex.Lock()
	defer mutex.Unlock()

	q := dir.newQueryWithoutID("find", resource)
  q.WhereClauses = whereClauses
	q.ReflectTypeOfResource()
	q.ReflectModelNameFromType()
	q.ApplyWhereClauses(true)
	q.BuildDirPath()
	q.ThwartIOBasePathEscape()
	q.ExitIfDirNotExist()
	q.BuildResourcePath()
	q.ReadGobFromDisk()
	q.DecryptGobBuffer()
	q.DecodeGobToResource()
	q.Log()
	return q.FatalError
}

// Replace replaces a serialized resource.
// TODO: update and Dir.Upsert indices
func (dir Directory) Replace(resource interface{}) error {
	mutex.Lock()
	defer mutex.Unlock()

	id, err := getID(resource)
	if err != nil {
		return err
	}
	q := dir.newQueryWithID("replace", resource, id)
	q.ReflectTypeOfResource()
	q.ReflectModelNameFromType()
	q.BuildDirPath()
	q.ThwartIOBasePathEscape()
	q.ExitIfDirNotExist()
	q.BuildResourcePath()
	q.ExitIfResourceNotExist()
	q.EncodeResourceToGob()
	q.EncryptGobBuffer()
	q.BuildResourcePath()
	q.WriteGobToDisk()
	q.Log()
	return q.FatalError
}

// Update partially updates a serialized resource with all non-zero values of the given resource.
// TODO: update and Dir.Upsert indices
func (dir Directory) Update(resource interface{}, id int) error {
	err := dir.Create(resource)
	if err != nil {
		return err
	}

	tmpID, err := getID(resource)
	if err != nil {
		return err
	}

	err = dir.Read(resource, id)
	if err != nil {
		return err
	}

	err = dir.Read(resource, tmpID)
	if err != nil {
		return err
	}

	err = setID(resource, id)
	if err != nil {
		return err
	}
	err = dir.Replace(resource)
	if err != nil {
		return err
	}

	err = setID(resource, tmpID)
	if err != nil {
		return err
	}
	err = dir.Delete(resource)
	return err
}

// Delete deletes a serialized resource.
func (dir Directory) Delete(resource interface{}) error {
	mutex.Lock()
	defer mutex.Unlock()

	id, err := getID(resource)
	if err != nil {
		return err
	}
	q := dir.newQueryWithID("delete", resource, id)
	q.ReflectTypeOfResource()
	q.ReflectModelNameFromType()
	q.BuildDirPath()
	q.ThwartIOBasePathEscape()
	q.ExitIfDirNotExist()
	q.BuildResourcePath()
	q.ThwartIOBasePathEscape()
	q.FatalError = os.Remove(q.ResourcePath) // TODO make this a Directory method
	q.Log()
	return q.FatalError
}

// DeleteAll deletes all serialized resources of the given type.
func (dir Directory) DeleteAll(resource interface{}) error {
	mutex.Lock()
	defer mutex.Unlock()
	var err error

	q := dir.newQueryWithoutID("delete", resource)
	q.ReflectTypeOfResource()
	q.ReflectModelNameFromType()
	q.BuildDirPath()
	q.ThwartIOBasePathEscape()
	q.ExitIfDirNotExist()
	q.ReadDirFileinfo()
	for _, f := range q.DirFileInfo {
		if f.IsDir() {
			continue
		}
		q.ID, err = strconv.Atoi(f.Name())
		if err != nil {
			continue
		}
		q.BuildResourcePath()
		q.DeleteFromDisk()
		q.Log()
	}
	return q.FatalError
}

// ResetCounter resets the resource counter to zero
func (dir Directory) ResetCounter(resource interface{}) error {
	mutex.Lock()
	defer mutex.Unlock()

	q := dir.newQueryWithoutID("reset counter", resource)
	q.ReflectTypeOfResource()
	q.ReflectModelNameFromType()
	q.BuildDirPath()
	q.ThwartIOBasePathEscape()
	q.BuildMetadataPath()
	q.CreateMetadataDirectoryIfNotExist()
	q.BuildCounterPath()
	q.SetCounterToZero()
	q.WriteCounterToDisk()
	q.Log()
	return q.FatalError
}

func (q *Query) ReflectTypeOfResource() {
	if q.FatalError != nil {
		return
	}
	if q.Resource == nil {
		q.FatalError = errors.New("Resource missing")
		return
	}
	q.ResourceType = reflect.TypeOf(q.Resource)
}

func (q *Query) ReflectModelNameFromType() {
	if q.FatalError != nil {
		return
	}
	if q.ResourceType == nil {
		q.FatalError = errors.New("Resource type missing")
		return
	}
	q.Model = q.ResourceType.String()[1:]
}

func (q *Query) UpdateIndices() {
	if q.FatalError != nil {
		return
	}
	if q.ResourceType == nil {
		q.FatalError = errors.New("Resource type missing")
		return
	}
	for i := 0; i < q.ResourceType.Elem().NumField(); i++ {
		field := q.ResourceType.Elem().Field(i)
		tag := field.Tag.Get("gorialize")
			if tag == "indexed" {
			value := reflect.Indirect(
				reflect.ValueOf(q.Resource),
			).FieldByName(field.Name).Interface()
			if _, ok := q.Dir.Indices[Model(q.Model)]; !ok {
				q.Dir.Indices[Model(q.Model)] = Index{}
			}
			if _, ok := q.Dir.Indices[Model(q.Model)][Field(field.Name)]; !ok {
				q.Dir.Indices[Model(q.Model)][Field(field.Name)] = map[Value][]ID{}
			}
			if _, ok := q.Dir.Indices[Model(q.Model)][Field(field.Name)][Value(value)] ; !ok {
				q.Dir.Indices[Model(q.Model)][Field(field.Name)][Value(value)] = []ID{}
			}
			q.Dir.Indices[Model(q.Model)][Field(field.Name)][Value(value)] = append(
				q.Dir.Indices[Model(q.Model)][Field(field.Name)][Value(value)], ID(q.ID),
			)
			q.IndexUpdateLog = append(
				q.IndexUpdateLog,
				fmt.Sprintf("%s %d: %s = %v", q.Model, q.ID, field.Name, value),
			)
		}
	}
}

func (q *Query) BuildDirPath() {
	if q.FatalError != nil {
		return
	}
	if q.Dir.Path == "" {
		q.FatalError = errors.New("Directory path missing")
		return
	}
	if q.Model == "" {
		q.FatalError = errors.New("Model name missing")
		return
	}
	q.DirPath = q.Dir.Path + "/" + q.Model
}

func (q *Query) BuildCustomDirPath(subdir string) {
	if q.FatalError != nil {
		return
	}
	if q.Dir.Path == "" {
		q.FatalError = errors.New("Directory path missing")
		return
	}
	q.DirPath = q.Dir.Path + "/" + subdir
}

func (q *Query) BuildMetadataPath() {
	if q.FatalError != nil {
		return
	}
	if q.DirPath == "" {
		q.FatalError = errors.New("Directory path missing")
		return
	}
	q.MetadataPath = q.DirPath + "/metadata"
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
	if q.DirPath == "" {
		q.FatalError = errors.New("Directory path missing")
		return
	}
	if q.ID < 1 {
		q.FatalError = errors.New("ID smaller than 1")
		return
	}
	filename := fmt.Sprintf("%07d", q.ID)
	q.ResourcePath = q.DirPath + "/" + filename
}

func (q *Query) ReadCounterFromDisk() {
	if q.FatalError != nil {
		return
	}
	if !q.SafeIOPath {
		q.FatalError = errors.New("IO path not marked as safe")
		return
	}
	if q.CounterPath == "" {
		q.FatalError = errors.New("Counter path missing")
		return
	}
	b, err := q.Dir.readFromDisk(q.CounterPath)
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
	if q.Resource == nil {
		q.FatalError = errors.New("Resource missing")
		return
	}
	q.Counter++
	q.ID = q.Counter
	q.FatalError = setID(q.Resource, q.ID)
}

func (q *Query) SetCounterToZero() {
	if q.FatalError != nil {
		return
	}
	q.Counter = 0
}

func (q *Query) EncodeResourceToGob() {
	if q.FatalError != nil {
		return
	}
	enc := gob.NewEncoder(&q.Writer)
	q.FatalError = enc.Encode(q.Resource)
	if q.FatalError == nil {
		q.GobBuffer = q.Writer.Bytes()
	}
}

func (q *Query) WriteGobToDisk() {
	if q.FatalError != nil {
		return
	}
	if !q.SafeIOPath {
		q.FatalError = errors.New("Write path not marked as safe")
		return
	}
	if q.ResourcePath == "" {
		q.FatalError = errors.New("Resource path missing")
		return
	}
	q.FatalError = q.Dir.writeToDisk(q.ResourcePath, q.GobBuffer)
}

func (q *Query) WriteCounterToDisk() {
	if q.FatalError != nil {
		return
	}
	if !q.SafeIOPath {
		q.FatalError = errors.New("Write path not marked as safe")
		return
	}
	if q.CounterPath == "" {
		q.FatalError = errors.New("Counter path missing")
		return
	}
	q.FatalError = q.Dir.writeToDisk(q.CounterPath, []byte(strconv.Itoa(q.Counter)))
}

func (q *Query) ExitIfDirNotExist() {
	if q.FatalError != nil {
		return
	}
	if q.DirPath == "" {
		q.FatalError = errors.New("Directory path missing")
		return
	}
	if _, err := os.Stat(q.DirPath); os.IsNotExist(err) {
		fmt.Println(q.DirPath)
		q.FatalError = errors.New("Directory does not exist")
	}
}

func (q *Query) ExitIfResourceNotExist() {
	if q.FatalError != nil {
		return
	}
	if q.ResourcePath == "" {
		q.FatalError = errors.New("Resource path missing")
		return
	}
	if _, err := os.Stat(q.ResourcePath); os.IsNotExist(err) {
		q.FatalError = errors.New("Resource does not exist")
	}
}

func (q *Query) ReadGobFromDisk() {
	if q.FatalError != nil {
		return
	}
	if !q.SafeIOPath {
		q.FatalError = errors.New("IO path not marked as safe")
		return
	}
	if q.ResourcePath == "" {
		q.FatalError = errors.New("Resource path missing")
		return
	}
	q.GobBuffer, q.FatalError = q.Dir.readFromDisk(q.ResourcePath)
}

func (q *Query) DecodeGobToResource() {
	if q.FatalError != nil {
		return
	}
	if len(q.GobBuffer) == 0 {
		q.FatalError = errors.New("Gob buffer empty")
		return
	}
	reader := bytes.NewReader(q.GobBuffer)
	dec := gob.NewDecoder(reader)
	q.FatalError = dec.Decode(q.Resource)
}

func (q *Query) ReadDirFileinfo() {
	if q.FatalError != nil {
		return
	}
	if !q.SafeIOPath {
		q.FatalError = errors.New("IO path not marked as safe")
		return
	}
	if q.DirPath == "" {
		q.FatalError = errors.New("Directory path missing")
		return
	}
	q.DirFileInfo, q.FatalError = ioutil.ReadDir(q.DirPath)
}

func (q *Query) PassResourceToCallback(callback func(resource interface{})) {
	if q.FatalError != nil {
		return
	}
	if q.Resource == nil {
		q.FatalError = errors.New("Resource missing")
		return
	}
	callback(q.Resource)
}

func (q *Query) DeleteFromDisk() {
	if q.FatalError != nil {
		return
	}
	if q.ResourcePath == "" {
		q.FatalError = errors.New("Resource path missing")
		return
	}
	q.FatalError = os.Remove(q.ResourcePath) // TODO make this a Directory method
}

func (q *Query) ThwartIOBasePathEscape() {
	if !strings.HasPrefix(q.DirPath, q.Dir.Path) {
		q.SafeIOPath = false
		q.Log()
		log.Fatal("Thwarted IO operation outside of ", q.Dir.Path)
	}
	if strings.Contains(q.DirPath, "..") {
		q.SafeIOPath = false
		q.Log()
		log.Fatal("Thwarted IO operation with path containing '..'")
	}
	q.SafeIOPath = true
}

func (q *Query) EncryptGobBuffer() {
	if q.FatalError != nil {
		return
	}
	if !q.Dir.Encrypted {
		return
	}
	if len(q.GobBuffer) == 0 {
		q.FatalError = errors.New("Gob buffer empty")
		return
	}
	if q.Dir.Key == nil {
		q.FatalError = errors.New("Encryption key missing")
	}
	var block cipher.Block
	block, q.FatalError = aes.NewCipher(q.Dir.Key[:])
	if q.FatalError != nil {
		return
	}
	var gcm cipher.AEAD
	gcm, q.FatalError = cipher.NewGCM(block)
	if q.FatalError != nil {
		return
	}
	nonce := make([]byte, gcm.NonceSize())
	_, q.FatalError = io.ReadFull(rand.Reader, nonce)
	if q.FatalError != nil {
		return
	}
	q.GobBuffer = gcm.Seal(nonce, nonce, q.GobBuffer, nil)
}

func (q *Query) DecryptGobBuffer() {
	if q.FatalError != nil {
		return
	}
	if !q.Dir.Encrypted {
		return
	}
	if q.Dir.Key == nil {
		q.FatalError = errors.New("Decryption key missing")
	}
	var block cipher.Block
	block, q.FatalError = aes.NewCipher(q.Dir.Key[:])
	if q.FatalError != nil {
		return
	}
	var gcm cipher.AEAD
	gcm, q.FatalError = cipher.NewGCM(block)
	if q.FatalError != nil {
		return
	}
	if len(q.GobBuffer) < gcm.NonceSize() {
		q.FatalError = errors.New("Gob buffer smaller than nonce")
		return
	}
	q.GobBuffer, q.FatalError = gcm.Open(
		nil,
		q.GobBuffer[:gcm.NonceSize()],
		q.GobBuffer[gcm.NonceSize():],
		nil,
	)
}

func (q *Query) ApplyWhereClauses(pickFirst bool) {
	if q.FatalError != nil {
		return
	}
	if q.Model == "" {
		q.FatalError = errors.New("Model name missing")
		return
	}
	cnt := len(q.WhereClauses)
	if cnt == 0 {
		q.FatalError = errors.New("Where clauses missing")
		return
	}
	modelIndices, ok := q.Dir.Indices[Model(q.Model)]
	if !ok {
		q.FatalError = errors.New("Model indices missing")
	}
	idMap := make(map[ID]int, 1)
	for _, clause := range q.WhereClauses {
		idx, ok := modelIndices[clause.Field]
		if !ok {
			q.FatalError = errors.New("Index missing")
		}
		ids, ok := idx[clause.Value]
		if ok {
			for _, id := range ids {
				idMap[id] += 1
				if pickFirst {
					if idMap[id] == cnt {
						q.Matches = append(q.Matches, id)
						q.ID = int(id)
						return
					}
				}
			}
		}
	}
	for id, v := range idMap {
		if v == cnt {
			q.Matches = append(q.Matches, id)
		}
	}
	if len(q.Matches) == 0 {
		q.FatalError = errors.New("No matching where clauses")
	}
}

func hashPassphrase(passphrase []byte) []byte {
	h := hmac.New(sha512.New512_256, []byte("key"))
	_, _ = h.Write(passphrase) // TODO find out if returned err should be checked
	return h.Sum(nil)
}

func (dir Directory) writeToDisk(path string, b []byte) error {
	err := ioutil.WriteFile(path, b, 0644)
	return err
}

func (dir Directory) readFromDisk(path string) ([]byte, error) {
	b, err := ioutil.ReadFile(path)
	return b, err
}

func getID(resource interface{}) (int, error) {
	val := reflect.ValueOf(resource).Elem()
	if val.Kind() != reflect.Struct {
		return 0, errors.New("resource is not a struct pointer")
	}

	idField := val.FieldByName("ID")
	if !idField.IsValid() || !idField.CanSet() || idField.Kind() != reflect.Int {
		return 0, errors.New("resource does not have an addressable ID int field")
	}
	return int(idField.Int()), nil
}

func setID(resource interface{}, id int) error {
	val := reflect.ValueOf(resource).Elem()
	if val.Kind() != reflect.Struct {
		return errors.New("resource is not a struct pointer")
	}

	idField := val.FieldByName("ID")
	if !idField.IsValid() || !idField.CanSet() || idField.Kind() != reflect.Int {
		return errors.New("resource does not have an addressable ID int field")
	}
	idField.SetInt(int64(id))
	return nil
}

func getOwnerID(resource interface{}, owner interface{}) (int, error) {
	if reflect.ValueOf(owner).Elem().Kind() != reflect.Struct {
		return 0, errors.New("owner is not a struct pointer")
	}
	subs := strings.Split(reflect.TypeOf(owner).String()[1:], ".")
	ownerModel := subs[len(subs)-1]

	val := reflect.ValueOf(resource).Elem()
	if val.Kind() != reflect.Struct {
		return 0, errors.New("resource is not a struct pointer")
	}

	idField := val.FieldByName(strings.Title(ownerModel) + "ID")
	if !idField.IsValid() || !idField.CanSet() || idField.Kind() != reflect.Int {
		return 0, errors.New(`resource does not have an addressable owner ID int field or the
 field doesn't follow the required 'FooID' naming convention where 'Foo' is the owner type`)
	}
	return int(idField.Int()), nil
}
