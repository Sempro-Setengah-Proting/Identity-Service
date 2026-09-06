package usecases

type RefreshTokenGenerator interface {
	GenerateRefreshToken() (string, error)
	GenerateHashToken(token string) string
	
}