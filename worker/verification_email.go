package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	db "github.com/shivangp0208/bank_application/db/sqlc"
	"github.com/shivangp0208/bank_application/util"
)

var logger = util.GetLogger()

const (
	TypeEmailDelivery = "task:send_verification_email"
)

type EmailDeliveryPayload struct {
	Username string `json:"username"`
}

func (producer *RedisTaskProducer) ProduceSendVerificationEmail(ctx context.Context, payload *EmailDeliveryPayload, opts ...asynq.Option) error {
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
		logger.Error().Bytes("payload", task.Payload()).Msgf("failed to unmarshal the json payload from async queue: %v", err)
		return fmt.Errorf("unable to unmarshal the json payload from async queue %w", err)
	}

	user, err := processor.store.GetUser(ctx, emailPayload.Username)
	if err != nil {
		logger.Error().Bytes("payload", task.Payload()).Str("username", user.Email).Msgf("failed to get the user detail from email payload username: %v", err)
		return fmt.Errorf("unable to get the user data from username %s %w", emailPayload.Username, err)
	}

	// creating a verify email db entry to maintain the verification of email record
	arg := db.CreateVerifiyEmailParams{
		Username:   user.Username,
		Email:      user.Email,
		SecretCode: util.GenerateRandomName(36),
	}

	sqlRes, err := processor.store.CreateVerifiyEmail(ctx, arg)
	id, err := sqlRes.LastInsertId()
	if err != nil {
		logger.Error().Bytes("payload", task.Payload()).Any("CreateVerifiyEmailParams", arg).Msgf("failed to record the data in db for verifying email: %v", err)
		return fmt.Errorf("failed to record the data in db for verifying email: %v", err)
	}

	createdVerifyEmail, err := processor.store.GetVerifiyEmail(ctx, uint64(id))
	if err != nil {
		logger.Error().Int("id", int(id)).Any("CreateVerifiyEmailParams", arg).Msgf("failed to get the saved record for verify email: %v", err)
		return fmt.Errorf("failed to get the save record for verify email: %v", err)
	}

	// email definition
	subject, msg := GetVerificationMail(createdVerifyEmail.Username, createdVerifyEmail.SecretCode, processor.config)

	// sending the email for verification
	processor.emailSender.SendEmail(subject, msg, []string{createdVerifyEmail.Email}, nil, nil)

	logger.Info().Str("type", task.Type()).Bytes("payload", task.Payload()).Str("username", user.Email).Msg("successfully processed send verification email")

	return nil
}

func (processor *RedisTaskProcessor) Start() error {
	asyncMux := asynq.NewServeMux()

	asyncMux.HandleFunc(TypeEmailDelivery, processor.ProcessSendVerificationEmail)

	return processor.server.Start(asyncMux)
}

func GetVerificationMail(username string, secret string, config *util.Config) (subject string, body string) {
	subject = "Verify Email For Bank"

	verificationLink := fmt.Sprintf("http://localhost:%s/api/v1/verify?username=%s&secret_code=%s", config.MailVerificationPort, username, secret)

	body = fmt.Sprintf(`<!DOCTYPE html>
<html>
<body style="margin:0;padding:0;font-family:Arial,sans-serif;background:#f4f4f4;">
  <table width="100%%" cellpadding="0" cellspacing="0" style="padding:40px 0;">
    <tr><td align="center">
      <table width="560" cellpadding="0" cellspacing="0" style="background:#ffffff;border-radius:8px;overflow:hidden;">
        
        <!-- Header -->
        <tr>
          <td style="background:#0C447C;padding:32px;text-align:center;">
            <span style="color:#ffffff;font-size:20px;font-weight:bold;">SecureBank</span>
          </td>
        </tr>

        <!-- Body -->
        <tr>
          <td style="padding:32px 40px;">
            <p style="margin:0 0 8px;font-size:14px;color:#555555;">Hi <strong>%s</strong>,</p>
            <h2 style="margin:0 0 12px;font-size:20px;color:#111111;font-weight:600;">Verify your email address</h2>
            <p style="margin:0 0 24px;font-size:14px;color:#555555;line-height:1.7;">
              Thank you for creating your SecureBank account. To complete your 
              registration and activate your account, please verify your email 
              address by clicking the button below.
            </p>

            <!-- Button -->
            <table cellpadding="0" cellspacing="0" width="100%%">
              <tr><td align="center" style="padding:8px 0 24px;">
                <a href="%s" style="background:#0C447C;color:#ffffff;text-decoration:none;padding:14px 32px;border-radius:6px;font-size:14px;font-weight:bold;display:inline-block;">
                  Verify my email address
                </a>
              </td></tr>
            </table>

            <p style="margin:0 0 16px;font-size:13px;color:#555555;line-height:1.7;">
              This link will expire in <strong>24 hours</strong>. If you did not 
              create a SecureBank account, you can safely ignore this email.
            </p>

            <!-- Fallback link box -->
            <table cellpadding="0" cellspacing="0" width="100%%">
              <tr><td style="background:#f8f8f8;border-radius:6px;padding:16px;">
                <p style="margin:0 0 6px;font-size:12px;color:#888888;font-weight:bold;">Or copy this link into your browser:</p>
                <p style="margin:0;font-size:12px;color:#185FA5;word-break:break-all;font-family:monospace;">%s</p>
              </td></tr>
            </table>
          </td>
        </tr>

        <!-- Footer -->
        <tr>
          <td style="border-top:1px solid #eeeeee;padding:20px 40px;text-align:center;">
            <p style="margin:0 0 4px;font-size:12px;color:#aaaaaa;">SecureBank Inc. · 123 Finance Street, Mumbai 400001</p>
            <p style="margin:0;font-size:12px;color:#aaaaaa;">Need help? <a href="mailto:support@securebank.com" style="color:#185FA5;">support@securebank.com</a></p>
          </td>
        </tr>

      </table>
    </td></tr>
  </table>
</body>
</html>`, username, verificationLink, verificationLink)

	return
}
