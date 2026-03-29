package models

import "time"

// AccountBilling representa una factura mensual.
// Period es un entero en formato YYYYMM, ej: 202603 = marzo 2026.
type AccountBilling struct {
	ID           string     `json:"id"`
	AccountID    string     `json:"account_id"`
	Period       int        `json:"period"`
	AmountBilled float64    `json:"amount_billed"`
	AmountPaid   float64    `json:"amount_paid"`
	IsPaid       bool       `json:"is_paid"`
	PaidAt       *time.Time `json:"paid_at,omitempty"`
	CarriedFrom  *string    `json:"carried_from,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}

// CreateBillingRequest — period en formato YYYYMM, ej: 202603
type CreateBillingRequest struct {
	Period       int     `json:"period"`
	AmountBilled float64 `json:"amount_billed"`
}

type UpdateBillingRequest struct {
	AmountPaid *float64 `json:"amount_paid"`
}
