package services

import (
	"github.com/fernandovalenzuela/api-home-pay/internal/models"
)

type ServiceAccountService interface {
	Create(userID string, account *models.ServiceAccount) error
	GetByID(userID string, id int) (*models.ServiceAccount, error)
	GetAll(userID string, companyID *int) ([]models.ServiceAccount, error)
	Update(userID string, account *models.ServiceAccount) error
	Delete(userID string, id int) error
}
