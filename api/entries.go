package api

import (
	"errors"
	"fmt"
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
	db "github.com/shivangp0208/bank_application/db/sqlc"
	"github.com/shivangp0208/bank_application/token"
	"github.com/shivangp0208/bank_application/util"
)

type AccountId struct {
	AccountID uint64 `uri:"id" binding:"required,min=1"`
}

func (s *Server) GetAllEntryForAccountID(c *gin.Context) {
	var req AccountId

	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	authPayload := c.MustGet(authorizationPayloadKey).(*token.Payload)

	// if user is not accountant he should be only able to see his own entries
	if authPayload.Role != util.Accountant {
		accountList, err := s.Store.ListAllAccountIdByUsername(c, authPayload.Username)
		if err != nil {
			err = errors.Join(fmt.Errorf("error getting the list of accounts for username %s: ", authPayload.Username), err)
			c.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}
		if !slices.Contains(accountList, req.AccountID) {
			err := fmt.Errorf("user %s is not allowed to see the entries for this account", authPayload.Username)
			c.JSON(http.StatusForbidden, errorResponse(err))
			return
		}
	}

	arg := db.ListEntriesByAccountIdAndUsernameParams{
		Username:  authPayload.Username,
		AccountID: req.AccountID,
	}

	entryList, err := s.Store.ListEntriesByAccountIdAndUsername(c, arg)
	if err != nil {
		err = errors.Join(fmt.Errorf("error getting the list of entries: "), err)
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	c.JSON(http.StatusOK, entryList)
}

func (s *Server) GetAllEntryByToAccount(c *gin.Context) {}

func (s *Server) GetAllEntries(c *gin.Context) {
	authPayload := c.MustGet(authorizationPayloadKey).(*token.Payload)

	if authPayload.Role != util.Accountant {
		err := fmt.Errorf("user %s is not allowed to see the entries for all accounts", authPayload.Username)
		c.JSON(http.StatusForbidden, errorResponse(err))
		return
	}

	entryList, err := s.Store.ListEntries(c)
	if err != nil {
		err = errors.Join(fmt.Errorf("error getting the list of entries: "), err)
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	c.JSON(http.StatusOK, entryList)
}
