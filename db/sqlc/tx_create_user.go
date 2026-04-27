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
func (store *SQLStore) CreateUserTx(ctx context.Context, arg CreateUserTxParams) (result CreateUserTxResult, err error) {

	err = store.execTx(ctx, func(q *Queries) error {

		_, err = q.CreateUser(ctx, arg.User)
		createdUser, err := q.GetUser(ctx, arg.User.Username)
		if err != nil {
			return fmt.Errorf("unable to get the created user: %v", err.Error())
		}
		logger.Info().Str("username", arg.User.Username).Str("email", arg.User.Email).Str("full_name", arg.User.FullName).Msg("successfully created user in db")

		result = CreateUserTxResult{User: createdUser}

		return arg.AfterCreateUser(createdUser)
	})

	return result, err
}
