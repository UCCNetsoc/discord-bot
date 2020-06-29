package emails

import (
	"github.com/sendgrid/rest"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/spf13/viper"
)

func SendEmail(from, to, subject, content string) (*rest.Response, error) {
	fromAddress := mail.NewEmail(from, from)
	toAddress := mail.NewEmail(to, to)
	message := mail.NewSingleEmail(fromAddress, subject, toAddress, content, content)
	client := sendgrid.NewSendClient(viper.GetString("sendgrid.token"))
	response, err := client.Send(message)
	return response, err
}
