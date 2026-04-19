package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/homepay/api/internal/middleware"
	"github.com/homepay/api/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock BillingService
type MockBillingService struct {
	mock.Mock
}

func (m *MockBillingService) Create(ctx context.Context, accountID, authUserID string, req *models.CreateBillingRequest) (*models.AccountBilling, error) {
	args := m.Called(ctx, accountID, authUserID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AccountBilling), args.Error(1)
}

func (m *MockBillingService) GetAllByAccount(ctx context.Context, accountID, authUserID string, p models.PaginationParams) ([]models.AccountBilling, int, error) {
	args := m.Called(ctx, accountID, authUserID, p)
	return args.Get(0).([]models.AccountBilling), args.Int(1), args.Error(2)
}

func (m *MockBillingService) GetAllByPeriod(ctx context.Context, authUserID string, period int, isPaid *bool, p models.PaginationParams) ([]models.AccountBillingWithDetails, int, error) {
	args := m.Called(ctx, authUserID, period, isPaid, p)
	return args.Get(0).([]models.AccountBillingWithDetails), args.Int(1), args.Error(2)
}

func (m *MockBillingService) GetByID(ctx context.Context, id, authUserID string) (*models.AccountBilling, error) {
	args := m.Called(ctx, id, authUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AccountBilling), args.Error(1)
}

func (m *MockBillingService) Update(ctx context.Context, id, authUserID string, req *models.UpdateBillingRequest) (*models.AccountBilling, error) {
	args := m.Called(ctx, id, authUserID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AccountBilling), args.Error(1)
}

func (m *MockBillingService) OpenPeriod(ctx context.Context, authUserID string, period int) (*models.OpenPeriodResponse, error) {
	args := m.Called(ctx, authUserID, period)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.OpenPeriodResponse), args.Error(1)
}

// Tests for BillingHandler
func TestBillingHandler_Create(t *testing.T) {
	mockSvc := new(MockBillingService)
	handler := NewBillingHandler(mockSvc)

	t.Run("success - create billing", func(t *testing.T) {
		mockSvc.On("Create", mock.Anything, "account-123", "user_123", mock.Anything).Return(&models.AccountBilling{
			ID:           "billing-123",
			AccountID:    "account-123",
			Period:       202603,
			AmountBilled: 15000,
			IsPaid:       false,
		}, nil)

		body := `{"period":202603,"amount_billed":15000}`
		req := httptest.NewRequest("POST", "/accounts/account-123/billings", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("accountID", "account-123")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserIDKey, "user_123"))
		w := httptest.NewRecorder()

		handler.Create(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("error - invalid body", func(t *testing.T) {
		body := `{"invalid`
		req := httptest.NewRequest("POST", "/accounts/account-123/billings", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("accountID", "account-123")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserIDKey, "user_123"))
		w := httptest.NewRecorder()

		handler.Create(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestBillingHandler_List(t *testing.T) {
	mockSvc := new(MockBillingService)
	handler := NewBillingHandler(mockSvc)

	t.Run("success - list billings", func(t *testing.T) {
		billings := []models.AccountBilling{
			{ID: "billing-1", Period: 202603, AmountBilled: 15000},
			{ID: "billing-2", Period: 202602, AmountBilled: 14000},
		}
		mockSvc.On("GetAllByAccount", mock.Anything, "account-123", "user_123", mock.Anything).Return(billings, 2, nil)

		req := httptest.NewRequest("GET", "/accounts/account-123/billings", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("accountID", "account-123")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserIDKey, "user_123"))
		w := httptest.NewRecorder()

		handler.List(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response, "data")
	})
}

func TestBillingHandler_GetOne(t *testing.T) {
	mockSvc := new(MockBillingService)
	handler := NewBillingHandler(mockSvc)

	t.Run("success - get billing", func(t *testing.T) {
		mockSvc.On("GetByID", mock.Anything, "billing-123", "user_123").Return(&models.AccountBilling{
			ID:           "billing-123",
			AccountID:    "account-123",
			Period:       202603,
			AmountBilled: 15000,
		}, nil)

		req := httptest.NewRequest("GET", "/billings/billing-123", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "billing-123")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserIDKey, "user_123"))
		w := httptest.NewRecorder()

		handler.GetOne(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("error - not found", func(t *testing.T) {
		mockSvc.On("GetByID", mock.Anything, "billing-999", "user_123").Return(nil, nil)

		req := httptest.NewRequest("GET", "/billings/billing-999", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "billing-999")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserIDKey, "user_123"))
		w := httptest.NewRecorder()

		handler.GetOne(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestBillingHandler_Update(t *testing.T) {
	mockSvc := new(MockBillingService)
	handler := NewBillingHandler(mockSvc)

	t.Run("success - update billing", func(t *testing.T) {
		mockSvc.On("Update", mock.Anything, "billing-123", "user_123", mock.Anything).Return(&models.AccountBilling{
			ID:           "billing-123",
			AmountBilled: 20000,
			IsPaid:       true,
		}, nil)

		body := `{"amount_billed":20000,"is_paid":true}`
		req := httptest.NewRequest("PUT", "/billings/billing-123", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "billing-123")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserIDKey, "user_123"))
		w := httptest.NewRecorder()

		handler.Update(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestBillingHandler_OpenPeriod(t *testing.T) {
	mockSvc := new(MockBillingService)
	handler := NewBillingHandler(mockSvc)

	t.Run("success - open period", func(t *testing.T) {
		mockSvc.On("OpenPeriod", mock.Anything, "user_123", 202603).Return(&models.OpenPeriodResponse{
			Period:  202603,
			Created: 5,
			Skipped: 2,
		}, nil)

		req := httptest.NewRequest("POST", "/periods/202603/open", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("period", "202603")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserIDKey, "user_123"))
		w := httptest.NewRecorder()

		handler.OpenPeriod(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("error - invalid period", func(t *testing.T) {
		mockSvc.On("OpenPeriod", mock.Anything, "user_123", 202613).Return(nil, assert.AnError)

		req := httptest.NewRequest("POST", "/periods/202613/open", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("period", "202613")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserIDKey, "user_123"))
		w := httptest.NewRecorder()

		handler.OpenPeriod(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestBillingHandler_ListByPeriod(t *testing.T) {
	mockSvc := new(MockBillingService)
	handler := NewBillingHandler(mockSvc)

	t.Run("success - list by period", func(t *testing.T) {
		details := []models.AccountBillingWithDetails{
			{AccountBilling: models.AccountBilling{ID: "billing-1", Period: 202603}},
		}
		mockSvc.On("GetAllByPeriod", mock.Anything, "user_123", 202603, (*bool)(nil), mock.Anything).Return(details, 1, nil)

		req := httptest.NewRequest("GET", "/periods/202603/billings", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("period", "202603")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserIDKey, "user_123"))
		w := httptest.NewRecorder()

		handler.ListByPeriod(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
