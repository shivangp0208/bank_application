package db

import (
	"context"
	"fmt"
)

type TransferTxParams struct {
	FromAccountID uint64 `json:"from_account_id"`
	ToAccountID   uint64 `json:"to_account_id"`
	Amount        int64  `json:"amount"`
}

type TransferTxResult struct {
	Transfer    Transfer `json:"transfer"`
	FromAccount Account  `json:"from_account"`
	FromEntry   Entry    `json:"from_entry"`
	ToAccount   Account  `json:"to_account"`
	ToEntry     Entry    `json:"to_entry"`
}

// execTx is a helper function which starts a transaction and run the function
// passed in it and then take rollback action in case of failure in any transaction
// and at the end commit the transaction in case of success in all transaction
func (store *SQLStore) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	queries := New(tx)
	err = fn(queries)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("rollback err : %v and tx err : %v", rbErr, err)
		}
		return err
	}

	return tx.Commit()
}

// TransferTx performs a money transfer from one account to other.
// It creates a transfer record, add account entries, and update accounts
func (store *SQLStore) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	var result TransferTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		// earlier the account updation was at bottom but due to that deadlock was occuring as
		// in mysql the update query in transaction is alreasy blocking which waits for 50sec
		// till other transaction completes it's updation so due to that i moved up the account
		// updation logic so that lock does not have to wait that much now

		// to counter the deadlock in this situation, the best way is to update the account balance
		// in consistent order

		// checking for accounts status
		if arg.FromAccountID < arg.ToAccountID {
			err := q.AddAccountBalance(ctx, AddAccountBalanceParams{
				ID:     arg.FromAccountID,
				Amount: -arg.Amount,
			})
			if err != nil {
				return err
			}
			result.FromAccount, err = q.GetAccount(ctx, arg.FromAccountID)
			if err != nil {
				return err
			}

			err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
				ID:     arg.ToAccountID,
				Amount: arg.Amount,
			})
			if err != nil {
				return err
			}
			result.ToAccount, err = q.GetAccount(ctx, arg.ToAccountID)
			if err != nil {
				return err
			}

		} else {
			err := q.AddAccountBalance(ctx, AddAccountBalanceParams{
				ID:     arg.ToAccountID,
				Amount: arg.Amount,
			})
			if err != nil {
				return err
			}
			result.ToAccount, err = q.GetAccount(ctx, arg.ToAccountID)
			if err != nil {
				return err
			}
			err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
				ID:     arg.FromAccountID,
				Amount: -arg.Amount,
			})
			if err != nil {
				return err
			}
			result.FromAccount, err = q.GetAccount(ctx, arg.FromAccountID)
			if err != nil {
				return err
			}
		}

		transferRes, err := q.CreateTransfers(ctx, CreateTransfersParams{
			FromAccountID: arg.FromAccountID,
			ToAccountID:   arg.ToAccountID,
			Amount:        arg.Amount,
		})
		if err != nil {
			return err
		}
		transferId, err := transferRes.LastInsertId()
		if err != nil {
			return err
		}
		result.Transfer, err = q.GetTransfers(ctx, uint64(transferId))
		if err != nil {
			return err
		}

		fromEntryRes, err := q.CreateEntries(ctx, CreateEntriesParams{
			AccountID: arg.FromAccountID,
			Amount:    -arg.Amount,
		})
		if err != nil {
			return err
		}
		fromEntryId, err := fromEntryRes.LastInsertId()
		if err != nil {
			return err
		}
		result.FromEntry, err = q.GetEntries(ctx, uint64(fromEntryId))
		if err != nil {
			return err
		}

		toEntryRes, err := q.CreateEntries(ctx, CreateEntriesParams{
			AccountID: arg.ToAccountID,
			Amount:    arg.Amount,
		})
		if err != nil {
			return err
		}
		toEntryId, err := toEntryRes.LastInsertId()
		if err != nil {
			return err
		}
		result.ToEntry, err = q.GetEntries(ctx, uint64(toEntryId))
		if err != nil {
			return err
		}

		return nil
	})

	return result, err
}
