package mailer

import (
	"fmt"
	"net/smtp"
	"strconv"

	"github.com/jordan-wright/email"
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
	smtpServerAddress string
}

func (sender *GmailSender) SendEmail(
	subject string,
	body string,
	to []string,
	cc []string,
	bcc []string,
) error {
	e := email.NewEmail()
	e.From = fmt.Sprintf("%s <%s>", sender.name, sender.fromEmailAddress)
	e.Subject = subject
	e.HTML = []byte(body)
	e.To = to
	e.Cc = cc
	e.Bcc = bcc

	auth := smtp.PlainAuth("", sender.fromEmailAddress, sender.fromEmailPass, sender.smtpAuthAddress)

	return e.Send(sender.smtpServerAddress, auth)
}

func NewGmailSender(name string, config util.Config) EmailSender {
	smtpServerAdd := config.EmailHost + ":" + strconv.Itoa(config.EmailHostPort)
	return &GmailSender{
		name:              name,
		fromEmailAddress:  config.FromEmailAddress,
		fromEmailPass:     config.FromEmailPass,
		smtpAuthAddress:   config.EmailHost,
		smtpServerAddress: smtpServerAdd,
	}
}
