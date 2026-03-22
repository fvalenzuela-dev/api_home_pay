package services

import (
	"fmt"
	"strings"

	"github.com/fernandovalenzuela/api-home-pay/internal/models"
	"github.com/fernandovalenzuela/api-home-pay/internal/repository"
	"github.com/fernandovalenzuela/api-home-pay/internal/utils"
)

type expenseService struct {
	repo repository.ExpenseRepository
}

func NewExpenseService(repo repository.ExpenseRepository) ExpenseService {
	return &expenseService{repo: repo}
}

func (s *expenseService) Create(userID string, expense *models.Expense) error {
	if strings.TrimSpace(expense.Description) == "" {
		return fmt.Errorf("description cannot be empty")
	}

	if expense.CurrentAmount <= 0 {
		return fmt.Errorf("current amount must be greater than zero")
	}

	if expense.AmountPaid < 0 {
		return fmt.Errorf("amount paid cannot be negative")
	}

	if expense.CurrentInstallment < 1 {
		expense.CurrentInstallment = 1
	}

	if expense.TotalInstallments < 1 {
		expense.TotalInstallments = 1
	}

	if expense.CurrentInstallment > expense.TotalInstallments {
		return fmt.Errorf("current installment cannot be greater than total installments")
	}

	if expense.DueDate != nil && *expense.DueDate != "" {
		if !utils.IsValidDate(*expense.DueDate) {
			return fmt.Errorf("invalid due date format. Use YYYY-MM-DD")
		}
	}

	exists, err := s.repo.CategoryExistsAndBelongsToUser(userID, expense.CategoryID)
	if err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	if !exists {
		return fmt.Errorf("category not found or access denied")
	}

	exists, err = s.repo.PeriodExistsAndBelongsToUser(userID, expense.PeriodID)
	if err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	if !exists {
		return fmt.Errorf("period not found or access denied")
	}

	if expense.AccountID != nil && *expense.AccountID > 0 {
		exists, err = s.repo.ServiceAccountExistsAndBelongsToUser(userID, *expense.AccountID)
		if err != nil {
			return fmt.Errorf("validation error: %w", err)
		}
		if !exists {
			return fmt.Errorf("service account not found or access denied")
		}
	}

	return s.repo.Create(userID, expense)
}

func (s *expenseService) GetByID(userID string, id int) (*models.Expense, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid expense ID")
	}

	expense, err := s.repo.GetByID(userID, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get expense: %w", err)
	}
	if expense == nil {
		return nil, fmt.Errorf("expense not found")
	}

	return expense, nil
}

func (s *expenseService) GetAll(userID string, filters repository.ExpenseFilters) ([]models.Expense, error) {
	if filters.PeriodID != nil && *filters.PeriodID > 0 {
		exists, err := s.repo.PeriodExistsAndBelongsToUser(userID, *filters.PeriodID)
		if err != nil {
			return nil, fmt.Errorf("validation error: %w", err)
		}
		if !exists {
			return nil, fmt.Errorf("period not found or access denied")
		}
	}

	if filters.CategoryID != nil && *filters.CategoryID > 0 {
		exists, err := s.repo.CategoryExistsAndBelongsToUser(userID, *filters.CategoryID)
		if err != nil {
			return nil, fmt.Errorf("validation error: %w", err)
		}
		if !exists {
			return nil, fmt.Errorf("category not found or access denied")
		}
	}

	if filters.AccountID != nil && *filters.AccountID > 0 {
		exists, err := s.repo.ServiceAccountExistsAndBelongsToUser(userID, *filters.AccountID)
		if err != nil {
			return nil, fmt.Errorf("validation error: %w", err)
		}
		if !exists {
			return nil, fmt.Errorf("service account not found or access denied")
		}
	}

	if filters.PaymentStatus != nil {
		validStatuses := map[string]bool{"paid": true, "partial": true, "pending": true}
		if !validStatuses[*filters.PaymentStatus] {
			return nil, fmt.Errorf("invalid payment status. Use: paid, partial, or pending")
		}
	}

	return s.repo.GetAll(userID, filters)
}

func (s *expenseService) Update(userID string, expense *models.Expense) error {
	if expense.ID <= 0 {
		return fmt.Errorf("invalid expense ID")
	}

	if strings.TrimSpace(expense.Description) == "" {
		return fmt.Errorf("description cannot be empty")
	}

	if expense.CurrentAmount <= 0 {
		return fmt.Errorf("current amount must be greater than zero")
	}

	if expense.AmountPaid < 0 {
		return fmt.Errorf("amount paid cannot be negative")
	}

	if expense.CurrentInstallment < 1 {
		expense.CurrentInstallment = 1
	}

	if expense.TotalInstallments < 1 {
		expense.TotalInstallments = 1
	}

	if expense.CurrentInstallment > expense.TotalInstallments {
		return fmt.Errorf("current installment cannot be greater than total installments")
	}

	if expense.DueDate != nil && *expense.DueDate != "" {
		if !utils.IsValidDate(*expense.DueDate) {
			return fmt.Errorf("invalid due date format. Use YYYY-MM-DD")
		}
	}

	existing, err := s.repo.GetByID(userID, expense.ID)
	if err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("expense not found")
	}

	exists, err := s.repo.CategoryExistsAndBelongsToUser(userID, expense.CategoryID)
	if err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	if !exists {
		return fmt.Errorf("category not found or access denied")
	}

	exists, err = s.repo.PeriodExistsAndBelongsToUser(userID, expense.PeriodID)
	if err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	if !exists {
		return fmt.Errorf("period not found or access denied")
	}

	if expense.AccountID != nil && *expense.AccountID > 0 {
		exists, err = s.repo.ServiceAccountExistsAndBelongsToUser(userID, *expense.AccountID)
		if err != nil {
			return fmt.Errorf("validation error: %w", err)
		}
		if !exists {
			return fmt.Errorf("service account not found or access denied")
		}
	}

	return s.repo.Update(userID, expense)
}

func (s *expenseService) Delete(userID string, id int) error {
	if id <= 0 {
		return fmt.Errorf("invalid expense ID")
	}

	existing, err := s.repo.GetByID(userID, id)
	if err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("expense not found")
	}

	return s.repo.Delete(userID, id)
}

func (s *expenseService) MarkAsPaid(userID string, id int) error {
	if id <= 0 {
		return fmt.Errorf("invalid expense ID")
	}

	existing, err := s.repo.GetByID(userID, id)
	if err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("expense not found")
	}

	return s.repo.MarkAsPaid(userID, id)
}

func (s *expenseService) GetPendingExpenses(userID string, daysAhead int, overdueOnly bool) ([]models.Expense, error) {
	if userID == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	if daysAhead < 0 {
		daysAhead = 0
	}

	return s.repo.GetPendingExpenses(userID, daysAhead, overdueOnly)
}
