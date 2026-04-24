package models

import "time"

type Account struct {
	ID             string     `json:"id"`
	CompanyID      string     `json:"company_id"`
	GroupID        *string    `json:"group_id,omitempty"`
	AccountNumber  *string    `json:"account_number,omitempty"`
	Name           string     `json:"name"`
	BillingDay     int        `json:"billing_day"`
	AutoAccumulate bool       `json:"auto_accumulate"`
	IsActive       bool       `json:"is_active"`
	CreatedAt      time.Time  `json:"created_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty"`
}

type CreateAccountRequest struct {
	GroupID        *string `json:"group_id"`
	AccountNumber  *string `json:"account_number"`
	Name           string  `json:"name"`
	BillingDay     int     `json:"billing_day"`
	AutoAccumulate bool    `json:"auto_accumulate"`
}

type UpdateAccountRequest struct {
	GroupID        *string `json:"group_id"`
	AccountNumber  *string `json:"account_number"`
	Name           *string `json:"name"`
	BillingDay     *int    `json:"billing_day"`
	AutoAccumulate *bool   `json:"auto_accumulate"`
}
