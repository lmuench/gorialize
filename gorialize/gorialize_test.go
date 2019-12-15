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

type userV3 struct {
	ID   int
	Name string `gorialize:"indexed"`
	Age  uint   `gorialize:"indexed"`
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
		Log:        false,
	})

	_ = dir.DeleteAll(&user{})
	_ = dir.DeleteAll(&userV2{})
	_ = dir.DeleteAll(&userV3{})
	_ = dir.DeleteAll(&todoList{})
	_ = dir.DeleteAll(&todoItem{})
}

func afterEach() {
	_ = dir.DeleteAll(&user{})
	_ = dir.DeleteAll(&userV2{})
	_ = dir.DeleteAll(&userV3{})
	_ = dir.DeleteAll(&todoList{})
	_ = dir.DeleteAll(&todoItem{})
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

func TestReadAllCB(t *testing.T) {
	beforeEach()

	newUsers := []user{}
	for i := 0; i < testIterationCount; i++ {
		newUser := user{
			Name: faker.Name().Name(),
			Age:  uint(faker.Number().NumberInt(2)),
		}
		err := dir.Create(&newUser)
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
	err := dir.ReadAllCB(&user{}, func(resource interface{}) {
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
	err = dir.ReadAllCB(&user{}, func(resource interface{}) {
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

func TestReadAll(t *testing.T) {
	beforeEach()

	newUsers := []user{}
	for i := 0; i < testIterationCount; i++ {
		newUser := user{
			Name: faker.Name().Name(),
			Age:  uint(faker.Number().NumberInt(2)),
		}
		err := dir.Create(&newUser)
		if err != nil {
			t.Fatal(err)
		}
		newUsers = append(newUsers, newUser)
	}

	serializedUsers := []user{}
	err := dir.ReadAll(&serializedUsers)
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

		ids := dir.Index.getIDs("gorialize.userV3", "Name", name)
		if len(ids) == 0 {
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

func TestIndexAfterDelete(t *testing.T) {
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

		serializedUser := &userV3{}
		err = dir.Read(serializedUser, newUser.ID)
		if err != nil {
			t.Fatal(err)
		}

		ids := dir.Index.getIDs("gorialize.userV3", "Name", name)
		if len(ids) == 0 {
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

		err = dir.Delete(serializedUser)
		if err != nil {
			t.Fatal(err)
		}

		ids2 := dir.Index.getIDs("gorialize.userV3", "Name", name)
		if len(ids2) > 0 {
			t.Fatal("Index still contains entry")
		}
	}

	afterEach()
}

func TestFindCB(t *testing.T) {
	beforeEach()
	newUsers := []userV3{}
	for _, age := range []int{17, 36, 23, 56, 19, 23} {
		user := userV3{
			Name: faker.Name().Name(),
			Age:  uint(age),
		}
		err := dir.Create(&user)
		if err != nil {
			t.Fatal(err)
		}
		newUsers = append(newUsers, user)
	}

	serializedUsers := []userV3{}
	err := dir.FindCB(&userV3{}, func(resource interface{}) {
		serializedUsers = append(serializedUsers, *resource.(*userV3))
	}, Where{Field: "Age", Equals: 23})
	if err != nil {
		t.Fatal(err)
	}

	actualCnt := len(serializedUsers)
	expectedCnt := 2
	if actualCnt != expectedCnt {
		t.Fatalf("Found: %d, expected: %d", actualCnt, expectedCnt)
	}

	expectedUsersMap := map[int]userV3{
		newUsers[2].ID: newUsers[2],
		newUsers[5].ID: newUsers[5],
	}
	serializedUsersMap := map[int]userV3{
		serializedUsers[0].ID: serializedUsers[0],
		serializedUsers[1].ID: serializedUsers[1],
	}
	if !reflect.DeepEqual(expectedUsersMap, serializedUsersMap) {
		t.Fatal("Found users don't match expected users")
	}

	afterEach()
}

func TestFind(t *testing.T) {
	beforeEach()

	newUsers := []userV3{}
	for _, age := range []int{17, 36, 23, 56, 19, 23} {
		user := userV3{
			Name: faker.Name().Name(),
			Age:  uint(age),
		}
		err := dir.Create(&user)
		if err != nil {
			t.Fatal(err)
		}
		newUsers = append(newUsers, user)
	}

	serializedUsers := []userV3{}
	err := dir.Find(&serializedUsers, Where{Field: "Age", Equals: 23})
	if err != nil {
		t.Fatal(err)
	}

	actualCnt := len(serializedUsers)
	expectedCnt := 2
	if actualCnt != expectedCnt {
		t.Fatalf("Found: %d, expected: %d", actualCnt, expectedCnt)
	}

	expectedUsersMap := map[int]userV3{
		newUsers[2].ID: newUsers[2],
		newUsers[5].ID: newUsers[5],
	}
	serializedUsersMap := map[int]userV3{
		serializedUsers[0].ID: serializedUsers[0],
		serializedUsers[1].ID: serializedUsers[1],
	}
	if !reflect.DeepEqual(expectedUsersMap, serializedUsersMap) {
		t.Fatal("Found users don't match expected users")
	}

	afterEach()
}

func TestFindWithANDedWhereClauses(t *testing.T) {
	beforeEach()

	newUsers := []userV3{}
	for _, age := range []int{17, 36, 23, 56, 19, 23} {
		user := userV3{
			Name: "John Doe",
			Age:  uint(age),
		}
		err := dir.Create(&user)
		if err != nil {
			t.Fatal(err)
		}
		newUsers = append(newUsers, user)
	}

	serializedUsers := []userV3{}
	err := dir.Find(&serializedUsers,
		Where{Field: "Name", Equals: "John Doe", And: &Where{Field: "Age", Equals: 23}},
	)
	if err != nil {
		t.Fatal(err)
	}

	actualCnt := len(serializedUsers)
	expectedCnt := 2
	if actualCnt != expectedCnt {
		t.Fatalf("Found: %d, expected: %d", actualCnt, expectedCnt)
	}

	expectedUsersMap := map[int]userV3{
		newUsers[2].ID: newUsers[2],
		newUsers[5].ID: newUsers[5],
	}
	serializedUsersMap := map[int]userV3{
		serializedUsers[0].ID: serializedUsers[0],
		serializedUsers[1].ID: serializedUsers[1],
	}
	if !reflect.DeepEqual(expectedUsersMap, serializedUsersMap) {
		t.Fatal("Found users don't match expected users")
	}

	afterEach()
}

func TestFindWithORedWhereClauses(t *testing.T) {
	beforeEach()

	newUsers := []userV3{}
	for _, age := range []int{17, 36, 23, 56, 19, 23} {
		user := userV3{
			Name: "John Doe",
			Age:  uint(age),
		}
		err := dir.Create(&user)
		if err != nil {
			t.Fatal(err)
		}
		newUsers = append(newUsers, user)
	}

	serializedUsers := []userV3{}
	err := dir.Find(&serializedUsers,
		Where{Field: "Age", Equals: 23},
		Where{Field: "Age", Equals: 17},
	)
	if err != nil {
		t.Fatal(err)
	}

	actualCnt := len(serializedUsers)
	expectedCnt := 3
	if actualCnt != expectedCnt {
		t.Fatalf("Found: %d, expected: %d", actualCnt, expectedCnt)
	}

	expectedUsersMap := map[int]userV3{
		newUsers[0].ID: newUsers[0],
		newUsers[2].ID: newUsers[2],
		newUsers[5].ID: newUsers[5],
	}
	serializedUsersMap := map[int]userV3{
		serializedUsers[0].ID: serializedUsers[0],
		serializedUsers[1].ID: serializedUsers[1],
		serializedUsers[2].ID: serializedUsers[2],
	}
	if !reflect.DeepEqual(expectedUsersMap, serializedUsersMap) {
		t.Fatal("Found users don't match expected users")
	}

	afterEach()
}


func TestFindWithIntPassedAsString(t *testing.T) {
	beforeEach()

	newUsers := []userV3{}
	for _, age := range []int{17, 36, 23, 56, 19, 23} {
		user := userV3{
			Name: "John Doe",
			Age:  uint(age),
		}
		err := dir.Create(&user)
		if err != nil {
			t.Fatal(err)
		}
		newUsers = append(newUsers, user)
	}

	serializedUsers := []userV3{}
	err := dir.Find(&serializedUsers,
		Where{Field: "Name", Equals: "John Doe", And: &Where{Field: "Age", Equals: "23"}},
	)
	if err != nil {
		t.Fatal(err)
	}

	actualCnt := len(serializedUsers)
	expectedCnt := 2
	if actualCnt != expectedCnt {
		t.Fatalf("Found: %d, expected: %d", actualCnt, expectedCnt)
	}

	expectedUsersMap := map[int]userV3{
		newUsers[2].ID: newUsers[2],
		newUsers[5].ID: newUsers[5],
	}
	serializedUsersMap := map[int]userV3{
		serializedUsers[0].ID: serializedUsers[0],
		serializedUsers[1].ID: serializedUsers[1],
	}
	if !reflect.DeepEqual(expectedUsersMap, serializedUsersMap) {
		t.Fatal("Found users don't match expected users")
	}

	afterEach()
}
