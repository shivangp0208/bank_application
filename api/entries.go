package api

import (
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strconv"

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
	var formReq ListDataReq

	pageNo := c.Query("pageNo")
	pageSize := c.Query("pageSize")

	var err error
	if pageNo == "" || pageSize == "" {
		formReq.PageNo = 1
		formReq.PageSize = 10
	} else {
		formReq.PageNo, err = strconv.Atoi(pageNo)
		formReq.PageSize, err = strconv.Atoi(pageSize)
		if err != nil {
			c.JSON(http.StatusBadRequest, err)
			return
		}
	}

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
		Limit:     int32(formReq.PageSize),
		Offset:    int32((formReq.PageNo) * (formReq.PageSize + 1)),
	}

	entryList, err := s.Store.ListEntriesByAccountIdAndUsername(c, arg)
	if err != nil {
		err = errors.Join(fmt.Errorf("error getting the list of entries: "), err)
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	c.JSON(http.StatusOK, entryList)
}

func (s *Server) GetAllEntries(c *gin.Context) {
	var req ListDataReq
	pageNo := c.Query("pageNo")
	pageSize := c.Query("pageSize")

	var err error
	if pageNo == "" || pageSize == "" {
		req.PageNo = 1
		req.PageSize = 10
	} else {
		req.PageNo, err = strconv.Atoi(pageNo)
		req.PageSize, err = strconv.Atoi(pageSize)
		if err != nil {
			c.JSON(http.StatusBadRequest, err)
			return
		}
	}

	authPayload := c.MustGet(authorizationPayloadKey).(*token.Payload)

	if authPayload.Role != util.Accountant {
		err := fmt.Errorf("user %s is not allowed to see the entries for all accounts", authPayload.Username)
		c.JSON(http.StatusForbidden, errorResponse(err))
		return
	}

	arg := db.ListEntriesParams{
		Limit:  int32(req.PageSize),
		Offset: int32((req.PageNo) * (req.PageSize + 1)),
	}
	entryList, err := s.Store.ListEntries(c, arg)
	if err != nil {
		err = errors.Join(fmt.Errorf("error getting the list of entries: "), err)
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	c.JSON(http.StatusOK, entryList)
}
