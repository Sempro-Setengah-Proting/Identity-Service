package request

type ForgotPasswordRequest struct {
	NewPassword       string `json:"new_password" validate:"required,min=8,max=128"`
	ConfirmPassword   string `json:"confirm_password" validate:"required,min=8,max=128"`
	RegistrationToken string `json:"registration_token" validate:"required"`
}
