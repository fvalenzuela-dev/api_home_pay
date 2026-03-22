package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fernandovalenzuela/api-home-pay/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockIncomeService is a mock implementation of IncomeService
type MockIncomeService struct {
	mock.Mock
}

func (m *MockIncomeService) Create(userID string, income *models.Income) error {
	args := m.Called(userID, income)
	return args.Error(0)
}

func (m *MockIncomeService) GetByID(userID string, id int) (*models.Income, error) {
	args := m.Called(userID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Income), args.Error(1)
}

func (m *MockIncomeService) GetAll(userID string, periodID *int) ([]models.Income, error) {
	args := m.Called(userID, periodID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Income), args.Error(1)
}

func (m *MockIncomeService) Update(userID string, income *models.Income) error {
	args := m.Called(userID, income)
	return args.Error(0)
}

func (m *MockIncomeService) Delete(userID string, id int) error {
	args := m.Called(userID, id)
	return args.Error(0)
}

func setupIncomeHandlerTest() (*IncomeHandler, *MockIncomeService, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockIncomeService)
	handler := NewIncomeHandler(mockService)
	router := gin.New()
	return handler, mockService, router
}

func TestIncomeHandler_Create_Success(t *testing.T) {
	handler, mockService, router := setupIncomeHandlerTest()
	_ = mockService

	router.POST("/incomes", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.Create(c)
	})

	income := models.Income{
		PeriodID:    1,
		Description: "Salary",
		Amount:      5000.00,
		IsRecurring: true,
		ReceivedAt:  "2024-06-01",
	}
	body, _ := json.Marshal(income)

	mockService.On("Create", "user123", mock.AnythingOfType("*models.Income")).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/incomes", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestIncomeHandler_Create_Unauthorized(t *testing.T) {
	handler, mockService, router := setupIncomeHandlerTest()
	_ = mockService

	router.POST("/incomes", func(c *gin.Context) {
		handler.Create(c)
	})

	income := models.Income{Description: "Salary", Amount: 5000.00}
	body, _ := json.Marshal(income)

	req := httptest.NewRequest(http.MethodPost, "/incomes", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestIncomeHandler_Create_InvalidJSON(t *testing.T) {
	handler, mockService, router := setupIncomeHandlerTest()
	_ = mockService

	router.POST("/incomes", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.Create(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/incomes", bytes.NewBuffer([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestIncomeHandler_Create_ValidationError(t *testing.T) {
	handler, mockService, router := setupIncomeHandlerTest()
	_ = mockService

	router.POST("/incomes", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.Create(c)
	})

	income := models.Income{Description: "Salary", Amount: 5000.00, PeriodID: 1}
	body, _ := json.Marshal(income)

	mockService.On("Create", "user123", mock.AnythingOfType("*models.Income")).Return(errors.New("period not found"))

	req := httptest.NewRequest(http.MethodPost, "/incomes", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}

func TestIncomeHandler_GetByID_Success(t *testing.T) {
	handler, mockService, router := setupIncomeHandlerTest()
	_ = mockService

	router.GET("/incomes/:id", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.GetByID(c)
	})

	expectedIncome := &models.Income{
		ID:          1,
		Description: "Salary",
		Amount:      5000.00,
	}
	mockService.On("GetByID", "user123", 1).Return(expectedIncome, nil)

	req := httptest.NewRequest(http.MethodGet, "/incomes/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestIncomeHandler_GetByID_Unauthorized(t *testing.T) {
	handler, mockService, router := setupIncomeHandlerTest()
	_ = mockService

	router.GET("/incomes/:id", func(c *gin.Context) {
		handler.GetByID(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/incomes/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestIncomeHandler_GetByID_InvalidID(t *testing.T) {
	handler, mockService, router := setupIncomeHandlerTest()
	_ = mockService

	router.GET("/incomes/:id", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.GetByID(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/incomes/invalid", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestIncomeHandler_GetByID_NotFound(t *testing.T) {
	handler, mockService, router := setupIncomeHandlerTest()
	_ = mockService

	router.GET("/incomes/:id", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.GetByID(c)
	})

	mockService.On("GetByID", "user123", 999).Return(nil, errors.New("income not found"))

	req := httptest.NewRequest(http.MethodGet, "/incomes/999", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

func TestIncomeHandler_GetAll_Success(t *testing.T) {
	handler, mockService, router := setupIncomeHandlerTest()
	_ = mockService

	router.GET("/incomes", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.GetAll(c)
	})

	expectedIncomes := []models.Income{
		{ID: 1, Description: "Salary"},
		{ID: 2, Description: "Bonus"},
	}
	mockService.On("GetAll", "user123", (*int)(nil)).Return(expectedIncomes, nil)

	req := httptest.NewRequest(http.MethodGet, "/incomes", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestIncomeHandler_GetAll_WithPeriodFilter(t *testing.T) {
	handler, mockService, router := setupIncomeHandlerTest()
	_ = mockService

	router.GET("/incomes", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.GetAll(c)
	})

	periodID := 1
	expectedIncomes := []models.Income{
		{ID: 1, PeriodID: 1, Description: "Salary"},
	}

	mockService.On("GetAll", "user123", &periodID).Return(expectedIncomes, nil)

	req := httptest.NewRequest(http.MethodGet, "/incomes?period_id=1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestIncomeHandler_GetAll_Unauthorized(t *testing.T) {
	handler, mockService, router := setupIncomeHandlerTest()
	_ = mockService

	router.GET("/incomes", func(c *gin.Context) {
		handler.GetAll(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/incomes", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestIncomeHandler_GetAll_ServiceError(t *testing.T) {
	handler, mockService, router := setupIncomeHandlerTest()
	_ = mockService

	router.GET("/incomes", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.GetAll(c)
	})

	mockService.On("GetAll", "user123", (*int)(nil)).Return(nil, errors.New("database error"))

	req := httptest.NewRequest(http.MethodGet, "/incomes", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

func TestIncomeHandler_Update_Success(t *testing.T) {
	handler, mockService, router := setupIncomeHandlerTest()
	_ = mockService

	router.PUT("/incomes/:id", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.Update(c)
	})

	income := models.Income{
		PeriodID:    1,
		Description: "Updated Income",
		Amount:      5500.00,
	}
	body, _ := json.Marshal(income)

	mockService.On("Update", "user123", mock.AnythingOfType("*models.Income")).Return(nil)

	req := httptest.NewRequest(http.MethodPut, "/incomes/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestIncomeHandler_Update_Unauthorized(t *testing.T) {
	handler, mockService, router := setupIncomeHandlerTest()
	_ = mockService

	router.PUT("/incomes/:id", func(c *gin.Context) {
		handler.Update(c)
	})

	income := models.Income{Description: "Updated"}
	body, _ := json.Marshal(income)

	req := httptest.NewRequest(http.MethodPut, "/incomes/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestIncomeHandler_Update_InvalidID(t *testing.T) {
	handler, mockService, router := setupIncomeHandlerTest()
	_ = mockService

	router.PUT("/incomes/:id", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.Update(c)
	})

	income := models.Income{Description: "Updated"}
	body, _ := json.Marshal(income)

	req := httptest.NewRequest(http.MethodPut, "/incomes/invalid", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestIncomeHandler_Update_InvalidJSON(t *testing.T) {
	handler, mockService, router := setupIncomeHandlerTest()
	_ = mockService

	router.PUT("/incomes/:id", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.Update(c)
	})

	req := httptest.NewRequest(http.MethodPut, "/incomes/1", bytes.NewBuffer([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestIncomeHandler_Update_ValidationError(t *testing.T) {
	handler, mockService, router := setupIncomeHandlerTest()
	_ = mockService

	router.PUT("/incomes/:id", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.Update(c)
	})

	income := models.Income{Description: "Updated", Amount: 5500.00, PeriodID: 1}
	body, _ := json.Marshal(income)

	mockService.On("Update", "user123", mock.AnythingOfType("*models.Income")).Return(errors.New("income not found"))

	req := httptest.NewRequest(http.MethodPut, "/incomes/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}

func TestIncomeHandler_Delete_Success(t *testing.T) {
	handler, mockService, router := setupIncomeHandlerTest()
	_ = mockService

	router.DELETE("/incomes/:id", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.Delete(c)
	})

	mockService.On("Delete", "user123", 1).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/incomes/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestIncomeHandler_Delete_Unauthorized(t *testing.T) {
	handler, mockService, router := setupIncomeHandlerTest()
	_ = mockService

	router.DELETE("/incomes/:id", func(c *gin.Context) {
		handler.Delete(c)
	})

	req := httptest.NewRequest(http.MethodDelete, "/incomes/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestIncomeHandler_Delete_InvalidID(t *testing.T) {
	handler, mockService, router := setupIncomeHandlerTest()
	_ = mockService

	router.DELETE("/incomes/:id", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.Delete(c)
	})

	req := httptest.NewRequest(http.MethodDelete, "/incomes/invalid", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestIncomeHandler_Delete_ValidationError(t *testing.T) {
	handler, mockService, router := setupIncomeHandlerTest()
	_ = mockService

	router.DELETE("/incomes/:id", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.Delete(c)
	})

	mockService.On("Delete", "user123", 1).Return(errors.New("income not found"))

	req := httptest.NewRequest(http.MethodDelete, "/incomes/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}
