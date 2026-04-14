package gapi

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	db "github.com/shivangp0208/bank_application/db/sqlc"
	"github.com/shivangp0208/bank_application/pb"
	"github.com/shivangp0208/bank_application/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var logger = util.GetLogger()

func (s *Server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {

	userPass, err := util.GenerateHashPassword(req.Password)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to hash the password: %v", err)
	}

	arg := db.CreateUserParams{
		Username:       req.Username,
		HashedPassword: userPass,
		FullName:       req.FullName,
		Email:          req.Email,
	}

	_, err = s.store.CreateUser(ctx, arg)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to create user: %v", err.Error())
	}

	createdUser, err := s.store.GetUser(ctx, req.Username)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to get the created user: %v", err.Error())
	}

	res := pb.CreateUserResponse{
		User: &pb.User{
			Username:          createdUser.Username,
			FullName:          createdUser.FullName,
			Email:             createdUser.Email,
			PasswordChangedAt: timestamppb.New(createdUser.PasswordChangedAt.Time),
			CreatedAt:         timestamppb.New(createdUser.CreatedAt),
		},
	}

	return &res, nil
}

func (s *Server) LoginUser(c context.Context, req *pb.LoginUserRequest) (*pb.LoginUserResponse, error) {

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

	var userAgent, clientIP string
	if md, ok := metadata.FromIncomingContext(c); ok {
		ua := md.Get("user-agent")
		cip := md.Get("client-ip")
		if len(ua) > 0 {
			userAgent = ua[0]
			logger.Printf("user agent provided by client side %s", userAgent)
		}
		if len(cip) > 0 {
			clientIP = cip[0]
			logger.Printf("client ip provided by client side %s", clientIP)
		}
	}

	_, err = s.store.CreateSession(c, db.CreateSessionParams{
		ID:           refreshPayload.ID.String(),
		Username:     accessPayload.Username,
		RefreshToken: refreshToken,
		UserAgent:    userAgent,
		ClientIp:     clientIP,
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
