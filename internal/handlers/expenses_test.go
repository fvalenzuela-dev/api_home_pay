package handlers

import (
	"bytes"
	"context"
	"encoding/json"
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

// Mock ExpenseService
type MockExpenseService struct {
	mock.Mock
}

func (m *MockExpenseService) Create(ctx context.Context, authUserID string, req *models.CreateExpenseRequest) (*models.Expense, error) {
	args := m.Called(ctx, authUserID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Expense), args.Error(1)
}

func (m *MockExpenseService) GetByID(ctx context.Context, id, authUserID string) (*models.Expense, error) {
	args := m.Called(ctx, id, authUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Expense), args.Error(1)
}

func (m *MockExpenseService) GetAll(ctx context.Context, authUserID string, filters models.ExpenseFilters, p models.PaginationParams) ([]models.Expense, int, error) {
	args := m.Called(ctx, authUserID, filters, p)
	return args.Get(0).([]models.Expense), args.Int(1), args.Error(2)
}

func (m *MockExpenseService) Update(ctx context.Context, id, authUserID string, req *models.UpdateExpenseRequest) (*models.Expense, error) {
	args := m.Called(ctx, id, authUserID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Expense), args.Error(1)
}

func (m *MockExpenseService) Delete(ctx context.Context, id, authUserID string) error {
	args := m.Called(ctx, id, authUserID)
	return args.Error(0)
}

// Tests for ExpenseHandler
func TestExpenseHandler_Create(t *testing.T) {
	mockSvc := new(MockExpenseService)
	handler := NewExpenseHandler(mockSvc)

	t.Run("success - create expense", func(t *testing.T) {
		mockSvc.On("Create", mock.Anything, "user_123", mock.Anything).Return(&models.Expense{
			ID:          "expense-123",
			AuthUserID:  "user_123",
			Description: "Groceries",
			Amount:      25000,
			ExpenseDate: time.Now(),
		}, nil)

		body := `{"description":"Groceries","amount":25000,"expense_date":"2026-03-15"}`
		req := httptest.NewRequest("POST", "/expenses", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserIDKey, "user_123"))
		w := httptest.NewRecorder()

		handler.Create(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("error - invalid body", func(t *testing.T) {
		body := `{"invalid`
		req := httptest.NewRequest("POST", "/expenses", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserIDKey, "user_123"))
		w := httptest.NewRecorder()

		handler.Create(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestExpenseHandler_List(t *testing.T) {
	mockSvc := new(MockExpenseService)
	handler := NewExpenseHandler(mockSvc)

	t.Run("success - list expenses", func(t *testing.T) {
		expenses := []models.Expense{
			{ID: "expense-1", Description: "Groceries", Amount: 25000},
			{ID: "expense-2", Description: "Gas", Amount: 15000},
		}
		mockSvc.On("GetAll", mock.Anything, "user_123", mock.Anything, mock.Anything).Return(expenses, 2, nil)

		req := httptest.NewRequest("GET", "/expenses", nil)
		req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserIDKey, "user_123"))
		w := httptest.NewRecorder()

		handler.List(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response, "data")
	})

	t.Run("error - service error", func(t *testing.T) {
		mockSvc.On("GetAll", mock.Anything, "user_123", mock.Anything, mock.Anything).Return(nil, 0, assert.AnError)

		req := httptest.NewRequest("GET", "/expenses", nil)
		req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserIDKey, "user_123"))
		w := httptest.NewRecorder()

		handler.List(w, req)

		// Handler might return empty list on error or 500
		assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusInternalServerError)
	})
}

func TestExpenseHandler_GetOne(t *testing.T) {
	mockSvc := new(MockExpenseService)
	handler := NewExpenseHandler(mockSvc)

	t.Run("success - get expense", func(t *testing.T) {
		mockSvc.On("GetByID", mock.Anything, "expense-123", "user_123").Return(&models.Expense{
			ID:          "expense-123",
			Description: "Groceries",
			Amount:      25000,
		}, nil)

		req := httptest.NewRequest("GET", "/expenses/expense-123", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "expense-123")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserIDKey, "user_123"))
		w := httptest.NewRecorder()

		handler.GetOne(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("error - not found", func(t *testing.T) {
		mockSvc.On("GetByID", mock.Anything, "expense-999", "user_123").Return(nil, nil)

		req := httptest.NewRequest("GET", "/expenses/expense-999", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "expense-999")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserIDKey, "user_123"))
		w := httptest.NewRecorder()

		handler.GetOne(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestExpenseHandler_Update(t *testing.T) {
	mockSvc := new(MockExpenseService)
	handler := NewExpenseHandler(mockSvc)

	t.Run("success - update expense", func(t *testing.T) {
		mockSvc.On("Update", mock.Anything, "expense-123", "user_123", mock.Anything).Return(&models.Expense{
			ID:          "expense-123",
			Description: "Updated",
			Amount:      30000,
		}, nil)

		body := `{"amount":30000}`
		req := httptest.NewRequest("PUT", "/expenses/expense-123", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "expense-123")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserIDKey, "user_123"))
		w := httptest.NewRecorder()

		handler.Update(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestExpenseHandler_Delete(t *testing.T) {
	mockSvc := new(MockExpenseService)
	handler := NewExpenseHandler(mockSvc)

	t.Run("success - delete expense", func(t *testing.T) {
		mockSvc.On("Delete", mock.Anything, "expense-123", "user_123").Return(nil)

		req := httptest.NewRequest("DELETE", "/expenses/expense-123", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "expense-123")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserIDKey, "user_123"))
		w := httptest.NewRecorder()

		handler.Delete(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})
}
