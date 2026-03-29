package models

import "time"

type Company struct {
	ID          string     `json:"id"`
	AuthUserID  string     `json:"auth_user_id"`
	Name        string     `json:"name"`
	Category    string     `json:"category"`
	IsActive    bool       `json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

type CreateCompanyRequest struct {
	Name     string `json:"name"`
	Category string `json:"category"`
}

type UpdateCompanyRequest struct {
	Name     *string `json:"name"`
	Category *string `json:"category"`
}
