package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/shivangp0208/bank_application/db/sqlc"
	"github.com/shivangp0208/bank_application/util"
)

type CreateUserReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
	Email    string `json:"email"`
}

func (s *Server) CreateUser(c *gin.Context) {
	var req CreateUserReq

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	userPass, err := util.GenerateHashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.CreateUserParams{
		Username:       req.Username,
		HashedPassword: userPass,
		FullName:       req.FullName,
		Email:          req.Email,
	}

	_, err = s.store.CreateUser(c, arg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(errors.New("unable to create user: "+err.Error())))
		return
	}

	createdUser, err := s.store.GetUser(c, req.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(errors.New("unable to get the created user: "+err.Error())))
		return
	}

	c.JSON(http.StatusCreated, createdUser)
}
