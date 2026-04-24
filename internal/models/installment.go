package models

import "time"

type InstallmentPlan struct {
	ID                string     `json:"id"`
	AuthUserID        string     `json:"auth_user_id"`
	Description       string     `json:"description"`
	TotalAmount       float64    `json:"total_amount"`
	TotalInstallments int        `json:"total_installments"`
	InstallmentsPaid  int        `json:"installments_paid"`
	StartDate         time.Time  `json:"start_date"`
	IsCompleted       bool       `json:"is_completed"`
	CreatedAt         time.Time  `json:"created_at"`
	DeletedAt         *time.Time `json:"deleted_at,omitempty"`
}

type InstallmentPayment struct {
	ID                 string     `json:"id"`
	PlanID             string     `json:"plan_id"`
	InstallmentNumber  int        `json:"installment_number"`
	Amount             float64    `json:"amount"`
	DueDate            time.Time  `json:"due_date"`
	IsPaid             bool       `json:"is_paid"`
	PaidAt             *time.Time `json:"paid_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	DeletedAt          *time.Time `json:"deleted_at,omitempty"`
}

type CreateInstallmentRequest struct {
	Description       string  `json:"description"`
	TotalAmount       float64 `json:"total_amount"`
	TotalInstallments int     `json:"total_installments"`
	StartDate         string  `json:"start_date"`
}

type InstallmentPlanWithPayments struct {
	InstallmentPlan
	Payments []InstallmentPayment `json:"payments"`
}
