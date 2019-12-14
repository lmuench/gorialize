// Package gorialize is an embedded database that stores Go structs serialized to gobs
package gorialize

import (
	"io/ioutil"
	"os"
)

func writeToDisk(path string, b []byte) error {
	err := ioutil.WriteFile(path, b, 0644)
	return err
}

func readFromDisk(path string) ([]byte, error) {
	b, err := ioutil.ReadFile(path)
	return b, err
}

func deleteFromDisk(path string) error {
	err := os.Remove(path)
	return err
}
