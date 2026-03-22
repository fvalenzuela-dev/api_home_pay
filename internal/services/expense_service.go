package services

import (
	"github.com/fernandovalenzuela/api-home-pay/internal/models"
	"github.com/fernandovalenzuela/api-home-pay/internal/repository"
)

type ExpenseService interface {
	Create(userID string, expense *models.Expense) error
	GetByID(userID string, id int) (*models.Expense, error)
	GetAll(userID string, filters repository.ExpenseFilters) ([]models.Expense, error)
	Update(userID string, expense *models.Expense) error
	Delete(userID string, id int) error
	MarkAsPaid(userID string, id int) error
	GetPendingExpenses(userID string, daysAhead int, overdueOnly bool) ([]models.Expense, error)
}
