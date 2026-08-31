package domain

//id-token berisi informasi dari goggle yang sudah di validasi
type GoogleIdentity struct {
	ProvideId string
	Email string
	EmailVerified bool
	DisplayName string
	AvatarUrl string
}

