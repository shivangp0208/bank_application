package worker

import (
	"context"

	"github.com/hibiken/asynq"
)

type TaskProducer interface {
	ProduceSendVerificationEmail(ctx context.Context, payload *EmailDeliveryPayload, opts ...asynq.Option) error
}

type RedisTaskProducer struct {
	client *asynq.Client
}

func NewRedisTaskProducer(redisOptClient asynq.RedisClientOpt) TaskProducer {
	redisClient := asynq.NewClient(redisOptClient)
	return &RedisTaskProducer{
		client: redisClient,
	}
}
