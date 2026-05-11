package gapi

import (
	"github.com/shivangp0208/bank_application/config"
	db "github.com/shivangp0208/bank_application/db/sqlc"
	"github.com/shivangp0208/bank_application/pb"
	"github.com/shivangp0208/bank_application/token"
	"github.com/shivangp0208/bank_application/util"
	"github.com/shivangp0208/bank_application/worker"
)

const (
	AuthorizationHeaderKey  = "authorization"
	AuthorizationType       = "bearer"
	AuthorizationPayloadKey = "authorization_payload"
)

var Logger = util.GetLogger()

type Server struct {
	pb.UnimplementedSimpleBankServer
	store        db.Store
	tokenMaker   token.Maker
	config       *config.Config
	taskProducer worker.TaskProducer
}

func NewServer(store db.Store, config config.Config, producer worker.TaskProducer) (*Server, error) {
	jwtMaker := token.GetJWTMaker()

	// defining a server with sb configuration
	server := &Server{
		config:       &config,
		store:        store,
		tokenMaker:   jwtMaker,
		taskProducer: producer,
	}

	return server, nil
}
