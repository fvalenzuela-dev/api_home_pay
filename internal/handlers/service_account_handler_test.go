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

// MockServiceAccountService is a mock implementation of ServiceAccountService
type MockServiceAccountService struct {
	mock.Mock
}

func (m *MockServiceAccountService) Create(userID string, account *models.ServiceAccount) error {
	args := m.Called(userID, account)
	return args.Error(0)
}

func (m *MockServiceAccountService) GetByID(userID string, id int) (*models.ServiceAccount, error) {
	args := m.Called(userID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ServiceAccount), args.Error(1)
}

func (m *MockServiceAccountService) GetAll(userID string, companyID *int) ([]models.ServiceAccount, error) {
	args := m.Called(userID, companyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.ServiceAccount), args.Error(1)
}

func (m *MockServiceAccountService) Update(userID string, account *models.ServiceAccount) error {
	args := m.Called(userID, account)
	return args.Error(0)
}

func (m *MockServiceAccountService) Delete(userID string, id int) error {
	args := m.Called(userID, id)
	return args.Error(0)
}

func setupServiceAccountHandlerTest() (*ServiceAccountHandler, *MockServiceAccountService, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockServiceAccountService)
	handler := NewServiceAccountHandler(mockService)
	router := gin.New()
	return handler, mockService, router
}

func TestServiceAccountHandler_Create_Success(t *testing.T) {
	handler, mockService, router := setupServiceAccountHandlerTest()
	_ = mockService

	router.POST("/service-accounts", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.Create(c)
	})

	account := models.ServiceAccount{
		CompanyID:         1,
		AccountIdentifier: "ACC123456",
		Alias:             "My Account",
	}
	body, _ := json.Marshal(account)

	mockService.On("Create", "user123", mock.AnythingOfType("*models.ServiceAccount")).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/service-accounts", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestServiceAccountHandler_Create_Unauthorized(t *testing.T) {
	handler, mockService, router := setupServiceAccountHandlerTest()
	_ = mockService

	router.POST("/service-accounts", func(c *gin.Context) {
		handler.Create(c)
	})

	account := models.ServiceAccount{CompanyID: 1, AccountIdentifier: "ACC123456"}
	body, _ := json.Marshal(account)

	req := httptest.NewRequest(http.MethodPost, "/service-accounts", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCompanyHandler_Create_ServiceAccount_InvalidJSON(t *testing.T) {
	handler, mockService, router := setupServiceAccountHandlerTest()
	_ = mockService

	router.POST("/service-accounts", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.Create(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/service-accounts", bytes.NewBuffer([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestServiceAccountHandler_Create_ValidationError(t *testing.T) {
	handler, mockService, router := setupServiceAccountHandlerTest()
	_ = mockService

	router.POST("/service-accounts", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.Create(c)
	})

	account := models.ServiceAccount{CompanyID: 1, AccountIdentifier: "ACC123456"}
	body, _ := json.Marshal(account)

	mockService.On("Create", "user123", mock.AnythingOfType("*models.ServiceAccount")).Return(errors.New("account already exists"))

	req := httptest.NewRequest(http.MethodPost, "/service-accounts", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}

func TestServiceAccountHandler_GetByID_Success(t *testing.T) {
	handler, mockService, router := setupServiceAccountHandlerTest()
	_ = mockService

	router.GET("/service-accounts/:id", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.GetByID(c)
	})

	expectedAccount := &models.ServiceAccount{
		ID:                1,
		CompanyID:         1,
		AccountIdentifier: "ACC123456",
		Alias:             "My Account",
	}
	mockService.On("GetByID", "user123", 1).Return(expectedAccount, nil)

	req := httptest.NewRequest(http.MethodGet, "/service-accounts/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestServiceAccountHandler_GetByID_Unauthorized(t *testing.T) {
	handler, mockService, router := setupServiceAccountHandlerTest()
	_ = mockService

	router.GET("/service-accounts/:id", func(c *gin.Context) {
		handler.GetByID(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/service-accounts/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestServiceAccountHandler_GetByID_InvalidID(t *testing.T) {
	handler, mockService, router := setupServiceAccountHandlerTest()
	_ = mockService

	router.GET("/service-accounts/:id", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.GetByID(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/service-accounts/invalid", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestServiceAccountHandler_GetByID_NotFound(t *testing.T) {
	handler, mockService, router := setupServiceAccountHandlerTest()
	_ = mockService

	router.GET("/service-accounts/:id", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.GetByID(c)
	})

	mockService.On("GetByID", "user123", 999).Return(nil, errors.New("service account not found"))

	req := httptest.NewRequest(http.MethodGet, "/service-accounts/999", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

func TestServiceAccountHandler_GetAll_Success(t *testing.T) {
	handler, mockService, router := setupServiceAccountHandlerTest()
	_ = mockService

	router.GET("/service-accounts", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.GetAll(c)
	})

	expectedAccounts := []models.ServiceAccount{
		{ID: 1, AccountIdentifier: "ACC001"},
		{ID: 2, AccountIdentifier: "ACC002"},
	}
	mockService.On("GetAll", "user123", (*int)(nil)).Return(expectedAccounts, nil)

	req := httptest.NewRequest(http.MethodGet, "/service-accounts", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestServiceAccountHandler_GetAll_WithCompanyFilter(t *testing.T) {
	handler, mockService, router := setupServiceAccountHandlerTest()
	_ = mockService

	router.GET("/service-accounts", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.GetAll(c)
	})

	companyID := 1
	expectedAccounts := []models.ServiceAccount{
		{ID: 1, CompanyID: 1, AccountIdentifier: "ACC001"},
	}

	mockService.On("GetAll", "user123", &companyID).Return(expectedAccounts, nil)

	req := httptest.NewRequest(http.MethodGet, "/service-accounts?company_id=1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestServiceAccountHandler_GetAll_Unauthorized(t *testing.T) {
	handler, mockService, router := setupServiceAccountHandlerTest()
	_ = mockService

	router.GET("/service-accounts", func(c *gin.Context) {
		handler.GetAll(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/service-accounts", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestServiceAccountHandler_GetAll_ServiceError(t *testing.T) {
	handler, mockService, router := setupServiceAccountHandlerTest()
	_ = mockService

	router.GET("/service-accounts", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.GetAll(c)
	})

	mockService.On("GetAll", "user123", (*int)(nil)).Return(nil, errors.New("database error"))

	req := httptest.NewRequest(http.MethodGet, "/service-accounts", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

func TestServiceAccountHandler_Update_Success(t *testing.T) {
	handler, mockService, router := setupServiceAccountHandlerTest()
	_ = mockService

	router.PUT("/service-accounts/:id", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.Update(c)
	})

	account := models.ServiceAccount{
		CompanyID:         1,
		AccountIdentifier: "UPDATED123",
		Alias:             "Updated Account",
	}
	body, _ := json.Marshal(account)

	mockService.On("Update", "user123", mock.AnythingOfType("*models.ServiceAccount")).Return(nil)

	req := httptest.NewRequest(http.MethodPut, "/service-accounts/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestServiceAccountHandler_Update_Unauthorized(t *testing.T) {
	handler, mockService, router := setupServiceAccountHandlerTest()
	_ = mockService

	router.PUT("/service-accounts/:id", func(c *gin.Context) {
		handler.Update(c)
	})

	account := models.ServiceAccount{CompanyID: 1, AccountIdentifier: "UPDATED123"}
	body, _ := json.Marshal(account)

	req := httptest.NewRequest(http.MethodPut, "/service-accounts/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestServiceAccountHandler_Update_InvalidID(t *testing.T) {
	handler, mockService, router := setupServiceAccountHandlerTest()
	_ = mockService

	router.PUT("/service-accounts/:id", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.Update(c)
	})

	account := models.ServiceAccount{CompanyID: 1, AccountIdentifier: "UPDATED123"}
	body, _ := json.Marshal(account)

	req := httptest.NewRequest(http.MethodPut, "/service-accounts/invalid", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestServiceAccountHandler_Update_InvalidJSON(t *testing.T) {
	handler, mockService, router := setupServiceAccountHandlerTest()
	_ = mockService

	router.PUT("/service-accounts/:id", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.Update(c)
	})

	req := httptest.NewRequest(http.MethodPut, "/service-accounts/1", bytes.NewBuffer([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestServiceAccountHandler_Update_ValidationError(t *testing.T) {
	handler, mockService, router := setupServiceAccountHandlerTest()
	_ = mockService

	router.PUT("/service-accounts/:id", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.Update(c)
	})

	account := models.ServiceAccount{CompanyID: 1, AccountIdentifier: "UPDATED123"}
	body, _ := json.Marshal(account)

	mockService.On("Update", "user123", mock.AnythingOfType("*models.ServiceAccount")).Return(errors.New("account already exists"))

	req := httptest.NewRequest(http.MethodPut, "/service-accounts/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}

func TestServiceAccountHandler_Delete_Success(t *testing.T) {
	handler, mockService, router := setupServiceAccountHandlerTest()
	_ = mockService

	router.DELETE("/service-accounts/:id", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.Delete(c)
	})

	mockService.On("Delete", "user123", 1).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/service-accounts/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestServiceAccountHandler_Delete_Unauthorized(t *testing.T) {
	handler, mockService, router := setupServiceAccountHandlerTest()
	_ = mockService

	router.DELETE("/service-accounts/:id", func(c *gin.Context) {
		handler.Delete(c)
	})

	req := httptest.NewRequest(http.MethodDelete, "/service-accounts/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestServiceAccountHandler_Delete_InvalidID(t *testing.T) {
	handler, mockService, router := setupServiceAccountHandlerTest()
	_ = mockService

	router.DELETE("/service-accounts/:id", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.Delete(c)
	})

	req := httptest.NewRequest(http.MethodDelete, "/service-accounts/invalid", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestServiceAccountHandler_Delete_ValidationError(t *testing.T) {
	handler, mockService, router := setupServiceAccountHandlerTest()
	_ = mockService

	router.DELETE("/service-accounts/:id", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.Delete(c)
	})

	mockService.On("Delete", "user123", 1).Return(errors.New("cannot delete service account with associated expenses"))

	req := httptest.NewRequest(http.MethodDelete, "/service-accounts/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}
