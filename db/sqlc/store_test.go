package db

import (
	"context"
	"testing"
	"time"

	"github.com/shivangp0208/bank_application/util"
	"github.com/stretchr/testify/require"
)

func TestTransferTx(t *testing.T) {
	store := NewStore(testDB)

	account1 := Account{
		ID:        1,
		Owner:     util.GenerateRandomName(),
		Balance:   util.GenerateRandomAmount(),
		Currency:  util.GenerateRandomCurrency(),
		CreatedAt: time.Now(),
	}
	account2 := Account{
		ID:        2,
		Owner:     util.GenerateRandomName(),
		Balance:   util.GenerateRandomAmount(),
		Currency:  util.GenerateRandomCurrency(),
		CreatedAt: time.Now(),
	}

	// run n concurrent transfer transactions
	n := 5
	amount := int64(10)

	errs := make(chan error)
	res := make(chan TransferTxResult)

	for range n {
		go func() {
			result, err := store.TransferTx(context.Background(), TransferTxParams{
				FromAccountID: account1.ID,
				ToAccountID:   account2.ID,
				Amount:        amount,
			})

			errs <- err
			res <- result
		}()
	}

	for range n {
		err := <-errs
		require.NoError(t, err)

		result := <-res
		require.NotEmpty(t, result)

		transfer := result.Transfer
		require.NotEmpty(t, transfer)
		require.Equal(t, account1.ID, transfer.FromAccountID)
		require.Equal(t, account2.ID, transfer.ToAccountID)
		require.Equal(t, amount, transfer.Amount)
		require.NotZero(t, transfer.ID)
		require.NotZero(t, transfer.CreatedAt)
		_, err = store.GetTransfers(context.Background(), transfer.ID)
		require.NoError(t, err)

		fromEntry := result.FromEntry
		require.NotEmpty(t, fromEntry)
		require.Equal(t, account1.ID, fromEntry.AccountID)
		require.Equal(t, -amount, fromEntry.Amount)
		require.NotZero(t, fromEntry.ID)
		require.NotZero(t, fromEntry.CreatedAt)
		_, err = store.GetEntries(context.Background(), fromEntry.ID)
		require.NoError(t, err)

		toEntry := result.ToEntry
		require.NotEmpty(t, toEntry)
		require.Equal(t, account2.ID, toEntry.AccountID)
		require.Equal(t, amount, toEntry.Amount)
		require.NotZero(t, toEntry.ID)
		require.NotZero(t, toEntry.CreatedAt)
		_, err = store.GetEntries(context.Background(), toEntry.ID)
		require.NoError(t, err)
	}
}
