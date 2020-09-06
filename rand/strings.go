package rand

import (
	"crypto/rand"
	"encoding/base64"
)

// RememberTokenBytes  is the constant used for generating remember tokens
const RememberTokenBytes = 32

// RememberToken is a helper function designed to generate remember tokens
// of a predetermined byte size.
func RememberToken() (string, error) {
	return String(RememberTokenBytes)
}

// Bytes generates n random bytes or return an error.
// This function uses crypto/rand package to safely remember tokens.
func Bytes(n int) ([]byte, error) {
	b := make([]byte, n)

	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

// String generates a byte slice of size nBytes and
// return a string that is the base64 URL encoded version
// of the byte slice.
func String(nBytes int) (string, error) {
	b, err := Bytes(nBytes)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(b), nil
}
