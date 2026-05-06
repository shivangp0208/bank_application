package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	mockdb "github.com/shivangp0208/bank_application/db/mock"
	db "github.com/shivangp0208/bank_application/db/sqlc"
	"github.com/shivangp0208/bank_application/token"
	"github.com/shivangp0208/bank_application/util"
	"github.com/stretchr/testify/require"
)

func TestGetAccountByID(t *testing.T) {

	username := "testUser"
	account := randomAccount(username)
	testCases := []struct {
		name      string
		accountId uint64
		setupAuth func(t *testing.T, req *http.Request, tokenMaker token.Maker)
		// buildStubs func takes a mock object as arg and configure that mock object
		// to return a specific response in a particular case to cover all edge case
		buildStubs func(store *mockdb.MockStore)
		// checkResponse func takes an testing object and a http recorder as arg
		// to check for the response stored in the recorder and send the required
		// test result from testing object
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "Unauthenticated",
			accountId: account.ID,
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				// no authorization needed
			},
			buildStubs: func(store *mockdb.MockStore) {
				// we don't need any stubs in this
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:      "OK",
			accountId: account.ID,
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, req, tokenMaker, authorizationHeaderType, username, util.User, 15*time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.
					EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(account, nil)
			}, checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireAccountMatch(t, recorder.Body, account)
			},
		},
		{
			name:      "Invalid ID",
			accountId: 0,
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, req, tokenMaker, authorizationHeaderType, username, util.User, 15*time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.
					EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
			}, checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:      "Not Found",
			accountId: account.ID,
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, req, tokenMaker, authorizationHeaderType, username, util.User, 15*time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.
					EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Account{}, sql.ErrNoRows)
			}, checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// creating a new gomock controller to control the mock objects like it's scope and lifecycle
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// this store is a new mock object created which is managed by ctrl mock controller
			store := mockdb.NewMockStore(ctrl)
			// this build the stubs to define about how this mock object should behave in case of
			// GetAccount func in DB is called
			tc.buildStubs(store)

			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/api/v1/accounts/%d", tc.accountId)
			req, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)
			tc.setupAuth(t, req, server.TokenMaker)

			server.Router.ServeHTTP(recorder, req)
			tc.checkResponse(t, recorder)
		})
	}
}

func randomAccount(username string) db.Account {
	return db.Account{
		ID:        util.GenerateRandomID(),
		Owner:     username,
		Balance:   util.GenerateRandomAmount(),
		Currency:  util.GenerateRandomCurrency(),
		CreatedAt: time.Now().UTC(),
	}
}

func requireAccountMatch(t *testing.T, body *bytes.Buffer, excpectedAccount db.Account) {

	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotAccount db.Account
	err = json.Unmarshal(data, &gotAccount)
	require.NoError(t, err)

	require.Equal(t, excpectedAccount, gotAccount)
}
