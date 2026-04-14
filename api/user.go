package api

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	db "github.com/shivangp0208/bank_application/db/sqlc"
	"github.com/shivangp0208/bank_application/util"
)

type CreateUserReq struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=6"`
	FullName string `json:"full_name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
}

type UserResponse struct {
	Username          string    `json:"username"`
	FullName          string    `json:"full_name"`
	Email             string    `json:"email"`
	PasswordChangedAt time.Time `json:"password_changed_at"`
	CreatedAt         time.Time `json:"created_at"`
}

func getUserResponse(user db.User) UserResponse {
	return UserResponse{
		Username:          user.Username,
		FullName:          user.FullName,
		Email:             user.Email,
		PasswordChangedAt: user.PasswordChangedAt.Time,
		CreatedAt:         user.CreatedAt,
	}
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

	res := getUserResponse(createdUser)

	c.JSON(http.StatusCreated, res)
}

type GetUserReq struct {
	username string `uri:"id" binding:"required,alphanum"`
}

func (s *Server) GetUser(c *gin.Context) {
	var req GetUserReq

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	user, err := s.store.GetUser(c, req.username)
	if !checkSqlErr(c, err) {
		return
	}

	res := UserResponse{
		Username:          user.Username,
		FullName:          user.FullName,
		Email:             user.Email,
		PasswordChangedAt: user.PasswordChangedAt.Time,
		CreatedAt:         user.CreatedAt,
	}
	c.JSON(http.StatusOK, res)
}

type ListUsersReq struct {
	PageNo   int `form:"pageNo"`
	PageSize int `form:"pageSize"`
}

func (s *Server) GetAllUser(c *gin.Context) {
	var req ListUsersReq

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

	arg := db.ListPagedUsersParams{
		Limit:  int32(req.PageSize),
		Offset: int32((req.PageNo - 1) * req.PageSize),
	}

	users, err := s.store.ListPagedUsers(c, arg)
	if !checkSqlErr(c, err) {
		return
	}

	res := []UserResponse{}
	for _, user := range users {
		res = append(res, UserResponse{
			Username:          user.Username,
			FullName:          user.FullName,
			Email:             user.Email,
			PasswordChangedAt: user.PasswordChangedAt.Time,
			CreatedAt:         user.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, res)
}

type LoginUserReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginUserRes struct {
	SessionID             uuid.UUID    `json:"session_id"`
	AccessToken           string       `json:"access_token"`
	AccessTokenExpiresAt  time.Time    `json:"access_token_expires_at"`
	RefreshToken          string       `json:"refresh_token"`
	RefreshTokenExpiresAt time.Time    `json:"refresh_token_expires_at"`
	User                  UserResponse `json:"user"`
}

func (s *Server) LoginUser(c *gin.Context) {

	var req LoginUserReq

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	user, err := s.store.GetUser(c, req.Username)
	if !checkSqlErr(c, err) {
		return
	}

	if err := util.ComparePasswords(user.HashedPassword, req.Password); err != nil {
		c.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	accessToken, accessPayload, err := s.tokenMaker.CreateToken(req.Username, s.config.AccessTokenExpirationTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	refreshToken, refreshPayload, err := s.tokenMaker.CreateToken(req.Username, s.config.RefreshTokenExpirationTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	_, err = s.store.CreateSession(c, db.CreateSessionParams{
		ID:           refreshPayload.ID.String(),
		Username:     accessPayload.Username,
		RefreshToken: refreshToken,
		UserAgent:    c.Request.UserAgent(),
		ClientIp:     c.ClientIP(),
		IsBlocked:    false,
		ExpiresAt:    refreshPayload.ExpiredAt,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	session, err := s.store.GetSession(c, refreshPayload.ID.String())
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	res := LoginUserRes{
		SessionID:             uuid.MustParse(session.ID),
		User:                  getUserResponse(user),
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessPayload.ExpiredAt,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshPayload.ExpiredAt,
	}

	c.JSON(http.StatusOK, res)
}

func checkSqlErr(c *gin.Context, err error) bool {
	if err != nil {
		if errors.Is(sql.ErrNoRows, err) {
			c.JSON(http.StatusNotFound, errorResponse(err))
			return false
		}
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return false
	}
	return true
}
