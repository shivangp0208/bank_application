package worker

import (
	"context"

	"github.com/hibiken/asynq"
	db "github.com/shivangp0208/bank_application/db/sqlc"
)

const (
	CriticalQueue = "critical"
	DefaultQueue  = "default"
)

type TaskProcessor interface {
	ProcessSendVerificationEmail(ctx context.Context, task *asynq.Task) error
	Start() error
}

type RedisTaskProcessor struct {
	server *asynq.Server
	store  db.Store
}

func NewRedisTaskProcessor(redisOpt asynq.RedisClientOpt, store db.Store) TaskProcessor {
	server := asynq.NewServer(redisOpt, asynq.Config{
		Queues: map[string]int{
			CriticalQueue: 10,
			DefaultQueue:  5,
		},
	})

	return &RedisTaskProcessor{
		server: server,
		store:  store,
	}
}
