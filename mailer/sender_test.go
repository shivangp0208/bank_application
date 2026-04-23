package mailer

import (
	"testing"

	"github.com/shivangp0208/bank_application/util"
	"github.com/stretchr/testify/require"
)

func TestSendEmail(t *testing.T) {

	mailer := NewGmailSender("Test User", util.GetConfig())

	body := `<div class="content"> <p>Hello,</p> <p>This is a dummy email template for testing purposes.</p> <p>You can use this layout to test email rendering, styles, and formatting.</p> <a href="#" class="button">Click Me</a> </div> <div class="footer"> <p>© 2026 Your Company. All rights reserved.</p> <p>This is an automated test email.</p> </div>`

	toList := []string{"patelshivang702@gmail.com"}

	err := mailer.SendEmail("Testing Mail", body, toList, nil, nil)

	require.NoError(t, err)
}
