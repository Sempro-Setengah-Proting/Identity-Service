package request

type OAuthSignInRequest struct {
	IDToken  string `json:"id_token" validate:"required"`
	DeviceID string `json:"device_id" validate:"required"`
}

type SignInRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
	DeviceID string `json:"device_id" validate:"required"`
}

type SignUpRequest struct {
	Email             string `json:"email" validate:"required,email"`
	Password          string `json:"password" validate:"required,min=6"`
	PhoneNumber       string `json:"phone_number" validate:"required"`
	DeviceID          string `json:"device_id" validate:"required"`
	Name              string `json:"name" validate:"required"`
	RegistrationToken string `json:"registration_token" validate:"required"`
}
