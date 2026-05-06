package db

import (
	"context"
	"testing"

	"github.com/shivangp0208/bank_application/util"
	"github.com/stretchr/testify/require"
)

// single test function to test create, get and delete an entry
func TestCreateEntries(t *testing.T) {

	user, _ := GenerateRandomDBUser()
	account := GenerateRandomDBAccount(user)

	// create entry test
	res, err := store.CreateUser(context.Background(), CreateUserParams{
		Username:       user.Username,
		HashedPassword: user.HashedPassword,
		FullName:       user.FullName,
		Email:          user.Email,
	})
	require.NoError(t, err)
	require.NotNil(t, res)

	res, err = store.CreateAccounts(context.Background(), CreateAccountsParams{
		Owner:    account.Owner,
		Balance:  account.Balance,
		Currency: account.Currency,
	})
	require.NoError(t, err)
	require.NotNil(t, res)

	accountId, err := res.LastInsertId()
	require.NoError(t, err)
	require.NotEmpty(t, accountId)

	arg := CreateEntriesParams{
		AccountID: uint64(accountId),
		Amount:    util.GenerateRandomAmount(),
	}
	entriyRes, err := store.CreateEntries(context.Background(), arg)
	require.NoError(t, err)

	entryId, err := entriyRes.LastInsertId()
	require.NoError(t, err)
	require.NotEmpty(t, entryId)

	// get entry test
	entry, err := store.GetEntries(context.Background(), uint64(entryId))
	require.NoError(t, err)
	require.NotEmpty(t, entry)

	require.Equal(t, arg.AccountID, entry.AccountID)
	require.Equal(t, arg.Amount, entry.Amount)

	require.NotEmpty(t, entry.ID)
	require.NotEmpty(t, entry.CreatedAt)

	// delete entry test
	err = store.DeleteEntries(context.Background(), uint64(entryId))
	require.NoError(t, err)
}

func TestListEntries(t *testing.T) {
	// list all accounts test
	entryList, err := store.ListEntries(context.Background())
	require.NoError(t, err)
	require.NotNil(t, entryList)
}
