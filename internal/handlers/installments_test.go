package handlers

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/homepay/api/internal/middleware"
	"github.com/homepay/api/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock InstallmentService
type MockInstallmentService struct {
	mock.Mock
}

func (m *MockInstallmentService) Create(ctx context.Context, authUserID string, req *models.CreateInstallmentRequest) (*models.InstallmentPlanWithPayments, error) {
	args := m.Called(ctx, authUserID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.InstallmentPlanWithPayments), args.Error(1)
}

func (m *MockInstallmentService) GetByID(ctx context.Context, id, authUserID string) (*models.InstallmentPlanWithPayments, error) {
	args := m.Called(ctx, id, authUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.InstallmentPlanWithPayments), args.Error(1)
}

func (m *MockInstallmentService) GetAll(ctx context.Context, authUserID string, p models.PaginationParams) ([]models.InstallmentPlanWithPayments, int, error) {
	args := m.Called(ctx, authUserID, p)
	return args.Get(0).([]models.InstallmentPlanWithPayments), args.Int(1), args.Error(2)
}

func (m *MockInstallmentService) PayInstallment(ctx context.Context, planID, paymentID, authUserID string) (*models.InstallmentPayment, error) {
	args := m.Called(ctx, planID, paymentID, authUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.InstallmentPayment), args.Error(1)
}

func (m *MockInstallmentService) Delete(ctx context.Context, id, authUserID string) error {
	args := m.Called(ctx, id, authUserID)
	return args.Error(0)
}

// Tests for InstallmentHandler
func TestInstallmentHandler_Create(t *testing.T) {
	mockSvc := new(MockInstallmentService)
	handler := NewInstallmentHandler(mockSvc)

	t.Run("success - create installment plan", func(t *testing.T) {
		mockSvc.On("Create", mock.Anything, "user_123", mock.Anything).Return(&models.InstallmentPlanWithPayments{
			InstallmentPlan: models.InstallmentPlan{
				ID:                "plan-123",
				Description:       "TV Payment",
				TotalAmount:       300000,
				TotalInstallments: 12,
			},
			Payments: []models.InstallmentPayment{},
		}, nil)

		body := `{"description":"TV Payment","total_amount":300000,"total_installments":12,"start_date":"2026-03-01"}`
		req := httptest.NewRequest("POST", "/installments", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserIDKey, "user_123"))
		w := httptest.NewRecorder()

		handler.Create(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("error - invalid body", func(t *testing.T) {
		body := `{"invalid`
		req := httptest.NewRequest("POST", "/installments", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserIDKey, "user_123"))
		w := httptest.NewRecorder()

		handler.Create(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})


}

func TestInstallmentHandler_GetOne(t *testing.T) {
	mockSvc := new(MockInstallmentService)
	handler := NewInstallmentHandler(mockSvc)

	t.Run("success - get installment plan", func(t *testing.T) {
		mockSvc.On("GetByID", mock.Anything, "plan-123", "user_123").Return(&models.InstallmentPlanWithPayments{
			InstallmentPlan: models.InstallmentPlan{
				ID:                "plan-123",
				Description:       "TV Payment",
				TotalAmount:       300000,
				TotalInstallments: 12,
			},
			Payments: []models.InstallmentPayment{},
		}, nil)

		req := httptest.NewRequest("GET", "/installments/plan-123", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "plan-123")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserIDKey, "user_123"))
		w := httptest.NewRecorder()

		handler.GetOne(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("error - not found", func(t *testing.T) {
		mockSvc.On("GetByID", mock.Anything, "plan-999", "user_123").Return(nil, nil)

		req := httptest.NewRequest("GET", "/installments/plan-999", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "plan-999")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserIDKey, "user_123"))
		w := httptest.NewRecorder()

		handler.GetOne(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestInstallmentHandler_List(t *testing.T) {
	mockSvc := new(MockInstallmentService)
	handler := NewInstallmentHandler(mockSvc)

	t.Run("success - list installment plans", func(t *testing.T) {
		plans := []models.InstallmentPlanWithPayments{
			{InstallmentPlan: models.InstallmentPlan{ID: "plan-1", Description: "TV Payment"}},
			{InstallmentPlan: models.InstallmentPlan{ID: "plan-2", Description: "Phone"}},
		}
		mockSvc.On("GetAll", mock.Anything, "user_123", mock.Anything).Return(plans, 2, nil)

		req := httptest.NewRequest("GET", "/installments", nil)
		req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserIDKey, "user_123"))
		w := httptest.NewRecorder()

		handler.List(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("success - empty list", func(t *testing.T) {
		mockSvc.On("GetAll", mock.Anything, "user_123", mock.Anything).Return([]models.InstallmentPlanWithPayments{}, 0, nil)

		req := httptest.NewRequest("GET", "/installments", nil)
		req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserIDKey, "user_123"))
		w := httptest.NewRecorder()

		handler.List(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("success - nil plans converts to empty slice", func(t *testing.T) {
		mockSvc.On("GetAll", mock.Anything, "user_123", mock.Anything).Return(nil, 0, nil)

		req := httptest.NewRequest("GET", "/installments", nil)
		req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserIDKey, "user_123"))
		w := httptest.NewRecorder()

		handler.List(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestInstallmentHandler_PayInstallment(t *testing.T) {
	mockSvc := new(MockInstallmentService)
	handler := NewInstallmentHandler(mockSvc)

	t.Run("success - pay installment", func(t *testing.T) {
		mockSvc.On("PayInstallment", mock.Anything, "plan-123", "payment-1", "user_123").Return(&models.InstallmentPayment{
			ID:                "payment-123",
			PlanID:            "plan-123",
			InstallmentNumber: 1,
			Amount:            25000,
			IsPaid:            true,
			PaidAt:            func() *time.Time { t := time.Now(); return &t }(),
		}, nil)

		req := httptest.NewRequest("PUT", "/installments/plan-123/payments/payment-1", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "plan-123")
		rctx.URLParams.Add("paymentID", "payment-1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserIDKey, "user_123"))
		w := httptest.NewRecorder()

		handler.PayInstallment(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("error - not found", func(t *testing.T) {
		mockSvc.On("PayInstallment", mock.Anything, "plan-999", "payment-1", "user_123").Return(nil, errors.New("not found"))

		req := httptest.NewRequest("PUT", "/installments/plan-999/payments/payment-1", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "plan-999")
		rctx.URLParams.Add("paymentID", "payment-1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserIDKey, "user_123"))
		w := httptest.NewRecorder()

		handler.PayInstallment(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})


}

func TestInstallmentHandler_Delete(t *testing.T) {
	mockSvc := new(MockInstallmentService)
	handler := NewInstallmentHandler(mockSvc)

	t.Run("success - delete installment plan", func(t *testing.T) {
		mockSvc.On("Delete", mock.Anything, "plan-123", "user_123").Return(nil)

		req := httptest.NewRequest("DELETE", "/installments/plan-123", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "plan-123")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserIDKey, "user_123"))
		w := httptest.NewRecorder()

		handler.Delete(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("error - not found", func(t *testing.T) {
		mockSvc.On("Delete", mock.Anything, "plan-999", "user_123").Return(errors.New("not found"))

		req := httptest.NewRequest("DELETE", "/installments/plan-999", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "plan-999")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserIDKey, "user_123"))
		w := httptest.NewRecorder()

		handler.Delete(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})


}
