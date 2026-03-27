package util

import (
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestPassword(t *testing.T) {
	pass := GenerateRandomName(10)

	hashedPass, err := GenerateHashPassword(pass)
	require.NoError(t, err)
	require.NotEmpty(t, hashedPass)

	err = ComparePasswords(hashedPass, pass)
	require.NoError(t, err)

	wrongPass := GenerateRandomName(10)
	err = ComparePasswords(hashedPass, wrongPass)
	require.EqualError(t, err, bcrypt.ErrMismatchedHashAndPassword.Error())

	hashedPass1, err := GenerateHashPassword(pass)
	require.NoError(t, err)
	require.NotEmpty(t, hashedPass1)
	require.NotEqual(t, hashedPass, hashedPass1)
}
