package db

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransferTx(t *testing.T) {

	// TODO: Instead of creating user, account for transfer transaction test, think of standardizing this
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

	fmt.Println(">> before:", account1.Balance, account2.Balance)

	t.Run("Concurrent Transfer", TestTransferTxDeadlocks)

	// run n concurrent transfer transactions
	n := 10
	amount := int64(10)

	errs := make(chan error)
	res := make(chan TransferTxResult)
	existed := make(map[int]bool)

	for range n {
		go func() {
			result, err := store.TransferTx(context.Background(), TransferTxParams{
				FromAccountID: uint64(accountId),
				ToAccountID:   uint64(accountId2),
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

		fmt.Println()
		// checking transfer status
		transfer := result.Transfer
		require.NotEmpty(t, transfer)
		require.Equal(t, uint64(accountId), transfer.FromAccountID)
		require.Equal(t, uint64(accountId2), transfer.ToAccountID)
		require.Equal(t, amount, transfer.Amount)
		require.NotZero(t, transfer.ID)
		require.NotZero(t, transfer.CreatedAt)
		_, err = store.GetTransfers(context.Background(), transfer.ID)
		require.NoError(t, err)

		// checking entries status
		fromEntry := result.FromEntry
		require.NotEmpty(t, fromEntry)
		require.Equal(t, uint64(accountId), fromEntry.AccountID)
		require.Equal(t, -amount, fromEntry.Amount)
		require.NotZero(t, fromEntry.ID)
		require.NotZero(t, fromEntry.CreatedAt)
		_, err = store.GetEntries(context.Background(), fromEntry.ID)
		require.NoError(t, err)

		toEntry := result.ToEntry
		require.NotEmpty(t, toEntry)
		require.Equal(t, uint64(accountId2), toEntry.AccountID)
		require.Equal(t, amount, toEntry.Amount)
		require.NotZero(t, toEntry.ID)
		require.NotZero(t, toEntry.CreatedAt)
		_, err = store.GetEntries(context.Background(), toEntry.ID)
		require.NoError(t, err)

		// checking account status
		fromAccount := result.FromAccount
		require.NotEmpty(t, fromAccount)
		require.Equal(t, uint64(accountId), fromAccount.ID)

		toAccount := result.ToAccount
		require.NotEmpty(t, toAccount)
		require.Equal(t, uint64(accountId2), toAccount.ID)

		// checking account balance
		diff1 := account1.Balance - fromAccount.Balance
		diff2 := toAccount.Balance - account2.Balance
		require.Equal(t, diff1, diff2)
		require.True(t, diff1 > 0)
		require.True(t, diff1%amount == 0)

		k := int(diff1 / amount)
		require.True(t, k >= 1 && k <= n)
		require.NotContains(t, existed, k)
		existed[k] = true
	}

	// check for update account balance
	updatedAccount1, err := store.GetAccount(context.Background(), uint64(accountId))
	require.NoError(t, err)

	updatedAccount2, err := store.GetAccount(context.Background(), uint64(accountId2))
	require.NoError(t, err)

	fmt.Println(">> after:", updatedAccount1.Balance, updatedAccount2.Balance)

	require.Equal(t, account1.Balance-(int64(n)*amount), updatedAccount1.Balance)
	require.Equal(t, account2.Balance+(int64(n)*amount), updatedAccount2.Balance)
}

func TestTransferTxDeadlocks(t *testing.T) {

	account1, err := store.GetAccount(context.Background(), 3)
	if err != nil {
		log.Fatalf("unable to get the sender account with id %v", 3)
	}
	account2, err := store.GetAccount(context.Background(), 4)
	if err != nil {
		log.Fatalf("unable to get the receiver account with id %v", 4)
	}
	fmt.Println(">> before:", account1.Balance, account2.Balance)

	// run n concurrent transfer transactions
	n := 10
	amount := int64(10)

	errs := make(chan error)

	for i := range n {
		fromAccountId := account1.ID
		toAccountId := account2.ID

		if i%2 == 1 {
			fromAccountId = account2.ID
			toAccountId = account1.ID
		}

		go func() {
			_, err := store.TransferTx(context.Background(), TransferTxParams{
				FromAccountID: fromAccountId,
				ToAccountID:   toAccountId,
				Amount:        amount,
			})

			errs <- err
		}()
	}

	for range n {
		err := <-errs
		require.NoError(t, err)
	}

	// check for update account balance
	updatedAccount1, err := store.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)

	updatedAccount2, err := store.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)

	require.Equal(t, account1.Balance, updatedAccount1.Balance)
	require.Equal(t, account2.Balance, updatedAccount2.Balance)
}
