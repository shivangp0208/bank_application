package interceptors

import (
	"context"
	"fmt"
	"strings"

	"github.com/shivangp0208/bank_application/gapi"
	"github.com/shivangp0208/bank_application/token"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	authorizationHeaderKey = "authorization"
	authorizationType      = "bearer"
)

var logger = gapi.Logger

func GRPCAuthInterceptor(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (resp any, err error) {

	mtdt, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unable to get metadata from context")
	}
	logger.Debug().Msgf("successfully got the metadata from the incoming context %v", mtdt)

	fields := mtdt.Get(authorizationHeaderKey)
	if len(fields) == 0 {
		return nil, status.Error(codes.Unauthenticated, "no authorization provided in metadata")
	}
	authorizationFields := strings.Fields(fields[0])
	if len(authorizationFields) < 2 {
		return nil, status.Error(codes.Unauthenticated, "invalid authorization header format in metadata")
	}
	logger.Debug().Msgf("successfully got the authorization header from metadata %v", authorizationFields)

	authType := strings.ToLower(authorizationFields[0])
	if authType != authorizationType {
		return nil, status.Error(codes.Unauthenticated, fmt.Sprintf("invalid authorization type in authorization %s", authType))
	}
	logger.Debug().Msg("successfully validated the authorization type from the header")

	tokenMaker := token.GetJWTMaker()
	payload, err := tokenMaker.VerifyToken(authorizationFields[1])
	if err != nil {
		return nil, err
	}
	logger.Debug().Msgf("successfully verified the token %v", authorizationFields[1])
	return payload, nil
}
