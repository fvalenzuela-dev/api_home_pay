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
	Period       int        `json:"period"`
	AmountBilled float64    `json:"amount_billed"`
	AmountPaid   *float64   `json:"amount_paid,omitempty"`  // opcional; si >= amount_billed se marca como pagada
	IsPaid       *bool      `json:"is_paid,omitempty"`       // opcional; fuerza estado pagado
	PaidAt       *time.Time `json:"paid_at,omitempty"`       // opcional; fecha de pago
	CarriedFrom  *string    `json:"carried_from,omitempty"`  // UUID de factura impaga anterior (carry-over manual)
}

type UpdateBillingRequest struct {
	AmountBilled *float64   `json:"amount_billed,omitempty"`
	AmountPaid   *float64   `json:"amount_paid,omitempty"`
	IsPaid       *bool      `json:"is_paid,omitempty"`
	PaidAt       *time.Time `json:"paid_at,omitempty"`
}

// OpenPeriodResponse — resultado de la apertura de un periodo.
type OpenPeriodResponse struct {
	Period  int `json:"period"`
	Created int `json:"created"`
	Skipped int `json:"skipped"`
}

// AccountBillingWithDetails — billing con nombre de categoría, empresa y cuenta (usado en GET /periods/{period}/billings).
type AccountBillingWithDetails struct {
	AccountBilling
	CategoryName string `json:"category_name"`
	CompanyName  string `json:"company_name"`
	AccountName  string `json:"account_name"`
}

// PeriodBillingInsert — datos necesarios para insertar un billing al abrir un periodo.
type PeriodBillingInsert struct {
	AccountID    string
	AmountBilled float64
	CarriedFrom  *string
}
