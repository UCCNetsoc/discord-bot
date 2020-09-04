package emails

import (
	"github.com/sendgrid/rest"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/spf13/viper"
)

// SendEmail to end user.
func SendEmail(from, to, subject, content, htmlContent string) (*rest.Response, error) {
	fromAddress := mail.NewEmail(from, from)
	toAddress := mail.NewEmail(to, to)
	message := mail.NewSingleEmail(fromAddress, subject, toAddress, content, htmlContent)
	client := sendgrid.NewSendClient(viper.GetString("sendgrid.token"))
	response, err := client.Send(message)
	return response, err
}
