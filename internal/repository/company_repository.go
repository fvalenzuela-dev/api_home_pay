package repository

import (
	"github.com/fernandovalenzuela/api-home-pay/internal/models"
)

type CompanyRepository interface {
	Create(userID string, company *models.Company) error
	GetByID(userID string, id int) (*models.Company, error)
	GetAll(userID string) ([]models.Company, error)
	Update(userID string, company *models.Company) error
	Delete(userID string, id int) error
	ExistsByName(userID string, name string) (bool, error)
	HasServiceAccounts(id int) (bool, error)
}
