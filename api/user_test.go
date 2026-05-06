package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	mockdb "github.com/shivangp0208/bank_application/db/mock"
	db "github.com/shivangp0208/bank_application/db/sqlc"
	"github.com/shivangp0208/bank_application/util"
	"github.com/stretchr/testify/require"
)

type eqCreateUserTxParams struct {
	arg      db.CreateUserTxParams
	password string
}

func (e eqCreateUserTxParams) Matches(x interface{}) bool {
	arg, ok := x.(db.CreateUserTxParams)
	if !ok {
		return false
	}
	arg.AfterCreateUser = nil

	err := util.ComparePasswords(arg.User.HashedPassword, e.password)
	if err != nil {
		return false
	}

	e.arg.User.HashedPassword = arg.User.HashedPassword
	return reflect.DeepEqual(e.arg, arg)
}

func (e eqCreateUserTxParams) String() string {
	return fmt.Sprintf("matches arg %v and password %v", e.arg, e.password)
}

func EqCreateUserParams(arg db.CreateUserTxParams, password string) gomock.Matcher {
	return eqCreateUserTxParams{arg, password}
}

func TestCreateUser(t *testing.T) {

	user1, user1Pass := randomUser(t)
	// user2, user2Pass := randomUser(t)

	testCases := []struct {
		name          string
		body          gin.H
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"username":  user1.Username,
				"password":  user1Pass,
				"full_name": user1.FullName,
				"email":     user1.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateUserTxParams{
					User: db.CreateUserParams{
						Username:       user1.Username,
						HashedPassword: user1.HashedPassword,
						FullName:       user1.FullName,
						Email:          user1.Email,
					},
				}
				store.
					EXPECT().
					CreateUserTx(gomock.Any(), EqCreateUserParams(arg, user1Pass)).
					Times(1).
					Return(db.CreateUserTxResult{
						User: user1,
					}, nil)

			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusCreated, recorder.Code)
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

			url := "/api/v1/users"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)
			request.Header.Set("Content-Type", "application/json")

			server.Router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}
