package services

import (
	"fmt"
	"strings"

	"github.com/fernandovalenzuela/api-home-pay/internal/models"
	"github.com/fernandovalenzuela/api-home-pay/internal/repository"
	"github.com/fernandovalenzuela/api-home-pay/internal/utils"
)

type incomeService struct {
	repo repository.IncomeRepository
}

func NewIncomeService(repo repository.IncomeRepository) IncomeService {
	return &incomeService{repo: repo}
}

func (s *incomeService) Create(userID string, income *models.Income) error {
	if strings.TrimSpace(income.Description) == "" {
		return fmt.Errorf("description cannot be empty")
	}

	if income.Amount <= 0 {
		return fmt.Errorf("amount must be greater than zero")
	}

	if income.ReceivedAt != "" {
		if !utils.IsValidDate(income.ReceivedAt) {
			return fmt.Errorf("invalid received_at date format. Use YYYY-MM-DD")
		}
	}

	exists, err := s.repo.PeriodExistsAndBelongsToUser(userID, income.PeriodID)
	if err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	if !exists {
		return fmt.Errorf("period not found or access denied")
	}

	return s.repo.Create(userID, income)
}

func (s *incomeService) GetByID(userID string, id int) (*models.Income, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid income ID")
	}

	income, err := s.repo.GetByID(userID, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get income: %w", err)
	}
	if income == nil {
		return nil, fmt.Errorf("income not found")
	}

	return income, nil
}

func (s *incomeService) GetAll(userID string, periodID *int) ([]models.Income, error) {
	if periodID != nil && *periodID > 0 {
		exists, err := s.repo.PeriodExistsAndBelongsToUser(userID, *periodID)
		if err != nil {
			return nil, fmt.Errorf("validation error: %w", err)
		}
		if !exists {
			return nil, fmt.Errorf("period not found or access denied")
		}
	}

	return s.repo.GetAll(userID, periodID)
}

func (s *incomeService) Update(userID string, income *models.Income) error {
	if income.ID <= 0 {
		return fmt.Errorf("invalid income ID")
	}

	if strings.TrimSpace(income.Description) == "" {
		return fmt.Errorf("description cannot be empty")
	}

	if income.Amount <= 0 {
		return fmt.Errorf("amount must be greater than zero")
	}

	if income.ReceivedAt != "" {
		if !utils.IsValidDate(income.ReceivedAt) {
			return fmt.Errorf("invalid received_at date format. Use YYYY-MM-DD")
		}
	}

	existing, err := s.repo.GetByID(userID, income.ID)
	if err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("income not found")
	}

	exists, err := s.repo.PeriodExistsAndBelongsToUser(userID, income.PeriodID)
	if err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	if !exists {
		return fmt.Errorf("period not found or access denied")
	}

	return s.repo.Update(userID, income)
}

func (s *incomeService) Delete(userID string, id int) error {
	if id <= 0 {
		return fmt.Errorf("invalid income ID")
	}

	existing, err := s.repo.GetByID(userID, id)
	if err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("income not found")
	}

	return s.repo.Delete(userID, id)
}
