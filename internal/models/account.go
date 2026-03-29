package models

import "time"

type Account struct {
	ID              string     `json:"id"`
	CompanyID       string     `json:"company_id"`
	Name            string     `json:"name"`
	BillingDay      int        `json:"billing_day"`
	AutoAccumulate  bool       `json:"auto_accumulate"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty"`
}

type CreateAccountRequest struct {
	Name           string `json:"name"`
	BillingDay     int    `json:"billing_day"`
	AutoAccumulate bool   `json:"auto_accumulate"`
}

type UpdateAccountRequest struct {
	Name           *string `json:"name"`
	BillingDay     *int    `json:"billing_day"`
	AutoAccumulate *bool   `json:"auto_accumulate"`
}
