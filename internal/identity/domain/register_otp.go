package domain

type OTPPurpose string

const (
	OTPRegister       OTPPurpose = "REGISTER"
	OTPForgotPassword OTPPurpose = "FORGOT_PASSWORD"
)

type RegisterOTPState struct {
	OTPHash      string
	AttemptCount int
}
