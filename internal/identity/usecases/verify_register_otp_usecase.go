package usecases

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"identityservice/internal/identity/domain"
	"identityservice/internal/identity/dto/request"
	"identityservice/internal/identity/repositories"

	"golang.org/x/crypto/bcrypt"
)

const registrationTokenGenerationAttempts = 3

type verifyRegisterOTPUseCase struct {
	otpStore repositories.RegisterOTPStore
}

type VerifyRegisterOTPUseCase interface {
	Verify(ctx context.Context, payload request.VerifyRegisterOTPRequest) (string, int64, error)
}

func NewVerifyRegisterOTPUseCase(otpStore repositories.RegisterOTPStore) VerifyRegisterOTPUseCase {
	return &verifyRegisterOTPUseCase{otpStore: otpStore}
}

func (u *verifyRegisterOTPUseCase) Verify(
	ctx context.Context,
	payload request.VerifyRegisterOTPRequest,
) (string, int64, error) {
	email := normalizeRegisterEmail(payload.Email)
	state, err := u.otpStore.GetRegisterOTP(ctx, email, domain.OTPRegister)
	if errors.Is(err, repositories.ErrRegisterOTPNotFound) {
		return "", 0, ErrOTPNotFound
	}
	if err != nil {
		return "", 0, ErrInternalError
	}
	if state.AttemptCount >= registerOTPMaxAttempts {
		return "", 0, ErrOTPMaxAttempts
	}

	if err = bcrypt.CompareHashAndPassword([]byte(state.OTPHash), []byte(payload.OTP)); err != nil {
		if !errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return "", 0, ErrInternalError
		}

		attemptCount, incrementErr := u.otpStore.IncrementRegisterOTPAttempt(
			ctx,
			email,
			domain.OTPRegister,
			state.OTPHash,
			registerOTPMaxAttempts,
		)
		if errors.Is(incrementErr, repositories.ErrRegisterOTPMaxAttempts) ||
			attemptCount >= registerOTPMaxAttempts {
			return "", 0, ErrOTPMaxAttempts
		}
		if errors.Is(incrementErr, repositories.ErrRegisterOTPNotFound) ||
			errors.Is(incrementErr, repositories.ErrRegisterOTPStateChanged) {
			return "", 0, ErrOTPNotFound
		}
		if incrementErr != nil {
			return "", 0, ErrInternalError
		}

		return "", 0, ErrOTPInvalid
	}

	for attempt := 0; attempt < registrationTokenGenerationAttempts; attempt++ {
		registrationToken, generateErr := generateRegistrationToken()
		if generateErr != nil {
			return "", 0, ErrInternalError
		}
		registrationTokenHash := HashRegistrationToken(registrationToken)

		err = u.otpStore.ConsumeRegisterOTP(
			ctx,
			email,
			domain.OTPRegister,
			state.OTPHash,
			registerOTPMaxAttempts,
			registrationTokenHash,
			registrationTokenTTL,
		)
		if errors.Is(err, repositories.ErrRegistrationProofExists) {
			continue
		}
		if errors.Is(err, repositories.ErrRegisterOTPMaxAttempts) {
			return "", 0, ErrOTPMaxAttempts
		}
		if errors.Is(err, repositories.ErrRegisterOTPNotFound) ||
			errors.Is(err, repositories.ErrRegisterOTPStateChanged) {
			return "", 0, ErrOTPNotFound
		}
		if err != nil {
			return "", 0, ErrInternalError
		}

		return registrationToken, int64(registrationTokenTTL.Seconds()), nil
	}

	return "", 0, ErrInternalError
}

func generateRegistrationToken() (string, error) {
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(randomBytes), nil
}

func HashRegistrationToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
