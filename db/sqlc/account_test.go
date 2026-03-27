package db

import (
	"context"
	"testing"

	"github.com/shivangp0208/bank_application/util"
	"github.com/stretchr/testify/require"
)

// single test function to test create, get and delete an account
func TestCreateAccounts(t *testing.T) {
	// create account test
	arg := CreateAccountsParams{
		Owner:    util.GenerateRandomName(10),
		Balance:  util.GenerateRandomAmount(),
		Currency: util.GenerateRandomCurrency(),
	}

	accountRes, err := testQueries.CreateAccounts(context.Background(), arg)
	require.NoError(t, err)

	accountId, err := accountRes.LastInsertId()
	require.NoError(t, err)
	require.NotEmpty(t, accountId)

	// get account test
	account, err := testQueries.GetAccount(context.Background(), uint64(accountId))
	require.NoError(t, err)
	require.NotEmpty(t, account)

	require.Equal(t, arg.Owner, account.Owner)
	require.Equal(t, arg.Balance, account.Balance)
	require.Equal(t, arg.Currency, account.Currency)

	require.NotEmpty(t, account.ID)
	require.NotEmpty(t, account.CreatedAt)

	// delete account test
	err = testQueries.DeleteAccounts(context.Background(), uint64(accountId))
	require.NoError(t, err)
}

func TestListAccounts(t *testing.T) {
	// list all accounts test
	accountsList, err := testQueries.ListAccounts(context.Background())
	require.NoError(t, err)
	require.NotNil(t, accountsList)
}
