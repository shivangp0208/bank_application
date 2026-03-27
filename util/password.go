package util

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func GenerateHashPassword(pass string) (hashedPass string, err error) {

	passByte, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password %s: %w", pass, err)
	}
	return string(passByte), err
}

func ComparePasswords(expectedHashPass string, gotPass string) error {
	return bcrypt.CompareHashAndPassword([]byte(expectedHashPass), []byte(gotPass))
}
