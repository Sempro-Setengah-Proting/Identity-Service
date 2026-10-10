package request

type RequestForgotPasswordOTPRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type VerifyForgotPasswordOTPRequest struct {
	Email string `json:"email" validate:"required,email"`
	OTP   string `json:"otp" validate:"required,len=6,numeric"`
}
