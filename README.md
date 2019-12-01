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

#### Find
Find reads the first serialized resources matching the given WHERE clauses
```Go
func (dir Directory) Find(resource interface{}, whereClauses ...Where) error {
```

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

#### GetOwner
```Go
func (dir Directory) GetOwner(resource interface{}, owner interface{}) error
```
GetOwner reads the serialized resource which owns the given resource.
The resource needs to have an addressable owner ID int field which
follows a 'FooID' naming convention where 'Foo' is the owner type.
