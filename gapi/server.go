package gapi

import (
	db "github.com/shivangp0208/bank_application/db/sqlc"
	"github.com/shivangp0208/bank_application/pb"
	"github.com/shivangp0208/bank_application/token"
	"github.com/shivangp0208/bank_application/util"
	"github.com/shivangp0208/bank_application/worker"
)

type Server struct {
	pb.UnimplementedSimpleBankServer
	store        db.Store
	tokenMaker   token.Maker
	config       *util.Config
	taskProducer worker.TaskProducer
}

func NewServer(store db.Store, config util.Config, producer worker.TaskProducer) (*Server, error) {
	jwtMaker, err := token.NewJwtMaker(config.AccessTokenSecretKey)
	if err != nil {
		return nil, err
	}

	// defining a server with sb configuration
	server := &Server{
		config:       &config,
		store:        store,
		tokenMaker:   jwtMaker,
		taskProducer: producer,
	}

	return server, nil
}
