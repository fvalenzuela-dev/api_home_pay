package models

import "time"

type Category struct {
	ID         int        `json:"id"`
	Name       string     `json:"name"`
	AuthUserID string     `json:"auth_user_id"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty"`
}

type CreateCategoryRequest struct {
	Name string `json:"name"`
}

type UpdateCategoryRequest struct {
	Name *string `json:"name"`
}
