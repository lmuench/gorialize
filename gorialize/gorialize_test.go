package gorialize

import (
	"reflect"
	"testing"

	"syreclabs.com/go/faker"
)

const testIterationCount = 10

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

type userV2 struct {
	ID        int
	Name      string
	Birthdate string
}

func (self *userV2) GetID() int {
	return self.ID
}

func (self *userV2) SetID(ID int) {
	self.ID = ID
}

var dir *Directory

func beforeEach() {
	dir = NewEncryptedDirectory(
		"/tmp/gorialize/gorialize_test",
		false,
		"password123",
	)

	_ = dir.DeleteAll(&user{})
}

func afterEach() {
	_ = dir.DeleteAll(&user{})
}

func TestCreateAndRead(t *testing.T) {
	beforeEach()

	for i := 0; i < testIterationCount; i++ {
		newUser := &user{
			Name: faker.Name().Name(),
			Age:  uint(faker.Number().NumberInt(2)),
		}
		err := dir.Create(newUser)
		if err != nil {
			t.Fatal(err)
		}

		serializedUser := &user{}
		err = dir.Read(serializedUser, newUser.GetID())
		if err != nil {
			t.Fatal(err)
		}

		if serializedUser.ID != newUser.ID {
			t.Fatal("IDs don't equal")
		}
		if serializedUser.Name != newUser.Name {
			t.Fatal("Names don't equal")
		}
		if serializedUser.Age != newUser.Age {
			t.Fatal("Ages don't equal")
		}
	}

	afterEach()
}

func TestReplace(t *testing.T) {
	beforeEach()

	for i := 0; i < testIterationCount; i++ {
		newUser := &user{
			Name: faker.Name().Name(),
			Age:  uint(faker.Number().NumberInt(2)),
		}
		err := dir.Create(newUser)
		if err != nil {
			t.Fatal(err)
		}

		serializedUser := &user{}
		err = dir.Read(serializedUser, newUser.GetID())
		if err != nil {
			t.Fatal(err)
		}

		newName := faker.Name().Name()
		newAge := uint(faker.Number().NumberInt(2))

		serializedUser.Name = newName
		serializedUser.Age = newAge

		err = dir.Replace(serializedUser)
		if err != nil {
			t.Fatal(err)
		}

		updatedUser := &user{}
		err = dir.Read(updatedUser, serializedUser.GetID())
		if err != nil {
			t.Fatal(err)
		}

		if updatedUser.Name != newName {
			t.Fatal("Names don't equal")
		}
		if updatedUser.Age != newAge {
			t.Fatal("Ages don't equal")
		}
	}

	afterEach()
}

func TestDelete(t *testing.T) {
	beforeEach()

	for i := 0; i < testIterationCount; i++ {
		newUser := &user{
			Name: faker.Name().Name(),
			Age:  uint(faker.Number().NumberInt(2)),
		}
		err := dir.Create(newUser)
		if err != nil {
			t.Fatal(err)
		}

		serializedUser := &user{}
		err = dir.Read(serializedUser, newUser.GetID())
		if err != nil {
			t.Fatal(err)
		}

		newName := faker.Name().Name()
		newAge := uint(faker.Number().NumberInt(2))

		serializedUser.Name = newName
		serializedUser.Age = newAge

		err = dir.Delete(serializedUser)
		if err != nil {
			t.Fatal(err)
		}

		err = dir.Read(&user{}, serializedUser.GetID())
		if err == nil {
			t.Fatal("Resource should have been deleted but was not:", *serializedUser)
		}
	}

	afterEach()
}

func TestReadAll(t *testing.T) {
	beforeEach()
	err := dir.ResetCounter(&user{})
	if err != nil {
		t.Fatal(err)
	}

	newUsers := []user{}

	for i := 0; i < testIterationCount; i++ {
		newUser := user{
			Name: faker.Name().Name(),
			Age:  uint(faker.Number().NumberInt(2)),
		}
		err = dir.Create(&newUser)
		if err != nil {
			t.Fatal(err)
		}
		newUsers = append(newUsers, newUser)
	}

	unorderedNewUsers := make(map[int]user)
	for _, u := range newUsers {
		unorderedNewUsers[u.GetID()] = u
	}

	unorderedSerializedUsers := make(map[int]user)
	err = dir.ReadAll(&user{}, func(resource interface{}) {
		user := *resource.(*user)
		unorderedSerializedUsers[user.GetID()] = user
	})
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(unorderedNewUsers, unorderedSerializedUsers) {
		t.Fatal("Users don't equal")
	}

	serializedUsers := []user{}
	err = dir.ReadAll(&user{}, func(resource interface{}) {
		serializedUsers = append(serializedUsers, *resource.(*user))
	})
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(newUsers, serializedUsers) {
		t.Fatal("Users are not in the correct order")
	}

	afterEach()
}

func TestReadAllIntoSlice(t *testing.T) {
	beforeEach()
	err := dir.ResetCounter(&user{})
	if err != nil {
		t.Fatal(err)
	}

	newUsers := []user{}

	for i := 0; i < testIterationCount; i++ {
		newUser := user{
			Name: faker.Name().Name(),
			Age:  uint(faker.Number().NumberInt(2)),
		}
		err = dir.Create(&newUser)
		if err != nil {
			t.Fatal(err)
		}
		newUsers = append(newUsers, newUser)
	}

	serializedUsers := []user{}
	err = dir.ReadAllIntoSlice(&serializedUsers)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(newUsers, serializedUsers) {
		t.Fatal("Users are not in the correct order")
	}

	afterEach()
}

func TestReadAfterResourceFieldsChanged(t *testing.T) {
	beforeEach()

	for i := 0; i < testIterationCount; i++ {
		newUser := &user{
			Name: faker.Name().Name(),
			Age:  uint(faker.Number().NumberInt(2)),
		}
		err := dir.Create(newUser)
		if err != nil {
			t.Fatal(err)
		}

		serializedUser := &userV2{}
		err = dir.readFromCustomSubdirectory(serializedUser, newUser.GetID(), "gorialize.user")
		if err != nil {
			t.Fatal(err)
		}

		if serializedUser.ID != newUser.ID {
			t.Fatal("IDs don't equal")
		}
		if serializedUser.Name != newUser.Name {
			t.Fatal("Names don't equal")
		}
		if serializedUser.Birthdate != "" {
			t.Fatal("Added field birthdate should be empty")
		}
	}

	afterEach()
}
