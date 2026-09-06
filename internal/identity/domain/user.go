package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID          uuid.UUID `gorm:"column:id;type:uuid;default:gen_random_uuid();primaryKey"`
	Username    string    `json:"username" gorm:"column:username;type:varchar(255);not null"`
	Email       string    `json:"email" gorm:"column:email;type:varchar(255);unique;not null"`
	Password    string    `json:"password" gorm:"column:password;type:varchar(255)"`
	PhoneNumber string    `json:"phone_number" gorm:"column:phone_number;type:varchar(20)"`
	Provider    string    `json:"provider" gorm:"column:provider;type:varchar(20);not null"`
	ProviderId  string    `json:"provider_id" gorm:"column:provider_id;type:varchar"`
	AvatarUrl   string    `json:"avatar_url" gorm:"column:avatar_url;type:varchar(255)"`
	CreatedAt   time.Time `json:"created_at" gorm:"column:created_at;type:timestamp;not null"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"column:updated_at;type:timestamp;not null"`
}
