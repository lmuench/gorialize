package gorialize

import (
	"testing"
)

type user struct {
	ID   int
	Name string
	Age  uint
}

func (self *user) GetID() int {
	return self.ID
}

func (self *user) SetID(ID int) {
	self.ID = ID
}

var dir *Directory

func beforeEach() {
	dir = NewEncryptedDirectory(
		"/tmp/gorialize/gorialize_test",
		true,
		"password123",
	)

	_ = dir.DeleteAll(&user{})
}

func TestCreateAndRead(t *testing.T) {
	beforeEach()

	newUser := &user{
		Name: "John Doe",
		Age:  42,
	}
	err := dir.Create(newUser)
	if err != nil {
		t.Error(err)
	}

	serializedUser := &user{}
	err = dir.Read(serializedUser, newUser.GetID())
	if err != nil {
		t.Error(err)
	}

	if serializedUser.ID != newUser.ID {
		t.Error("IDs don't equal")
	}
	if serializedUser.Name != newUser.Name {
		t.Error("Names don't equal")
	}
	if serializedUser.Age != newUser.Age {
		t.Error("Ages don't equal")
	}
}
