package api

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/shivangp0208/bank_application/token"
)

const (
	authorizationHeaderKey  = "authorization"
	authorizationHeaderType = "bearer"
	authorizationPayloadKey = "authorization_payload"
)

func authMiddleware(tokenMaker token.Maker) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authrizationHeader := ctx.GetHeader(authorizationHeaderKey)

		if len(authrizationHeader) == 0 {
			err := fmt.Errorf("invalid authorization header cannot have empty header")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		headerParts := strings.Split(authrizationHeader, " ")
		if len(headerParts) < 2 {
			err := errors.New("invalid authorization header format")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		if strings.ToLower(headerParts[0]) != authorizationHeaderType {
			err := fmt.Errorf("invalid authorization header unsupported authprization type")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		accessToken := headerParts[1]
		payload, err := tokenMaker.VerifyToken(accessToken)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		ctx.Set(authorizationPayloadKey, payload)
		ctx.Next()
	}
}
