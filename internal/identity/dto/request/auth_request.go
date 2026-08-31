package request

type OAuthSignInRequest struct {
	IDToken string `json:"id_token" validate:"required"`
}