package services

import (
	"fmt"

	"github.com/fernandovalenzuela/api-home-pay/internal/models"
	"github.com/fernandovalenzuela/api-home-pay/internal/repository"
)

type periodService struct {
	repo repository.PeriodRepository
}

func NewPeriodService(repo repository.PeriodRepository) PeriodService {
	return &periodService{repo: repo}
}

func (s *periodService) Create(userID string, period *models.Period) error {
	if err := validatePeriod(period); err != nil {
		return err
	}

	exists, err := s.repo.ExistsByMonthYear(userID, period.MonthNumber, period.YearNumber)
	if err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	if exists {
		return fmt.Errorf("period for month %d and year %d already exists", period.MonthNumber, period.YearNumber)
	}

	return s.repo.Create(userID, period)
}

func (s *periodService) GetByID(userID string, id int) (*models.Period, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid period ID")
	}

	period, err := s.repo.GetByID(userID, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get period: %w", err)
	}
	if period == nil {
		return nil, fmt.Errorf("period not found")
	}

	return period, nil
}

func (s *periodService) GetAll(userID string) ([]models.Period, error) {
	return s.repo.GetAll(userID)
}

func (s *periodService) Update(userID string, period *models.Period) error {
	if period.ID <= 0 {
		return fmt.Errorf("invalid period ID")
	}

	if err := validatePeriod(period); err != nil {
		return err
	}

	existing, err := s.repo.GetByID(userID, period.ID)
	if err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("period not found")
	}

	if existing.MonthNumber != period.MonthNumber || existing.YearNumber != period.YearNumber {
		exists, err := s.repo.ExistsByMonthYear(userID, period.MonthNumber, period.YearNumber)
		if err != nil {
			return fmt.Errorf("validation error: %w", err)
		}
		if exists {
			return fmt.Errorf("period for month %d and year %d already exists", period.MonthNumber, period.YearNumber)
		}
	}

	return s.repo.Update(userID, period)
}

func (s *periodService) Delete(userID string, id int) error {
	if id <= 0 {
		return fmt.Errorf("invalid period ID")
	}

	hasDependencies, err := s.repo.HasExpensesOrIncomes(id)
	if err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	if hasDependencies {
		return fmt.Errorf("cannot delete period with associated expenses or incomes")
	}

	return s.repo.Delete(userID, id)
}

func validatePeriod(period *models.Period) error {
	if period.MonthNumber < 1 || period.MonthNumber > 12 {
		return fmt.Errorf("month must be between 1 and 12")
	}

	if period.YearNumber < 1 {
		return fmt.Errorf("year must be a positive number")
	}

	return nil
}
