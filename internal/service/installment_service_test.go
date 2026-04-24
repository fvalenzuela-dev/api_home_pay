package service

import (
	"context"
	"fmt"
	"testing"
	"time"

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

func TestInstallmentService_Create_HappyPath(t *testing.T) {
	mockRepo := new(MockInstallmentRepoForTest)
	svc := NewInstallmentService(mockRepo)

	req := &models.CreateInstallmentRequest{
		Description:       "Test Plan",
		TotalAmount:       120000,
		TotalInstallments: 3,
		StartDate:         "2026-03-01",
	}

	createdPlan := &models.InstallmentPlan{
		ID:                "plan-1",
		Description:       req.Description,
		TotalAmount:       req.TotalAmount,
		TotalInstallments: req.TotalInstallments,
		StartDate:         time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
	}

	mockRepo.On("CreatePlan", mock.Anything, "user_123", mock.Anything).Return(createdPlan, nil)
	mockRepo.On("CreatePayments", mock.Anything, mock.Anything).Return(nil)

	result, err := svc.Create(context.Background(), "user_123", req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "plan-1", result.ID)
	assert.Len(t, result.Payments, 3)
	assert.Equal(t, 40000.0, result.Payments[0].Amount) // 120000 / 3

	// Verify first payment
	assert.Equal(t, 1, result.Payments[0].InstallmentNumber)
	assert.Equal(t, "plan-1", result.Payments[0].PlanID)
	assert.Equal(t, time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC), result.Payments[0].DueDate)

	// Verify second payment (one month later)
	assert.Equal(t, 2, result.Payments[1].InstallmentNumber)
	assert.Equal(t, time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC), result.Payments[1].DueDate)

	// Verify third payment (two months later)
	assert.Equal(t, 3, result.Payments[2].InstallmentNumber)
	assert.Equal(t, time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC), result.Payments[2].DueDate)

	mockRepo.AssertExpectations(t)
}

func TestInstallmentService_Create_PaymentsError_Rollback(t *testing.T) {
	mockRepo := new(MockInstallmentRepoForTest)
	svc := NewInstallmentService(mockRepo)

	req := &models.CreateInstallmentRequest{
		Description:       "Test Plan",
		TotalAmount:       120000,
		TotalInstallments: 3,
		StartDate:         "2026-03-01",
	}

	createdPlan := &models.InstallmentPlan{
		ID:                "plan-1",
		Description:       req.Description,
		TotalAmount:       req.TotalAmount,
		TotalInstallments: req.TotalInstallments,
		StartDate:         time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
	}

	mockRepo.On("CreatePlan", mock.Anything, "user_123", mock.Anything).Return(createdPlan, nil)
	mockRepo.On("CreatePayments", mock.Anything, mock.Anything).Return(fmt.Errorf("database error"))

	result, err := svc.Create(context.Background(), "user_123", req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "create payments")
	mockRepo.AssertExpectations(t)
}

func TestInstallmentService_GetByID_RepoError(t *testing.T) {
	mockRepo := new(MockInstallmentRepoForTest)
	svc := NewInstallmentService(mockRepo)

	mockRepo.On("GetPlan", mock.Anything, "p1", "user_123").Return(nil, fmt.Errorf("repo error"))

	result, err := svc.GetByID(context.Background(), "p1", "user_123")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "repo error")
}

func TestInstallmentService_GetAll_Empty(t *testing.T) {
	mockRepo := new(MockInstallmentRepoForTest)
	svc := NewInstallmentService(mockRepo)

	mockRepo.On("GetAllPlans", mock.Anything, "user_123", mock.Anything).Return([]models.InstallmentPlan{}, 0, nil)

	result, total, err := svc.GetAll(context.Background(), "user_123", models.PaginationParams{})

	assert.NoError(t, err)
	assert.Equal(t, 0, total)
	assert.Len(t, result, 0)
	mockRepo.AssertExpectations(t)
}

func TestInstallmentService_PayInstallment_PaymentNotFound(t *testing.T) {
	mockRepo := new(MockInstallmentRepoForTest)
	svc := NewInstallmentService(mockRepo)

	mockRepo.On("GetPlan", mock.Anything, "p1", "user_123").Return(&models.InstallmentPlan{ID: "p1", TotalInstallments: 12}, nil)
	mockRepo.On("UpdatePayment", mock.Anything, "p1", "pmt_notfound", "user_123").Return(nil, nil)

	result, err := svc.PayInstallment(context.Background(), "p1", "pmt_notfound", "user_123")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "not found or already paid")
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

	t.Run("error from repo", func(t *testing.T) {
		mockRepo := new(MockInstallmentRepoForTest)
		svc := NewInstallmentService(mockRepo)
		mockRepo.On("SoftDeletePlan", mock.Anything, "p1", "user_123").Return(fmt.Errorf("db error"))

		err := svc.Delete(context.Background(), "p1", "user_123")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db error")
	})
}

func TestInstallmentService_GetByID_PaymentsError(t *testing.T) {
	mockRepo := new(MockInstallmentRepoForTest)
	svc := NewInstallmentService(mockRepo)

	mockRepo.On("GetPlan", mock.Anything, "p1", "user_123").Return(&models.InstallmentPlan{ID: "p1"}, nil)
	mockRepo.On("GetPaymentsByPlan", mock.Anything, "p1", mock.Anything).Return([]models.InstallmentPayment{}, 0, fmt.Errorf("repo error"))

	result, err := svc.GetByID(context.Background(), "p1", "user_123")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "repo error")
}

func TestInstallmentService_GetAll_PaymentsError(t *testing.T) {
	mockRepo := new(MockInstallmentRepoForTest)
	svc := NewInstallmentService(mockRepo)

	plans := []models.InstallmentPlan{{ID: "p1"}}
	mockRepo.On("GetAllPlans", mock.Anything, "user_123", mock.Anything).Return(plans, 1, nil)
	mockRepo.On("GetPaymentsByPlan", mock.Anything, "p1", mock.Anything).Return([]models.InstallmentPayment{}, 0, fmt.Errorf("repo error"))

	result, total, err := svc.GetAll(context.Background(), "user_123", models.PaginationParams{})

	assert.Error(t, err)
	assert.Equal(t, 0, total)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "repo error")
}

func TestInstallmentService_PayInstallment_IncrementError(t *testing.T) {
	mockRepo := new(MockInstallmentRepoForTest)
	svc := NewInstallmentService(mockRepo)

	mockRepo.On("GetPlan", mock.Anything, "p1", "user_123").Return(&models.InstallmentPlan{ID: "p1", TotalInstallments: 12}, nil)
	mockRepo.On("UpdatePayment", mock.Anything, "p1", "pmt1", "user_123").Return(&models.InstallmentPayment{ID: "pmt1", IsPaid: true}, nil)
	mockRepo.On("IncrementPaid", mock.Anything, "p1", 12).Return(fmt.Errorf("increment error"))

	result, err := svc.PayInstallment(context.Background(), "p1", "pmt1", "user_123")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "increment error")
}

func TestInstallmentService_PayInstallment_GetPlanError(t *testing.T) {
	mockRepo := new(MockInstallmentRepoForTest)
	svc := NewInstallmentService(mockRepo)

	mockRepo.On("GetPlan", mock.Anything, "p1", "user_123").Return(nil, assert.AnError)

	result, err := svc.PayInstallment(context.Background(), "p1", "pmt1", "user_123")

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestInstallmentService_GetAll_MultiplePlans(t *testing.T) {
	t.Run("success with multiple plans", func(t *testing.T) {
		mockRepo := new(MockInstallmentRepoForTest)
		svc := NewInstallmentService(mockRepo)

		plans := []models.InstallmentPlan{
			{ID: "p1", TotalInstallments: 12},
			{ID: "p2", TotalInstallments: 6},
		}
		mockRepo.On("GetAllPlans", mock.Anything, "user_123", mock.Anything).Return(plans, 2, nil)
		mockRepo.On("GetPaymentsByPlan", mock.Anything, "p1", mock.Anything).Return([]models.InstallmentPayment{{ID: "pmt1"}}, 1, nil)
		mockRepo.On("GetPaymentsByPlan", mock.Anything, "p2", mock.Anything).Return([]models.InstallmentPayment{{ID: "pmt2"}}, 1, nil)

		result, total, err := svc.GetAll(context.Background(), "user_123", models.PaginationParams{})

		assert.NoError(t, err)
		assert.Equal(t, 2, total)
		assert.Len(t, result, 2)
	})
}
