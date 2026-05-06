package api

import (
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	db "github.com/shivangp0208/bank_application/db/sqlc"
	"github.com/shivangp0208/bank_application/util"
	"github.com/shivangp0208/bank_application/worker"
	"github.com/stretchr/testify/require"
)

func newTestServer(t *testing.T, store db.Store) *Server {
	config := util.Config{
		AccessTokenSecretKey:      util.GenerateString(32),
		AccessTokenExpirationTime: time.Minute,
	}
	server, err := NewServer(store, config, &worker.RedisTaskProducer{})
	require.NoError(t, err)

	return server
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)

	os.Exit(m.Run())
}
