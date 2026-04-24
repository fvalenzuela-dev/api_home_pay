package models

import "time"

type Company struct {
	ID         string     `json:"id"`
	AuthUserID string     `json:"auth_user_id"`
	CategoryID int        `json:"category_id"`
	Name       string     `json:"name"`
	Website    *string    `json:"website,omitempty"`
	Phone      *string    `json:"phone,omitempty"`
	IsActive   bool       `json:"is_active"`
	CreatedAt  time.Time  `json:"created_at"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty"`
}

type CreateCompanyRequest struct {
	Name       string  `json:"name"`
	CategoryID int     `json:"category_id"`
	Website    *string `json:"website"`
	Phone      *string `json:"phone"`
}

type UpdateCompanyRequest struct {
	Name       *string `json:"name"`
	CategoryID *int    `json:"category_id"`
	Website    *string `json:"website"`
	Phone      *string `json:"phone"`
}
