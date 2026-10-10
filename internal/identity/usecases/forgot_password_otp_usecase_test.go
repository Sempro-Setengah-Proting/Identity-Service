package usecases

import (
	"context"
	"errors"
	"identityservice/internal/identity/domain"
	"identityservice/internal/identity/dto/request"
	"identityservice/internal/identity/repositories"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type recoveryUserRepo struct {
	repositories.UserRepository
	provider     string
	lookupErr    error
	passwordHash string
	lookupEmail  string
}

func (r *recoveryUserRepo) FindEmail(_ context.Context, email string) (*domain.User, error) {
	r.lookupEmail = email
	return &domain.User{Email: email, Provider: r.provider}, r.lookupErr
}

func (r *recoveryUserRepo) ForgetPassword(_ context.Context, email, hash string) error {
	r.passwordHash = hash
	return nil
}

type recoveryMailer struct {
	otp, email string
	err        error
}

func (m *recoveryMailer) SendForgotPasswordOTP(_ context.Context, email, otp string) error {
	m.email, m.otp = email, otp
	return m.err
}

type recoveryProof struct {
	email   string
	purpose domain.OTPPurpose
}

type recoveryOTPStore struct {
	repositories.RegisterOTPStore
	states                        map[domain.OTPPurpose]*domain.RegisterOTPState
	proofs                        map[string]recoveryProof
	otpTTL, cooldownTTL, proofTTL time.Duration
}

func newRecoveryOTPStore() *recoveryOTPStore {
	return &recoveryOTPStore{states: make(map[domain.OTPPurpose]*domain.RegisterOTPState), proofs: make(map[string]recoveryProof)}
}

func (s *recoveryOTPStore) SaveRegisterOTP(_ context.Context, _ string, purpose domain.OTPPurpose, hash string, ttl, cooldown time.Duration) error {
	if s.states[purpose] != nil {
		return repositories.ErrRegisterOTPCooldown
	}
	s.states[purpose] = &domain.RegisterOTPState{OTPHash: hash}
	s.otpTTL, s.cooldownTTL = ttl, cooldown
	return nil
}

func (s *recoveryOTPStore) GetRegisterOTP(_ context.Context, _ string, purpose domain.OTPPurpose) (*domain.RegisterOTPState, error) {
	state := s.states[purpose]
	if state == nil {
		return nil, repositories.ErrRegisterOTPNotFound
	}
	return state, nil
}

func (s *recoveryOTPStore) IncrementRegisterOTPAttempt(_ context.Context, _ string, purpose domain.OTPPurpose, _ string, _ int) (int, error) {
	s.states[purpose].AttemptCount++
	return s.states[purpose].AttemptCount, nil
}

func (s *recoveryOTPStore) ConsumeRegisterOTP(_ context.Context, email string, purpose domain.OTPPurpose, _ string, _ int, tokenHash string, ttl time.Duration) error {
	s.proofs[tokenHash] = recoveryProof{email, purpose}
	s.proofTTL = ttl
	delete(s.states, purpose)
	return nil
}

func (s *recoveryOTPStore) DeleteRegisterOTP(_ context.Context, _ string, purpose domain.OTPPurpose, _ string) error {
	delete(s.states, purpose)
	return nil
}

func (s *recoveryOTPStore) ConsumeRegistrationProof(_ context.Context, hash string, purpose domain.OTPPurpose) (string, error) {
	proof, ok := s.proofs[hash]
	if !ok || proof.purpose != purpose {
		return "", repositories.ErrRegistrationTokenNotFound
	}
	delete(s.proofs, hash)
	return proof.email, nil
}

func TestForgotPasswordOTPFlowAndPurposeIsolation(t *testing.T) {
	ctx := context.Background()
	store := newRecoveryOTPStore()
	users := &recoveryUserRepo{provider: "LOCAL"}
	mailer := &recoveryMailer{}
	requester := NewRequestForgotPasswordOTPUseCase(users, store, mailer)
	payload := request.RequestForgotPasswordOTPRequest{Email: " User@Example.com "}
	if err := requester.Request(ctx, payload); err != nil {
		t.Fatal(err)
	}
	if mailer.email != "user@example.com" || users.lookupEmail != mailer.email || len(mailer.otp) != 6 {
		t.Fatal("OTP email was not normalized or OTP length is invalid")
	}
	state := store.states[domain.OTPForgotPassword]
	if state == nil || bcrypt.CompareHashAndPassword([]byte(state.OTPHash), []byte(mailer.otp)) != nil {
		t.Fatal("password-recovery OTP was not hashed under its own purpose")
	}
	if store.otpTTL != 5*time.Minute || store.cooldownTTL != time.Minute {
		t.Fatal("incorrect OTP expiration or cooldown")
	}
	if err := requester.Request(ctx, payload); !errors.Is(err, ErrOTPResendCooldown) {
		t.Fatalf("expected cooldown, got %v", err)
	}
	if _, _, err := NewVerifyRegisterOTPUseCase(store).Verify(ctx, request.VerifyRegisterOTPRequest{Email: mailer.email, OTP: mailer.otp}); !errors.Is(err, ErrOTPNotFound) {
		t.Fatalf("registration must not read password-recovery OTP: %v", err)
	}
	// Leave an independent registration OTP in place during recovery verification.
	store.states[domain.OTPRegister] = &domain.RegisterOTPState{OTPHash: "registration-hash"}
	verifier := NewVerifyForgotPasswordOTPUseCase(store)
	verifyPayload := request.VerifyForgotPasswordOTPRequest{Email: mailer.email, OTP: mailer.otp}
	token, ttl, err := verifier.Verify(ctx, verifyPayload)
	if err != nil || token == "" || ttl != 600 || store.proofTTL != 10*time.Minute {
		t.Fatalf("unexpected verification result: token empty=%v ttl=%d error=%v", token == "", ttl, err)
	}
	if store.states[domain.OTPRegister] == nil || store.states[domain.OTPForgotPassword] != nil {
		t.Fatal("wrong OTP purpose was consumed")
	}
	if _, err := store.ConsumeRegistrationProof(ctx, HashRegistrationToken(token), domain.OTPRegister); !errors.Is(err, repositories.ErrRegistrationTokenNotFound) {
		t.Fatal("recovery proof must not work for registration")
	}
	if _, _, err := verifier.Verify(ctx, verifyPayload); !errors.Is(err, ErrOTPNotFound) {
		t.Fatal("OTP must be single-use")
	}
	localAuth := NewLocalAuthUseCase(nil, users, nil, nil, nil, store)
	reset := request.ForgotPasswordRequest{RegistrationToken: token, NewPassword: "new-password", ConfirmPassword: "new-password"}
	if err := localAuth.ForgotPassword(ctx, reset); err != nil {
		t.Fatal(err)
	}
	if bcrypt.CompareHashAndPassword([]byte(users.passwordHash), []byte(reset.NewPassword)) != nil {
		t.Fatal("reset did not store a new password hash")
	}
	if err := localAuth.ForgotPassword(ctx, reset); !errors.Is(err, ErrInvalidRegistrationToken) {
		t.Fatal("recovery proof must be single-use")
	}
	registrationToken := "registration-only-token"
	store.proofs[HashRegistrationToken(registrationToken)] = recoveryProof{mailer.email, domain.OTPRegister}
	reset.RegistrationToken = registrationToken
	if err := localAuth.ForgotPassword(ctx, reset); !errors.Is(err, ErrInvalidRegistrationToken) {
		t.Fatal("registration proof must not reset passwords")
	}
}

func TestForgotPasswordOTPAccountAndEmailFailures(t *testing.T) {
	for _, tt := range []struct {
		name, provider           string
		lookupErr, mailErr, want error
	}{
		{"unknown account", "LOCAL", repositories.ErrUserNotFound, nil, ErrUserNotFound},
		{"google account", "GOOGLE", nil, nil, ErrDifferentAuthProvider},
		{"database failure", "LOCAL", errors.New("database unavailable"), nil, ErrInternalError},
		{"email failure", "LOCAL", nil, errors.New("SMTP unavailable"), ErrSmtpError},
	} {
		t.Run(tt.name, func(t *testing.T) {
			store := newRecoveryOTPStore()
			mailer := &recoveryMailer{err: tt.mailErr}
			users := &recoveryUserRepo{provider: tt.provider, lookupErr: tt.lookupErr}
			err := NewRequestForgotPasswordOTPUseCase(users, store, mailer).Request(context.Background(), request.RequestForgotPasswordOTPRequest{Email: "user@example.com"})
			if !errors.Is(err, tt.want) {
				t.Fatalf("expected %v, got %v", tt.want, err)
			}
			if len(store.states) != 0 {
				t.Fatal("failed recovery request left a usable OTP")
			}
			if tt.mailErr == nil && mailer.otp != "" {
				t.Fatal("invalid account received an OTP")
			}
		})
	}
}

func TestForgotPasswordOTPAttemptLimit(t *testing.T) {
	store := newRecoveryOTPStore()
	hash, err := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	store.states[domain.OTPForgotPassword] = &domain.RegisterOTPState{OTPHash: string(hash)}
	verifier := NewVerifyForgotPasswordOTPUseCase(store)
	for i := 1; i <= registerOTPMaxAttempts; i++ {
		_, _, err := verifier.Verify(context.Background(), request.VerifyForgotPasswordOTPRequest{Email: "user@example.com", OTP: "999999"})
		want := ErrOTPInvalid
		if i == registerOTPMaxAttempts {
			want = ErrOTPMaxAttempts
		}
		if !errors.Is(err, want) {
			t.Fatalf("attempt %d: expected %v, got %v", i, want, err)
		}
	}
	if _, _, err := verifier.Verify(context.Background(), request.VerifyForgotPasswordOTPRequest{Email: "user@example.com", OTP: "123456"}); !errors.Is(err, ErrOTPMaxAttempts) {
		t.Fatal("correct OTP was accepted after attempt limit")
	}
	if len(store.proofs) != 0 {
		t.Fatal("failed verification created a proof")
	}
}
