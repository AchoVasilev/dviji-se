package email

import (
	"bytes"
	"fmt"
	"html/template"
	"log/slog"
	"net/smtp"
	"server/internal/config"
)

type EmailService struct {
	host     string
	port     string
	username string
	password string
	from     string
	baseURL  string
}

func NewEmailService() *EmailService {
	return &EmailService{
		host:     config.SMTPHost(),
		port:     config.SMTPPort(),
		username: config.SMTPUsername(),
		password: config.SMTPPassword(),
		from:     config.SMTPFrom(),
		baseURL:  config.BaseURL(),
	}
}

func (s *EmailService) SendPasswordResetEmail(toEmail, token string) error {
	resetLink := fmt.Sprintf("%s/reset-password?token=%s", s.baseURL, token)

	subject := "Заявка за смяна на парола - Движи се"
	body, err := s.renderPasswordResetTemplate(resetLink)
	if err != nil {
		return err
	}

	return s.sendEmail(toEmail, subject, body)
}

func (s *EmailService) renderPasswordResetTemplate(resetLink string) (string, error) {
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h2 style="color: #2563eb;">Смяна на парола</h2>
        <p>Получихме заявка за смяна на паролата на вашия акаунт.</p>
        <p>Кликнете на бутона по-долу, за да зададете нова парола:</p>
        <p style="margin: 30px 0;">
            <a href="{{.ResetLink}}"
               style="background-color: #2563eb; color: white; padding: 12px 24px; text-decoration: none; border-radius: 6px; display: inline-block;">
                Смяна на паролата
            </a>
        </p>
        <p style="color: #666; font-size: 14px;">
            Този линк е валиден 1 час. Ако не сте заявили смяна на паролата, игнорирайте този имейл.
        </p>
        <hr style="border: none; border-top: 1px solid #eee; margin: 30px 0;">
        <p style="color: #999; font-size: 12px;">
            Движи се - Фитнес блог
        </p>
    </div>
</body>
</html>`

	t, err := template.New("passwordReset").Parse(tmpl)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = t.Execute(&buf, map[string]string{"ResetLink": resetLink})
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (s *EmailService) sendEmail(to, subject, htmlBody string) error {
	// If SMTP is not configured, log the email instead (for development)
	if s.host == "" || s.from == "" {
		slog.Info("Email would be sent (SMTP not configured)",
			"to", to,
			"subject", subject,
		)
		slog.Debug("Email body", "body", htmlBody)
		return nil
	}

	headers := make(map[string]string)
	headers["From"] = s.from
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"

	var msg bytes.Buffer
	for k, v := range headers {
		msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	msg.WriteString("\r\n")
	msg.WriteString(htmlBody)

	auth := smtp.PlainAuth("", s.username, s.password, s.host)
	addr := fmt.Sprintf("%s:%s", s.host, s.port)

	err := smtp.SendMail(addr, auth, s.from, []string{to}, msg.Bytes())
	if err != nil {
		slog.Error("Failed to send email", "error", err, "to", to)
		return err
	}

	slog.Info("Email sent successfully", "to", to, "subject", subject)
	return nil
}
