package services

import (
	"github.com/fernandovalenzuela/api-home-pay/internal/models"
)

type CompanyService interface {
	Create(userID string, company *models.Company) error
	GetByID(userID string, id int) (*models.Company, error)
	GetAll(userID string) ([]models.Company, error)
	Update(userID string, company *models.Company) error
	Delete(userID string, id int) error
}
