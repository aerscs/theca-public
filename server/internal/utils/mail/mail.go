package mail

import (
	"bytes"
	"context"
	"fmt"
	"text/template"
	"time"

	"github.com/OxytocinGroup/theca-v3/internal/config"
	"github.com/resend/resend-go/v2"
)

// Mailer интерфейс для отправки почты
type Mailer interface {
	SendVerificationEmail(email, code, username string) error
	SendResetEmail(email, username, token string) error
}

// Mail структура для данных письма
type Mail struct {
	Email    string
	Username string
	Code     string
}

// mailer реализация интерфейса Mailer
type mailer struct {
	client *resend.Client
	from   string
}

// NewMailer создает новый экземпляр mailer
func NewMailer(cfg *config.Config) Mailer {
	return &mailer{
		client: resend.NewClient(cfg.SMTPAPIKey),
		from:   "Theca <no-reply@theca.oxytocingroup.com>",
	}
}

// sendEmail общий метод для отправки почты с таймаутом 10 секунд
func (m *mailer) sendEmail(to, subject, templatePath string, data Mail) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	var tpl bytes.Buffer
	if err := tmpl.Execute(&tpl, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	params := &resend.SendEmailRequest{
		From:    m.from,
		To:      []string{to},
		Html:    tpl.String(),
		Subject: subject,
	}

	_, err = m.client.Emails.SendWithContext(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// SendVerificationEmail отправляет письмо для верификации
func (m *mailer) SendVerificationEmail(email, code, username string) error {
	return m.sendEmail(
		email,
		fmt.Sprintf("%s | Verification Code", code),
		"templates/verifyMail.html",
		Mail{Username: username, Code: code},
	)
}

// SendResetEmail отправляет письмо для сброса пароля
func (m *mailer) SendResetEmail(email, username, token string) error {
	return m.sendEmail(
		email,
		"Theca | Reset Password",
		"templates/resetEmail.html",
		Mail{Username: username, Code: token},
	)
}
