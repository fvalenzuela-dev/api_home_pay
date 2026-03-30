package models

import "time"

type Expense struct {
	ID          string     `json:"id"`
	AuthUserID  string     `json:"auth_user_id"`
	CompanyID   *string    `json:"company_id,omitempty"`
	Description string     `json:"description"`
	Amount      float64    `json:"amount"`
	ExpenseDate time.Time  `json:"expense_date"`
	CreatedAt   time.Time  `json:"created_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

type CreateExpenseRequest struct {
	CompanyID   *string `json:"company_id"`
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
	ExpenseDate string  `json:"expense_date"`
}

type UpdateExpenseRequest struct {
	CompanyID   *string  `json:"company_id"`
	Description *string  `json:"description"`
	Amount      *float64 `json:"amount"`
	ExpenseDate *string  `json:"expense_date"`
}

type ExpenseFilters struct {
	Month     *int
	Year      *int
	CompanyID *string
}
