package repositories

import (
	"context"
	"errors"
	"identityservice/internal/identity/domain"
	"time"
)

var (
	ErrRegisterOTPNotFound     = errors.New("register OTP not found")
	ErrRegisterOTPCooldown     = errors.New("register OTP resend cooldown")
	ErrRegisterOTPStateChanged = errors.New("register OTP state changed")
	ErrRegisterOTPMaxAttempts  = errors.New("register OTP max attempts")
	ErrRegistrationProofExists = errors.New("registration proof already exists")
	ErrRegistrationTokenNotFound = errors.New("registration token not found")
)

type RegisterOTPStore interface {
	SaveRegisterOTP(
		ctx context.Context,
		email string,
		otpHash string,
		otpTTL time.Duration,
		cooldownTTL time.Duration,
	) error
	GetRegisterOTP(ctx context.Context, email string) (*domain.RegisterOTPState, error)
	IncrementRegisterOTPAttempt(
		ctx context.Context,
		email string,
		expectedOTPHash string,
		maxAttempts int,
	) (int, error)
	ConsumeRegisterOTP(
		ctx context.Context,
		email string,
		expectedOTPHash string,
		maxAttempts int,
		registrationTokenHash string,
		registrationTokenTTL time.Duration,
	) error
	DeleteRegisterOTP(ctx context.Context, email string, expectedOTPHash string) error
	FindRegistrationProof(
		ctx context.Context,
		tokenHash string,
	) (string, error)
}
