package service

import (
	"context"
	"testing"

	"github.com/homepay/api/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockExpenseRepoForTest struct {
	mock.Mock
}

func (m *MockExpenseRepoForTest) Create(ctx context.Context, authUserID string, req *models.CreateExpenseRequest) (*models.Expense, error) {
	args := m.Called(ctx, authUserID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Expense), args.Error(1)
}

func (m *MockExpenseRepoForTest) GetByID(ctx context.Context, id, authUserID string) (*models.Expense, error) {
	args := m.Called(ctx, id, authUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Expense), args.Error(1)
}

func (m *MockExpenseRepoForTest) GetAll(ctx context.Context, authUserID string, filters models.ExpenseFilters, p models.PaginationParams) ([]models.Expense, int, error) {
	args := m.Called(ctx, authUserID, filters, p)
	return args.Get(0).([]models.Expense), args.Int(1), args.Error(2)
}

func (m *MockExpenseRepoForTest) Update(ctx context.Context, id, authUserID string, req *models.UpdateExpenseRequest) (*models.Expense, error) {
	args := m.Called(ctx, id, authUserID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Expense), args.Error(1)
}

func (m *MockExpenseRepoForTest) SoftDelete(ctx context.Context, id, authUserID string) error {
	args := m.Called(ctx, id, authUserID)
	return args.Error(0)
}

func TestExpenseService_Create(t *testing.T) {
	mockRepo := new(MockExpenseRepoForTest)
	svc := NewExpenseService(mockRepo)

	t.Run("error - amount must be greater than 0", func(t *testing.T) {
		req := &models.CreateExpenseRequest{
			Description: "Groceries",
			Amount:     0,
			ExpenseDate: "2026-03-15",
		}

		result, err := svc.Create(context.Background(), "user_123", req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "amount must be greater than 0")
	})

	t.Run("error - expense_date is required", func(t *testing.T) {
		req := &models.CreateExpenseRequest{
			Description: "Groceries",
			Amount:      25000,
			ExpenseDate: "",
		}

		result, err := svc.Create(context.Background(), "user_123", req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "expense_date is required")
	})
}
