package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fernandovalenzuela/api-home-pay/internal/models"
	"github.com/fernandovalenzuela/api-home-pay/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockExpenseService is a mock implementation of ExpenseService
type MockExpenseService struct {
	mock.Mock
}

func (m *MockExpenseService) Create(userID string, expense *models.Expense) error {
	args := m.Called(userID, expense)
	return args.Error(0)
}

func (m *MockExpenseService) GetByID(userID string, id int) (*models.Expense, error) {
	args := m.Called(userID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Expense), args.Error(1)
}

func (m *MockExpenseService) GetAll(userID string, filters repository.ExpenseFilters) ([]models.Expense, error) {
	args := m.Called(userID, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Expense), args.Error(1)
}

func (m *MockExpenseService) Update(userID string, expense *models.Expense) error {
	args := m.Called(userID, expense)
	return args.Error(0)
}

func (m *MockExpenseService) Delete(userID string, id int) error {
	args := m.Called(userID, id)
	return args.Error(0)
}

func (m *MockExpenseService) MarkAsPaid(userID string, id int) error {
	args := m.Called(userID, id)
	return args.Error(0)
}

func (m *MockExpenseService) GetPendingExpenses(userID string, daysAhead int, overdueOnly bool) ([]models.Expense, error) {
	args := m.Called(userID, daysAhead, overdueOnly)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Expense), args.Error(1)
}

func setupExpenseHandlerTest() (*ExpenseHandler, *MockExpenseService, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockExpenseService)
	handler := NewExpenseHandler(mockService)
	router := gin.New()
	return handler, mockService, router
}

func TestExpenseHandler_Create_Success(t *testing.T) {
	handler, mockService, router := setupExpenseHandlerTest()
	_ = mockService

	router.POST("/expenses", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.Create(c)
	})

	accountID := 1
	dueDate := "2024-06-15"
	expense := models.Expense{
		CategoryID:        1,
		PeriodID:          1,
		AccountID:         &accountID,
		Description:       "Test Expense",
		DueDate:           &dueDate,
		CurrentAmount:     100.00,
		AmountPaid:        0,
		TotalInstallments: 1,
	}
	body, _ := json.Marshal(expense)

	mockService.On("Create", "user123", mock.AnythingOfType("*models.Expense")).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/expenses", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestExpenseHandler_Create_Unauthorized(t *testing.T) {
	handler, mockService, router := setupExpenseHandlerTest()
	_ = mockService

	router.POST("/expenses", func(c *gin.Context) {
		handler.Create(c)
	})

	expense := models.Expense{Description: "Test"}
	body, _ := json.Marshal(expense)

	req := httptest.NewRequest(http.MethodPost, "/expenses", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestExpenseHandler_Create_InvalidJSON(t *testing.T) {
	handler, mockService, router := setupExpenseHandlerTest()
	_ = mockService

	router.POST("/expenses", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.Create(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/expenses", bytes.NewBuffer([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestExpenseHandler_Create_ValidationError(t *testing.T) {
	handler, mockService, router := setupExpenseHandlerTest()
	_ = mockService

	router.POST("/expenses", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.Create(c)
	})

	body, _ := json.Marshal(map[string]interface{}{"description": "Test"})

	req := httptest.NewRequest(http.MethodPost, "/expenses", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertNotCalled(t, "Create")
}

func TestExpenseHandler_GetByID_Success(t *testing.T) {
	handler, mockService, router := setupExpenseHandlerTest()
	_ = mockService

	router.GET("/expenses/:id", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.GetByID(c)
	})

	expectedExpense := &models.Expense{
		ID:            1,
		Description:   "Test Expense",
		CurrentAmount: 100.00,
	}
	mockService.On("GetByID", "user123", 1).Return(expectedExpense, nil)

	req := httptest.NewRequest(http.MethodGet, "/expenses/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestExpenseHandler_GetByID_Unauthorized(t *testing.T) {
	handler, mockService, router := setupExpenseHandlerTest()
	_ = mockService

	router.GET("/expenses/:id", func(c *gin.Context) {
		handler.GetByID(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/expenses/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestExpenseHandler_GetByID_InvalidID(t *testing.T) {
	handler, mockService, router := setupExpenseHandlerTest()
	_ = mockService

	router.GET("/expenses/:id", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.GetByID(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/expenses/invalid", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestExpenseHandler_GetByID_NotFound(t *testing.T) {
	handler, mockService, router := setupExpenseHandlerTest()
	_ = mockService

	router.GET("/expenses/:id", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.GetByID(c)
	})

	mockService.On("GetByID", "user123", 999).Return(nil, errors.New("expense not found"))

	req := httptest.NewRequest(http.MethodGet, "/expenses/999", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

func TestExpenseHandler_GetAll_Success(t *testing.T) {
	handler, mockService, router := setupExpenseHandlerTest()
	_ = mockService

	router.GET("/expenses", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.GetAll(c)
	})

	expectedExpenses := []models.Expense{
		{ID: 1, Description: "Expense 1"},
		{ID: 2, Description: "Expense 2"},
	}
	mockService.On("GetAll", "user123", repository.ExpenseFilters{}).Return(expectedExpenses, nil)

	req := httptest.NewRequest(http.MethodGet, "/expenses", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestExpenseHandler_GetAll_WithFilters(t *testing.T) {
	handler, mockService, router := setupExpenseHandlerTest()
	_ = mockService

	router.GET("/expenses", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.GetAll(c)
	})

	periodID := 1
	categoryID := 2
	accountID := 3
	paymentStatus := "paid"

	filters := repository.ExpenseFilters{
		PeriodID:      &periodID,
		CategoryID:    &categoryID,
		AccountID:     &accountID,
		PaymentStatus: &paymentStatus,
	}

	expectedExpenses := []models.Expense{{ID: 1, Description: "Filtered Expense"}}
	mockService.On("GetAll", "user123", filters).Return(expectedExpenses, nil)

	req := httptest.NewRequest(http.MethodGet, "/expenses?period_id=1&category_id=2&account_id=3&payment_status=paid", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestExpenseHandler_GetAll_Unauthorized(t *testing.T) {
	handler, mockService, router := setupExpenseHandlerTest()
	_ = mockService

	router.GET("/expenses", func(c *gin.Context) {
		handler.GetAll(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/expenses", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestExpenseHandler_GetAll_ServiceError(t *testing.T) {
	handler, mockService, router := setupExpenseHandlerTest()
	_ = mockService

	router.GET("/expenses", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.GetAll(c)
	})

	mockService.On("GetAll", "user123", repository.ExpenseFilters{}).Return(nil, errors.New("database error"))

	req := httptest.NewRequest(http.MethodGet, "/expenses", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

func TestExpenseHandler_Update_Success(t *testing.T) {
	handler, mockService, router := setupExpenseHandlerTest()
	_ = mockService

	router.PUT("/expenses/:id", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.Update(c)
	})

	body, _ := json.Marshal(map[string]interface{}{
		"category_id":    1,
		"period_id":      1,
		"description":    "Updated Expense",
		"current_amount": 150.00,
	})

	mockService.On("Update", "user123", mock.AnythingOfType("*models.Expense")).Return(nil)

	req := httptest.NewRequest(http.MethodPut, "/expenses/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestExpenseHandler_Update_Unauthorized(t *testing.T) {
	handler, mockService, router := setupExpenseHandlerTest()
	_ = mockService

	router.PUT("/expenses/:id", func(c *gin.Context) {
		handler.Update(c)
	})

	expense := models.Expense{Description: "Updated"}
	body, _ := json.Marshal(expense)

	req := httptest.NewRequest(http.MethodPut, "/expenses/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestExpenseHandler_Update_InvalidID(t *testing.T) {
	handler, mockService, router := setupExpenseHandlerTest()
	_ = mockService

	router.PUT("/expenses/:id", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.Update(c)
	})

	expense := models.Expense{Description: "Updated"}
	body, _ := json.Marshal(expense)

	req := httptest.NewRequest(http.MethodPut, "/expenses/invalid", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestExpenseHandler_Update_InvalidJSON(t *testing.T) {
	handler, mockService, router := setupExpenseHandlerTest()
	_ = mockService

	router.PUT("/expenses/:id", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.Update(c)
	})

	req := httptest.NewRequest(http.MethodPut, "/expenses/1", bytes.NewBuffer([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestExpenseHandler_Update_ValidationError(t *testing.T) {
	handler, mockService, router := setupExpenseHandlerTest()
	_ = mockService

	router.PUT("/expenses/:id", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.Update(c)
	})

	body, _ := json.Marshal(map[string]interface{}{"description": "Updated"})

	req := httptest.NewRequest(http.MethodPut, "/expenses/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertNotCalled(t, "Update")
}

func TestExpenseHandler_Delete_Success(t *testing.T) {
	handler, mockService, router := setupExpenseHandlerTest()
	_ = mockService

	router.DELETE("/expenses/:id", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.Delete(c)
	})

	mockService.On("Delete", "user123", 1).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/expenses/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestExpenseHandler_Delete_Unauthorized(t *testing.T) {
	handler, mockService, router := setupExpenseHandlerTest()
	_ = mockService

	router.DELETE("/expenses/:id", func(c *gin.Context) {
		handler.Delete(c)
	})

	req := httptest.NewRequest(http.MethodDelete, "/expenses/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestExpenseHandler_Delete_InvalidID(t *testing.T) {
	handler, mockService, router := setupExpenseHandlerTest()
	_ = mockService

	router.DELETE("/expenses/:id", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.Delete(c)
	})

	req := httptest.NewRequest(http.MethodDelete, "/expenses/invalid", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestExpenseHandler_Delete_ValidationError(t *testing.T) {
	handler, mockService, router := setupExpenseHandlerTest()
	_ = mockService

	router.DELETE("/expenses/:id", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.Delete(c)
	})

	mockService.On("Delete", "user123", 1).Return(errors.New("expense not found"))

	req := httptest.NewRequest(http.MethodDelete, "/expenses/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}

func TestExpenseHandler_MarkAsPaid_Success(t *testing.T) {
	handler, mockService, router := setupExpenseHandlerTest()
	_ = mockService

	router.POST("/expenses/:id/pay", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.MarkAsPaid(c)
	})

	mockService.On("MarkAsPaid", "user123", 1).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/expenses/1/pay", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestExpenseHandler_MarkAsPaid_Unauthorized(t *testing.T) {
	handler, mockService, router := setupExpenseHandlerTest()
	_ = mockService

	router.POST("/expenses/:id/pay", func(c *gin.Context) {
		handler.MarkAsPaid(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/expenses/1/pay", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestExpenseHandler_MarkAsPaid_InvalidID(t *testing.T) {
	handler, mockService, router := setupExpenseHandlerTest()
	_ = mockService

	router.POST("/expenses/:id/pay", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.MarkAsPaid(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/expenses/invalid/pay", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestExpenseHandler_MarkAsPaid_ValidationError(t *testing.T) {
	handler, mockService, router := setupExpenseHandlerTest()
	_ = mockService

	router.POST("/expenses/:id/pay", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.MarkAsPaid(c)
	})

	mockService.On("MarkAsPaid", "user123", 1).Return(errors.New("expense not found"))

	req := httptest.NewRequest(http.MethodPost, "/expenses/1/pay", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}

func TestExpenseHandler_GetPending_Success(t *testing.T) {
	handler, mockService, router := setupExpenseHandlerTest()
	_ = mockService

	router.GET("/expenses/pending", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.GetPending(c)
	})

	expectedExpenses := []models.Expense{
		{ID: 1, Description: "Pending 1"},
		{ID: 2, Description: "Pending 2"},
	}
	mockService.On("GetPendingExpenses", "user123", 7, false).Return(expectedExpenses, nil)

	req := httptest.NewRequest(http.MethodGet, "/expenses/pending", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestExpenseHandler_GetPending_WithDaysAhead(t *testing.T) {
	handler, mockService, router := setupExpenseHandlerTest()
	_ = mockService

	router.GET("/expenses/pending", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.GetPending(c)
	})

	expectedExpenses := []models.Expense{{ID: 1, Description: "Pending"}}
	mockService.On("GetPendingExpenses", "user123", 14, false).Return(expectedExpenses, nil)

	req := httptest.NewRequest(http.MethodGet, "/expenses/pending?days_ahead=14", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestExpenseHandler_GetPending_OverdueOnly(t *testing.T) {
	handler, mockService, router := setupExpenseHandlerTest()
	_ = mockService

	router.GET("/expenses/pending", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.GetPending(c)
	})

	expectedExpenses := []models.Expense{{ID: 1, Description: "Overdue"}}
	mockService.On("GetPendingExpenses", "user123", 7, true).Return(expectedExpenses, nil)

	req := httptest.NewRequest(http.MethodGet, "/expenses/pending?overdue_only=true", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestExpenseHandler_GetPending_Unauthorized(t *testing.T) {
	handler, mockService, router := setupExpenseHandlerTest()
	_ = mockService

	router.GET("/expenses/pending", func(c *gin.Context) {
		handler.GetPending(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/expenses/pending", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestExpenseHandler_GetPending_ServiceError(t *testing.T) {
	handler, mockService, router := setupExpenseHandlerTest()
	_ = mockService

	router.GET("/expenses/pending", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.GetPending(c)
	})

	mockService.On("GetPendingExpenses", "user123", 7, false).Return(nil, errors.New("database error"))

	req := httptest.NewRequest(http.MethodGet, "/expenses/pending", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}
