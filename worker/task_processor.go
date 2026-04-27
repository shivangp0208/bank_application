package worker

import (
	"context"

	"github.com/hibiken/asynq"
	db "github.com/shivangp0208/bank_application/db/sqlc"
	"github.com/shivangp0208/bank_application/mailer"
	"github.com/shivangp0208/bank_application/util"
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
	server      *asynq.Server
	store       db.Store
	emailSender mailer.EmailSender
	config      *util.Config
}

func NewRedisTaskProcessor(redisOpt asynq.RedisClientOpt, store db.Store, emailSender mailer.EmailSender, config *util.Config) TaskProcessor {
	server := asynq.NewServer(redisOpt, asynq.Config{
		Queues: map[string]int{
			CriticalQueue: 10,
			DefaultQueue:  5,
		},
		ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
			logger.Error().Str("task_type", task.Type()).Str("task_payload", string(task.Payload())).Msgf("unable to send the verification email %v", err)
		}),
	})

	return &RedisTaskProcessor{
		server:      server,
		store:       store,
		emailSender: emailSender,
		config:      config,
	}
}
