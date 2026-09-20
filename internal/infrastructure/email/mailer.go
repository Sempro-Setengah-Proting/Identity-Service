package email

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"identityservice/internal/infrastructure/config"
	"io"
	"log"
	"strings"

	gomail "gopkg.in/gomail.v2"
)

type Mailer struct {
	cfg *config.Config
}

//go:embed travpal-logo.png
var travPalLogo []byte

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
	if strings.Contains(body, "cid:travpal-logo.png") {
		message.Embed("travpal-logo.png", gomail.SetCopyFunc(func(w io.Writer) error {
			_, err := w.Write(travPalLogo)
			return err
		}))
	}

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
<body style="margin: 0; padding: 0; background-color: #f6f6f6; font-family: Arial, Helvetica, sans-serif;">
    <table role="presentation" width="100%%" cellspacing="0" cellpadding="0" border="0" style="width: 100%%; margin: 0; padding: 0; background-color: #f6f6f6;">
        <tr>
            <td align="center" style="padding: 38px 20px 42px;">
                <table role="presentation" width="334" cellspacing="0" cellpadding="0" border="0" style="width: 100%%; max-width: 334px;">
                    <tr>
                        <td align="center" style="padding: 0 0 18px;">
                            <img src="cid:travpal-logo.png" width="88" height="50" alt="TravPal" style="display: block; width: 88px; height: 50px; margin: 0 auto; border: 0; outline: none; text-decoration: none;">
                            <div style="margin-top: 9px; color: #66767a; font-size: 19px; font-weight: 700; line-height: 23px; letter-spacing: -0.6px;">TravPal</div>
                        </td>
                    </tr>
                    <tr>
                        <td style="padding: 24px 20px 22px; background-color: #ffffff; border-radius: 12px; box-shadow: 0 8px 20px rgba(0, 0, 0, 0.05);">
                            <h2 style="margin: 0 0 14px; color: #111111; font-size: 16px; font-weight: 400; line-height: 20px;">Verify Your Email</h2>
                            <p style="margin: 0; color: #111111; font-size: 13px; font-weight: 400; line-height: 15px;">
                                Thank you for registering. Use the verification code below to continue.
                            </p>
                            <div style="margin: 20px auto; text-align: center;">
                                <p style="display: none; margin: 0; color: #111111; font-size: 11px; font-weight: 600; line-height: 13px; text-transform: uppercase; letter-spacing: 0.25px;">Your Verification Code</p>
                                <div style="display: inline-block; padding: 11px 18px; background-color: #d9dcdd; color: #101d1f; font-size: 23px; font-weight: 700; line-height: 24px; letter-spacing: 6px; white-space: nowrap;">%s</div>
                            </div>
                            <p style="margin: 0 0 14px; color: #111111; font-size: 13px; font-weight: 400; line-height: 15px;">
                                This code expires in <strong style="font-weight: 700;">5 minutes</strong>. Do not share this code with anyone.
                            </p>
                            <p style="margin: 0; color: #111111; font-size: 13px; font-weight: 400; line-height: 15px;">
                                If you did not request this email, you can safely ignore it.
                            </p>
                        </td>
                    </tr>
                    <tr>
                        <td align="center" style="padding: 12px 0 0; color: #a8afb1; font-size: 11px; font-weight: 400; line-height: 14px;">
                            ©&nbsp; 2026 TravPal. All rights reserved.
                        </td>
                    </tr>
                </table>
            </td>
        </tr>
    </table>
</body>
</html>
`, otp)

	return m.SendEmail(ctx, to, "Your Registration Verification Code", body)
}
