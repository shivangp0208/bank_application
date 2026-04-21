package main

import (
	"context"
	"database/sql"
	"net"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/shivangp0208/bank_application/api"
	db "github.com/shivangp0208/bank_application/db/sqlc"
	"github.com/shivangp0208/bank_application/gapi"
	"github.com/shivangp0208/bank_application/pb"
	"github.com/shivangp0208/bank_application/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
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
	store := db.NewStore(conn)

	go startGRPCSever(store)
	go startGRPCGatewaySever(store)
	startGinSever(store)
}

func startGinSever(store db.Store) {
	server, err := api.NewServer(store, config)
	if err != nil {
		logger.Fatalf("unable to create Gin server due to err %v", err)
	}

	err = server.Start(config.GinHTTPServerAddress)
	logger.Printf("Gin server listnening on address %s", config.GinHTTPServerAddress)
	if err != nil {
		logger.Fatalf("unable to start the Gin server with address %s due to err %v", config.GinHTTPServerAddress, err)
	}
}

func startGRPCSever(store db.Store) {

	server, err := gapi.NewServer(store, config)
	if err != nil {
		logger.Fatalf("unable to create the grpc server due to %v", err)
	}

	// creating a new grpc server instance
	grpcServer := grpc.NewServer()

	// grpc.UnaryInterceptor()

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

func startGRPCGatewaySever(store db.Store) {

	server, err := gapi.NewServer(store, config)
	if err != nil {
		logger.Fatalf("unable to create the grpc gateway server due to %v", err)
	}

	// this is a code snippet provided by grpc gateway to make the json format to snake case
	jsonOption := runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			UseProtoNames: true,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	})

	// creating a mux which is a handler for hadling all the REST req
	gatewayMux := runtime.NewServeMux(jsonOption)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// registering the gateway handler to the grpc server
	err = pb.RegisterSimpleBankHandlerServer(ctx, gatewayMux, server)
	if err != nil {
		logger.Fatalf("unable to register the server to grpc gateway handler %v", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", gatewayMux)

	swaggerHandler := http.FileServer(http.Dir("doc/swagger"))
	mux.Handle("/api/swagger/ui", http.StripPrefix("/swagger/", swaggerHandler))

	// listen on a tcp port to handle grpc req
	lis, err := net.Listen("tcp", config.GRPCGatewayServerAddress)
	if err != nil {
		logger.Fatalf("unable to create net listner due to err %v", err)
	}

	logger.Printf("grpc gateway server listnening on address %s", config.GRPCGatewayServerAddress)
	if err := http.Serve(lis, mux); err != nil {
		logger.Fatalf("unable to start the grpc gateway server with address %s due to err %v", config.GRPCGatewayServerAddress, err)
	}
}
