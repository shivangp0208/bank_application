package main

import (
	"database/sql"
	"net"
	"sync"

	_ "github.com/go-sql-driver/mysql"
	"github.com/shivangp0208/bank_application/api"
	db "github.com/shivangp0208/bank_application/db/sqlc"
	"github.com/shivangp0208/bank_application/gapi"
	"github.com/shivangp0208/bank_application/pb"
	"github.com/shivangp0208/bank_application/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var conn *sql.DB
var err error
var config util.Config
var logger = util.GetLogger()

func init() {
	config = util.GetConfig()
	conn, err = sql.Open(config.DBDriver, config.DBSource)
	logger.Printf("init main: dbDriver: %s and dbSource: %s", config.DBDriver, config.DBSource)
	if err != nil {
		logger.Fatalf("unable to open db connection: %v", err)
	}
	logger.Printf("successfull opening up the db connection")
}

func main() {
	var wg sync.WaitGroup
	store := db.NewStore(conn)

	wg.Go(func() {
		startGinSever(store)
	})
	wg.Go(func() {
		startGRPCSever(store)
	})

	wg.Wait()
}

func startGinSever(store db.Store) {
	server, err := api.NewServer(store, config)
	if err != nil {
		logger.Fatalf("unable to create Gin server due to err %v", err)
	}

	err = server.Start(config.HTTPServerAddress)
	logger.Printf("Gin server listnening on address %s", config.HTTPServerAddress)
	if err != nil {
		logger.Fatalf("unable to start the Gin server with address %s due to err %v", config.HTTPServerAddress, err)
	}
}

func startGRPCSever(store db.Store) {

	server, err := gapi.NewServer(store, config)
	if err != nil {
		logger.Fatalf("unable to create the grpc server due to %v", err)
	}

	// creating a new grpc server instance
	grpcServer := grpc.NewServer()

	// registering the grpc server by giving an grpc server instance and a server instance conatining all api's
	pb.RegisterSimpleBankServer(grpcServer, server)
	reflection.Register(grpcServer)

	// listen on a tcp port to handle grpc req
	lis, err := net.Listen("tcp", config.GRPCServerAddress)
	if err != nil {
		logger.Fatalf("unable to create grpc listner due to err %v", err)
	}

	logger.Printf("grpc server listnening on address %s", config.GRPCServerAddress)
	if err := grpcServer.Serve(lis); err != nil {
		logger.Fatalf("unable to start the grpc server with address %s due to err %v", config.GRPCServerAddress, err)
	}
}
