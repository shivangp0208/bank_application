package util

import "golang.org/x/crypto/bcrypt"

func GenerateHashPassword(pass string) (hashedPass string, err error) {

	passByte, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		return
	}
	return string(passByte), err
}

func ComparePasswords(expectedPass string, gotPass string) error {
	return bcrypt.CompareHashAndPassword([]byte(expectedPass), []byte(gotPass))
}
