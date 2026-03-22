package repository

import (
	"github.com/fernandovalenzuela/api-home-pay/internal/models"
)

type IncomeRepository interface {
	Create(userID string, income *models.Income) error
	GetByID(userID string, id int) (*models.Income, error)
	GetAll(userID string, periodID *int) ([]models.Income, error)
	Update(userID string, income *models.Income) error
	Delete(userID string, id int) error
	PeriodExistsAndBelongsToUser(userID string, periodID int) (bool, error)
	GetTotalByPeriod(userID string, periodID int) (float64, int, error)
}

type IncomeSummary struct {
	TotalAmount float64
	IncomeCount int
}
