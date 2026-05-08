package gapi

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/shivangp0208/bank_application/pb"
)

// TODO: implement this accounts grpc api
func (s *Server) CreateAccount(ctx context.Context, req *pb.CreateAccountRequest) (*pb.Account, error) {
	return nil, nil
}
func (s *Server) GetAccountByID(ctx context.Context, req *pb.AccountIDReq) (*pb.Account, error) {
	return nil, nil
}
func (s *Server) GetAllAccount(ctx context.Context, req *pb.PaginationReq) (*pb.AccountList, error) {
	return nil, nil
}
func (s *Server) DeleteAccount(ctx context.Context, req *pb.AccountIDReq) (*empty.Empty, error) {
	return nil, nil
}
