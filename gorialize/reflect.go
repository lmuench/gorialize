// Package gorialize is an embedded database that stores Go structs serialized to gobs
package gorialize

import (
	"errors"
	"reflect"
	"strings"
)

func getID(resource interface{}) (int, error) {
	val := reflect.ValueOf(resource).Elem()
	if val.Kind() != reflect.Struct {
		return 0, errors.New("resource is not a struct pointer")
	}

	idField := val.FieldByName("ID")
	if !idField.IsValid() || !idField.CanSet() || idField.Kind() != reflect.Int {
		return 0, errors.New("resource does not have an addressable ID int field")
	}
	return int(idField.Int()), nil
}

func setID(resource interface{}, id int) error {
	val := reflect.ValueOf(resource).Elem()
	if val.Kind() != reflect.Struct {
		return errors.New("resource is not a struct pointer")
	}

	idField := val.FieldByName("ID")
	if !idField.IsValid() || !idField.CanSet() || idField.Kind() != reflect.Int {
		return errors.New("resource does not have an addressable ID int field")
	}
	idField.SetInt(int64(id))
	return nil
}

func getOwnerID(resource interface{}, owner interface{}) (int, error) {
	if reflect.ValueOf(owner).Elem().Kind() != reflect.Struct {
		return 0, errors.New("owner is not a struct pointer")
	}
	subs := strings.Split(reflect.TypeOf(owner).String()[1:], ".")
	ownerModel := subs[len(subs)-1]

	val := reflect.ValueOf(resource).Elem()
	if val.Kind() != reflect.Struct {
		return 0, errors.New("resource is not a struct pointer")
	}

	idField := val.FieldByName(strings.Title(ownerModel) + "ID")
	if !idField.IsValid() || !idField.CanSet() || idField.Kind() != reflect.Int {
		return 0, errors.New(`resource does not have an addressable owner ID int field or the
 field doesn't follow the required 'FooID' naming convention where 'Foo' is the owner type`)
	}
	return int(idField.Int()), nil
}
