package providers

import "context"

type EmailSender interface {
	SendRegisterOTP(ctx context.Context, email string, otp string) error
}
