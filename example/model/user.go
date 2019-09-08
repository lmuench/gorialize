package model

import "github.com/lmuench/gorialize/gorialize"

// User implements gorialize.Resource interface
type User struct {
	ID   int
	Name string
	Age  int
}

func (self *User) GetID() int {
	return self.ID
}

func (self *User) SetID(ID int) {
	self.ID = ID
}

// SELECT * FROM USERS
func GetAllUsersMap(db *gorialize.Directory) (map[int]User, error) {
	users := make(map[int]User)

	err := db.ReadAll(&User{}, func(resource interface{}) {
		user := *resource.(*User)
		users[user.GetID()] = user
	})
	return users, err
}

// SELECT * FROM USERS
func GetAllUsers(db *gorialize.Directory) ([]User, error) {
	users := []User{}

	err := db.ReadAll(&User{}, func(resource interface{}) {
		user := *resource.(*User)
		users = append(users, user)
	})
	return users, err
}

// SELECT * FROM USERS WHERE AGE > 42
func GetAllUsersOver42(db *gorialize.Directory) ([]User, error) {
	users := []User{}

	err := db.ReadAll(&User{}, func(resource interface{}) {
		user := *resource.(*User)
		if user.Age > 42 {
			users = append(users, user)
		}
	})
	return users, err
}
