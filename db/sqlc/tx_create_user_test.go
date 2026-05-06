package db

import (
	"context"
	"testing"

	"github.com/shivangp0208/bank_application/util"
	"github.com/stretchr/testify/require"
)

func TestCreateUserTx(t *testing.T) {
	user, userPass := GenerateRandomDBUser()

	testCases := []struct {
		name       string
		argument   CreateUserTxParams
		testResult func(t *testing.T, res CreateUserTxResult, err error)
	}{{
		name: "OK",
		argument: CreateUserTxParams{
			User: CreateUserParams{
				Username:       user.Username,
				HashedPassword: user.HashedPassword,
				FullName:       user.FullName,
				Email:          user.Email,
			},
			AfterCreateUser: func(user User) error {
				return nil
			},
		},
		testResult: func(t *testing.T, res CreateUserTxResult, err error) {
			require.NoError(t, err)
			require.NotNil(t, res)
			require.Equal(t, res.User.Username, user.Username)
			require.NoError(t, err)
			err = util.ComparePasswords(res.User.HashedPassword, userPass)
			require.NoError(t, err)
			require.Equal(t, res.User.HashedPassword, user.HashedPassword)
		},
	}}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := store.CreateUserTx(context.Background(), tc.argument)
			tc.testResult(t, res, err)
		})
	}
}
