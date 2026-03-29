package models

import "time"

type User struct {
	ID         string     `json:"id"`
	AuthUserID string     `json:"auth_user_id"`
	Email      string     `json:"email"`
	FullName   string     `json:"full_name"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty"`
}
