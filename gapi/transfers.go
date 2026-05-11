package gapi

import (
	"context"
	"errors"
	"fmt"

	db "github.com/shivangp0208/bank_application/db/sqlc"
	"github.com/shivangp0208/bank_application/pb"
	"github.com/shivangp0208/bank_application/token"
	"github.com/shivangp0208/bank_application/util/validator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Server) TransferMoney(ctx context.Context, req *pb.TransferMoneyRequest) (*pb.TransferMoneyResponse, error) {

	payload, ok := ctx.Value(AuthorizationPayloadKey).(*token.Payload)
	if !ok {
		return nil, status.Errorf(codes.Internal, "invalid req")
	}

	if !validator.ValidateCurrency(req.Currency) {
		err := fmt.Errorf("invalid currency, currency %s not supported", req.Currency)
		Logger.Error().Msgf("invalid currency, currency %s not supported", req.Currency)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	fromAccount, err := s.store.GetAccount(ctx, req.FromAccountId)
	if err := checkSqlErr(err); err != nil {
		err = errors.Join(fmt.Errorf("unable to get the account details:"), err)
		Logger.Error().Msgf("unable to get the account details: %v", err)
		return nil, err
	}

	if fromAccount.Owner != payload.Username {
		err := errors.New("account doesn't belong to the authenticated user")
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	toAccount, err := s.store.GetAccount(ctx, req.ToAccountId)
	if err := checkSqlErr(err); err != nil {
		return nil, err
	}

	if fromAccount.Currency != toAccount.Currency {
		err := fmt.Errorf("unable to do transaction, currency mismatching between sender %s and receiver %s", fromAccount.Currency, toAccount.Currency)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	arg := db.TransferTxParams{
		FromAccountID: req.FromAccountId,
		ToAccountID:   req.ToAccountId,
		Amount:        int64(req.Amount),
	}

	transferRes, err := s.store.TransferTx(ctx, arg)
	if err := checkSqlErr(err); err != nil {
		err = errors.Join(fmt.Errorf("unable to do the transaction:"), err)
		return nil, err
	}

	res := &pb.TransferMoneyResponse{
		Transfer: &pb.Transfer{
			Id:            transferRes.Transfer.ID,
			FromAccountId: transferRes.Transfer.FromAccountID,
			ToAccountId:   transferRes.Transfer.ToAccountID,
			Amount:        uint64(transferRes.Transfer.Amount),
			CreatedAt:     timestamppb.New(transferRes.Transfer.CreatedAt),
		},
		FromAccount: &pb.Account{
			Id:        transferRes.FromAccount.ID,
			Owner:     transferRes.FromAccount.Owner,
			Currency:  transferRes.FromAccount.Currency,
			CreatedAt: timestamppb.New(transferRes.FromAccount.CreatedAt),
		},
		FromEntry: &pb.Entry{
			Id:        transferRes.FromEntry.ID,
			AccountId: transferRes.FromEntry.AccountID,
			Amount:    uint64(transferRes.FromEntry.Amount),
			CreatedAt: timestamppb.New(transferRes.FromEntry.CreatedAt),
		},
		ToAccount: &pb.Account{
			Id:        transferRes.ToAccount.ID,
			Owner:     transferRes.ToAccount.Owner,
			Currency:  transferRes.ToAccount.Currency,
			CreatedAt: timestamppb.New(transferRes.ToAccount.CreatedAt),
		},
		ToEntry: &pb.Entry{
			Id:        transferRes.ToEntry.ID,
			AccountId: transferRes.ToEntry.AccountID,
			Amount:    uint64(transferRes.ToEntry.Amount),
			CreatedAt: timestamppb.New(transferRes.ToEntry.CreatedAt),
		},
	}
	return res, nil
}
