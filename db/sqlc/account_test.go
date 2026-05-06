package db

import (
	"context"
	"testing"

	"github.com/shivangp0208/bank_application/util"
	"github.com/stretchr/testify/require"
)

// single test function to test create, get and delete an account
func TestCreateAccounts(t *testing.T) {
	user, _ := GenerateRandomDBUser()
	// create account test
	arg := CreateAccountsParams{
		Owner:    user.Username,
		Balance:  util.GenerateRandomAmount(),
		Currency: util.GenerateRandomCurrency(),
	}

	res, err := store.CreateUser(context.Background(), CreateUserParams{
		Username:       user.Username,
		HashedPassword: user.HashedPassword,
		FullName:       user.FullName,
		Email:          user.Email,
	})
	require.NoError(t, err)
	require.NotNil(t, res)

	accountRes, err := store.CreateAccounts(context.Background(), arg)
	require.NoError(t, err)

	accountId, err := accountRes.LastInsertId()
	require.NoError(t, err)
	require.NotEmpty(t, accountId)

	// get account test
	account, err := store.GetAccount(context.Background(), uint64(accountId))
	require.NoError(t, err)
	require.NotEmpty(t, account)

	require.Equal(t, arg.Owner, account.Owner)
	require.Equal(t, arg.Balance, account.Balance)
	require.Equal(t, arg.Currency, account.Currency)

	require.NotEmpty(t, account.ID)
	require.NotEmpty(t, account.CreatedAt)

	// delete account test
	err = store.DeleteAccounts(context.Background(), uint64(accountId))
	require.NoError(t, err)
}

func TestListAccounts(t *testing.T) {
	// list all accounts test
	accountsList, err := store.ListAccounts(context.Background())
	require.NoError(t, err)
	require.NotNil(t, accountsList)
}
