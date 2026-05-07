package gapi

import (
	"context"
	"errors"
	"fmt"
	"slices"

	db "github.com/shivangp0208/bank_application/db/sqlc"
	"github.com/shivangp0208/bank_application/pb"
	"github.com/shivangp0208/bank_application/util"
	"github.com/shivangp0208/bank_application/util/validator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Server) GetAllEntryForAccountID(ctx context.Context, req *pb.AccountID) (*pb.EntryListResponse, error) {
	logger.Info().Msgf("GetAllEntryForAccountID called with account_id: %d", req.AccountId)

	payload, err := s.authorizeUser(ctx)
	if err != nil {
		logger.Error().Msgf("unauthenticated req error %v", err)
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	logger.Info().Msgf("authorized user: %s with role: %s", payload.Username, payload.Role)

	if err := validator.ValidateUsername(payload.Username); err != nil {
		logger.Error().Msgf("invalid username in req %s", payload.Username)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	logger.Info().Msgf("username %s validated successfully", payload.Username)

	if payload.Role != util.Accountant {
		logger.Info().Msgf("user %s is not an accountant, checking account ownership", payload.Username)

		accountList, err := s.store.ListAllAccountIdByUsername(ctx, payload.Username)
		if err != nil {
			err = errors.Join(fmt.Errorf("error getting the list of accounts for username %s: ", payload.Username), err)
			return nil, status.Error(codes.Internal, err.Error())
		}
		logger.Info().Msgf("fetched %d accounts for user %s", len(accountList), payload.Username)

		if !slices.Contains(accountList, uint64(req.AccountId)) {
			err := fmt.Errorf("user %s is not allowed to see the entries for this account", payload.Username)
			logger.Error().Msgf("permission denied: user %s does not own account_id %d", payload.Username, req.AccountId)
			return nil, status.Error(codes.PermissionDenied, err.Error())
		}
		logger.Info().Msgf("account_id %d ownership verified for user %s", req.AccountId, payload.Username)
	} else {
		logger.Info().Msgf("user %s is an accountant, skipping ownership check", payload.Username)
	}

	arg := db.ListEntriesByAccountIdAndUsernameParams{
		Username:  payload.Username,
		AccountID: uint64(req.AccountId),
	}
	logger.Info().Msgf("fetching entries for account_id: %d, username: %s", req.AccountId, payload.Username)

	entryList, err := s.store.ListEntriesByAccountIdAndUsername(ctx, arg)
	if err != nil {
		err = errors.Join(fmt.Errorf("error getting the list of entries: "), err)
		return nil, status.Error(codes.Internal, err.Error())
	}
	logger.Info().Msgf("fetched %d entries for account_id: %d", len(entryList), req.AccountId)

	res := &pb.EntryListResponse{}
	for _, entry := range entryList {
		res.Entries = append(res.Entries, &pb.Entry{
			Id:        entry.ID,
			AccountId: entry.AccountID,
			Amount:    uint64(entry.Amount),
			CreatedAt: timestamppb.New(entry.CreatedAt),
		})
	}

	logger.Info().Msgf("GetAllEntryForAccountID completed successfully for account_id: %d, returning %d entries", req.AccountId, len(res.Entries))
	return res, nil
}
