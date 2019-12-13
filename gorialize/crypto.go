// Package gorialize is an embedded database that stores Go structs serialized to gobs
package gorialize

import (
	"crypto/hmac"
	"crypto/sha512"
)

func hashPassphrase(passphrase []byte) []byte {
	h := hmac.New(sha512.New512_256, []byte("key"))
	_, _ = h.Write(passphrase) // TODO find out if returned err should be checked
	return h.Sum(nil)
}
