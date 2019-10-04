# Gorialize
Gorialize is a serialization framework for Go. It aims to provide an embedded persistence layer for applications that do not require all the features of a database. Gorialize lets you serialize your structs and other data types to [gobs](https://golang.org/pkg/encoding/gob/) while organizing the serialized data like a database. It provides a CRUD API that accepts any struct with an addressable `ID` field of type `int`. Those types of structs have to be passed to Gorialize's methods by reference and are named `resource` in the method header.

#### Example Resource Type
```Go
type User struct {
	ID   int     // <-- required field
	Name string
}
```

#### Directory
```Go
type Directory struct {
    Path      string
    Log       bool
    Encrypted bool
    Key       *[32]byte
}
```
Directory exposes methods to read and write serialized data inside a base directory.

#### DirectoryConfig
```Go
type DirectoryConfig struct {
    Path       string
    Encrypted  bool
    Passphrase string
    Log        bool
}
```
DirectoryConfig holds parameters to be passed to NewDirectory().

#### NewDirectory
```Go
func NewDirectory(config DirectoryConfig) *Directory
```
NewDirectory returns a new Directory struct for the given configuration.

#### Create
```Go
func (dir Directory) Create(resource interface{}) error
```
Create creates a new serialized resource and sets its ID.

#### Read
```Go
func (dir Directory) Read(resource interface{}, id int) error
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
func (dir Directory) Replace(resource interface{}) error
```
Replace replaces a serialized resource

#### Update
```Go
func (dir Directory) Update(resource interface{}, id int) error
```
Update partially updates a serialized resource with all non-zero values of the given resource.

#### Delete
```Go
func (dir Directory) Delete(resource interface{}) error
```
Delete deletes a serialized resource.

#### DeleteAll
```Go
func (dir Directory) DeleteAll(resource interface{}) error
```
DeleteAll deletes all serialized resources of the given type.
