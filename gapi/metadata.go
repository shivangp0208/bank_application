package gapi

import (
	"context"

	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

type Metadata struct {
	UserAgent string
	ClientIP  string
}

const (
	grpcUserAgentKey        = "user-agent"
	grpcGatewayUserAgentKey = "grpcgateway-user-agent"
	grpcGatewayClientIPKey  = "x-forwarded-for"
)

func (s *Server) extractMetadata(ctx context.Context) *Metadata {
	var md Metadata
	if data, ok := metadata.FromIncomingContext(ctx); ok {
		logger.Printf("got the meta data from the context \n%v", data)
		if userAgents := data.Get(grpcGatewayUserAgentKey); len(userAgents) > 0 {
			logger.Printf("grpcgateway-user-agent val in context : %v", userAgents)
			md.UserAgent = userAgents[0]
		} else if userAgents := data.Get(grpcUserAgentKey); len(userAgents) > 0 {
			logger.Printf("user-agent val in context : %v", userAgents)
			md.UserAgent = userAgents[0]
		}

		if clientIP := data.Get(grpcGatewayClientIPKey); len(clientIP) > 0 {
			logger.Printf("x-forwarded-for val in context : %v", clientIP)
			md.ClientIP = clientIP[0]
		}
	}
	if p, ok := peer.FromContext(ctx); ok {
		logger.Printf("the peer info from the context \n%v", p)
		if p.Addr != nil {
			logger.Printf("client ip val in context : %v", p.Addr.String())
			md.ClientIP = p.Addr.String()
		}
	}
	return &md
}
