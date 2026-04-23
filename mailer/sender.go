package mailer

import (
	"strconv"

	"gopkg.in/gomail.v2"

	"github.com/shivangp0208/bank_application/util"
)

type EmailSender interface {
	SendEmail(
		subject string,
		body string,
		to []string,
		cc []string,
		bcc []string,
	) error
}

var logger = util.GetLogger()

type GmailSender struct {
	name             string
	fromEmailAddress string
	fromEmailPass    string

	smtpAuthAddress   string
	smtpServerPort    int
	smtpServerAddress string
}

func (sender *GmailSender) SendEmail(
	subject string,
	body string,
	to []string,
	cc []string,
	bcc []string,
) error {

	m := gomail.NewMessage()
	m.SetHeader("From", sender.fromEmailAddress)
	m.SetHeader("To", to...)
	m.SetHeader("Cc", cc...)
	m.SetHeader("Bcc", bcc...)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	dialer := gomail.NewDialer(sender.smtpAuthAddress, sender.smtpServerPort, sender.fromEmailAddress, sender.fromEmailPass)

	if err := dialer.DialAndSend(m); err != nil {
		logger.Error().Str("from", sender.fromEmailAddress).Msgf("unable to send the email: %v", err)
		return err
	}
	logger.Info().Str("from", sender.fromEmailAddress).Str("subject", subject).Msgf("success sending the email")
	return nil
}

func NewGmailSender(name string, config util.Config) EmailSender {
	smtpServerAdd := config.EmailHost + ":" + strconv.Itoa(config.EmailHostPort)
	return &GmailSender{
		name:              name,
		fromEmailAddress:  config.FromEmailAddress,
		fromEmailPass:     config.FromEmailPass,
		smtpAuthAddress:   config.EmailHost,
		smtpServerAddress: smtpServerAdd,
		smtpServerPort:    config.EmailHostPort,
	}
}
