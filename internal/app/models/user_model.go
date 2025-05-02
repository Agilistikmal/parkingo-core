package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        int            `json:"id"`
	Username  string         `json:"username"`
	FullName  string         `json:"full_name"`
	Email     string         `json:"email"`
	AvatarURL string         `json:"avatar_url"`
	GoogleID  string         `json:"google_id"`
	Role      string         `json:"role"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at"`
}

type CreateUserRequest struct {
	Username  string `json:"username" validate:"required,min=3,max=50"`
	FullName  string `json:"full_name" validate:"required,min=3,max=100"`
	Email     string `json:"email" validate:"required,email"`
	AvatarURL string `json:"avatar_url" validate:"omitempty,url"`
	GoogleID  string `json:"google_id" validate:"required"`
}

type UpdateUserRequest struct {
	Username  string `json:"username" validate:"omitempty,min=3,max=50"`
	AvatarURL string `json:"avatar_url" validate:"omitempty,url"`
	FullName  string `json:"full_name" validate:"omitempty,min=3,max=100"`
}

func (u *User) IsAdmin() bool {
	return u.Role == "ADMIN"
}
