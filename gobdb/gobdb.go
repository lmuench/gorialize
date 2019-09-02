package gobdb

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

type Resource interface {
	GetID() int
	SetID(ID int)
}

type DB struct {
	Path      string
	Log       bool
	Encrypted bool
	Key       *[32]byte
}

func NewDB(path string, log bool) *DB {
	db := &DB{
		Path: path,
		Log:  log,
	}
	return db
}

func NewEncryptedDB(path string, log bool, passphrase string) *DB {
	h := HashPassphrase([]byte(passphrase))
	var key32B [32]byte
	copy(key32B[:], h[:32])
	db := &DB{
		Path:      path,
		Log:       log,
		Encrypted: true,
		Key:       &key32B,
	}
	return db
}

type Query struct {
	FatalError    error
	DB            DB
	Operation     string
	Writer        bytes.Buffer
	GobBuffer     []byte
	Model         string
	Resource      Resource
	ID            int
	Counter       int
	CounterPath   string
	MetadataPath  string
	ResourcePath  string
	TablePath     string
	SafeIOPath    bool
	TableFileInfo []os.FileInfo
}

func (q Query) Log() {
	if q.DB.Log {
		fmt.Println()
		fmt.Println("Operation    :", q.Operation)
		fmt.Println("Model        :", q.Model)
		fmt.Println("ID           :", q.ID)
		fmt.Println("Resource     :", q.Resource)
		fmt.Println("Table Path   :", q.TablePath)
		fmt.Println("Fatal Error  :", q.FatalError)
		fmt.Println()
	}
}

func (db DB) NewQueryWithoutID(operation string, resource Resource) *Query {
	return &Query{
		DB:        db,
		Operation: operation,
		Resource:  resource,
	}
}

func (db DB) NewQueryWithID(operation string, resource Resource, id int) *Query {
	return &Query{
		DB:        db,
		Operation: operation,
		Resource:  resource,
		ID:        id,
	}
}

func (db DB) Insert(resource Resource) error {
	mutex.Lock()
	defer mutex.Unlock()

	q := db.NewQueryWithoutID("insert", resource)
	q.ReflectModelNameFromResource()
	q.BuildTablePath()
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

func (db DB) Get(resource Resource, id int) error {
	mutex.Lock()
	defer mutex.Unlock()

	q := db.NewQueryWithID("get", resource, id)
	q.ReflectModelNameFromResource()
	q.BuildTablePath()
	q.ThwartIOBasePathEscape()
	q.ExitIfTableNotExist()
	q.BuildResourcePath()
	q.ReadGobFromDisk()
	q.DecryptGobBuffer()
	q.DecodeGobToResource()
	q.Log()
	return q.FatalError
}

func (db DB) GetAll(resource Resource, callback func(resource interface{})) error {
	mutex.Lock()
	defer mutex.Unlock()
	var err error

	q := db.NewQueryWithoutID("get", resource)
	q.ReflectModelNameFromResource()
	q.BuildTablePath()
	q.ThwartIOBasePathEscape()
	q.ExitIfTableNotExist()
	q.ReadTableFileinfo()
	for _, f := range q.TableFileInfo {
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

func (db DB) Update(resource Resource) error {
	mutex.Lock()
	defer mutex.Unlock()

	q := db.NewQueryWithID("update", resource, resource.GetID())
	q.ReflectModelNameFromResource()
	q.BuildTablePath()
	q.ThwartIOBasePathEscape()
	q.ExitIfTableNotExist()
	q.BuildResourcePath()
	q.ExitIfResourceNotExist()
	q.EncodeResourceToGob()
	q.EncryptGobBuffer()
	q.BuildResourcePath()
	q.WriteGobToDisk()
	q.Log()
	return q.FatalError
}

func (db DB) Upsert(resource Resource) error {
	mutex.Lock()
	defer mutex.Unlock()

	q := db.NewQueryWithID("upsert", resource, resource.GetID())
	q.ReflectModelNameFromResource()
	q.BuildTablePath()
	q.ThwartIOBasePathEscape()
	q.ExitIfTableNotExist()
	q.EncodeResourceToGob()
	q.EncryptGobBuffer()
	q.BuildResourcePath()
	q.WriteGobToDisk()
	q.Log()
	return q.FatalError
}

func (db DB) Delete(resource Resource) error {
	mutex.Lock()
	defer mutex.Unlock()

	q := db.NewQueryWithID("delete", resource, resource.GetID())
	q.ReflectModelNameFromResource()
	q.BuildTablePath()
	q.ThwartIOBasePathEscape()
	q.ExitIfTableNotExist()
	q.BuildResourcePath()
	q.ThwartIOBasePathEscape()
	q.FatalError = os.Remove(q.ResourcePath)
	q.Log()
	return q.FatalError
}

func (db DB) DeleteAll(resource Resource) error {
	mutex.Lock()
	defer mutex.Unlock()
	var err error

	q := db.NewQueryWithoutID("delete", resource)
	q.ReflectModelNameFromResource()
	q.BuildTablePath()
	q.ThwartIOBasePathEscape()
	q.ExitIfTableNotExist()
	q.ReadTableFileinfo()
	for _, f := range q.TableFileInfo {
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
	if !q.SafeIOPath {
		q.FatalError = errors.New("IO path not marked as safe")
		return
	}
	if q.CounterPath == "" {
		q.FatalError = errors.New("Counter path missing")
		return
	}
	b, err := q.DB.ReadFromDisk(q.CounterPath)
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
	q.FatalError = q.DB.WriteToDisk(q.ResourcePath, q.GobBuffer)
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
	q.FatalError = q.DB.WriteToDisk(q.CounterPath, []byte(strconv.Itoa(q.Counter)))
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
	q.GobBuffer, q.FatalError = q.DB.ReadFromDisk(q.ResourcePath)
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

func (q *Query) ReadTableFileinfo() {
	if q.FatalError != nil {
		return
	}
	if !q.SafeIOPath {
		q.FatalError = errors.New("IO path not marked as safe")
		return
	}
	if q.TablePath == "" {
		q.FatalError = errors.New("Table path missing")
		return
	}
	q.TableFileInfo, q.FatalError = ioutil.ReadDir(q.TablePath)
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
	q.FatalError = os.Remove(q.ResourcePath)
}

func (q *Query) ThwartIOBasePathEscape() {
	if !strings.HasPrefix(q.TablePath, q.DB.Path) {
		q.SafeIOPath = false
		q.Log()
		log.Fatal("Thwarted IO operation outside of ", q.DB.Path)
	}
	if strings.Contains(q.TablePath, "..") {
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
	if !q.DB.Encrypted {
		return
	}
	if len(q.GobBuffer) == 0 {
		q.FatalError = errors.New("Gob buffer empty")
		return
	}
	if q.DB.Key == nil {
		q.FatalError = errors.New("Encryption key missing")
	}
	var block cipher.Block
	block, q.FatalError = aes.NewCipher(q.DB.Key[:])
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
	if !q.DB.Encrypted {
		return
	}
	if q.DB.Key == nil {
		q.FatalError = errors.New("Decryption key missing")
	}
	var block cipher.Block
	block, q.FatalError = aes.NewCipher(q.DB.Key[:])
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

func HashPassphrase(passphrase []byte) []byte {
	h := hmac.New(sha512.New512_256, []byte("key"))
	h.Write(passphrase)
	return h.Sum(nil)
}

func (db DB) WriteToDisk(path string, b []byte) error {
	err := ioutil.WriteFile(path, b, 0644)
	return err
}

func (db DB) ReadFromDisk(path string) ([]byte, error) {
	b, err := ioutil.ReadFile(path)
	return b, err
}
