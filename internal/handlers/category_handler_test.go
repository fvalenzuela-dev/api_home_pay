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

// MockCategoryService is a mock implementation of CategoryService
type MockCategoryService struct {
	mock.Mock
}

func (m *MockCategoryService) Create(userID string, category *models.Category) error {
	args := m.Called(userID, category)
	return args.Error(0)
}

func (m *MockCategoryService) GetByID(userID string, id int) (*models.Category, error) {
	args := m.Called(userID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Category), args.Error(1)
}

func (m *MockCategoryService) GetAll(userID string) ([]models.Category, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Category), args.Error(1)
}

func (m *MockCategoryService) Update(userID string, category *models.Category) error {
	args := m.Called(userID, category)
	return args.Error(0)
}

func (m *MockCategoryService) Delete(userID string, id int) error {
	args := m.Called(userID, id)
	return args.Error(0)
}

func setupCategoryHandlerTest() (*CategoryHandler, *MockCategoryService, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockCategoryService)
	handler := NewCategoryHandler(mockService)
	router := gin.New()
	return handler, mockService, router
}

func TestCategoryHandler_Create_Success(t *testing.T) {
	handler, mockService, router := setupCategoryHandlerTest()
	_ = mockService

	router.POST("/categories", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.Create(c)
	})

	category := models.Category{Name: "Groceries"}
	body, _ := json.Marshal(category)

	mockService.On("Create", "user123", mock.AnythingOfType("*models.Category")).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestCategoryHandler_Create_Unauthorized(t *testing.T) {
	handler, mockService, router := setupCategoryHandlerTest()
	_ = mockService

	router.POST("/categories", func(c *gin.Context) {
		// Don't set user_id - simulating unauthorized
		handler.Create(c)
	})

	category := models.Category{Name: "Groceries"}
	body, _ := json.Marshal(category)

	req := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCategoryHandler_Create_InvalidJSON(t *testing.T) {
	handler, mockService, router := setupCategoryHandlerTest()
	_ = mockService

	router.POST("/categories", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.Create(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCategoryHandler_Create_ValidationError(t *testing.T) {
	handler, mockService, router := setupCategoryHandlerTest()
	_ = mockService

	router.POST("/categories", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.Create(c)
	})

	category := models.Category{Name: ""}
	body, _ := json.Marshal(category)

	mockService.On("Create", "user123", mock.AnythingOfType("*models.Category")).Return(errors.New("name cannot be empty"))

	req := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}

func TestCategoryHandler_GetByID_Success(t *testing.T) {
	handler, mockService, router := setupCategoryHandlerTest()
	_ = mockService

	router.GET("/categories/:id", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.GetByID(c)
	})

	expectedCategory := &models.Category{ID: 1, Name: "Groceries"}
	mockService.On("GetByID", "user123", 1).Return(expectedCategory, nil)

	req := httptest.NewRequest(http.MethodGet, "/categories/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestCategoryHandler_GetByID_Unauthorized(t *testing.T) {
	handler, mockService, router := setupCategoryHandlerTest()
	_ = mockService

	router.GET("/categories/:id", func(c *gin.Context) {
		// Don't set user_id
		handler.GetByID(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/categories/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCategoryHandler_GetByID_InvalidID(t *testing.T) {
	handler, mockService, router := setupCategoryHandlerTest()
	_ = mockService

	router.GET("/categories/:id", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.GetByID(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/categories/invalid", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCategoryHandler_GetByID_NotFound(t *testing.T) {
	handler, mockService, router := setupCategoryHandlerTest()
	_ = mockService

	router.GET("/categories/:id", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.GetByID(c)
	})

	mockService.On("GetByID", "user123", 999).Return(nil, errors.New("category not found"))

	req := httptest.NewRequest(http.MethodGet, "/categories/999", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

func TestCategoryHandler_GetAll_Success(t *testing.T) {
	handler, mockService, router := setupCategoryHandlerTest()
	_ = mockService

	router.GET("/categories", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.GetAll(c)
	})

	expectedCategories := []models.Category{
		{ID: 1, Name: "Groceries"},
		{ID: 2, Name: "Utilities"},
	}
	mockService.On("GetAll", "user123").Return(expectedCategories, nil)

	req := httptest.NewRequest(http.MethodGet, "/categories", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestCategoryHandler_GetAll_Unauthorized(t *testing.T) {
	handler, mockService, router := setupCategoryHandlerTest()
	_ = mockService

	router.GET("/categories", func(c *gin.Context) {
		// Don't set user_id
		handler.GetAll(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/categories", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCategoryHandler_GetAll_ServiceError(t *testing.T) {
	handler, mockService, router := setupCategoryHandlerTest()
	_ = mockService

	router.GET("/categories", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.GetAll(c)
	})

	mockService.On("GetAll", "user123").Return(nil, errors.New("database error"))

	req := httptest.NewRequest(http.MethodGet, "/categories", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

func TestCategoryHandler_Update_Success(t *testing.T) {
	handler, mockService, router := setupCategoryHandlerTest()
	_ = mockService

	router.PUT("/categories/:id", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.Update(c)
	})

	category := models.Category{Name: "Updated Category"}
	body, _ := json.Marshal(category)

	mockService.On("Update", "user123", mock.AnythingOfType("*models.Category")).Return(nil)

	req := httptest.NewRequest(http.MethodPut, "/categories/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestCategoryHandler_Update_Unauthorized(t *testing.T) {
	handler, mockService, router := setupCategoryHandlerTest()
	_ = mockService

	router.PUT("/categories/:id", func(c *gin.Context) {
		// Don't set user_id
		handler.Update(c)
	})

	category := models.Category{Name: "Updated"}
	body, _ := json.Marshal(category)

	req := httptest.NewRequest(http.MethodPut, "/categories/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCategoryHandler_Update_InvalidID(t *testing.T) {
	handler, mockService, router := setupCategoryHandlerTest()
	_ = mockService

	router.PUT("/categories/:id", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.Update(c)
	})

	category := models.Category{Name: "Updated"}
	body, _ := json.Marshal(category)

	req := httptest.NewRequest(http.MethodPut, "/categories/invalid", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCategoryHandler_Update_InvalidJSON(t *testing.T) {
	handler, mockService, router := setupCategoryHandlerTest()
	_ = mockService

	router.PUT("/categories/:id", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.Update(c)
	})

	req := httptest.NewRequest(http.MethodPut, "/categories/1", bytes.NewBuffer([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCategoryHandler_Update_ValidationError(t *testing.T) {
	handler, mockService, router := setupCategoryHandlerTest()
	_ = mockService

	router.PUT("/categories/:id", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.Update(c)
	})

	category := models.Category{Name: ""}
	body, _ := json.Marshal(category)

	mockService.On("Update", "user123", mock.AnythingOfType("*models.Category")).Return(errors.New("name cannot be empty"))

	req := httptest.NewRequest(http.MethodPut, "/categories/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}

func TestCategoryHandler_Delete_Success(t *testing.T) {
	handler, mockService, router := setupCategoryHandlerTest()
	_ = mockService

	router.DELETE("/categories/:id", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.Delete(c)
	})

	mockService.On("Delete", "user123", 1).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/categories/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestCategoryHandler_Delete_Unauthorized(t *testing.T) {
	handler, mockService, router := setupCategoryHandlerTest()
	_ = mockService

	router.DELETE("/categories/:id", func(c *gin.Context) {
		// Don't set user_id
		handler.Delete(c)
	})

	req := httptest.NewRequest(http.MethodDelete, "/categories/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCategoryHandler_Delete_InvalidID(t *testing.T) {
	handler, mockService, router := setupCategoryHandlerTest()
	_ = mockService

	router.DELETE("/categories/:id", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.Delete(c)
	})

	req := httptest.NewRequest(http.MethodDelete, "/categories/invalid", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCategoryHandler_Delete_ValidationError(t *testing.T) {
	handler, mockService, router := setupCategoryHandlerTest()
	_ = mockService

	router.DELETE("/categories/:id", func(c *gin.Context) {
		c.Set("user_id", "user123")
		handler.Delete(c)
	})

	mockService.On("Delete", "user123", 1).Return(errors.New("cannot delete category with associated expenses"))

	req := httptest.NewRequest(http.MethodDelete, "/categories/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}
