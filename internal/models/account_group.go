package models

import "time"

type AccountGroup struct {
	ID         string     `json:"id"`
	AuthUserID string     `json:"auth_user_id"`
	Name       string     `json:"name"`
	CreatedAt  time.Time  `json:"created_at"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty"`
}

type CreateAccountGroupRequest struct {
	Name string `json:"name"`
}

type UpdateAccountGroupRequest struct {
	Name *string `json:"name"`
}
