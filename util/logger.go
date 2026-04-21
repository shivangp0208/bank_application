package util

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

var logger zerolog.Logger

func GetLogger() zerolog.Logger {
	logger = zerolog.New(os.Stderr).With().Timestamp().Logger().Output(zerolog.ConsoleWriter{Out: os.Stderr})
	return logger
}

func GRPCLoggerInterceptor(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (resp any, err error) {

	logger.Info().Msgf("intercepting the method %v", info.FullMethod)
	startTime := time.Now()
	res, err := handler(ctx, req)

	endTime := time.Since(startTime)
	logger.Info().Msgf("time taken for the method %v : %v", info.FullMethod, endTime)
	return res, err
}

func HTTPLogger(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		logger.Info().Msgf("intercepting the method %v", r.Method)
		startTime := time.Now()

		handler.ServeHTTP(w, r)

		endTime := time.Since(startTime)
		logger.Info().Msgf("time taken for the method %v : %v", r.Method, endTime)
	})
}
