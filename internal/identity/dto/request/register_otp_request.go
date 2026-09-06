package request

type RequestRegisterOTPRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type VerifyRegisterOTPRequest struct {
	Email string `json:"email" validate:"required,email"`
	OTP   string `json:"otp" validate:"required,len=6,numeric"`
}
