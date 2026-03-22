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

// MockCompanyService is a mock implementation of CompanyService
type MockCompanyService struct {
	mock.Mock
}

func (m *MockCompanyService) Create(userID string, company *models.Company) error {
	args := m.Called(userID, company)
	return args.Error(0)
}

func (m *MockCompanyService) GetByID(userID string, id int) (*models.Company, error) {
	args := m.Called(userID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Company), args.Error(1)
}

func (m *MockCompanyService) GetAll(userID string) ([]models.Company, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Company), args.Error(1)
}

func (m *MockCompanyService) Update(userID string, company *models.Company) error {
	args := m.Called(userID, company)
	return args.Error(0)
}

func (m *MockCompanyService) Delete(userID string, id int) error {
	args := m.Called(userID, id)
	return args.Error(0)
}

func setupCompanyHandlerTest() (*CompanyHandler, *MockCompanyService, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockCompanyService)
	handler := NewCompanyHandler(mockService)
	router := gin.New()
	return handler, mockService, router
}

func TestCompanyHandler_Create_Success(t *testing.T) {
	handler, mockService, router := setupCompanyHandlerTest()
	_ = mockService

	router.POST("/companies", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.Create(c)
	})

	company := models.Company{Name: "Acme Corp", WebsiteURL: "https://acme.com"}
	body, _ := json.Marshal(company)

	mockService.On("Create", "user123", mock.AnythingOfType("*models.Company")).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/companies", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestCompanyHandler_Create_Unauthorized(t *testing.T) {
	handler, mockService, router := setupCompanyHandlerTest()
	_ = mockService

	router.POST("/companies", func(c *gin.Context) {
		handler.Create(c)
	})

	company := models.Company{Name: "Acme Corp"}
	body, _ := json.Marshal(company)

	req := httptest.NewRequest(http.MethodPost, "/companies", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCompanyHandler_Create_InvalidJSON(t *testing.T) {
	handler, mockService, router := setupCompanyHandlerTest()
	_ = mockService

	router.POST("/companies", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.Create(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/companies", bytes.NewBuffer([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCompanyHandler_Create_ValidationError(t *testing.T) {
	handler, mockService, router := setupCompanyHandlerTest()
	_ = mockService

	router.POST("/companies", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.Create(c)
	})

	company := models.Company{Name: "Acme Corp"}
	body, _ := json.Marshal(company)

	mockService.On("Create", "user123", mock.AnythingOfType("*models.Company")).Return(errors.New("company already exists"))

	req := httptest.NewRequest(http.MethodPost, "/companies", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}

func TestCompanyHandler_GetByID_Success(t *testing.T) {
	handler, mockService, router := setupCompanyHandlerTest()
	_ = mockService

	router.GET("/companies/:id", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.GetByID(c)
	})

	expectedCompany := &models.Company{ID: 1, Name: "Acme Corp", WebsiteURL: "https://acme.com"}
	mockService.On("GetByID", "user123", 1).Return(expectedCompany, nil)

	req := httptest.NewRequest(http.MethodGet, "/companies/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestCompanyHandler_GetByID_Unauthorized(t *testing.T) {
	handler, mockService, router := setupCompanyHandlerTest()
	_ = mockService

	router.GET("/companies/:id", func(c *gin.Context) {
		handler.GetByID(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/companies/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCompanyHandler_GetByID_InvalidID(t *testing.T) {
	handler, mockService, router := setupCompanyHandlerTest()
	_ = mockService

	router.GET("/companies/:id", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.GetByID(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/companies/invalid", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCompanyHandler_GetByID_NotFound(t *testing.T) {
	handler, mockService, router := setupCompanyHandlerTest()
	_ = mockService

	router.GET("/companies/:id", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.GetByID(c)
	})

	mockService.On("GetByID", "user123", 999).Return(nil, errors.New("company not found"))

	req := httptest.NewRequest(http.MethodGet, "/companies/999", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

func TestCompanyHandler_GetAll_Success(t *testing.T) {
	handler, mockService, router := setupCompanyHandlerTest()
	_ = mockService

	router.GET("/companies", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.GetAll(c)
	})

	expectedCompanies := []models.Company{
		{ID: 1, Name: "Acme Corp"},
		{ID: 2, Name: "Tech Inc"},
	}
	mockService.On("GetAll", "user123").Return(expectedCompanies, nil)

	req := httptest.NewRequest(http.MethodGet, "/companies", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestCompanyHandler_GetAll_Unauthorized(t *testing.T) {
	handler, mockService, router := setupCompanyHandlerTest()
	_ = mockService

	router.GET("/companies", func(c *gin.Context) {
		handler.GetAll(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/companies", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCompanyHandler_GetAll_ServiceError(t *testing.T) {
	handler, mockService, router := setupCompanyHandlerTest()
	_ = mockService

	router.GET("/companies", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.GetAll(c)
	})

	mockService.On("GetAll", "user123").Return(nil, errors.New("database error"))

	req := httptest.NewRequest(http.MethodGet, "/companies", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

func TestCompanyHandler_Update_Success(t *testing.T) {
	handler, mockService, router := setupCompanyHandlerTest()
	_ = mockService

	router.PUT("/companies/:id", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.Update(c)
	})

	company := models.Company{Name: "Updated Corp", WebsiteURL: "https://updated.com"}
	body, _ := json.Marshal(company)

	mockService.On("Update", "user123", mock.AnythingOfType("*models.Company")).Return(nil)

	req := httptest.NewRequest(http.MethodPut, "/companies/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestCompanyHandler_Update_Unauthorized(t *testing.T) {
	handler, mockService, router := setupCompanyHandlerTest()
	_ = mockService

	router.PUT("/companies/:id", func(c *gin.Context) {
		handler.Update(c)
	})

	company := models.Company{Name: "Updated Corp"}
	body, _ := json.Marshal(company)

	req := httptest.NewRequest(http.MethodPut, "/companies/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCompanyHandler_Update_InvalidID(t *testing.T) {
	handler, mockService, router := setupCompanyHandlerTest()
	_ = mockService

	router.PUT("/companies/:id", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.Update(c)
	})

	company := models.Company{Name: "Updated Corp"}
	body, _ := json.Marshal(company)

	req := httptest.NewRequest(http.MethodPut, "/companies/invalid", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCompanyHandler_Update_InvalidJSON(t *testing.T) {
	handler, mockService, router := setupCompanyHandlerTest()
	_ = mockService

	router.PUT("/companies/:id", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.Update(c)
	})

	req := httptest.NewRequest(http.MethodPut, "/companies/1", bytes.NewBuffer([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCompanyHandler_Update_ValidationError(t *testing.T) {
	handler, mockService, router := setupCompanyHandlerTest()
	_ = mockService

	router.PUT("/companies/:id", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.Update(c)
	})

	company := models.Company{Name: "Updated Corp"}
	body, _ := json.Marshal(company)

	mockService.On("Update", "user123", mock.AnythingOfType("*models.Company")).Return(errors.New("company already exists"))

	req := httptest.NewRequest(http.MethodPut, "/companies/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}

func TestCompanyHandler_Delete_Success(t *testing.T) {
	handler, mockService, router := setupCompanyHandlerTest()
	_ = mockService

	router.DELETE("/companies/:id", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.Delete(c)
	})

	mockService.On("Delete", "user123", 1).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/companies/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestCompanyHandler_Delete_Unauthorized(t *testing.T) {
	handler, mockService, router := setupCompanyHandlerTest()
	_ = mockService

	router.DELETE("/companies/:id", func(c *gin.Context) {
		handler.Delete(c)
	})

	req := httptest.NewRequest(http.MethodDelete, "/companies/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCompanyHandler_Delete_InvalidID(t *testing.T) {
	handler, mockService, router := setupCompanyHandlerTest()
	_ = mockService

	router.DELETE("/companies/:id", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.Delete(c)
	})

	req := httptest.NewRequest(http.MethodDelete, "/companies/invalid", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCompanyHandler_Delete_ValidationError(t *testing.T) {
	handler, mockService, router := setupCompanyHandlerTest()
	_ = mockService

	router.DELETE("/companies/:id", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.Delete(c)
	})

	mockService.On("Delete", "user123", 1).Return(errors.New("cannot delete company with associated service accounts"))

	req := httptest.NewRequest(http.MethodDelete, "/companies/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}
