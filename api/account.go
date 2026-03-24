package api

import (
	"net/http"

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

	c.JSON(http.StatusOK, acount)
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

func (s *Server) GetAllAccount(c *gin.Context) {
	accounts, err := s.store.ListAccounts(c)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse(err))
	}

	c.JSON(http.StatusOK, accounts)
}

type ListAccountsReq struct {
	PageNo   int32 `form:"pageNo"`
	PageSize int32 `form:"pageSize"`
}

func (s *Server) GetAllPagedAccount(c *gin.Context) {
	var req ListAccountsReq

	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.ListPagedAccountsParams{
		Limit:  req.PageSize,
		Offset: (req.PageNo - 1) * req.PageSize,
	}

	accounts, err := s.store.ListPagedAccounts(c, arg)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse(err))
		return
	}

	c.JSON(http.StatusOK, accounts)
}
