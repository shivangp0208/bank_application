package gapi

import (
	"context"
	"slices"

	"github.com/golang/protobuf/ptypes/empty"
	db "github.com/shivangp0208/bank_application/db/sqlc"
	"github.com/shivangp0208/bank_application/pb"
	"github.com/shivangp0208/bank_application/util/validator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Server) CreateAccount(ctx context.Context, req *pb.CreateAccountRequest) (*pb.Account, error) {

	if !validator.ValidateCurrency(req.Currency) {
		return nil, status.Error(codes.InvalidArgument, "invalid argument currency, validation failed")
	}

	arg := db.CreateAccountsParams{
		Owner:    req.Owner,
		Balance:  0,
		Currency: req.Currency,
	}

	accountRes, err := s.store.CreateAccounts(ctx, arg)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	accountId, err := accountRes.LastInsertId()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	account, err := s.store.GetAccount(ctx, uint64(accountId))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	res := &pb.Account{
		Id:        account.ID,
		Owner:     account.Owner,
		Currency:  account.Currency,
		CreatedAt: timestamppb.New(account.CreatedAt),
	}
	return res, nil
}

func (s *Server) GetAccountByID(ctx context.Context, req *pb.AccountIDReq) (*pb.Account, error) {

	payload, err := s.authorizeUser(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	userAccountList, err := s.store.ListAllAccountIdByUsername(ctx, payload.Username)
	if err := checkSqlErr(err); err != nil {
		return nil, err
	}

	if !slices.Contains(userAccountList, req.Id) {
		return nil, status.Error(codes.PermissionDenied, "unauthorized req")
	}

	account, err := s.store.GetAccount(ctx, req.Id)
	if err := checkSqlErr(err); err != nil {
		return nil, err
	}

	res := &pb.Account{
		Id:        account.ID,
		Owner:     account.Owner,
		Currency:  account.Currency,
		CreatedAt: timestamppb.New(account.CreatedAt),
	}
	return res, nil
}

func (s *Server) GetAllAccount(ctx context.Context, req *pb.PaginationReq) (*pb.AccountList, error) {
	payload, err := s.authorizeUser(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	if req.PageSize == 0 {
		req.PageSize = 5
	}

	arg := db.ListPagedAccountsParams{
		Owner:  payload.Username,
		Limit:  int32(req.PageSize),
		Offset: int32((req.PageNum - 1) * uint32(req.PageSize)),
	}

	accounts, err := s.store.ListPagedAccounts(ctx, arg)
	if err := checkSqlErr(err); err != nil {
		return nil, err
	}

	res := &pb.AccountList{}
	for _, acc := range accounts {
		pbAcc := &pb.Account{
			Id:        acc.ID,
			Owner:     acc.Owner,
			Currency:  acc.Currency,
			CreatedAt: timestamppb.New(acc.CreatedAt),
		}
		res.Accounts = append(res.Accounts, pbAcc)
	}

	return res, nil
}

func (s *Server) DeleteAccount(ctx context.Context, req *pb.AccountIDReq) (*empty.Empty, error) {
	payload, err := s.authorizeUser(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	account, err := s.store.GetAccount(ctx, req.Id)
	if err := checkSqlErr(err); err != nil {
		return nil, err
	}

	if payload.Username != account.Owner {
		return nil, status.Errorf(codes.PermissionDenied, "unauthorized req")
	}

	if err := s.store.DeleteAccounts(ctx, req.Id); err != nil {
		if err := checkSqlErr(err); err != nil {
			return nil, err
		}
	}

	return &emptypb.Empty{}, nil
}
