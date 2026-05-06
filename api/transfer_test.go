package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	mockdb "github.com/shivangp0208/bank_application/db/mock"
	db "github.com/shivangp0208/bank_application/db/sqlc"
	"github.com/shivangp0208/bank_application/token"
	"github.com/shivangp0208/bank_application/util"
	"github.com/stretchr/testify/require"
)

func TestTransferMoney(t *testing.T) {

	amount := int64(100)

	user1, _ := randomUser(t)
	user2, _ := randomUser(t)
	// user3, _ := randomUser(t)

	account1 := randomAccount(user1.Username)
	account2 := randomAccount(user2.Username)
	// account3 := randomAccount(user3.Username)

	account1.Currency = util.USD
	account2.Currency = util.USD
	// account3.Currency = util.EUR

	testCases := []struct {
		name          string
		body          gin.H
		setupAuth     func(t *testing.T, req *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"fromAccountId": account1.ID,
				"toAccountId":   account2.ID,
				"amount":        amount,
				"currency":      util.USD,
			},
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, req, tokenMaker, authorizationHeaderType, user1.Username, user1.Role, 15*time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.
					EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account1.ID)).
					Times(1).
					Return(account1, nil)
				store.
					EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account2.ID)).
					Times(1).
					Return(account2, nil)

				arg := db.TransferTxParams{
					FromAccountID: account1.ID,
					ToAccountID:   account2.ID,
					Amount:        amount,
				}

				store.
					EXPECT().
					TransferTx(gomock.Any(), gomock.Eq(arg)).
					Times(1)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/api/v1/transfer"
			req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)
			tc.setupAuth(t, req, server.TokenMaker)

			server.Router.ServeHTTP(recorder, req)
			tc.checkResponse(recorder)
		})
	}

}

func randomUser(t *testing.T) (user db.User, password string) {
	password = util.GenerateString(6)
	hashedPassword, err := util.GenerateHashPassword(password)
	require.NoError(t, err)

	user = db.User{
		Username:          util.GenerateRandomUsername(6),
		HashedPassword:    hashedPassword,
		Role:              util.User,
		IsVerified:        true,
		PasswordChangedAt: sql.NullTime{},
		CreatedAt:         time.Now(),
		FullName:          util.GenerateRandomFullName(6),
		Email:             util.GenerateRandomEmail(),
	}
	return
}
