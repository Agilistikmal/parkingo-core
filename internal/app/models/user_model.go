package models

import "time"

type User struct {
	ID        int        `json:"id"`
	Username  string     `json:"username"`
	FullName  string     `json:"full_name"`
	Email     string     `json:"email"`
	GoogleID  string     `json:"google_id"`
	Role      string     `json:"role"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

type CreateUserRequest struct {
	Username string `json:"username"`
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	GoogleID string `json:"google_id"`
}

type UpdateUserRequest struct {
	Username string `json:"username"`
	FullName string `json:"full_name"`
	Email    string `json:"email"`
}

func (u *User) IsAdmin() bool {
	return u.Role == "ADMIN"
}
