package gorialize

import (
	"reflect"
	"testing"

	"syreclabs.com/go/faker"
)

const testIterationCount = 3

type user struct {
	ID   int
	Name string
	Age  uint
}

type userV2 struct {
	ID        int
	Name      string
	Birthdate string
}

type todoList struct {
	ID    int
	Title string
}

type todoItem struct {
	ID         int
	TodoListID int
	Text       string
}

var dir *Directory

func beforeEach() {
	dir = NewDirectory(DirectoryConfig{
		Path:       "/tmp/gorialize/gorialize_test",
		Encrypted:  true,
		Passphrase: "password123",
		Log:        true,
	})

	_ = dir.DeleteAll(&user{})
}

func afterEach() {
	_ = dir.DeleteAll(&user{})
}

func TestGetID(t *testing.T) {
	beforeEach()

	u := &user{}
	id, err := getID(u)
	if err != nil {
		t.Fatal(err)
	}
	if id != u.ID {
		t.Fatal("getID does not return correct ID")
	}

	afterEach()
}

func TestSetID(t *testing.T) {
	beforeEach()

	u := &user{}
	err := setID(u, 42)
	if err != nil {
		t.Fatal(err)
	}
	if u.ID != 42 {
		t.Fatal("setID does not set ID correctly")
	}

	afterEach()
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
		err = dir.Read(serializedUser, newUser.ID)
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
		err = dir.Read(serializedUser, newUser.ID)
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
		err = dir.Read(updatedUser, serializedUser.ID)
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

func TestUpdate(t *testing.T) {
	beforeEach()

	for i := 0; i < testIterationCount; i++ {
		userAgeOnly := &user{
			Age: uint(faker.Number().NumberInt(2)),
		}
		err := dir.Create(userAgeOnly)
		if err != nil {
			t.Fatal(err)
		}

		userNameOnly := &user{
			Name: faker.Name().Name(),
		}

		err = dir.Update(userNameOnly, userAgeOnly.ID)
		if err != nil {
			t.Fatal(err)
		}

		updatedUser := &user{}
		err = dir.Read(updatedUser, userAgeOnly.ID)
		if err != nil {
			t.Fatal(err)
		}

		if updatedUser.Age != userAgeOnly.Age {
			t.Fatal("Ages don't equal")
		}
		if updatedUser.Name != userNameOnly.Name {
			t.Fatal("Names don't equal")
		}
	}

	for i := 0; i < testIterationCount; i++ {
		userAgeAndName := &user{
			Age:  uint(faker.Number().NumberInt(2)),
			Name: faker.Name().Name(),
		}
		err := dir.Create(userAgeAndName)
		if err != nil {
			t.Fatal(err)
		}

		userNameOnly := &user{
			Name: faker.Name().Name(),
		}

		err = dir.Update(userNameOnly, userAgeAndName.ID)
		if err != nil {
			t.Fatal(err)
		}

		updatedUser := &user{}
		err = dir.Read(updatedUser, userAgeAndName.ID)
		if err != nil {
			t.Fatal(err)
		}

		if updatedUser.Age != userAgeAndName.Age {
			t.Fatal("Ages don't equal")
		}
		if updatedUser.Name != userNameOnly.Name {
			t.Fatal("Names don't equal")
		}
	}

	for i := 0; i < testIterationCount; i++ {
		userAgeAndName := &user{
			Age:  uint(faker.Number().NumberInt(2)),
			Name: faker.Name().Name(),
		}
		err := dir.Create(userAgeAndName)
		if err != nil {
			t.Fatal(err)
		}

		userAgeOnly := &user{
			Age: uint(faker.Number().NumberInt(2)),
		}

		err = dir.Update(userAgeOnly, userAgeAndName.ID)
		if err != nil {
			t.Fatal(err)
		}

		updatedUser := &user{}
		err = dir.Read(updatedUser, userAgeAndName.ID)
		if err != nil {
			t.Fatal(err)
		}

		if updatedUser.Age != userAgeOnly.Age {
			t.Fatal("Ages don't equal")
		}
		if updatedUser.Name != userAgeAndName.Name {
			t.Fatal("Names don't equal")
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
		err = dir.Read(serializedUser, newUser.ID)
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

		err = dir.Read(&user{}, serializedUser.ID)
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
		unorderedNewUsers[u.ID] = u
	}

	unorderedSerializedUsers := make(map[int]user)
	err = dir.ReadAll(&user{}, func(resource interface{}) {
		user := *resource.(*user)
		unorderedSerializedUsers[user.ID] = user
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
		err = dir.readFromCustomSubdirectory(serializedUser, newUser.ID, "gorialize.user")
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

func TestGetOwner(t *testing.T) {
	beforeEach()

	for i := 0; i < testIterationCount; i++ {
		newTodoList := &todoList{
			Title: faker.Lorem().Sentence(3),
		}
		err := dir.Create(newTodoList)
		if err != nil {
			t.Fatal(err)
		}

		if newTodoList.ID == 0 {
			t.Fatal("ID still has zero-value after dir.Create()")
		}

		newTodoItem := &todoItem{
			TodoListID: newTodoList.ID,
			Text:       faker.Lorem().Sentence(7),
		}
		err = dir.Create(newTodoItem)
		if err != nil {
			t.Fatal(err)
		}

		serializedTodoList := &todoList{}
		err = dir.GetOwner(newTodoItem, serializedTodoList)
		if err != nil {
			t.Fatal(err)
		}

		if serializedTodoList.ID != newTodoList.ID {
			t.Fatal("IDs don't equal")
		}
		if serializedTodoList.Title != newTodoList.Title {
			t.Fatal("Titles don't equal")
		}
	}

	afterEach()
}

type userV3 struct {
	ID   int
	Name string `gorialize:"indexed"`
	Age  uint
}

func TestIndexAfterCreate(t *testing.T) {
	beforeEach()

	for i := 0; i < testIterationCount; i++ {
		name := faker.Name().Name()
		newUser := &userV3{
			Name: name,
			Age:  uint(faker.Number().NumberInt(2)),
		}
		err := dir.Create(newUser)
		if err != nil {
			t.Fatal(err)
		}

		ids, ok := dir.Index.getIDs("gorialize.userV3", "Name", name)
		if !ok {
			t.Fatal("Index doesn't contain entry")
		}
		if len(ids) < 1 {
			t.Fatal("Index points to empty ID slice")
		}
		// TODO: Test duplicate names.
		// Below would fail if faker would randomly create the same name twice.
		if ids[0] != newUser.ID {
			t.Fatal("Indexed name doesn't point to correct ID")
		}
	}

	afterEach()
}

func TestFind(t *testing.T) {
	beforeEach()

	for i := 0; i < testIterationCount; i++ {
		name := faker.Name().Name()
		newUser := &userV3{
			Name: name,
			Age:  uint(faker.Number().NumberInt(2)),
		}
		err := dir.Create(newUser)
		if err != nil {
			t.Fatal(err)
		}

		foundUser := &userV3{}
		err = dir.Find(foundUser, Where{Field: "Name", Value: name})
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(foundUser, newUser) {
			t.Fatal("Users don't equal")
		}
	}

	afterEach()
}
