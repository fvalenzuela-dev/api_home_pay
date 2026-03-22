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

// MockPeriodService is a mock implementation of PeriodService
type MockPeriodService struct {
	mock.Mock
}

func (m *MockPeriodService) Create(userID string, period *models.Period) error {
	args := m.Called(userID, period)
	return args.Error(0)
}

func (m *MockPeriodService) GetByID(userID string, id int) (*models.Period, error) {
	args := m.Called(userID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Period), args.Error(1)
}

func (m *MockPeriodService) GetAll(userID string) ([]models.Period, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Period), args.Error(1)
}

func (m *MockPeriodService) Update(userID string, period *models.Period) error {
	args := m.Called(userID, period)
	return args.Error(0)
}

func (m *MockPeriodService) Delete(userID string, id int) error {
	args := m.Called(userID, id)
	return args.Error(0)
}

func setupPeriodHandlerTest() (*PeriodHandler, *MockPeriodService, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockPeriodService)
	handler := NewPeriodHandler(mockService)
	router := gin.New()
	return handler, mockService, router
}

func TestPeriodHandler_Create_Success(t *testing.T) {
	handler, mockService, router := setupPeriodHandlerTest()
	_ = mockService

	router.POST("/periods", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.Create(c)
	})

	period := models.Period{MonthNumber: 6, YearNumber: 2024}
	body, _ := json.Marshal(period)

	mockService.On("Create", "user123", mock.AnythingOfType("*models.Period")).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/periods", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestPeriodHandler_Create_Unauthorized(t *testing.T) {
	handler, mockService, router := setupPeriodHandlerTest()
	_ = mockService

	router.POST("/periods", func(c *gin.Context) {
		handler.Create(c)
	})

	period := models.Period{MonthNumber: 6, YearNumber: 2024}
	body, _ := json.Marshal(period)

	req := httptest.NewRequest(http.MethodPost, "/periods", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestPeriodHandler_Create_InvalidJSON(t *testing.T) {
	handler, mockService, router := setupPeriodHandlerTest()
	_ = mockService

	router.POST("/periods", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.Create(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/periods", bytes.NewBuffer([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPeriodHandler_Create_ValidationError(t *testing.T) {
	handler, mockService, router := setupPeriodHandlerTest()
	_ = mockService

	router.POST("/periods", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.Create(c)
	})

	period := models.Period{MonthNumber: 6, YearNumber: 2024}
	body, _ := json.Marshal(period)

	mockService.On("Create", "user123", mock.AnythingOfType("*models.Period")).Return(errors.New("period already exists"))

	req := httptest.NewRequest(http.MethodPost, "/periods", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}

func TestPeriodHandler_GetByID_Success(t *testing.T) {
	handler, mockService, router := setupPeriodHandlerTest()
	_ = mockService

	router.GET("/periods/:id", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.GetByID(c)
	})

	expectedPeriod := &models.Period{ID: 1, MonthNumber: 6, YearNumber: 2024}
	mockService.On("GetByID", "user123", 1).Return(expectedPeriod, nil)

	req := httptest.NewRequest(http.MethodGet, "/periods/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestPeriodHandler_GetByID_Unauthorized(t *testing.T) {
	handler, mockService, router := setupPeriodHandlerTest()
	_ = mockService

	router.GET("/periods/:id", func(c *gin.Context) {
		handler.GetByID(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/periods/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestPeriodHandler_GetByID_InvalidID(t *testing.T) {
	handler, mockService, router := setupPeriodHandlerTest()
	_ = mockService

	router.GET("/periods/:id", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.GetByID(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/periods/invalid", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPeriodHandler_GetByID_NotFound(t *testing.T) {
	handler, mockService, router := setupPeriodHandlerTest()
	_ = mockService

	router.GET("/periods/:id", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.GetByID(c)
	})

	mockService.On("GetByID", "user123", 999).Return(nil, errors.New("period not found"))

	req := httptest.NewRequest(http.MethodGet, "/periods/999", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

func TestPeriodHandler_GetAll_Success(t *testing.T) {
	handler, mockService, router := setupPeriodHandlerTest()
	_ = mockService

	router.GET("/periods", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.GetAll(c)
	})

	expectedPeriods := []models.Period{
		{ID: 1, MonthNumber: 1, YearNumber: 2024},
		{ID: 2, MonthNumber: 2, YearNumber: 2024},
	}
	mockService.On("GetAll", "user123").Return(expectedPeriods, nil)

	req := httptest.NewRequest(http.MethodGet, "/periods", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestPeriodHandler_GetAll_Unauthorized(t *testing.T) {
	handler, mockService, router := setupPeriodHandlerTest()
	_ = mockService

	router.GET("/periods", func(c *gin.Context) {
		handler.GetAll(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/periods", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestPeriodHandler_GetAll_ServiceError(t *testing.T) {
	handler, mockService, router := setupPeriodHandlerTest()
	_ = mockService

	router.GET("/periods", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.GetAll(c)
	})

	mockService.On("GetAll", "user123").Return(nil, errors.New("database error"))

	req := httptest.NewRequest(http.MethodGet, "/periods", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

func TestPeriodHandler_Update_Success(t *testing.T) {
	handler, mockService, router := setupPeriodHandlerTest()
	_ = mockService

	router.PUT("/periods/:id", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.Update(c)
	})

	period := models.Period{MonthNumber: 7, YearNumber: 2024}
	body, _ := json.Marshal(period)

	mockService.On("Update", "user123", mock.AnythingOfType("*models.Period")).Return(nil)

	req := httptest.NewRequest(http.MethodPut, "/periods/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestPeriodHandler_Update_Unauthorized(t *testing.T) {
	handler, mockService, router := setupPeriodHandlerTest()
	_ = mockService

	router.PUT("/periods/:id", func(c *gin.Context) {
		handler.Update(c)
	})

	period := models.Period{MonthNumber: 7, YearNumber: 2024}
	body, _ := json.Marshal(period)

	req := httptest.NewRequest(http.MethodPut, "/periods/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestPeriodHandler_Update_InvalidID(t *testing.T) {
	handler, mockService, router := setupPeriodHandlerTest()
	_ = mockService

	router.PUT("/periods/:id", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.Update(c)
	})

	period := models.Period{MonthNumber: 7, YearNumber: 2024}
	body, _ := json.Marshal(period)

	req := httptest.NewRequest(http.MethodPut, "/periods/invalid", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPeriodHandler_Update_InvalidJSON(t *testing.T) {
	handler, mockService, router := setupPeriodHandlerTest()
	_ = mockService

	router.PUT("/periods/:id", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.Update(c)
	})

	req := httptest.NewRequest(http.MethodPut, "/periods/1", bytes.NewBuffer([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPeriodHandler_Update_ValidationError(t *testing.T) {
	handler, mockService, router := setupPeriodHandlerTest()
	_ = mockService

	router.PUT("/periods/:id", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.Update(c)
	})

	period := models.Period{MonthNumber: 7, YearNumber: 2024}
	body, _ := json.Marshal(period)

	mockService.On("Update", "user123", mock.AnythingOfType("*models.Period")).Return(errors.New("period already exists"))

	req := httptest.NewRequest(http.MethodPut, "/periods/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}

func TestPeriodHandler_Delete_Success(t *testing.T) {
	handler, mockService, router := setupPeriodHandlerTest()
	_ = mockService

	router.DELETE("/periods/:id", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.Delete(c)
	})

	mockService.On("Delete", "user123", 1).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/periods/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestPeriodHandler_Delete_Unauthorized(t *testing.T) {
	handler, mockService, router := setupPeriodHandlerTest()
	_ = mockService

	router.DELETE("/periods/:id", func(c *gin.Context) {
		handler.Delete(c)
	})

	req := httptest.NewRequest(http.MethodDelete, "/periods/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestPeriodHandler_Delete_InvalidID(t *testing.T) {
	handler, mockService, router := setupPeriodHandlerTest()
	_ = mockService

	router.DELETE("/periods/:id", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.Delete(c)
	})

	req := httptest.NewRequest(http.MethodDelete, "/periods/invalid", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPeriodHandler_Delete_ValidationError(t *testing.T) {
	handler, mockService, router := setupPeriodHandlerTest()
	_ = mockService

	router.DELETE("/periods/:id", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.Delete(c)
	})

	mockService.On("Delete", "user123", 1).Return(errors.New("cannot delete period with associated expenses or incomes"))

	req := httptest.NewRequest(http.MethodDelete, "/periods/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}
