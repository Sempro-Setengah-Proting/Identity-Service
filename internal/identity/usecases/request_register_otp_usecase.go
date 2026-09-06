package usecases

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"identityservice/internal/identity/dto/request"
	"identityservice/internal/identity/providers"
	"identityservice/internal/identity/repositories"
	"log"
	"math/big"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const (
	registerOTPTTL            = 5 * time.Minute
	registerOTPResendCooldown = 60 * time.Second
	registerOTPMaxAttempts    = 5
	registrationTokenTTL      = 10 * time.Minute
)

var (
	ErrEmailAlreadyRegistered = errors.New("email already registered")
	ErrOTPNotFound            = errors.New("OTP not found")
	ErrOTPInvalid             = errors.New("invalid OTP")
	ErrOTPMaxAttempts         = errors.New("OTP max attempts")
	ErrOTPResendCooldown      = errors.New("OTP resend cooldown")
	ErrSmtpError                = errors.New("smtp error")
)

type requestRegisterOTPUseCase struct {
	userRepo    repositories.UserRepository
	otpStore    repositories.RegisterOTPStore
	emailSender providers.EmailSender
}

type RequestRegisterOTPUseCase interface {
	Request(ctx context.Context, payload request.RequestRegisterOTPRequest) error
}

func NewRequestRegisterOTPUseCase(
	userRepo repositories.UserRepository,
	otpStore repositories.RegisterOTPStore,
	emailSender providers.EmailSender,
) RequestRegisterOTPUseCase {
	return &requestRegisterOTPUseCase{
		userRepo:    userRepo,
		otpStore:    otpStore,
		emailSender: emailSender,
	}
}

func (u *requestRegisterOTPUseCase) Request(ctx context.Context, payload request.RequestRegisterOTPRequest) error {
	email := normalizeRegisterEmail(payload.Email)

	user, err := u.userRepo.FindEmail(ctx, email)
	if user != nil {
		return ErrEmailAlreadyRegistered
	}
	if err != nil && !errors.Is(err, repositories.ErrUserNotFound) {
		log.Printf("error is: %v", err)
		log.Printf("user: %v", &user)
		return ErrInternalError
	}

	otp, err := generateNumericOTP()
	if err != nil {
		log.Println("position line 71")
		log.Printf("error is: %v", err)
		return ErrInternalError
	}
	otpHash, err := bcrypt.GenerateFromPassword([]byte(otp), bcrypt.DefaultCost)
	if err != nil {
		log.Println("position line 77")
		return ErrInternalError
	}

	if err = u.otpStore.SaveRegisterOTP(
		ctx,
		email,
		string(otpHash),
		registerOTPTTL,
		registerOTPResendCooldown,
	); err != nil {
		if errors.Is(err, repositories.ErrRegisterOTPCooldown) {
			return ErrOTPResendCooldown
		}
		log.Println("position line 91")
		return ErrInternalError
	}

	if err = u.emailSender.SendRegisterOTP(ctx, email, otp); err != nil {
		log.Println(err)
		_ = u.otpStore.DeleteRegisterOTP(ctx, email, string(otpHash))
		return ErrSmtpError
	}

	return nil
}

func normalizeRegisterEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func generateNumericOTP() (string, error) {
	value, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%06d", value.Int64()), nil
}
