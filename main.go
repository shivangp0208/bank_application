package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/hibiken/asynq"
	"github.com/shivangp0208/bank_application/api"
	db "github.com/shivangp0208/bank_application/db/sqlc"
	"github.com/shivangp0208/bank_application/gapi"
	"github.com/shivangp0208/bank_application/mailer"
	"github.com/shivangp0208/bank_application/pb"
	"github.com/shivangp0208/bank_application/util"
	"github.com/shivangp0208/bank_application/worker"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
)

var interruptSignals = []os.Signal{
	os.Interrupt,
	syscall.SIGTERM,
	syscall.SIGINT,
}

var conn *sql.DB
var err error
var config util.Config
var logger = util.GetLogger()

func init() {
	gin.SetMode(gin.ReleaseMode)
	config = util.GetConfig()
	conn, err = sql.Open(config.DBDriver, config.DBSource)
	logger.Info().Msgf("init main: dbDriver: %s and dbSource: %s", config.DBDriver, config.DBSource)
	if err != nil {
		logger.Err(fmt.Errorf("unable to open db connection: %v", err))
	}
	logger.Info().Msgf("successfull opening up the db connection")
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), interruptSignals...)
	defer stop()

	store := db.NewStore(conn)

	redisOpt := asynq.RedisClientOpt{
		Addr: config.RedisServerAddress,
	}

	taskProducer := worker.NewRedisTaskProducer(redisOpt)

	waitGroup, ctx := errgroup.WithContext(ctx)

	runTaskProcessorServer(ctx, waitGroup, redisOpt, store)
	startGRPCSever(ctx, waitGroup, store, taskProducer)
	startGRPCGatewaySever(ctx, waitGroup, store, taskProducer)
	startGinSever(ctx, waitGroup, store, taskProducer)

	err = waitGroup.Wait()
	if err != nil {
		logger.Fatal().Msgf("error from wait group: %v", err)
	}
}

func runTaskProcessorServer(ctx context.Context, waitGroup *errgroup.Group, redisOpt asynq.RedisClientOpt, store db.Store) {

	emailSender := mailer.NewGmailSender(config)

	taskProcessor := worker.NewRedisTaskProcessor(redisOpt, store, emailSender, &config)
	logger.Info().Msg("initailizing and starting task processor server")

	if err := taskProcessor.Start(); err != nil {
		logger.Fatal().Msgf("unable to start the task processor server %v", err)
	}

	waitGroup.Go(func() error {
		<-ctx.Done()
		logger.Info().Msg("gracefully shutting down the task processor")

		taskProcessor.Shutdown()
		logger.Info().Msg("task processor gracefully shutted down")

		return nil
	})
}

func startGinSever(ctx context.Context, waitGroup *errgroup.Group, store db.Store, taskProducer worker.TaskProducer) {
	server, err := api.NewServer(store, config, taskProducer)
	if err != nil {
		logger.Err(fmt.Errorf("unable to create Gin server due to err %v", err))
	}

	httpServer := &http.Server{
		Addr:    config.GinHTTPServerAddress,
		Handler: server.Router,
	}

	waitGroup.Go(func() error {
		err = httpServer.ListenAndServe()
		// err = server.Start(config.GinHTTPServerAddress)
		logger.Info().Msgf("Gin server listnening on address %s", config.GinHTTPServerAddress)
		if err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return nil
			}
			logger.Err(fmt.Errorf("unable to start the Gin server with address %s due to err %v", config.GinHTTPServerAddress, err))
			return err
		}
		return nil
	})

	waitGroup.Go(func() error {
		<-ctx.Done()
		logger.Info().Msg("gracefully shutting down the gin server")

		httpServer.Shutdown(context.Background())
		logger.Info().Msg("gin server gracefully shutted down")

		return nil
	})
}

func startGRPCSever(ctx context.Context, waitGroup *errgroup.Group, store db.Store, taskProducer worker.TaskProducer) {

	server, err := gapi.NewServer(store, config, taskProducer)
	if err != nil {
		logger.Err(fmt.Errorf("unable to create the grpc server due to %v", err))
	}

	// creating a new grpc server instance
	grpcLogger := grpc.UnaryInterceptor(util.GRPCLoggerInterceptor)
	grpcServer := grpc.NewServer(grpcLogger)

	// registering the grpc server by giving an grpc server instance and a server instance conatining all api's
	pb.RegisterSimpleBankServer(grpcServer, server)
	reflection.Register(grpcServer)

	// listen on a tcp port to handle grpc req
	lis, err := net.Listen("tcp", config.GRPCServerAddress)
	if err != nil {
		logger.Err(fmt.Errorf("unable to create grpc listner due to err %v", err))
	}

	waitGroup.Go(func() error {
		logger.Info().Msgf("grpc server listnening on address %s", config.GRPCServerAddress)
		if err := grpcServer.Serve(lis); err != nil {
			if errors.Is(err, grpc.ErrServerStopped) {
				return nil
			}
			logger.Err(fmt.Errorf("unable to start the grpc server with address %s due to err %v", config.GRPCServerAddress, err))
			return err
		}
		return nil
	})

	waitGroup.Go(func() error {
		<-ctx.Done()
		logger.Info().Msgf("gracefully shutting down the grpc server with host: %s", config.GRPCServerAddress)

		grpcServer.GracefulStop()
		logger.Info().Msg("grpc server gracefully shutted down")

		return nil
	})
}

func startGRPCGatewaySever(ctx context.Context, waitGroup *errgroup.Group, store db.Store, taskProducer worker.TaskProducer) {

	server, err := gapi.NewServer(store, config, taskProducer)
	if err != nil {
		logger.Err(fmt.Errorf("unable to create the grpc gateway server due to %v", err))
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

	// registering the gateway handler to the grpc server
	err = pb.RegisterSimpleBankHandlerServer(ctx, gatewayMux, server)
	if err != nil {
		logger.Err(fmt.Errorf("unable to register the server to grpc gateway handler %v", err))
	}

	mux := http.NewServeMux()
	mux.Handle("/", gatewayMux)

	swaggerHandler := http.FileServer(http.Dir("doc/swagger"))
	mux.Handle("/api/swagger/ui", http.StripPrefix("/swagger/", swaggerHandler))

	httpServer := &http.Server{
		Addr:    config.GRPCGatewayServerAddress,
		Handler: util.HTTPLogger(mux),
	}

	waitGroup.Go(func() error {
		logger.Info().Msgf("grpc gateway server listnening on address %s", config.GRPCGatewayServerAddress)
		if err := httpServer.ListenAndServe(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return nil
			}
			logger.Err(fmt.Errorf("unable to start the grpc gateway server with address %s due to err %v", config.GRPCGatewayServerAddress, err))
			return err
		}
		return nil
	})

	waitGroup.Go(func() error {
		<-ctx.Done()
		logger.Info().Msgf("gracefully shutting down th grpc gateway server with host: %s", config.GRPCServerAddress)

		err = httpServer.Shutdown(context.Background())
		if err != nil {
			logger.Error().Msgf("unable to gracefully shutdown the grpc gateway server with host %s", config.GRPCGatewayServerAddress)
			return err
		}
		logger.Info().Msg("grpc gateway server gracefully shutted down")

		return nil
	})
}
