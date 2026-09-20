package request

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=128,password"`
	ConfirmPassword string `json:"confirm_password" validate:"required,min=8,max=128,password"`
}
