package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/shivangp0208/bank_application/db/sqlc"
)

type TransferMoneyReq struct {
	FromAccountID uint64 `json:"fromAccountId" binding:"required,min=1"`
	ToAccountID   uint64 `json:"toAccountId" binding:"required,min=1"`
	Amount        int64  `json:"amount" binding:"required,min=1"`
	Currency      string `json:"currency" binding:"required,currency"`
}

func (s *Server) TransferMoney(c *gin.Context) {
	var req TransferMoneyReq

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	fromAccount, err := s.store.GetAccount(c, req.FromAccountID)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse(err))
		return
	}

	toAccount, err := s.store.GetAccount(c, req.ToAccountID)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse(err))
		return
	}

	if fromAccount.Currency != toAccount.Currency {
		c.JSON(http.StatusBadRequest, errors.New("unable to do transaction, currency mismatching between sender and receiver"))
		return
	}

	arg := db.TransferTxParams{
		FromAccountID: req.FromAccountID,
		ToAccountID:   req.ToAccountID,
		Amount:        req.Amount,
	}

	transferRes, err := s.store.TransferTx(c, arg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(errors.New("unable to do the transaction: "+err.Error())))
		return
	}

	c.JSON(http.StatusOK, transferRes)
}
