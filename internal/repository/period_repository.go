package repository

import (
	"github.com/fernandovalenzuela/api-home-pay/internal/models"
)

type PeriodRepository interface {
	Create(userID string, period *models.Period) error
	GetByID(userID string, id int) (*models.Period, error)
	GetAll(userID string) ([]models.Period, error)
	Update(userID string, period *models.Period) error
	Delete(userID string, id int) error
	ExistsByMonthYear(userID string, monthNumber, yearNumber int) (bool, error)
	HasExpensesOrIncomes(id int) (bool, error)
}
