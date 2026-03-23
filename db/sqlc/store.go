package db

import (
	"context"
	"database/sql"
	"fmt"
)

type Store struct {
	*Queries
	db *sql.DB
}

func NewStore(db *sql.DB) Store {
	return Store{
		db:      db,
		Queries: New(db),
	}
}

// execTx is a helper function which starts a transaction and run the function
// passed in it and then take rollback action in case of failure in any transaction
// and at the end commit the transaction in case of success in all transaction
func (store *Store) execTx(ctx context.Context, fn func(*Queries) error) error {
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

// TransferTx performs a money transfer from one account to other.
// It creates a transfer record, add account entries, and update accounts
func (store *Store) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	var result TransferTxResult

	err := store.execTx(ctx, func(q *Queries) error {

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
