package token

import (
	"testing"
	"time"

	"github.com/shivangp0208/bank_application/util"
	"github.com/stretchr/testify/require"
)

func TestJWTMaker(t *testing.T) {
	jwtMaker, err := NewJwtMaker(util.GenerateRandomName(32))
	require.NoError(t, err)

	username := util.GenerateRandomName(8)
	duration := time.Minute
	issuedAt := time.Now()
	expiredAt := issuedAt.Add(duration)

	jwtToken, jwtPayload, err := jwtMaker.CreateToken(username, duration)
	require.NoError(t, err)
	require.NotEmpty(t, jwtToken)
	require.NotEmpty(t, jwtPayload)

	payload, err := jwtMaker.VerifyToken(jwtToken)
	require.NoError(t, err)
	require.NotEmpty(t, payload)

	require.NotZero(t, payload.ID)
	require.Equal(t, username, payload.Username)
	require.WithinDuration(t, issuedAt, payload.IssuedAt, time.Second)
	require.WithinDuration(t, expiredAt, payload.ExpiredAt, time.Second)
}

func TestExpiredToken(t *testing.T) {
	jwtMaker, err := NewJwtMaker(util.GenerateRandomName(32))
	require.NoError(t, err)

	username := util.GenerateRandomName(8)
	duration := -time.Second

	jwtToken, jwtPayload, err := jwtMaker.CreateToken(username, duration)
	require.NoError(t, err)
	require.NotEmpty(t, jwtToken)
	require.NotEmpty(t, jwtPayload)

	payload, err := jwtMaker.VerifyToken(jwtToken)
	require.Error(t, err)
	require.EqualError(t, err, ErrExpiredToken.Error())
	require.Nil(t, payload)
}
