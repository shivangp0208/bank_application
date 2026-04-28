package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/shivangp0208/bank_application/util"
)

var logger = util.GetLogger()

type VerifyUserTxParams struct {
	Username   string
	SecretCode string
}

type VerifyUserTxResult struct {
	User User
}

// VerifyUserTx performs a email verification process by validating the expiration and matching the secret code, then updating the detail in the db
func (store *SQLStore) VerifyUserEmailTx(ctx context.Context, arg VerifyUserTxParams) (result VerifyUserTxResult, err error) {

	err = store.execTx(ctx, func(q *Queries) error {
		// first update the verify email info in db to set the is_used to true and make all validation
		emailArg := UpdateVerifyEmailParams{
			Username:   arg.Username,
			SecretCode: arg.SecretCode,
		}
		_, err := q.UpdateVerifyEmail(ctx, emailArg)
		updatedEmail, err := q.GetVerifiyEmailByUsername(ctx, emailArg.Username)
		if err != nil {
			logger.Error().Str("username", emailArg.Username).Msgf("unable to update the email info: %v", err)
			return err
		}
		logger.Info().Msgf("success updating and validating email info %v", updatedEmail)

		// second update the user is_verified to true in case of no error
		userArg := UpdateUserParams{
			Username: arg.Username,
			IsVerified: sql.NullBool{
				Bool:  true,
				Valid: true,
			},
		}
		err = q.UpdateUser(ctx, userArg)
		updatedUser, err := q.GetUser(ctx, arg.Username)
		if err != nil {
			logger.Error().Msgf("unable to update the user info: %v", err)
			return err
		}
		logger.Info().Msgf("success updating and validating user info %v", updatedUser)

		if updatedEmail.Username != updatedUser.Username {
			logger.Error().Msgf("invalid situation where updated email's username %s and updated user's username %s is not matching", updatedEmail.Username, updatedUser.Username)
			return fmt.Errorf("invalid situation where updated email's username %s and updated user's username %s is not matching", updatedEmail.Username, updatedUser.Username)
		}

		result = VerifyUserTxResult{
			User: updatedUser,
		}
		logger.Info().Str("email", updatedUser.Email).Msgf("successfully verified user %s", arg.Username)

		return nil
	})

	return result, err
}
