# Gorialize
Gorialize is a serialization framework for Go. It aims to provide an embedded persistence layer for applications that do not require all the features of a database. Gorialize lets you serialize your structs and other data types to [gobs](https://golang.org/pkg/encoding/gob/) while organizing the serialized data like a database. It provides a CRUD API that accepts any type that implements the Gorialize `Resource` interface:
```Go
type Resource interface {
	GetID() int
	SetID(ID int)
}
```
