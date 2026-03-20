package db

import (
	"context"
	"testing"

	"github.com/shivangp0208/bank_application/util"
	"github.com/stretchr/testify/require"
)

// single test function to test create, get and delete an transfer
func TestCreateTransfer(t *testing.T) {
	// create transfer test
	arg := CreateTransfersParams{
		FromAccountID: 1,
		ToAccountID:   2,
		Amount:        util.GenerateRandomAmount(),
	}

	transferRes, err := testQueries.CreateTransfers(context.Background(), arg)
	require.NoError(t, err)

	transferId, err := transferRes.LastInsertId()
	require.NoError(t, err)
	require.NotEmpty(t, transferId)

	// get transfer test
	transfer, err := testQueries.GetTransfers(context.Background(), uint64(transferId))
	require.NoError(t, err)
	require.NotEmpty(t, transfer)

	require.Equal(t, arg.FromAccountID, transfer.FromAccountID)
	require.Equal(t, arg.ToAccountID, transfer.ToAccountID)
	require.Equal(t, arg.Amount, transfer.Amount)

	require.NotEmpty(t, transfer.ID)
	require.NotEmpty(t, transfer.CreatedAt)

	// delete transfer test
	err = testQueries.DeleteTransfers(context.Background(), uint64(transferId))
	require.NoError(t, err)
}

func TestListTransfer(t *testing.T) {
	// list all transfer test
	transferList, err := testQueries.ListTransfers(context.Background())
	require.NoError(t, err)
	require.NotNil(t, transferList)
}
