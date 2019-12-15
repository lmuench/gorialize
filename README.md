# Gorialize
Gorialize is an embedded database that stores Go structs serialized to [gobs](https://golang.org/pkg/encoding/gob/).

#### Example Resource Type
```Go
type User struct {
    ID   int  // <-- required field
    Name string `gorialize:"indexed"`
}
```

#### Directory
```Go
type Directory struct {
    Path      string
    Encrypted bool
    Key       *[32]byte
    Log       bool
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

#### Where
```Go
type Where struct {
    Field Field
    Value Value
}
```
Where clauses can be passed to Find()

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

#### ReadAll
```Go
func (dir Directory) ReadAll(slice interface{}) error
```
ReadAll reads all serialized resources of the given slice's element type and appends them to the slice.

#### ReadAllCB
```Go
func (dir Directory) ReadAllCB(resource interface{}, callback func(resource interface{})) error
```
ReadAllCB reads all serialized resources of the given type and calls the provided callback function on each.

#### Find
Find reads the first serialized resources matching the given WHERE clauses
```Go
func (dir Directory) Find(resource interface{}, clauses ...Where) error
```

## FindAllCB
FindAllCB finds all serialized resource of the given type matching all
provided WHERE clauses and calls the provided callback function on each.
```Go
func (dir Directory) FindAllCB(resource interface{}, callback func(resource interface{}), clauses ...Where) error
```

#### Replace
```Go
func (dir Directory) Replace(resource interface{}) error
```
Replace replaces a serialized resource

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

#### GetOwner
```Go
func (dir Directory) GetOwner(resource interface{}, owner interface{}) error
```
GetOwner reads the serialized resource which owns the given resource.
The resource needs to have an addressable owner ID int field which
follows a 'FooID' naming convention where 'Foo' is the owner type.
