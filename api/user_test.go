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

type eqCreateUserParams struct {
	arg      db.CreateUserParams
	password string
}

func (e eqCreateUserParams) Matches(x interface{}) bool {
	arg, ok := x.(db.CreateUserParams)
	if !ok {
		return false
	}

	err := util.ComparePasswords(arg.HashedPassword, e.password)
	if err != nil {
		return false
	}

	e.arg.HashedPassword = arg.HashedPassword
	return reflect.DeepEqual(e.arg, arg)
}

func (e eqCreateUserParams) String() string {
	return fmt.Sprintf("matches arg %v and password %v", e.arg, e.password)
}

func EqCreateUserParams(arg db.CreateUserParams, password string) gomock.Matcher {
	return eqCreateUserParams{arg, password}
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
				arg := db.CreateUserParams{
					Username:       user1.Username,
					HashedPassword: user1.HashedPassword,
					FullName:       user1.FullName,
					Email:          user1.Email,
				}
				store.
					EXPECT().
					CreateUser(gomock.Any(), EqCreateUserParams(arg, user1Pass)).
					Times(1)

				store.
					EXPECT().
					GetUser(gomock.Any(), gomock.Eq(arg.Username)).
					Times(1).
					Return(db.User{
						Username:          arg.Username,
						HashedPassword:    arg.HashedPassword,
						FullName:          arg.FullName,
						Email:             arg.Email,
						PasswordChangedAt: user1.PasswordChangedAt,
						CreatedAt:         user1.CreatedAt,
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

			server := NewServer(store)
			recorder := httptest.NewRecorder()

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/api/v1/users"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}
