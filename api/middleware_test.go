package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shivangp0208/bank_application/token"
	"github.com/shivangp0208/bank_application/util"
	"github.com/stretchr/testify/require"
)

// addAuthorization is a helper function which takes http req and token maker and claims for token as argument to create and test the token then appending the token to the req's authorization header
func addAuthorization(t *testing.T, req *http.Request, tokenMaker token.Maker, authorizationType string, username string, role string, duration time.Duration) {
	token, payload, err := tokenMaker.CreateToken(username, role, duration)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.NotEmpty(t, payload)

	authorizationHeader := fmt.Sprintf("%s %s", authorizationType, token)
	req.Header.Set(authorizationHeaderKey, authorizationHeader)
}

func TestAuthMiddleware(t *testing.T) {
	username := util.GenerateRandomUsername(8)

	testCases := []struct {
		name          string
		setupAuth     func(t *testing.T, req *http.Request, tokenMaker token.Maker)
		checkResponse func(t *testing.T, res *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, req, tokenMaker, authorizationHeaderType, username, util.User, time.Minute)
			},
			checkResponse: func(t *testing.T, res *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, res.Code)
			},
		},
		{
			name: "NoAuthorization",
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, req, tokenMaker, "", username, util.User, time.Minute)
			},
			checkResponse: func(t *testing.T, res *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, res.Code)
			},
		},
		{
			name: "UnsupportedAuthorization",
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, req, tokenMaker, "unsupported", username, util.User, time.Minute)
			},
			checkResponse: func(t *testing.T, res *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, res.Code)
			},
		},
		{
			name: "ExpiredToken",
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, req, tokenMaker, authorizationHeaderType, username, util.User, -time.Minute)
			},
			checkResponse: func(t *testing.T, res *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, res.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := newTestServer(t, nil)
			authPath := "/auth"

			server.Router.GET(
				authPath,
				authMiddleware(server.TokenMaker),
				func(ctx *gin.Context) {
					ctx.JSON(http.StatusOK, gin.H{})
				})

			recorder := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodGet, authPath, nil)
			require.NoError(t, err)

			tc.setupAuth(t, req, server.TokenMaker)
			server.Router.ServeHTTP(recorder, req)
			tc.checkResponse(t, recorder)
		})
	}

}
