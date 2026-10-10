package usecases

import (
	"context"
	"errors"
	"fmt"
	"identityservice/internal/identity/domain"
	"identityservice/internal/identity/dto/request"
	"identityservice/internal/identity/providers"
	"identityservice/internal/identity/repositories"

	"golang.org/x/crypto/bcrypt"
)

type RequestForgotPasswordOTPUseCase interface {
	Request(ctx context.Context, payload request.RequestForgotPasswordOTPRequest) error
}

type VerifyForgotPasswordOTPUseCase interface {
	Verify(ctx context.Context, payload request.VerifyForgotPasswordOTPRequest) (string, int64, error)
}

type requestForgotPasswordOTPUseCase struct {
	userRepo    repositories.UserRepository
	otpStore    repositories.RegisterOTPStore
	emailSender providers.ForgotPasswordEmailSender
}

func NewRequestForgotPasswordOTPUseCase(userRepo repositories.UserRepository, otpStore repositories.RegisterOTPStore, emailSender providers.ForgotPasswordEmailSender) RequestForgotPasswordOTPUseCase {
	return &requestForgotPasswordOTPUseCase{userRepo: userRepo, otpStore: otpStore, emailSender: emailSender}
}

func (u *requestForgotPasswordOTPUseCase) Request(ctx context.Context, payload request.RequestForgotPasswordOTPRequest) error {
	email := normalizeRegisterEmail(payload.Email)
	user, err := u.userRepo.FindEmail(ctx, email)
	if errors.Is(err, repositories.ErrUserNotFound) {
		return ErrUserNotFound
	}
	if err != nil {
		return fmt.Errorf("%w: find password-recovery account: %w", ErrInternalError, err)
	}
	if user == nil {
		return ErrUserNotFound
	}
	if user.Provider != "LOCAL" {
		return ErrDifferentAuthProvider
	}
	otp, err := generateNumericOTP()
	if err != nil {
		return fmt.Errorf("%w: generate password-recovery OTP: %w", ErrInternalError, err)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(otp), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("%w: hash password-recovery OTP: %w", ErrInternalError, err)
	}
	if err := u.otpStore.SaveRegisterOTP(ctx, email, domain.OTPForgotPassword, string(hash), registerOTPTTL, registerOTPResendCooldown); err != nil {
		if errors.Is(err, repositories.ErrRegisterOTPCooldown) {
			return ErrOTPResendCooldown
		}
		return fmt.Errorf("%w: save password-recovery OTP: %w", ErrInternalError, err)
	}
	if err := u.emailSender.SendForgotPasswordOTP(ctx, email, otp); err != nil {
		cleanupErr := u.otpStore.DeleteRegisterOTP(ctx, email, domain.OTPForgotPassword, string(hash))
		return fmt.Errorf("%w: send password-recovery OTP: %w", ErrSmtpError, errors.Join(err, cleanupErr))
	}
	return nil
}

type verifyForgotPasswordOTPUseCase struct {
	otpStore repositories.RegisterOTPStore
}

func NewVerifyForgotPasswordOTPUseCase(otpStore repositories.RegisterOTPStore) VerifyForgotPasswordOTPUseCase {
	return &verifyForgotPasswordOTPUseCase{otpStore: otpStore}
}

func (u *verifyForgotPasswordOTPUseCase) Verify(ctx context.Context, payload request.VerifyForgotPasswordOTPRequest) (string, int64, error) {
	return verifyOTP(ctx, u.otpStore, payload.Email, payload.OTP, domain.OTPForgotPassword)
}
