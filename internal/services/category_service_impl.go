package services

import (
	"fmt"
	"strings"

	"github.com/fernandovalenzuela/api-home-pay/internal/models"
	"github.com/fernandovalenzuela/api-home-pay/internal/repository"
)

type categoryService struct {
	repo repository.CategoryRepository
}

func NewCategoryService(repo repository.CategoryRepository) CategoryService {
	return &categoryService{repo: repo}
}

func (s *categoryService) Create(userID string, category *models.Category) error {
	if strings.TrimSpace(category.Name) == "" {
		return fmt.Errorf("category name cannot be empty")
	}

	exists, err := s.repo.ExistsByName(userID, category.Name)
	if err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	if exists {
		return fmt.Errorf("category with name '%s' already exists", category.Name)
	}

	return s.repo.Create(userID, category)
}

func (s *categoryService) GetByID(userID string, id int) (*models.Category, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid category ID")
	}

	category, err := s.repo.GetByID(userID, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get category: %w", err)
	}
	if category == nil {
		return nil, fmt.Errorf("category not found")
	}

	return category, nil
}

func (s *categoryService) GetAll(userID string) ([]models.Category, error) {
	return s.repo.GetAll(userID)
}

func (s *categoryService) Update(userID string, category *models.Category) error {
	if category.ID <= 0 {
		return fmt.Errorf("invalid category ID")
	}

	if strings.TrimSpace(category.Name) == "" {
		return fmt.Errorf("category name cannot be empty")
	}

	existing, err := s.repo.GetByID(userID, category.ID)
	if err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("category not found")
	}

	if existing.Name != category.Name {
		exists, err := s.repo.ExistsByName(userID, category.Name)
		if err != nil {
			return fmt.Errorf("validation error: %w", err)
		}
		if exists {
			return fmt.Errorf("category with name '%s' already exists", category.Name)
		}
	}

	return s.repo.Update(userID, category)
}

func (s *categoryService) Delete(userID string, id int) error {
	if id <= 0 {
		return fmt.Errorf("invalid category ID")
	}

	hasExpenses, err := s.repo.HasExpenses(id)
	if err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	if hasExpenses {
		return fmt.Errorf("cannot delete category with associated expenses")
	}

	return s.repo.Delete(userID, id)
}
