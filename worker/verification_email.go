package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/shivangp0208/bank_application/util"
)

var logger = util.GetLogger()

const (
	TypeEmailDelivery = "task:send_verification_email"
)

type EmailDeliveryPayload struct {
	Username string `json:"username"`
}

func (producer *RedisTaskProducer) ProduceSendVerificationEmail(ctx context.Context, username string, payload *EmailDeliveryPayload, opts ...asynq.Option) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("unable to marshal the json payload for sending verification email %w", err)
	}

	newTask := asynq.NewTask(TypeEmailDelivery, jsonPayload, opts...)
	taskInfo, err := producer.client.EnqueueContext(ctx, newTask)
	if err != nil {
		return fmt.Errorf("unable to enque task in the async client %w", err)
	}

	logger.Info().Str("type", taskInfo.Type).Bytes("payload", taskInfo.Payload).Str("queue_name", taskInfo.Queue).Msg("successfully produced send verification email event as async operation")
	return nil
}

func (processor *RedisTaskProcessor) ProcessSendVerificationEmail(ctx context.Context, task *asynq.Task) error {

	var emailPayload EmailDeliveryPayload

	if err := json.Unmarshal(task.Payload(), &emailPayload); err != nil {
		return fmt.Errorf("unable to unmarshal the json payload from async queue %w", err)
	}

	user, err := processor.store.GetUser(ctx, emailPayload.Username)
	if err != nil {
		return fmt.Errorf("unable to get the user data from username %s %w", emailPayload.Username, err)
	}

	logger.Info().Str("type", task.Type()).Bytes("payload", task.Payload()).Str("username", user.Email).Msg("successfully processed send verification email")

	return nil
}

func (processor *RedisTaskProcessor) Start() error {
	asyncMux := asynq.NewServeMux()

	asyncMux.HandleFunc(TypeEmailDelivery, processor.ProcessSendVerificationEmail)

	return processor.server.Start(asyncMux)
}
