package emails

import (
	"github.com/sendgrid/rest"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/spf13/viper"
)

// SendEmail to end user.
func SendEmail(fromName, from, toName, to, subject, content, htmlContent string) (*rest.Response, error) {
	fromAddress := mail.NewEmail(fromName, from)
	toAddress := mail.NewEmail(toName, to)
	message := mail.NewSingleEmail(fromAddress, subject, toAddress, content, htmlContent)
	message.SetReplyTo(fromAddress)
	client := sendgrid.NewSendClient(viper.GetString("sendgrid.token"))
	response, err := client.Send(message)
	return response, err
}
