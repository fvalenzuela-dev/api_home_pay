package repository

import (
	"github.com/fernandovalenzuela/api-home-pay/internal/models"
)

type CategoryRepository interface {
	Create(userID string, category *models.Category) error
	GetByID(userID string, id int) (*models.Category, error)
	GetAll(userID string) ([]models.Category, error)
	Update(userID string, category *models.Category) error
	Delete(userID string, id int) error
	ExistsByName(userID string, name string) (bool, error)
	HasExpenses(id int) (bool, error)
}
