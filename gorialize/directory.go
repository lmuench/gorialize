// Package gorialize is an embedded database that stores Go structs serialized to gobs
package gorialize

import (
	"bufio"
	"log"
	"os"
	"reflect"
	"strconv"
	"sync"
)

var mutex sync.Mutex

// Directory exposes methods to read and write serialized data inside a base directory.
type Directory struct {
	Path         string
	Encrypted    bool
	Key          *[32]byte
	Log          bool
	Index        Index
	IndexLogPath string
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
		Path:         config.Path,
		Log:          config.Log,
		Index:        NewIndex(),
		IndexLogPath: config.Path + "/.idxlog",
	}

	if config.Encrypted {
		var key [32]byte
		h := hashPassphrase([]byte(config.Passphrase))
		copy(key[:], h[:32])
		dir.Key = &key
		dir.Encrypted = true
	}

	dir.ReplayIndexLog()
	return dir
}

func (dir Directory) ReplayIndexLog() {
	f, err := os.Open(dir.IndexLogPath)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		log.Fatal(err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) < 4 {
			log.Fatalf("IndexLog contains unprocessable line: %s", line)
		}
		var id []byte
		var kv string
		for i := len(line) - 1; i >= 0; i-- {
			if line[i] == '=' || line[i] == ':' {
				kv = line[1:i]
				break
			} else {
				id = append(id, line[i])
			}
		}
		ID, err := strconv.Atoi(string(id))
		if err != nil {
			log.Fatalf("IndexLog contains unprocessable line: %s", line)
		}
		op := line[0]
		switch op {
		case '+':
			err := dir.Index.addDirectly(kv, ID)
			if err != nil {
				log.Fatalf("IndexLog contains unprocessable line: %s", line)
			}
		case '-':
			dir.Index.removeDirectly(kv, ID)
		default:
			log.Fatalf("IndexLog contains unprocessable line: %s", line)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
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
	q.UpdateIndex('+')
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

// ReadAll reads all serialized resources of the given slice's elements's type and appends them to the slice.
func (dir Directory) ReadAll(slice interface{}) error {
	slicePtr := reflect.ValueOf(slice)
	sliceVal := reflect.Indirect(slicePtr)
	resourceTyp := reflect.TypeOf(slice).Elem().Elem()
	resourceVal := reflect.New(resourceTyp)
	resource := resourceVal.Interface()

	err := dir.ReadAllCB(resource, func(resource interface{}) {
		resourcePtr := reflect.ValueOf(resource)
		resourceVal := reflect.Indirect(resourcePtr)
		sliceVal.Set(reflect.Append(sliceVal, resourceVal))
	})
	return err
}

// ReadAllCB reads all serialized resources of the given type and calls the provided callback function on each.
func (dir Directory) ReadAllCB(resource interface{}, callback func(resource interface{})) error {
	mutex.Lock()
	defer mutex.Unlock()
	var err error

	q := dir.newQueryWithoutID("read all", resource)
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
func (dir Directory) Find(resource interface{}, clauses ...Where) error {
	mutex.Lock()
	defer mutex.Unlock()

	q := dir.newQueryWithoutID("find", resource)
	q.WhereClauses = clauses
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

// FindAllCB finds all serialized resource of the given type matching all
// provided WHERE clauses and calls the provided callback function on each.
func (dir Directory) FindAllCB(resource interface{}, callback func(resource interface{}), clauses ...Where) error {
	mutex.Lock()
	defer mutex.Unlock()

	q := dir.newQueryWithoutID("find all", resource)
	q.WhereClauses = clauses
	q.ReflectTypeOfResource()
	q.ReflectModelNameFromType()
	q.ApplyWhereClauses(false)
	q.BuildDirPath()
	q.ThwartIOBasePathEscape()
	q.ExitIfDirNotExist()
	for _, id := range q.MatchedIDs {
		q.ID = id
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
// TODO: update index and append to index change log
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
	q.UpdateIndex('x')
	q.Log()
	return q.FatalError
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
	q.DeleteFromDisk()
	q.UpdateIndex('-')
	q.Log()
	return q.FatalError
}

// DeleteAll deletes all serialized resources of the given type.
func (dir Directory) DeleteAll(resource interface{}) error {
	mutex.Lock()
	defer mutex.Unlock()
	var err error

	q := dir.newQueryWithoutID("delete all", resource)
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
		q.UpdateIndex('-')
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
