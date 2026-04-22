package gapi

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	db "github.com/shivangp0208/bank_application/db/sqlc"
	"github.com/shivangp0208/bank_application/pb"
	"github.com/shivangp0208/bank_application/util"
	"github.com/shivangp0208/bank_application/util/validator"
	"github.com/shivangp0208/bank_application/worker"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var logger = util.GetLogger()

func (s *Server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {

	if violations := validator.ValidateCreateUserReq(req); violations != nil {
		logger.Info().Msgf("validation failed for input arguments for create user req")
		return nil, validator.InvalidArgumentError(violations)
	}
	logger.Info().Msgf("validation passed for all input arguments for create user req")

	userPass, err := util.GenerateHashPassword(req.Password)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to hash the password: %v", err)
	}

	arg := db.CreateUserTxParams{
		User: db.CreateUserParams{
			Username:       req.Username,
			HashedPassword: userPass,
			FullName:       req.FullName,
			Email:          req.Email,
		},
		AfterCreateUser: func(user db.User) error {
			jsonPayload := &worker.EmailDeliveryPayload{
				Username: user.Username,
			}

			opts := []asynq.Option{
				asynq.MaxRetry(5),
				asynq.ProcessIn(10 * time.Second),
				asynq.Queue(worker.CriticalQueue),
			}

			return s.taskProducer.ProduceSendVerificationEmail(ctx, jsonPayload, opts...)
		},
	}

	txRes, err := s.store.CreateUserTx(ctx, arg)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to create user: %v", err.Error())
	}

	res := pb.CreateUserResponse{
		User: &pb.User{
			Username:          txRes.User.Username,
			FullName:          txRes.User.FullName,
			Email:             txRes.User.Email,
			PasswordChangedAt: timestamppb.New(txRes.User.PasswordChangedAt.Time),
			CreatedAt:         timestamppb.New(txRes.User.CreatedAt),
		},
	}

	return &res, nil
}

func (s *Server) LoginUser(c context.Context, req *pb.LoginUserRequest) (*pb.LoginUserResponse, error) {

	if violations := validator.ValidateLoginUserReq(req); violations != nil {
		logger.Info().Msgf("validation failed for input arguments for login user req")
		return nil, validator.InvalidArgumentError(violations)
	}
	logger.Info().Msgf("validation passed for all input arguments for login user req")

	user, err := s.store.GetUser(c, req.Username)
	if ok, err := checkSqlErr(err); !ok {
		return nil, err
	}

	if err := util.ComparePasswords(user.HashedPassword, req.Password); err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "password not matched %v", err)
	}

	accessToken, accessPayload, err := s.tokenMaker.CreateToken(req.Username, s.config.AccessTokenExpirationTime)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to generate the access token %v", err)
	}

	refreshToken, refreshPayload, err := s.tokenMaker.CreateToken(req.Username, s.config.RefreshTokenExpirationTime)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to generate the refresh token %v", err)
	}

	md := s.extractMetadata(c)
	_, err = s.store.CreateSession(c, db.CreateSessionParams{
		ID:           refreshPayload.ID.String(),
		Username:     accessPayload.Username,
		RefreshToken: refreshToken,
		UserAgent:    md.UserAgent,
		ClientIp:     md.ClientIP,
		IsBlocked:    false,
		ExpiresAt:    refreshPayload.ExpiredAt,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to create session %v", err)
	}

	session, err := s.store.GetSession(c, refreshPayload.ID.String())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to get the created session %v", err)
	}

	res := pb.LoginUserResponse{
		SessionId:             []byte(uuid.MustParse(session.ID).String()),
		User:                  getUserResponse(user),
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  timestamppb.New(accessPayload.ExpiredAt),
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: timestamppb.New(refreshPayload.ExpiredAt),
	}

	return &res, nil
}

func (s *Server) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {

	payload, err := s.authorizeUser(ctx)
	if err != nil {
		logger.Info().Msgf("authorization of user failed with req %v err %v", req, err)
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	if payload.Username != req.Username {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token, username mismatch in token %s and req %s", payload.Username, req.Username)
	}

	if violations := validator.ValidateUpdateUserReq(req); violations != nil {
		logger.Info().Msgf("validation failed for input arguments for update user req")
		return nil, validator.InvalidArgumentError(violations)
	}
	logger.Info().Msgf("validation passed for all input arguments for update user req")

	arg := db.UpdateUserParams{
		Username: req.Username,
	}

	if req.FullName != nil {
		arg.FullName = sql.NullString{
			String: *req.FullName,
			Valid:  len(*req.FullName) > 0,
		}
	}

	if req.Email != nil {
		arg.Email = sql.NullString{
			String: *req.Email,
			Valid:  len(*req.Email) > 0,
		}
	}

	if req.Password != nil && len(*req.Password) > 0 {
		pass, err := util.GenerateHashPassword(*req.Password)
		if err != nil {
			logger.Info().Msgf("unable to generate the hash password for %s", *req.Password)
			return nil, status.Errorf(codes.Internal, "unable to generate the hash password for %s", *req.Password)
		}
		logger.Info().Msgf("successfully generated the hashed password %v", arg.HashedPassword.String)

		arg.HashedPassword = sql.NullString{
			String: pass,
			Valid:  true,
		}
		arg.PasswordChangedAt = sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		}
	}

	err = s.store.UpdateUser(ctx, arg)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to update user: %v", err.Error())
	}

	updatedUser, err := s.store.GetUser(ctx, req.Username)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to get the updated user: %v", err.Error())
	}

	res := pb.UpdateUserResponse{
		User: &pb.User{
			Username:          updatedUser.Username,
			FullName:          updatedUser.FullName,
			Email:             updatedUser.Email,
			PasswordChangedAt: timestamppb.New(updatedUser.PasswordChangedAt.Time),
			CreatedAt:         timestamppb.New(updatedUser.CreatedAt),
		},
	}

	return &res, nil
}

func checkSqlErr(err error) (bool, error) {
	if err != nil {
		if errors.Is(sql.ErrNoRows, err) {
			return false, status.Errorf(codes.NotFound, "unable to find the user with given username, %v", err)
		}
		return false, status.Error(codes.Internal, err.Error())
	}
	return true, nil
}

func getUserResponse(user db.User) *pb.User {
	return &pb.User{
		Username:          user.Username,
		FullName:          user.FullName,
		Email:             user.Email,
		PasswordChangedAt: timestamppb.New(user.PasswordChangedAt.Time),
		CreatedAt:         timestamppb.New(user.CreatedAt),
	}
}
