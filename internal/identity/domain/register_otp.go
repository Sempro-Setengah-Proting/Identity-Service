package domain

type RegisterOTPState struct {
	OTPHash      string
	AttemptCount int
}
