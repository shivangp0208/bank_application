package api

import "github.com/gin-gonic/gin"

type TransferMoneyReq struct {
	FromAccountID uint64 `json:"fromAccountId" binding:"required,min=1"`
	ToAccountID   uint64 `json:"toAccountId" binding:"required,min=1"`
	Amount        int64  `json:"amount" binding:"required,min=1"`
	Currency      string `json:"currency" binding:"required,currency"`
}

func (s *Server) TransferMoney(c *gin.Context) {

}
