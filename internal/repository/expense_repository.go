package repository

import (
	"github.com/fernandovalenzuela/api-home-pay/internal/models"
)

type ExpenseRepository interface {
	Create(userID string, expense *models.Expense) error
	GetByID(userID string, id int) (*models.Expense, error)
	GetAll(userID string, filters ExpenseFilters) ([]models.Expense, error)
	Update(userID string, expense *models.Expense) error
	Delete(userID string, id int) error
	MarkAsPaid(userID string, id int) error
	UpdateAmountPaid(userID string, id int, amount float64) error
	CategoryExistsAndBelongsToUser(userID string, categoryID int) (bool, error)
	PeriodExistsAndBelongsToUser(userID string, periodID int) (bool, error)
	ServiceAccountExistsAndBelongsToUser(userID string, accountID int) (bool, error)
	GetPendingExpenses(userID string, daysAhead int, overdueOnly bool) ([]models.Expense, error)
	GetSummaryByPeriod(userID string, periodID int) (*ExpenseSummary, error)
}

type ExpenseFilters struct {
	PeriodID      *int
	CategoryID    *int
	AccountID     *int
	PaymentStatus *string // paid, partial, pending
}

type ExpenseSummary struct {
	TotalAmount   float64
	PaidAmount    float64
	PendingAmount float64
	ExpenseCount  int
}
