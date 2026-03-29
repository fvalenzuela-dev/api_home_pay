package service

import (
	"context"
	"fmt"

	"github.com/homepay/api/internal/models"
	"github.com/homepay/api/internal/repository"
	"github.com/jackc/pgx/v5"
)

type ExpenseService interface {
	Create(ctx context.Context, authUserID string, req *models.CreateExpenseRequest) (*models.Expense, error)
	GetAll(ctx context.Context, authUserID string, filters models.ExpenseFilters) ([]models.Expense, error)
	Update(ctx context.Context, id, authUserID string, req *models.UpdateExpenseRequest) (*models.Expense, error)
	Delete(ctx context.Context, id, authUserID string) error
}

type expenseService struct {
	expenses repository.ExpenseRepository
}

func NewExpenseService(expenses repository.ExpenseRepository) ExpenseService {
	return &expenseService{expenses: expenses}
}

func (s *expenseService) Create(ctx context.Context, authUserID string, req *models.CreateExpenseRequest) (*models.Expense, error) {
	if req.Description == "" {
		return nil, fmt.Errorf("description is required")
	}
	if req.Amount <= 0 {
		return nil, fmt.Errorf("amount must be greater than 0")
	}
	if req.ExpenseDate == "" {
		return nil, fmt.Errorf("expense_date is required")
	}
	return s.expenses.Create(ctx, authUserID, req)
}

func (s *expenseService) GetAll(ctx context.Context, authUserID string, filters models.ExpenseFilters) ([]models.Expense, error) {
	return s.expenses.GetAll(ctx, authUserID, filters)
}

func (s *expenseService) Update(ctx context.Context, id, authUserID string, req *models.UpdateExpenseRequest) (*models.Expense, error) {
	expense, err := s.expenses.Update(ctx, id, authUserID, req)
	if err != nil {
		return nil, err
	}
	if expense == nil {
		return nil, fmt.Errorf("not found")
	}
	return expense, nil
}

func (s *expenseService) Delete(ctx context.Context, id, authUserID string) error {
	if err := s.expenses.SoftDelete(ctx, id, authUserID); err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("not found")
		}
		return err
	}
	return nil
}
