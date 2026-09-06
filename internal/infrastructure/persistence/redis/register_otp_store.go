package redisstore

import (
	"context"
	"errors"
	"fmt"
	"identityservice/internal/identity/domain"
	"identityservice/internal/identity/repositories"
	"strconv"
	"time"

	redisClient "github.com/redis/go-redis/v9"
)

const (
	otpHashField      = "otp_hash"
	attemptCountField = "attempt_count"
)

var saveRegisterOTPScript = redisClient.NewScript(`
if redis.call("EXISTS", KEYS[2]) == 1 then
    return 0
end
redis.call("HSET", KEYS[1], "otp_hash", ARGV[1], "attempt_count", 0)
redis.call("PEXPIRE", KEYS[1], ARGV[2])
redis.call("SET", KEYS[2], "1", "PX", ARGV[3])
return 1
`)

var incrementRegisterOTPAttemptScript = redisClient.NewScript(`
if redis.call("EXISTS", KEYS[1]) == 0 then
    return -1
end
if redis.call("HGET", KEYS[1], "otp_hash") ~= ARGV[1] then
    return -2
end
local attempts = tonumber(redis.call("HGET", KEYS[1], "attempt_count") or "0")
if attempts >= tonumber(ARGV[2]) then
    return -3
end
return redis.call("HINCRBY", KEYS[1], "attempt_count", 1)
`)

var consumeRegisterOTPScript = redisClient.NewScript(`
if redis.call("EXISTS", KEYS[1]) == 0 then
    return -1
end
if redis.call("HGET", KEYS[1], "otp_hash") ~= ARGV[1] then
    return -2
end
local attempts = tonumber(redis.call("HGET", KEYS[1], "attempt_count") or "0")
if attempts >= tonumber(ARGV[2]) then
    return -3
end
local saved = redis.call("SET", KEYS[3], ARGV[3], "PX", ARGV[4], "NX")
if not saved then
    return -4
end
redis.call("DEL", KEYS[1], KEYS[2])
return 1
`)

var deleteRegisterOTPScript = redisClient.NewScript(`
if redis.call("EXISTS", KEYS[1]) == 0 then
    return 0
end
if redis.call("HGET", KEYS[1], "otp_hash") ~= ARGV[1] then
    return -1
end
redis.call("DEL", KEYS[1], KEYS[2])
return 1
`)

type registerOTPStore struct {
	client redisClient.UniversalClient
}

func NewRegisterOTPStore(client redisClient.UniversalClient) repositories.RegisterOTPStore {
	return &registerOTPStore{client: client}
}

func (s *registerOTPStore) SaveRegisterOTP(
	ctx context.Context,
	email string,
	otpHash string,
	otpTTL time.Duration,
	cooldownTTL time.Duration,
) error {
	result, err := saveRegisterOTPScript.Run(
		ctx,
		s.client,
		[]string{registerOTPKey(email), registerOTPCooldownKey(email)},
		otpHash,
		otpTTL.Milliseconds(),
		cooldownTTL.Milliseconds(),
	).Int64()
	if err != nil {
		return err
	}
	if result == 0 {
		return repositories.ErrRegisterOTPCooldown
	}

	return nil
}

func (s *registerOTPStore) GetRegisterOTP(ctx context.Context, email string) (*domain.RegisterOTPState, error) {
	values, err := s.client.HGetAll(ctx, registerOTPKey(email)).Result()
	if err != nil {
		return nil, err
	}
	if len(values) == 0 {
		return nil, repositories.ErrRegisterOTPNotFound
	}

	otpHash, ok := values[otpHashField]
	if !ok || otpHash == "" {
		return nil, errors.New("register OTP hash is missing")
	}
	attemptCount, err := strconv.Atoi(values[attemptCountField])
	if err != nil {
		return nil, fmt.Errorf("parse register OTP attempt count: %w", err)
	}

	return &domain.RegisterOTPState{
		OTPHash:      otpHash,
		AttemptCount: attemptCount,
	}, nil
}

func (s *registerOTPStore) IncrementRegisterOTPAttempt(
	ctx context.Context,
	email string,
	expectedOTPHash string,
	maxAttempts int,
) (int, error) {
	result, err := incrementRegisterOTPAttemptScript.Run(
		ctx,
		s.client,
		[]string{registerOTPKey(email)},
		expectedOTPHash,
		maxAttempts,
	).Int64()
	if err != nil {
		return 0, err
	}

	switch result {
	case -1:
		return 0, repositories.ErrRegisterOTPNotFound
	case -2:
		return 0, repositories.ErrRegisterOTPStateChanged
	case -3:
		return maxAttempts, repositories.ErrRegisterOTPMaxAttempts
	default:
		return int(result), nil
	}
}

func (s *registerOTPStore) ConsumeRegisterOTP(
	ctx context.Context,
	email string,
	expectedOTPHash string,
	maxAttempts int,
	registrationTokenHash string,
	registrationTokenTTL time.Duration,
) error {
	result, err := consumeRegisterOTPScript.Run(
		ctx,
		s.client,
		[]string{
			registerOTPKey(email),
			registerOTPCooldownKey(email),
			registrationProofKey(registrationTokenHash),
		},
		expectedOTPHash,
		maxAttempts,
		email,
		registrationTokenTTL.Milliseconds(),
	).Int64()
	if err != nil {
		return err
	}

	switch result {
	case -1:
		return repositories.ErrRegisterOTPNotFound
	case -2:
		return repositories.ErrRegisterOTPStateChanged
	case -3:
		return repositories.ErrRegisterOTPMaxAttempts
	case -4:
		return repositories.ErrRegistrationProofExists
	default:
		return nil
	}
}

func (s *registerOTPStore) DeleteRegisterOTP(ctx context.Context, email string, expectedOTPHash string) error {
	result, err := deleteRegisterOTPScript.Run(
		ctx,
		s.client,
		[]string{registerOTPKey(email), registerOTPCooldownKey(email)},
		expectedOTPHash,
	).Int64()
	if err != nil {
		return err
	}
	if result == -1 {
		return repositories.ErrRegisterOTPStateChanged
	}

	return nil
}
func (r *registerOTPStore) FindRegistrationProof(
	ctx context.Context,
	tokenHash string,
) (string, error) {

	key := fmt.Sprintf(
		"register:verified:%s",
		tokenHash,
	)

	email, err := r.client.Get(ctx, key).Result()
	if err != nil {

		if errors.Is(err, redisClient.Nil) {
			return "", repositories.ErrRegistrationTokenNotFound
		}

		return "", err
	}

	return email, nil
}

func registerOTPKey(email string) string {
	return "otp:register:{" + email + "}"
}

func registerOTPCooldownKey(email string) string {
	return "otp:register:cooldown:{" + email + "}"
}

func registrationProofKey(registrationTokenHash string) string {
	return "register:verified:" + registrationTokenHash
}
