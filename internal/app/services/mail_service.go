package services

import (
	"bytes"
	"text/template"

	"github.com/agilistikmal/parkingo-core/internal/app/templates"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gopkg.in/gomail.v2"
)

type MailService struct {
	Dialer   *gomail.Dialer
	Template *template.Template
}

func NewMailService() *MailService {
	host := viper.GetString("mail.host")
	port := viper.GetInt("mail.port")
	username := viper.GetString("mail.username")
	password := viper.GetString("mail.password")

	dialer := gomail.NewDialer(host, port, username, password)

	tmpl, err := template.ParseFS(templates.EmailTemplateFS, "email.html")
	if err != nil {
		logrus.Fatal("Failed to parse email template: ", err)
	}

	return &MailService{
		Dialer:   dialer,
		Template: tmpl,
	}
}

func (s *MailService) SendMail(to string, subject string, content string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", s.Dialer.Username)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)

	data := map[string]interface{}{
		"Subject": subject,
		"Content": content,
	}

	var body bytes.Buffer
	if err := s.Template.Execute(&body, data); err != nil {
		logrus.Fatal("Gagal render template:", err)
	}

	m.SetBody("text/html", body.String())

	err := s.Dialer.DialAndSend(m)
	if err != nil {
		return err
	}

	return nil
}
