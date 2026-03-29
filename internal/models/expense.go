package models

import "time"

type Expense struct {
	ID          string     `json:"id"`
	UserID      string     `json:"user_id"`
	Description string     `json:"description"`
	Amount      float64    `json:"amount"`
	Category    string     `json:"category"`
	ExpenseDate time.Time  `json:"expense_date"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

type CreateExpenseRequest struct {
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
	Category    string  `json:"category"`
	ExpenseDate string  `json:"expense_date"`
}

type UpdateExpenseRequest struct {
	Description *string  `json:"description"`
	Amount      *float64 `json:"amount"`
	Category    *string  `json:"category"`
	ExpenseDate *string  `json:"expense_date"`
}

type ExpenseFilters struct {
	Month    *int
	Year     *int
	Category *string
}
