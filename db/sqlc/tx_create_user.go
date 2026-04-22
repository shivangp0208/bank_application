package db

import (
	"context"
	"fmt"
)

type CreateUserTxParams struct {
	User            CreateUserParams
	AfterCreateUser func(user User) error
}

type CreateUserTxResult struct {
	User User
}

// CreateUserTx performs a creating user with sending verification email to the request user's email in a single transaction.
func (store *SQLStore) CreateUserTx(ctx context.Context, arg CreateUserTxParams) (CreateUserTxResult, error) {
	var result CreateUserTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		_, err = q.CreateUser(ctx, arg.User)
		if err != nil {
			return err
		}

		createdUser, err := q.GetUser(ctx, arg.User.Username)
		if err != nil {
			return fmt.Errorf("unable to get the created user: %v", err.Error())
		}

		result = CreateUserTxResult{User: createdUser}

		return arg.AfterCreateUser(createdUser)
	})

	return result, err
}
