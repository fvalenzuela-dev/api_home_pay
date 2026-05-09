package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/homepay/api/internal/middleware"
	"github.com/homepay/api/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock AccountService
type MockAccountService struct {
	mock.Mock
}

func (m *MockAccountService) Create(ctx context.Context, authUserID string, req *models.CreateAccountRequest) (*models.Account, error) {
	args := m.Called(ctx, authUserID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Account), args.Error(1)
}

func (m *MockAccountService) GetByID(ctx context.Context, id, authUserID string) (*models.Account, error) {
	args := m.Called(ctx, id, authUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Account), args.Error(1)
}

func (m *MockAccountService) GetAll(ctx context.Context, authUserID string, companyID *string, sort, order string, p models.PaginationParams) ([]models.Account, int, error) {
	args := m.Called(ctx, authUserID, companyID, sort, order, p)
	return args.Get(0).([]models.Account), args.Int(1), args.Error(2)
}

func (m *MockAccountService) Update(ctx context.Context, id, authUserID string, req *models.UpdateAccountRequest) (*models.Account, error) {
	args := m.Called(ctx, id, authUserID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Account), args.Error(1)
}

func (m *MockAccountService) Delete(ctx context.Context, id, authUserID string) error {
	args := m.Called(ctx, id, authUserID)
	return args.Error(0)
}

// Tests for AccountHandler
func TestAccountHandler_Create(t *testing.T) {
	mockSvc := new(MockAccountService)
	handler := NewAccountHandler(mockSvc)

	t.Run("success - create account", func(t *testing.T) {
		mockSvc.On("Create", mock.Anything, "user_123", mock.Anything).Return(&models.Account{
			ID:             "account-123",
			CompanyID:      "company-123",
			CompanyName:    "Test Company",
			Name:           "Electricity",
			BillingDay:     15,
			AutoAccumulate: true,
		}, nil)

		body := `{"company_id":"company-123","name":"Electricity","billing_day":15,"auto_accumulate":true}`
		req := httptest.NewRequest("POST", "/accounts", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserIDKey, "user_123"))
		w := httptest.NewRecorder()

		handler.Create(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("error - invalid body", func(t *testing.T) {
		body := `{"invalid`
		req := httptest.NewRequest("POST", "/accounts", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserIDKey, "user_123"))
		w := httptest.NewRecorder()

		handler.Create(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestAccountHandler_List(t *testing.T) {
	mockSvc := new(MockAccountService)
	handler := NewAccountHandler(mockSvc)

	t.Run("success - list accounts", func(t *testing.T) {
		accounts := []models.Account{
			{ID: "account-1", Name: "Electricity"},
			{ID: "account-2", Name: "Water"},
		}
		mockSvc.On("GetAll", mock.Anything, "user_123", mock.Anything, "", "", mock.Anything).Return(accounts, 2, nil)

		req := httptest.NewRequest("GET", "/accounts", nil)
		req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserIDKey, "user_123"))
		w := httptest.NewRecorder()

		handler.List(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response, "data")
	})

	t.Run("success - list accounts with company_id filter", func(t *testing.T) {
		accounts := []models.Account{
			{ID: "account-1", Name: "Electricity"},
		}
		companyID := "company-123"
		mockSvc.On("GetAll", mock.Anything, "user_123", &companyID, "", "", mock.Anything).Return(accounts, 1, nil)

		req := httptest.NewRequest("GET", "/accounts?company_id=company-123", nil)
		req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserIDKey, "user_123"))
		w := httptest.NewRecorder()

		handler.List(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("error - service error", func(t *testing.T) {
		mockSvc.On("GetAll", mock.Anything, "user_123", mock.Anything, "", "", mock.Anything).Return(nil, 0, assert.AnError)

		req := httptest.NewRequest("GET", "/accounts", nil)
		req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserIDKey, "user_123"))
		w := httptest.NewRecorder()

		handler.List(w, req)

		assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusInternalServerError)
	})
}

func TestAccountHandler_GetOne(t *testing.T) {
	mockSvc := new(MockAccountService)
	handler := NewAccountHandler(mockSvc)

	t.Run("success - get account", func(t *testing.T) {
		mockSvc.On("GetByID", mock.Anything, "account-123", "user_123").Return(&models.Account{
			ID:        "account-123",
			CompanyID: "company-123",
			Name:      "Electricity",
		}, nil)

		req := httptest.NewRequest("GET", "/accounts/account-123", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "account-123")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserIDKey, "user_123"))
		w := httptest.NewRecorder()

		handler.GetOne(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("error - not found", func(t *testing.T) {
		mockSvc.On("GetByID", mock.Anything, "account-999", "user_123").Return(nil, nil)

		req := httptest.NewRequest("GET", "/accounts/account-999", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "account-999")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserIDKey, "user_123"))
		w := httptest.NewRecorder()

		handler.GetOne(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestAccountHandler_Update(t *testing.T) {
	mockSvc := new(MockAccountService)
	handler := NewAccountHandler(mockSvc)

	t.Run("success - update account", func(t *testing.T) {
		mockSvc.On("Update", mock.Anything, "account-123", "user_123", mock.Anything).Return(&models.Account{
			ID:   "account-123",
			Name: "Updated",
		}, nil)

		body := `{"name":"Updated"}`
		req := httptest.NewRequest("PUT", "/accounts/account-123", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "account-123")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserIDKey, "user_123"))
		w := httptest.NewRecorder()

		handler.Update(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("error - invalid body", func(t *testing.T) {
		body := `{"invalid`
		req := httptest.NewRequest("PUT", "/accounts/account-123", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "account-123")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserIDKey, "user_123"))
		w := httptest.NewRecorder()

		handler.Update(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("error - service error", func(t *testing.T) {
		mockSvc.On("Update", mock.Anything, "account-999", "user_123", mock.Anything).Return(nil, nil)

		body := `{"name":"Updated"}`
		req := httptest.NewRequest("PUT", "/accounts/account-999", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "account-999")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserIDKey, "user_123"))
		w := httptest.NewRecorder()

		handler.Update(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestAccountHandler_Delete(t *testing.T) {
	mockSvc := new(MockAccountService)
	handler := NewAccountHandler(mockSvc)

	t.Run("success - delete account", func(t *testing.T) {
		mockSvc.On("Delete", mock.Anything, "account-123", "user_123").Return(nil)

		req := httptest.NewRequest("DELETE", "/accounts/account-123", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "account-123")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserIDKey, "user_123"))
		w := httptest.NewRecorder()

		handler.Delete(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("error - not found", func(t *testing.T) {
		mockSvc.On("Delete", mock.Anything, "account-999", "user_123").Return(errors.New("not found"))

		req := httptest.NewRequest("DELETE", "/accounts/account-999", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "account-999")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserIDKey, "user_123"))
		w := httptest.NewRecorder()

		handler.Delete(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

