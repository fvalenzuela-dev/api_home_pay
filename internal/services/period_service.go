package services

import (
	"github.com/fernandovalenzuela/api-home-pay/internal/models"
)

type PeriodService interface {
	Create(userID string, period *models.Period) error
	GetByID(userID string, id int) (*models.Period, error)
	GetAll(userID string) ([]models.Period, error)
	Update(userID string, period *models.Period) error
	Delete(userID string, id int) error
}
