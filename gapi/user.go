package gapi

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	db "github.com/shivangp0208/bank_application/db/sqlc"
	"github.com/shivangp0208/bank_application/pb"
	"github.com/shivangp0208/bank_application/util"
	"github.com/shivangp0208/bank_application/util/validator"
	"github.com/shivangp0208/bank_application/worker"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var logger = util.GetLogger()

func (s *Server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {

	if violations := validator.ValidateCreateUserReq(req); violations != nil {
		logger.Error().Msgf("validation failed for input arguments for create user req")
		return nil, validator.InvalidArgumentError(violations)
	}
	logger.Info().Msgf("validation passed for all input arguments for create user req")

	userPass, err := util.GenerateHashPassword(req.Password)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to hash the password: %v", err)
	}
	logger.Debug().Msg("successfully generated the hash password from the req password")

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
				// There is a very big imp of this delay in this task, as the create user task is wrapped inside a transaction with a execTx func so in that first the db call is being made but the user data is still not stored as it will be stored only after committing the transaction, so assume that this execTx took too long to commit the transaction but was able to run the func inside it quickly then if there will be no delay then the process email will throw the error user not found as at that timr the user data is not stored in the db
				asynq.ProcessIn(10 * time.Second),
				asynq.Queue(worker.DefaultQueue),
			}

			err := s.taskProducer.ProduceSendVerificationEmail(ctx, jsonPayload, opts...)
			if err != nil {
				logger.Error().Str("username", jsonPayload.Username).Msgf("unable to produce the verification email task in the async queue")
			}
			logger.Info().Str("username", jsonPayload.Username).Msgf("successfully produce the verification email task in the async queue")
			return nil
		},
	}

	txRes, err := s.store.CreateUserTx(ctx, arg)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to create user: %v", err.Error())
	}
	logger.Info().
		Str("full_name", txRes.User.FullName).
		Str("email", txRes.User.Email).
		Str("role", txRes.User.Role).
		Str("password_changed_at", txRes.User.PasswordChangedAt.Time.String()).
		Str("created_at", txRes.User.CreatedAt.String()).
		Msgf("successfully created user with username %s", txRes.User.Username)

	res := &pb.CreateUserResponse{
		User: &pb.User{
			Username:          txRes.User.Username,
			FullName:          txRes.User.FullName,
			Email:             txRes.User.Email,
			Role:              txRes.User.Role,
			IsVerified:        txRes.User.IsVerified,
			PasswordChangedAt: timestamppb.New(txRes.User.PasswordChangedAt.Time),
			CreatedAt:         timestamppb.New(txRes.User.CreatedAt),
		},
	}

	return res, nil
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

	accessToken, accessPayload, err := s.tokenMaker.CreateToken(user.Username, user.Role, s.config.AccessTokenExpirationTime)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to generate the access token %v", err)
	}

	refreshToken, refreshPayload, err := s.tokenMaker.CreateToken(user.Username, user.Role, s.config.RefreshTokenExpirationTime)
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

	if payload.Role != util.Accountant && payload.Username != req.Username {
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
			Role:              updatedUser.Role,
			IsVerified:        updatedUser.IsVerified,
			PasswordChangedAt: timestamppb.New(updatedUser.PasswordChangedAt.Time),
			CreatedAt:         timestamppb.New(updatedUser.CreatedAt),
		},
	}

	return &res, nil
}

// TODO: create a seperate transactio for updating user's password
func (s *Server) UpdateUserPassword(ctx context.Context, req *pb.UpdatePasswordRequest) (*empty.Empty, error) {
	payload, err := s.authorizeUser(ctx)
	if err != nil {
		logger.Warn().Msgf("unauthorized user")
		return nil, status.Errorf(codes.Unauthenticated, "unauthorized user: %v", err)
	}
	logger.Info().Msg("authentication and authorization successfull")

	if err := validator.ValidatePassword(req.OldPassword); err != nil {
		logger.Warn().Msgf("validation failed for old password in req")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := validator.ValidatePassword(req.NewPassword); err != nil {
		logger.Warn().Msgf("validation failed for new password in req")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if req.OldPassword == req.NewPassword {
		logger.Warn().Msgf("validation failed for new password in req")
		return nil, status.Errorf(codes.InvalidArgument, "cannot have new password same as old password")
	}
	logger.Info().Msg("successfully validated the user req for updating password")

	user, err := s.store.GetUser(ctx, payload.Username)
	if ok, err := checkSqlErr(err); !ok {
		logger.Error().Msg("unable to get the user details from token payload")
		return nil, status.Error(codes.Internal, err.Error())
	}

	if err := util.ComparePasswords(user.HashedPassword, req.OldPassword); err != nil {
		logger.Error().Msg("old password does not match with required pass")
		return nil, status.Errorf(codes.InvalidArgument, "wrong password, old password does not match with required pass")
	}

	newPass, err := util.GenerateHashPassword(req.NewPassword)
	if err != nil {
		logger.Error().Msgf("unable to generate hashed password for new password: %v", err)
		return nil, status.Errorf(codes.Internal, "unable to generate hashed password for new password: %v", err)
	}

	var arg db.UpdateUserParams = db.UpdateUserParams{
		Username: user.Username,
		HashedPassword: sql.NullString{
			String: newPass,
			Valid:  true,
		},
		PasswordChangedAt: sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		},
	}

	if err := s.store.UpdateUser(ctx, arg); err != nil {
		logger.Error().Msgf("unable to update the user's password %v", err)
		return nil, status.Errorf(codes.Internal, "unable to update user's password: %v", err)
	}
	logger.Info().Msg("successfully updated user's password")

	return &emptypb.Empty{}, nil
}

func (s *Server) VerifyUserEmail(ctx context.Context, req *pb.VerifyEmailRequest) (*pb.VerifyEmailResponse, error) {

	logger.Info().Msgf("req to verify the email by the user's username %s", req.Username)
	// validate all field according to format in the given req
	if violations := validator.ValidateVerifyUserEmailReq(req); violations != nil {
		logger.Warn().Msgf("validation failed for input arguments for verify user req")
		return nil, validator.InvalidArgumentError(violations)
	}
	logger.Info().Msgf("validation passed for all input arguments for verify user req")

	// update the db for verify emails by checking all validation
	verifiedUser, err := s.store.VerifyUserEmailTx(ctx, db.VerifyUserTxParams{
		Username:   req.Username,
		SecretCode: req.SecretCode,
	})
	if err != nil {
		logger.Error().Msgf("unable to verify the user: %v", err)
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	logger.Info().Msgf("success verifying user, updated all db records")

	result := &pb.VerifyEmailResponse{
		Username: verifiedUser.User.Username,
		Message:  "User Verified Successfully",
	}

	return result, nil
}

func (s *Server) GetUserByUsername(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	logger.Info().Msgf("GET req to get the user by the user's username %s", req.Username)
	if err := validator.ValidateUsername(req.GetUsername()); err != nil {
		logger.Info().Msgf("validation failed for input arguments for get user by username req")
		return nil, err
	}
	logger.Info().Msgf("validation passed for all input arguments for get user req")

	user, err := s.store.GetUser(ctx, req.GetUsername())
	if err != nil {
		logger.Error().Str("username", req.Username).Msgf("unable to get the user: %v", err)
		return nil, err
	}

	res := &pb.GetUserResponse{
		User: &pb.User{
			Username:          user.Username,
			FullName:          user.FullName,
			Email:             user.Email,
			Role:              user.Role,
			IsVerified:        user.IsVerified,
			PasswordChangedAt: timestamppb.New(user.PasswordChangedAt.Time),
			CreatedAt:         timestamppb.New(user.CreatedAt),
		},
	}

	return res, nil
}

func (s *Server) GetAllUser(ctx context.Context, req *pb.GetAllUserRequest) (*pb.GetAllUserResponse, error) {

	payload, err := s.authorizeUser(ctx)
	if err != nil {
		logger.Error().Msgf("unable to authorize user's token: %v", err)
		return nil, err
	}

	// setting up a default page size
	if req.PageSize == 0 {
		req.PageSize = 5
	}
	logger.Debug().Msgf("GET req to get all the users info with pageNum %d and pageSize %d", req.PageNum, req.PageSize)

	arg := db.ListPagedUsersParams{
		Limit:  int32(req.PageSize),
		Offset: int32(req.PageNum) * int32(req.PageSize+1),
	}

	userList, err := s.store.ListPagedUsers(ctx, arg)
	if err != nil {
		logger.Error().Msgf("unable to get all user details: %v", err)
		return nil, err
	}
	logger.Debug().Msgf("successfully retrieved all user list with length %d", len(userList))

	var pbUserList []*pb.User
	for _, user := range userList {
		// we are only going to show the account the full user list, while if any other normal user tries to access this we are going to return only his info
		if payload.Role == util.Accountant || payload.Username == user.Username {
			pbUser := &pb.User{
				Username:          user.Username,
				FullName:          user.FullName,
				Email:             user.Email,
				Role:              user.Role,
				IsVerified:        user.IsVerified,
				PasswordChangedAt: timestamppb.New(user.PasswordChangedAt.Time),
				CreatedAt:         timestamppb.New(user.CreatedAt),
			}
			pbUserList = append(pbUserList, pbUser)
		}
	}

	res := &pb.GetAllUserResponse{User: pbUserList}

	return res, nil
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
		Role:              user.Role,
		IsVerified:        user.IsVerified,
		PasswordChangedAt: timestamppb.New(user.PasswordChangedAt.Time),
		CreatedAt:         timestamppb.New(user.CreatedAt),
	}
}
