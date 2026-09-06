package request

type RequestRefreshToken struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
	DeviceID     string `json:"device_id" binding:"required"`
}