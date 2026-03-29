package models

import "time"

type AccountBilling struct {
	ID           string     `json:"id"`
	AccountID    string     `json:"account_id"`
	Month        int        `json:"month"`
	Year         int        `json:"year"`
	AmountBilled float64    `json:"amount_billed"`
	AmountPaid   float64    `json:"amount_paid"`
	IsPaid       bool       `json:"is_paid"`
	PaidAt       *time.Time `json:"paid_at,omitempty"`
	CarriedFrom  *string    `json:"carried_from,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}

type CreateBillingRequest struct {
	Month        int     `json:"month"`
	Year         int     `json:"year"`
	AmountBilled float64 `json:"amount_billed"`
}

type UpdateBillingRequest struct {
	AmountPaid *float64 `json:"amount_paid"`
	IsPaid     *bool    `json:"is_paid"`
}
