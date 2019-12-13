// Package gorialize is an embedded database that stores Go structs serialized to gobs
package gorialize

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
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
)

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
	MatchedIDs   []int
	IndexUpdates []string
}

type Where struct {
	Field string
	Value interface{}
}

func (q Query) Log() {
	if q.Dir.Log {
		fmt.Println()
		fmt.Println("Operation     :", q.Operation)
		fmt.Println("Model         :", q.Model)
		fmt.Println("ID            :", q.ID)
		fmt.Println("Resource      :", q.Resource)
		fmt.Println("Directory     :", q.DirPath)
		fmt.Println("Fatal Error   :", q.FatalError)
		fmt.Println("Index Updates :", q.IndexUpdates)
		fmt.Println()
	}
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

func (q *Query) UpdateIndex(operator rune) {
	if q.FatalError != nil {
		return
	}
	if operator != '+' && operator != '-' {
		q.FatalError = errors.New("Unknown index operator")
	}
	if q.ResourceType == nil {
		q.FatalError = errors.New("Resource type missing")
		return
	}
	f, err := os.OpenFile(q.Dir.IndexLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		q.FatalError = err
		return
	}
	defer f.Close()

	for i := 0; i < q.ResourceType.Elem().NumField(); i++ {
		field := q.ResourceType.Elem().Field(i)
		tag := field.Tag.Get("gorialize")
		if tag == "indexed" {
			value := reflect.Indirect(
				reflect.ValueOf(q.Resource),
			).FieldByName(field.Name).Interface()

			logEntry := fmt.Sprintf("%c%s:%s:%v=%d", operator, q.Model, field.Name, value, q.ID)
			_, err = f.WriteString(logEntry + "\n")
			if err != nil {
				q.FatalError = err
				return
			}
			q.IndexUpdates = append(q.IndexUpdates, logEntry)

			if operator == '+' {
				q.Dir.Index.appendID(q.Model, field.Name, value, q.ID)
			} else {
				q.Dir.Index.removeID(q.Model, field.Name, value, q.ID)
			}
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
	b, err := readFromDisk(q.CounterPath)
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
	q.FatalError = writeToDisk(q.ResourcePath, q.GobBuffer)
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
	q.FatalError = writeToDisk(q.CounterPath, []byte(strconv.Itoa(q.Counter)))
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
	q.GobBuffer, q.FatalError = readFromDisk(q.ResourcePath)
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
	q.FatalError = deleteFromDisk(q.ResourcePath)
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
		return
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
		return
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
	whereClauseCount := len(q.WhereClauses)
	if whereClauseCount == 0 {
		q.FatalError = errors.New("Where clauses missing")
		return
	}
	idCountMap := make(map[int]int, 1)
	for _, clause := range q.WhereClauses {
		ids, ok := q.Dir.Index.getIDs(q.Model, clause.Field, clause.Value)
		if ok {
			for _, id := range ids {
				idCountMap[id] += 1
				if pickFirst {
					if idCountMap[id] == whereClauseCount {
						q.MatchedIDs = append(q.MatchedIDs, id)
						q.ID = int(id)
						return
					}
				}
			}
		}
	}
	for id, v := range idCountMap {
		if v == whereClauseCount {
			q.MatchedIDs = append(q.MatchedIDs, id)
		}
	}
	if len(q.MatchedIDs) == 0 {
		q.FatalError = errors.New("No matching where clauses")
	}
}
