package db

import (
	"context"
	"testing"

	"github.com/shivangp0208/bank_application/util"
	"github.com/stretchr/testify/require"
)

// single test function to test create, get and delete an entry
func TestCreateEntries(t *testing.T) {
	// create entry test
	arg := CreateEntriesParams{
		AccountID: 1,
		Amount:    util.GenerateRandomAmount(),
	}

	entriyRes, err := testQueries.CreateEntries(context.Background(), arg)
	require.NoError(t, err)

	entryId, err := entriyRes.LastInsertId()
	require.NoError(t, err)
	require.NotEmpty(t, entryId)

	// get entry test
	entry, err := testQueries.GetEntries(context.Background(), uint64(entryId))
	require.NoError(t, err)
	require.NotEmpty(t, entry)

	require.Equal(t, arg.AccountID, entry.AccountID)
	require.Equal(t, arg.Amount, entry.Amount)

	require.NotEmpty(t, entry.ID)
	require.NotEmpty(t, entry.CreatedAt)

	// delete entry test
	err = testQueries.DeleteEntries(context.Background(), uint64(entryId))
	require.NoError(t, err)
}

func TestListEntries(t *testing.T) {
	// list all accounts test
	entryList, err := testQueries.ListEntries(context.Background())
	require.NoError(t, err)
	require.NotNil(t, entryList)
}
