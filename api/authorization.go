package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shivangp0208/bank_application/token"
)

func getPayloadFromToken(c *gin.Context) (payload *token.Payload, err error) {
	payload, ok := c.Value(authorizationPayloadKey).(*token.Payload)
	if !ok {
		myLogger.Error().Msgf("UpdateUser: unable to get the token info from the gin context")
		err = fmt.Errorf("unable to get the token info from the gin context")
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return nil, err
	}
	return payload, nil
}

func authorizeUser(c *gin.Context, username string, payload *token.Payload) (err error) {
	if payload.Username != username {
		myLogger.Error().Str("req_username", username).Str("token_usernam", payload.Username).Msgf("UpdateUser: username mismatch in token and req")
		err = fmt.Errorf("username mismatch in token and req, cannot change other user's info")
		c.JSON(http.StatusUnauthorized, errorResponse(err))
		return err
	}
	return nil
}
