package db

import (
	"context"
	"testing"

	"github.com/shivangp0208/bank_application/util"
	"github.com/stretchr/testify/require"
)

// single test function to test create, get and delete an transfer
func TestCreateTransfer(t *testing.T) {
	user1, _ := GenerateRandomDBUser()
	user2, _ := GenerateRandomDBUser()
	account1 := GenerateRandomDBAccount(user1)
	account2 := GenerateRandomDBAccount(user2)

	userres1, err := store.CreateUser(context.Background(), CreateUserParams{
		Username:       user1.Username,
		HashedPassword: user1.HashedPassword,
		FullName:       user1.FullName,
		Email:          user1.Email,
	})
	userres2, err := store.CreateUser(context.Background(), CreateUserParams{
		Username:       user2.Username,
		HashedPassword: user2.HashedPassword,
		FullName:       user2.FullName,
		Email:          user2.Email,
	})
	require.NoError(t, err)
	require.NotNil(t, userres1)
	require.NotNil(t, userres2)

	accres1, err := store.CreateAccounts(context.Background(), CreateAccountsParams{
		Owner:    account1.Owner,
		Balance:  account1.Balance,
		Currency: account1.Currency,
	})
	accres2, err := store.CreateAccounts(context.Background(), CreateAccountsParams{
		Owner:    account2.Owner,
		Balance:  account2.Balance,
		Currency: account2.Currency,
	})
	require.NoError(t, err)
	require.NotNil(t, accres1)
	require.NotNil(t, accres2)

	accountId, err := accres1.LastInsertId()
	accountId2, err := accres2.LastInsertId()
	require.NoError(t, err)
	require.NotEmpty(t, accountId)
	require.NotEmpty(t, accountId2)

	// create transfer test
	arg := CreateTransfersParams{
		FromAccountID: uint64(accountId),
		ToAccountID:   uint64(accountId2),
		Amount:        util.GenerateRandomAmount(),
	}

	transferRes, err := store.CreateTransfers(context.Background(), arg)
	require.NoError(t, err)

	transferId, err := transferRes.LastInsertId()
	require.NoError(t, err)
	require.NotEmpty(t, transferId)

	// get transfer test
	transfer, err := store.GetTransfers(context.Background(), uint64(transferId))
	require.NoError(t, err)
	require.NotEmpty(t, transfer)

	require.Equal(t, arg.FromAccountID, transfer.FromAccountID)
	require.Equal(t, arg.ToAccountID, transfer.ToAccountID)
	require.Equal(t, arg.Amount, transfer.Amount)

	require.NotEmpty(t, transfer.ID)
	require.NotEmpty(t, transfer.CreatedAt)

	// delete transfer test
	err = store.DeleteTransfers(context.Background(), uint64(transferId))
	require.NoError(t, err)
}

func TestListTransfer(t *testing.T) {
	// list all transfer test
	transferList, err := store.ListTransfers(context.Background())
	require.NoError(t, err)
	require.NotNil(t, transferList)
}
