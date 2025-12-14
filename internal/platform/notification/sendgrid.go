package notification

import (
	util "game-platform/internal/platform/utils"

	"github.com/sendgrid/rest"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

func SendGridMail(name, email, subject, fileName, token string) (*rest.Response, error) {
	from := mail.NewEmail("admin", "admin@vng.com.vn")
	to := mail.NewEmail(name, email)
	subjectMail := subject
	template := util.ParseHtml(fileName, map[string]string{
		"to":    email,
		"token": token,
	})

	message := mail.NewSingleEmail(from, subjectMail, to, "", template)
	client := sendgrid.NewSendClient(util.GodotEnv("SG_API_KEY"))
	return client.Send(message)
}
