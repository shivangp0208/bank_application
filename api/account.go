package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	db "github.com/shivangp0208/bank_application/db/sqlc"
)

type CreateAccountReq struct {
	Owner    string `json:"owner" binding:"required,min=2"`
	Currency string `json:"currency" binding:"required,oneof=INR USD EUR CAD"`
}

func (s *Server) CreateAccount(c *gin.Context) {
	var req CreateAccountReq

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.CreateAccountsParams{
		Owner:    req.Owner,
		Balance:  0,
		Currency: req.Currency,
	}

	res, err := s.store.CreateAccounts(c, arg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	accountId, err := res.LastInsertId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	acount, err := s.store.GetAccount(c, uint64(accountId))
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	c.JSON(http.StatusCreated, acount)
}

type GetAccountReq struct {
	ID uint64 `uri:"id" binding:"required,min=1"`
}

func (s *Server) GetAccountByID(c *gin.Context) {

	var req GetAccountReq

	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	account, err := s.store.GetAccount(c, req.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse(err))
		return
	}

	c.JSON(http.StatusOK, account)
}

type ListAccountsReq struct {
	PageNo   int `form:"pageNo"`
	PageSize int `form:"pageSize"`
}

func (s *Server) GetAllAccount(c *gin.Context) {
	var req ListAccountsReq

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

	arg := db.ListPagedAccountsParams{
		Limit:  int32(req.PageSize),
		Offset: int32((req.PageNo - 1) * req.PageSize),
	}

	accounts, err := s.store.ListPagedAccounts(c, arg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, accounts)
}

type UpdateAccountReq struct {
	ID       uint64 `uri:"id" binding:"required,min=1"`
	Owner    string `json:"owner" binding:"required,min=2"`
	Balance  int64  `json:"balance" binding:"required,min=0"`
	Currency string `json:"currency" binding:"required,oneof=INR USD EUR CAD"`
}

func (s *Server) UpdateAccount(c *gin.Context) {

	var req UpdateAccountReq

	if err := c.ShouldBindUri(req.ID); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.UpdateAccountParams{
		ID:       req.ID,
		Owner:    req.Owner,
		Balance:  req.Balance,
		Currency: req.Currency,
	}

	if err := s.store.UpdateAccount(c, arg); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	updatedAccount, err := s.store.GetAccount(c, req.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	c.JSON(http.StatusCreated, updatedAccount)
}

type DeleteAccountReq struct {
	ID uint64 `uri:"id" binding:"required,min=1"`
}

func (s *Server) DeleteAccount(c *gin.Context) {
	var req DeleteAccountReq

	if err := c.ShouldBindUri(req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if err := s.store.DeleteAccounts(c, req.ID); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	c.JSON(http.StatusNoContent, DeleteAccountReq{})
}
