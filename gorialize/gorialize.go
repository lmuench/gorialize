// Package gorialize is a serialization framework for Go. It aims to provide an embedded persistence layer
// for applications that do not require all the features of a database. Gorialize lets you serialize
// your structs and other data types to gobs while organizing the serialized data like database.
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

// Resource is the interface any type that should be serialized with gorialize has to implement.
type Resource interface {
	GetID() int
	SetID(ID int)
}

// Directory exposes methods to read and write serialized data inside a base directory.
type Directory struct {
	Path      string
	Log       bool
	Encrypted bool
	Key       *[32]byte
}

// NewDirectory returns a new unencrypted directory.
func NewDirectory(path string, log bool) *Directory {
	dir := &Directory{
		Path: path,
		Log:  log,
	}
	return dir
}

// NewDirectory returns a new encrypted directory.
func NewEncryptedDirectory(path string, log bool, passphrase string) *Directory {
	h := hashPassphrase([]byte(passphrase))
	var key32B [32]byte
	copy(key32B[:], h[:32])
	dir := &Directory{
		Path:      path,
		Log:       log,
		Encrypted: true,
		Key:       &key32B,
	}
	return dir
}

type Query struct {
	FatalError   error
	Dir          Directory
	Operation    string
	Writer       bytes.Buffer
	GobBuffer    []byte
	Model        string
	Resource     Resource
	ID           int
	Counter      int
	CounterPath  string
	MetadataPath string
	ResourcePath string
	DirPath      string
	SafeIOPath   bool
	DirFileInfo  []os.FileInfo
}

func (q Query) Log() {
	if q.Dir.Log {
		fmt.Println()
		fmt.Println("Operation    :", q.Operation)
		fmt.Println("Model        :", q.Model)
		fmt.Println("ID           :", q.ID)
		fmt.Println("Resource     :", q.Resource)
		fmt.Println("Directory    :", q.DirPath)
		fmt.Println("Fatal Error  :", q.FatalError)
		fmt.Println()
	}
}

func (dir Directory) newQueryWithoutID(operation string, resource Resource) *Query {
	return &Query{
		Dir:       dir,
		Operation: operation,
		Resource:  resource,
	}
}

func (dir Directory) newQueryWithID(operation string, resource Resource, id int) *Query {
	return &Query{
		Dir:       dir,
		Operation: operation,
		Resource:  resource,
		ID:        id,
	}
}

// Create creates a new serialized resource and sets its ID.
func (dir Directory) Create(resource Resource) error {
	mutex.Lock()
	defer mutex.Unlock()

	q := dir.newQueryWithoutID("create", resource)
	q.ReflectModelNameFromResource()
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
	q.Log()
	return q.FatalError
}

// Read reads the serialized resource with the given ID.
func (dir Directory) Read(resource Resource, id int) error {
	mutex.Lock()
	defer mutex.Unlock()

	q := dir.newQueryWithID("read", resource, id)
	q.ReflectModelNameFromResource()
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

// readFromCustomSubdirectory reads the serialized resource with the given ID from a custom subdirectory.
// This method is intended for testing purposes.
func (dir Directory) readFromCustomSubdirectory(resource Resource, id int, subdir string) error {
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

// ReadAll reads all serialized resource of the given type and calls the provided callback function on each.
func (dir Directory) ReadAll(resource Resource, callback func(resource interface{})) error {
	mutex.Lock()
	defer mutex.Unlock()
	var err error

	q := dir.newQueryWithoutID("read", resource)
	q.ReflectModelNameFromResource()
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

// Replace replaces a serialized resource.
func (dir Directory) Replace(resource Resource) error {
	mutex.Lock()
	defer mutex.Unlock()

	q := dir.newQueryWithID("replace", resource, resource.GetID())
	q.ReflectModelNameFromResource()
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

// func (dir Directory) CreateOrReplace(resource Resource) error {
// 	mutex.Lock()
// 	defer mutex.Unlock()

// 	q := dir.newQueryWithID("create or replace", resource, resource.GetID())
// 	q.ReflectModelNameFromResource()
// 	q.BuildDirPath()
// 	q.ThwartIOBasePathEscape()
// 	q.ExitIfDirNotExist()
// 	q.EncodeResourceToGob()
// 	q.EncryptGobBuffer()
// 	q.BuildResourcePath()
// 	q.WriteGobToDisk()
// 	q.Log()
// 	return q.FatalError
// }

// Delete deletes a serialized resource.
func (dir Directory) Delete(resource Resource) error {
	mutex.Lock()
	defer mutex.Unlock()

	q := dir.newQueryWithID("delete", resource, resource.GetID())
	q.ReflectModelNameFromResource()
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
func (dir Directory) DeleteAll(resource Resource) error {
	mutex.Lock()
	defer mutex.Unlock()
	var err error

	q := dir.newQueryWithoutID("delete", resource)
	q.ReflectModelNameFromResource()
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
func (dir Directory) ResetCounter(resource Resource) error {
	mutex.Lock()
	defer mutex.Unlock()

	q := dir.newQueryWithoutID("reset counter", resource)
	q.ReflectModelNameFromResource()
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
	q.Resource.SetID(q.ID)
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
