package response

import "github.com/google/uuid"

type UserResponse struct {
	ID          uuid.UUID `json:"id"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	PhoneNumber string    `json:"phone_number"`
	Provider    string    `json:"provider"`
	ProviderID  string    `json:"provider_id"`
	AvatarURL   string    `json:"avatar_url"`
}
