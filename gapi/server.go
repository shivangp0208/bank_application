package gapi

import (
	db "github.com/shivangp0208/bank_application/db/sqlc"
	"github.com/shivangp0208/bank_application/pb"
	"github.com/shivangp0208/bank_application/token"
	"github.com/shivangp0208/bank_application/util"
)

type Server struct {
	pb.UnimplementedSimpleBankServer
	store      db.Store
	tokenMaker token.Maker
	config     *util.Config
}

func NewServer(store db.Store, config util.Config) (*Server, error) {
	jwtMaker, err := token.NewJwtMaker(config.AccessTokenSecretKey)
	if err != nil {
		return nil, err
	}

	// defining a server with sb configuration
	server := &Server{
		config:     &config,
		store:      store,
		tokenMaker: jwtMaker,
	}

	return server, nil
}
