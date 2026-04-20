package service

import (
	"context"
	"testing"

	"github.com/homepay/api/internal/models"
	"github.com/jackc/pgx/v5"
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

func TestExpenseService_GetAll(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockRepo := new(MockExpenseRepoForTest)
		svc := NewExpenseService(mockRepo)
		mockRepo.On("GetAll", mock.Anything, "user_123", mock.Anything, mock.Anything).Return(
			[]models.Expense{{ID: "e1"}}, 1, nil)

		result, total, err := svc.GetAll(context.Background(), "user_123", models.ExpenseFilters{}, models.PaginationParams{})

		assert.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, result, 1)
	})
}

func TestExpenseService_GetByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockRepo := new(MockExpenseRepoForTest)
		svc := NewExpenseService(mockRepo)
		mockRepo.On("GetByID", mock.Anything, "e1", "user_123").Return(&models.Expense{ID: "e1"}, nil)

		result, err := svc.GetByID(context.Background(), "e1", "user_123")

		assert.NoError(t, err)
		assert.Equal(t, "e1", result.ID)
	})

	t.Run("not found", func(t *testing.T) {
		mockRepo := new(MockExpenseRepoForTest)
		svc := NewExpenseService(mockRepo)
		mockRepo.On("GetByID", mock.Anything, "notfound", "user_123").Return(nil, nil)

		result, err := svc.GetByID(context.Background(), "notfound", "user_123")

		assert.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestExpenseService_Update(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockRepo := new(MockExpenseRepoForTest)
		svc := NewExpenseService(mockRepo)
		amount := 100.0
		req := &models.UpdateExpenseRequest{Amount: &amount}
		mockRepo.On("Update", mock.Anything, "e1", "user_123", req).Return(&models.Expense{ID: "e1", Amount: 100}, nil)

		result, err := svc.Update(context.Background(), "e1", "user_123", req)

		assert.NoError(t, err)
		assert.Equal(t, 100.0, result.Amount)
	})

	t.Run("not found", func(t *testing.T) {
		mockRepo := new(MockExpenseRepoForTest)
		svc := NewExpenseService(mockRepo)
		req := &models.UpdateExpenseRequest{}
		mockRepo.On("Update", mock.Anything, "notfound", "user_123", req).Return(nil, nil)

		result, err := svc.Update(context.Background(), "notfound", "user_123", req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestExpenseService_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockRepo := new(MockExpenseRepoForTest)
		svc := NewExpenseService(mockRepo)
		mockRepo.On("SoftDelete", mock.Anything, "e1", "user_123").Return(nil)

		err := svc.Delete(context.Background(), "e1", "user_123")

		assert.NoError(t, err)
	})

	t.Run("not found", func(t *testing.T) {
		mockRepo := new(MockExpenseRepoForTest)
		svc := NewExpenseService(mockRepo)
		mockRepo.On("SoftDelete", mock.Anything, "notfound", "user_123").Return(pgx.ErrNoRows)

		err := svc.Delete(context.Background(), "notfound", "user_123")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}
