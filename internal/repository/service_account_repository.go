package repository

import (
	"github.com/fernandovalenzuela/api-home-pay/internal/models"
)

type ServiceAccountRepository interface {
	Create(userID string, account *models.ServiceAccount) error
	GetByID(userID string, id int) (*models.ServiceAccount, error)
	GetAll(userID string, companyID *int) ([]models.ServiceAccount, error)
	Update(userID string, account *models.ServiceAccount) error
	Delete(userID string, id int) error
	ExistsByIdentifier(userID string, companyID int, identifier string) (bool, error)
	HasExpenses(id int) (bool, error)
	CompanyExistsAndBelongsToUser(userID string, companyID int) (bool, error)
}
