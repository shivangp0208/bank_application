package api

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	db "github.com/shivangp0208/bank_application/db/sqlc"
	"github.com/shivangp0208/bank_application/util"
	"github.com/shivangp0208/bank_application/worker"
)

var myLogger = util.GetLogger()

type CreateUserReq struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=6"`
	FullName string `json:"full_name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
}

type UserResponse struct {
	Username          string    `json:"username"`
	FullName          string    `json:"full_name"`
	Role              string    `json:"role"`
	Email             string    `json:"email"`
	PasswordChangedAt time.Time `json:"password_changed_at"`
	CreatedAt         time.Time `json:"created_at"`
}

func getUserResponse(user db.User) UserResponse {
	return UserResponse{
		Username:          user.Username,
		FullName:          user.FullName,
		Email:             user.Email,
		Role:              user.Role,
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

	arg := db.CreateUserTxParams{
		User: db.CreateUserParams{
			Username:       req.Username,
			HashedPassword: userPass,
			FullName:       req.FullName,
			Email:          req.Email,
		},
		AfterCreateUser: func(user db.User) error {

			asynqOpts := []asynq.Option{
				asynq.MaxRetry(5),
				asynq.ProcessIn(10 * time.Second),
				asynq.Queue(worker.DefaultQueue),
			}
			emailPayload := &worker.EmailDeliveryPayload{
				Username: req.Username,
			}

			if err := s.TaskProducer.ProduceSendVerificationEmail(c, emailPayload, asynqOpts...); err != nil {
				return err
			}

			return nil
		},
	}

	txRes, err := s.Store.CreateUserTx(c, arg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(errors.New("unable to create user: "+err.Error())))
		return
	}

	res := getUserResponse(txRes.User)

	c.JSON(http.StatusCreated, res)
}

type GetUserReq struct {
	username string `uri:"username" binding:"required,alphanum"`
}

func (s *Server) GetUser(c *gin.Context) {
	var req GetUserReq

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if err := authorizeUser(c, req.username); err != nil {
		return
	}

	user, err := s.Store.GetUser(c, req.username)
	if !checkSqlErr(c, err) {
		return
	}

	res := getUserResponse(user)
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

	users, err := s.Store.ListPagedUsers(c, arg)
	if !checkSqlErr(c, err) {
		return
	}

	var res []UserResponse
	for _, user := range users {
		res = append(res, getUserResponse(user))
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

	user, err := s.Store.GetUser(c, req.Username)
	if !checkSqlErr(c, err) {
		return
	}

	if err := util.ComparePasswords(user.HashedPassword, req.Password); err != nil {
		c.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	accessToken, accessPayload, err := s.TokenMaker.CreateToken(user.Username, user.Role, s.Config.AccessTokenExpirationTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	refreshToken, refreshPayload, err := s.TokenMaker.CreateToken(user.Username, user.Role, s.Config.RefreshTokenExpirationTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	_, err = s.Store.CreateSession(c, db.CreateSessionParams{
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

	session, err := s.Store.GetSession(c, refreshPayload.ID.String())
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

type VerifyUserEmailReq struct {
	Username   string `form:"username" binding:"required,alphanum,min=2"`
	SecretCode string `form:"secre_code" binding:"required,min=30"`
}

type VerifyUserEmailRes struct {
	Username string `json:"username"`
	Message  string `json:"message"`
}

func (s *Server) VerifyUserEmail(c *gin.Context) {

	var req VerifyUserEmailReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	myLogger.Info().Msgf("GET req to verify the email by the user's username %s", req.Username)
	myLogger.Info().Msgf("validation passed for all input arguments for verify user req")

	// update the db for verify emails by checking all validation
	verifiedUser, err := s.Store.VerifyUserEmailTx(c, db.VerifyUserTxParams{
		Username:   req.Username,
		SecretCode: req.SecretCode,
	})
	if err != nil {
		myLogger.Error().Msgf("unable to verify the user: %v", err)
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	myLogger.Info().Msgf("success verifying user, updated all db records")

	result := &VerifyUserEmailRes{
		Username: verifiedUser.User.Username,
		Message:  "User Verified Successfully",
	}

	c.JSON(http.StatusOK, result)
}

type UpdateUserBodyReq struct {
	FullName string `json:"full_name" binding:"omitempty"`
	Email    string `json:"email" binding:"omitempty,email"`
}

type UpdateUserURLReq struct {
	Username string `uri:"username" binding:"required,alphanum"`
}

func (s *Server) UpdateUser(c *gin.Context) {
	myLogger.Println("UpdateUser: PATCH req for updating user")

	if err := checkForbiddenFields(c, []string{"username", "hashed_password", "password"}); err != nil {
		myLogger.Info().Msgf("UpdateUser: %v", err)
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var bodyReq UpdateUserBodyReq
	var urlReq UpdateUserURLReq

	if err := c.ShouldBindJSON(&bodyReq); err != nil {
		myLogger.Println("UpdateUser: unable to validate the JSON body req")
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	if err := c.ShouldBindUri(&urlReq); err != nil {
		myLogger.Println("UpdateUser: unable to validate the URL req")
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	myLogger.Println("UpdateUser: successfully validated json body req and url")

	if err := authorizeUser(c, urlReq.Username); err != nil {
		return
	}

	arg := db.UpdateUserParams{
		FullName: sql.NullString{
			String: bodyReq.FullName,
			Valid:  len(bodyReq.FullName) > 0,
		},
		Email: sql.NullString{
			String: bodyReq.Email,
			Valid:  len(bodyReq.Email) > 0,
		},
		Username: urlReq.Username,
	}

	// we should not allow the user to change the password from this update user api for security reason
	// if len(bodyReq.Password) > 0 {
	// 	pass, err := util.GenerateHashPassword(bodyReq.Password)
	// 	if err != nil {
	// 		myLogger.Info().Msgf("UpdateUser: unable to generate the hashed password for given pass %s", bodyReq.Password)
	// 		c.JSON(http.StatusInternalServerError, errorResponse(err))
	// 		return
	// 	}
	// 	myLogger.Info().Msgf("UpdateUser: success generating the hashed password %v", arg.HashedPassword.String)

	// 	arg.HashedPassword = sql.NullString{
	// 		String: pass,
	// 		Valid:  true,
	// 	}
	// 	arg.PasswordChangedAt = sql.NullTime{
	// 		Time:  time.Now(),
	// 		Valid: true,
	// 	}
	// }

	if err := s.Store.UpdateUser(c, arg); err != nil {
		myLogger.Info().Msgf("UpdateUser: unable to Store the updated user in db")
		if checkSqlErr(c, err) {
			return
		}
	}
	myLogger.Info().Msgf("UpdateUser: successfully Stored the updated user %v", arg)

	updatedUser, err := s.Store.GetUser(c, urlReq.Username)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse(err))
		return
	}
	myLogger.Info().Msgf("UpdateUser: successfully fetched the updated user %v", updatedUser)

	c.JSON(http.StatusOK, updatedUser)
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
