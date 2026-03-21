package services

import (
	"github.com/fernandovalenzuela/api-home-pay/internal/models"
)

type IncomeService interface {
	Create(userID string, income *models.Income) error
	GetByID(userID string, id int) (*models.Income, error)
	GetAll(userID string, periodID *int) ([]models.Income, error)
	Update(userID string, income *models.Income) error
	Delete(userID string, id int) error
}
