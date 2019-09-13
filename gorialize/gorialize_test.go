package gorialize

import (
	"testing"

	"syreclabs.com/go/faker"
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

	for i := 0; i < 100; i++ {
		newUser := &user{
			Name: faker.Name().Name(),
			Age:  uint(faker.Number().NumberInt(2)),
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
}

func TestReplace(t *testing.T) {
	beforeEach()

	for i := 0; i < 100; i++ {
		newUser := &user{
			Name: faker.Name().Name(),
			Age:  uint(faker.Number().NumberInt(2)),
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

		newName := faker.Name().Name()
		newAge := uint(faker.Number().NumberInt(2))

		serializedUser.Name = newName
		serializedUser.Age = newAge

		err = dir.Replace(serializedUser)
		if err != nil {
			t.Error(err)
		}

		updatedUser := &user{}
		err = dir.Read(updatedUser, serializedUser.GetID())
		if err != nil {
			t.Error(err)
		}

		if updatedUser.Name != newName {
			t.Error("Names don't equal")
		}
		if updatedUser.Age != newAge {
			t.Error("Ages don't equal")
		}
	}
}