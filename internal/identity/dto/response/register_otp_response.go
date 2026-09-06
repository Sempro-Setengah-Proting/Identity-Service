package response

type VerifyRegisterOTPResponse struct {
	RegistrationToken string `json:"registration_token"`
	ExpiresIn         int64  `json:"expires_in"`
}
