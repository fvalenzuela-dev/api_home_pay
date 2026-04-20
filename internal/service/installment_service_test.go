package service

import (
	"context"
	"testing"

	"github.com/homepay/api/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockInstallmentRepoForTest struct {
	mock.Mock
}

func (m *MockInstallmentRepoForTest) CreatePlan(ctx context.Context, authUserID string, plan *models.InstallmentPlan) (*models.InstallmentPlan, error) {
	args := m.Called(ctx, authUserID, plan)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.InstallmentPlan), args.Error(1)
}

func (m *MockInstallmentRepoForTest) CreatePayments(ctx context.Context, payments []models.InstallmentPayment) error {
	args := m.Called(ctx, payments)
	return args.Error(0)
}

func (m *MockInstallmentRepoForTest) GetPlan(ctx context.Context, id, authUserID string) (*models.InstallmentPlan, error) {
	args := m.Called(ctx, id, authUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.InstallmentPlan), args.Error(1)
}

func (m *MockInstallmentRepoForTest) GetAllPlans(ctx context.Context, authUserID string, p models.PaginationParams) ([]models.InstallmentPlan, int, error) {
	args := m.Called(ctx, authUserID, p)
	return args.Get(0).([]models.InstallmentPlan), args.Int(1), args.Error(2)
}

func (m *MockInstallmentRepoForTest) GetPaymentsByPlan(ctx context.Context, planID string, p models.PaginationParams) ([]models.InstallmentPayment, int, error) {
	args := m.Called(ctx, planID, p)
	return args.Get(0).([]models.InstallmentPayment), args.Int(1), args.Error(2)
}

func (m *MockInstallmentRepoForTest) GetActivePaymentsByMonth(ctx context.Context, authUserID string, month, year int) ([]models.InstallmentPayment, error) {
	args := m.Called(ctx, authUserID, month, year)
	return args.Get(0).([]models.InstallmentPayment), args.Error(1)
}

func (m *MockInstallmentRepoForTest) UpdatePayment(ctx context.Context, planID, paymentID, authUserID string) (*models.InstallmentPayment, error) {
	args := m.Called(ctx, planID, paymentID, authUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.InstallmentPayment), args.Error(1)
}

func (m *MockInstallmentRepoForTest) IncrementPaid(ctx context.Context, planID string, total int) error {
	args := m.Called(ctx, planID, total)
	return args.Error(0)
}

func (m *MockInstallmentRepoForTest) SoftDeletePlan(ctx context.Context, id, authUserID string) error {
	args := m.Called(ctx, id, authUserID)
	return args.Error(0)
}

func TestInstallmentService_Create(t *testing.T) {
	mockRepo := new(MockInstallmentRepoForTest)
	svc := NewInstallmentService(mockRepo)

	t.Run("error - description is required", func(t *testing.T) {
		req := &models.CreateInstallmentRequest{
			Description:       "",
			TotalAmount:      120000,
			TotalInstallments: 12,
			StartDate:        "2026-03-01",
		}

		result, err := svc.Create(context.Background(), "user_123", req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "description is required")
	})

	t.Run("error - total_amount must be greater than 0", func(t *testing.T) {
		req := &models.CreateInstallmentRequest{
			Description:       "Test Plan",
			TotalAmount:      0,
			TotalInstallments: 12,
			StartDate:        "2026-03-01",
		}

		result, err := svc.Create(context.Background(), "user_123", req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "total_amount must be greater than 0")
	})

	t.Run("error - total_installments must be greater than 0", func(t *testing.T) {
		req := &models.CreateInstallmentRequest{
			Description:       "Test Plan",
			TotalAmount:      120000,
			TotalInstallments: 0,
			StartDate:        "2026-03-01",
		}

		result, err := svc.Create(context.Background(), "user_123", req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "total_installments must be greater than 0")
	})

	t.Run("error - start_date is required", func(t *testing.T) {
		req := &models.CreateInstallmentRequest{
			Description:       "Test Plan",
			TotalAmount:      120000,
			TotalInstallments: 12,
			StartDate:        "",
		}

		result, err := svc.Create(context.Background(), "user_123", req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "start_date is required")
	})

	t.Run("error - invalid start_date format", func(t *testing.T) {
		req := &models.CreateInstallmentRequest{
			Description:       "Test Plan",
			TotalAmount:      120000,
			TotalInstallments: 12,
			StartDate:        "01-03-2026",
		}

		result, err := svc.Create(context.Background(), "user_123", req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "invalid start_date format")
	})
}

func TestInstallmentService_GetAll(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockRepo := new(MockInstallmentRepoForTest)
		svc := NewInstallmentService(mockRepo)
		mockRepo.On("GetAllPlans", mock.Anything, "user_123", mock.Anything).Return([]models.InstallmentPlan{{ID: "p1"}}, 1, nil)
		mockRepo.On("GetPaymentsByPlan", mock.Anything, "p1", mock.Anything).Return([]models.InstallmentPayment{{ID: "pmt1"}}, 1, nil)

		result, total, err := svc.GetAll(context.Background(), "user_123", models.PaginationParams{})

		assert.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, result, 1)
	})
}

func TestInstallmentService_GetByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockRepo := new(MockInstallmentRepoForTest)
		svc := NewInstallmentService(mockRepo)
		mockRepo.On("GetPlan", mock.Anything, "p1", "user_123").Return(&models.InstallmentPlan{ID: "p1"}, nil)
		mockRepo.On("GetPaymentsByPlan", mock.Anything, "p1", mock.Anything).Return([]models.InstallmentPayment{{ID: "pmt1"}}, 1, nil)

		result, err := svc.GetByID(context.Background(), "p1", "user_123")

		assert.NoError(t, err)
		assert.Equal(t, "p1", result.ID)
	})

	t.Run("not found", func(t *testing.T) {
		mockRepo := new(MockInstallmentRepoForTest)
		svc := NewInstallmentService(mockRepo)
		mockRepo.On("GetPlan", mock.Anything, "notfound", "user_123").Return(nil, nil)

		result, err := svc.GetByID(context.Background(), "notfound", "user_123")

		assert.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestInstallmentService_PayInstallment(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockRepo := new(MockInstallmentRepoForTest)
		svc := NewInstallmentService(mockRepo)
		mockRepo.On("GetPlan", mock.Anything, "p1", "user_123").Return(&models.InstallmentPlan{ID: "p1", TotalInstallments: 12}, nil)
		mockRepo.On("UpdatePayment", mock.Anything, "p1", "pmt1", "user_123").Return(&models.InstallmentPayment{ID: "pmt1", IsPaid: true}, nil)
		mockRepo.On("IncrementPaid", mock.Anything, "p1", 12).Return(nil)

		result, err := svc.PayInstallment(context.Background(), "p1", "pmt1", "user_123")

		assert.NoError(t, err)
		assert.True(t, result.IsPaid)
	})

	t.Run("plan not found", func(t *testing.T) {
		mockRepo := new(MockInstallmentRepoForTest)
		svc := NewInstallmentService(mockRepo)
		mockRepo.On("GetPlan", mock.Anything, "notfound", "user_123").Return(nil, nil)

		result, err := svc.PayInstallment(context.Background(), "notfound", "pmt1", "user_123")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestInstallmentService_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockRepo := new(MockInstallmentRepoForTest)
		svc := NewInstallmentService(mockRepo)
		mockRepo.On("SoftDeletePlan", mock.Anything, "p1", "user_123").Return(nil)

		err := svc.Delete(context.Background(), "p1", "user_123")

		assert.NoError(t, err)
	})

	t.Run("not found", func(t *testing.T) {
		mockRepo := new(MockInstallmentRepoForTest)
		svc := NewInstallmentService(mockRepo)
		mockRepo.On("SoftDeletePlan", mock.Anything, "notfound", "user_123").Return(pgx.ErrNoRows)

		err := svc.Delete(context.Background(), "notfound", "user_123")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}
