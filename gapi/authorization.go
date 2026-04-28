package gapi

import (
	"context"
	"fmt"
	"strings"

	"github.com/shivangp0208/bank_application/token"
	"google.golang.org/grpc/metadata"
)

const (
	authorizationHeaderKey = "authorization"
	authorizationType      = "bearer"
)

func (s *Server) authorizeUser(ctx context.Context) (*token.Payload, error) {
	mtdt, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("unable to get metadata from context %v", mtdt)
	}
	logger.Debug().Msgf("successfully got the metadata from the incoming context %v", mtdt)

	fields := mtdt.Get(authorizationHeaderKey)
	authorizationFields := strings.Fields(fields[0])
	if len(authorizationFields) < 2 {
		return nil, fmt.Errorf("invalid authorization header format in metadata")
	}
	logger.Debug().Msgf("successfully got the authorization header from metadata %v", authorizationFields)

	authType := strings.ToLower(authorizationFields[0])
	if authType != authorizationType {
		return nil, fmt.Errorf("invalid authorization type in authorization %s", authType)
	}
	logger.Debug().Msg("successfully validated the authorization type from the header")

	payload, err := s.tokenMaker.VerifyToken(authorizationFields[1])
	if err != nil {
		return nil, err
	}
	logger.Debug().Msgf("successfully verified the token %v", authorizationFields[1])
	return payload, nil
}
