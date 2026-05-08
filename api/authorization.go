package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shivangp0208/bank_application/token"
)

func authorizeUser(c *gin.Context, username string, payload *token.Payload) (err error) {
	if payload.Username != username {
		myLogger.Error().Str("req_username", username).Str("token_usernam", payload.Username).Msgf("UpdateUser: username mismatch in token and req")
		err = fmt.Errorf("account doesn't belong to the authenticated user, username mismatch, unauthorized")
		c.JSON(http.StatusForbidden, errorResponse(err))
		return err
	}
	return nil
}
