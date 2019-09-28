# Gorialize
Gorialize is a serialization framework for Go. It aims to provide an embedded persistence layer for applications that do not require all the features of a database. Gorialize lets you serialize your structs and other data types to [gobs](https://golang.org/pkg/encoding/gob/) while organizing the serialized data like a database. It provides a CRUD API that accepts any type that implements the Gorialize `Resource` interface:
```Go
type Resource interface {
	GetID() int
	SetID(ID int)
}
```

## API

### Directory
```Go
type Directory struct {
    Path      string
    Log       bool
    Encrypted bool
    Key       *[32]byte
}
```
Directory exposes methods to read and write serialized data inside a base directory.

#### NewDirectory
```Go
func NewDirectory(path string, log bool) *Directory
```
NewDirectory returns a new unencrypted directory.

#### NewEncryptedDirectory
```Go
func NewEncryptedDirectory(path string, log bool, passphrase string) *Directory
```
NewDirectory returns a new encrypted directory.

#### Create
```Go
func (dir Directory) Create(resource Resource) error
```
Create creates a new serialized resource and sets its ID.

#### Read
```Go
func (dir Directory) Read(resource Resource, id int) error
```
Read reads the serialized resource with the given ID.

#### ReadAllIntoSlice
```Go
func (dir Directory) ReadAllIntoSlice(slice interface{}) error {
```
ReadAllIntoSlice reads all serialized resources of the given slice's elements's type and writes them into the slice.

#### ReadAll
```Go
func (dir Directory) ReadAll(resource interface{}, callback func(resource interface{})) error
```
ReadAll reads all serialized resource of the given type and calls the provided callback function on each.

#### Replace
```Go
func (dir Directory) Replace(resource Resource) error
```
Replace replaces a serialized resource

#### Delete
```Go
func (dir Directory) Delete(resource Resource) error
```
Delete deletes a serialized resource.

#### DeleteAll
```Go
func (dir Directory) DeleteAll(resource Resource) error
```
DeleteAll deletes all serialized resources of the given type.
