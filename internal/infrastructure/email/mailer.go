package email

import (
	"context"
	"errors"
	"fmt"
	"identityservice/internal/infrastructure/config"
	"log"
	"strings"

	gomail "gopkg.in/gomail.v2"
)

type Mailer struct {
	cfg *config.Config
}

func NewMailer(cfg *config.Config) *Mailer {
	return &Mailer{cfg: cfg}
}

func (m *Mailer) SendEmail(ctx context.Context, to string, subject string, body string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if m == nil || m.cfg == nil {
		return errors.New("mailer is not configured")
	}
	if m.cfg.SMTPHost == "" {
		return errors.New("SMTP host is empty")
	}
	if m.cfg.SMTPPort == 0 {
		return errors.New("SMTP port is empty")
	}
	if m.cfg.SMTPEmail == "" {
		return errors.New("SMTP email is empty")
	}
	if m.cfg.SMTPPassword == "" {
		return errors.New("SMTP password is empty")
	}

	log.Printf(
		"SMTP delivery: host=%s port=%d sender=%s",
		m.cfg.SMTPHost,
		m.cfg.SMTPPort,
		m.cfg.SMTPEmail,
	)

	message := gomail.NewMessage()
	message.SetHeader("From", m.cfg.SMTPEmail)
	message.SetHeader("To", to)
	message.SetHeader("Subject", subject)
	message.SetBody("text/html", body)

	dialer := gomail.NewDialer(
		m.cfg.SMTPHost,
		m.cfg.SMTPPort,
		m.cfg.SMTPEmail,
		m.cfg.SMTPPassword,
	)
	if err := dialer.DialAndSend(message); err != nil {
		log.Printf("SMTP delivery failed: %+v", err)

		errMessage := strings.ToLower(err.Error())
		switch {
		case strings.Contains(errMessage, "535"),
			strings.Contains(errMessage, "5.7.8"),
			strings.Contains(errMessage, "authentication"):
			return errors.New("failed to authenticate SMTP account")
		case strings.Contains(errMessage, "connection refused"),
			strings.Contains(errMessage, "dial tcp"):
			return errors.New("failed to connect SMTP server")
		case strings.Contains(errMessage, "timeout"),
			strings.Contains(errMessage, "i/o timeout"):
			return errors.New("SMTP connection timeout")
		case strings.Contains(errMessage, "tls"),
			strings.Contains(errMessage, "certificate"):
			return errors.New("SMTP TLS or authentication issue")
		default:
			return fmt.Errorf("SMTP delivery failed: %s", errMessage)
		}
	}

	return nil
}

func (m *Mailer) SendRegisterOTP(ctx context.Context, to string, otp string) error {
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="font-family: Arial, sans-serif; background-color: #f4f4f5; margin: 0; padding: 20px;">
    <div style="max-width: 600px; margin: 0 auto; background-color: #ffffff; padding: 40px; border-radius: 8px;">
        <h2 style="color: #333333; text-align: center;">Verify Your Email</h2>
        <p style="color: #555555; font-size: 16px; line-height: 1.5;">
            Thank you for registering. Use the verification code below to continue.
        </p>
        <div style="background-color: #f9fafb; padding: 20px; border-radius: 8px; margin: 30px 0; text-align: center; border: 1px solid #e5e7eb;">
            <p style="color: #6b7280; font-size: 14px; text-transform: uppercase; font-weight: 600;">Your Verification Code</p>
            <div style="font-size: 32px; font-weight: bold; letter-spacing: 10px; color: #111827;">%s</div>
        </div>
        <p style="color: #555555; font-size: 14px; line-height: 1.5;">
            This code expires in <strong>5 minutes</strong>. Do not share this code with anyone.
        </p>
        <p style="color: #9ca3af; font-size: 12px; text-align: center;">
            If you did not request this email, you can safely ignore it.
        </p>
    </div>
</body>
</html>
`, otp)

	return m.SendEmail(ctx, to, "Your Registration Verification Code", body)
}
