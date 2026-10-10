package request

type EditProfileRequest struct {
	Name        *string `json:"name" validate:"omitempty,min=3,max=128"`
	PhoneNumber *string `json:"phone_number" validate:"omitempty,min=3,max=15"`
}