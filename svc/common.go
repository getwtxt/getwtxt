package svc

import "golang.org/x/crypto/bcrypt"

// HashPass returns the bcrypt hash of the provided string.
// If an empty string is provided, return an empty string.
func HashPass(s string) (string, error) {
	if s == "" {
		return "", nil
	}
	h, err := bcrypt.GenerateFromPassword([]byte(s), 14)
	if err != nil {
		return "", err
	}
	return string(h), nil
}
