package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/homepay/api/internal/middleware"
	"github.com/homepay/api/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock CompanyService
type MockCompanyService struct {
	mock.Mock
}

func (m *MockCompanyService) Create(ctx context.Context, authUserID string, req *models.CreateCompanyRequest) (*models.Company, error) {
	args := m.Called(ctx, authUserID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Company), args.Error(1)
}

func (m *MockCompanyService) GetByID(ctx context.Context, id, authUserID string) (*models.Company, error) {
	args := m.Called(ctx, id, authUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Company), args.Error(1)
}

func (m *MockCompanyService) GetAll(ctx context.Context, authUserID string, p models.PaginationParams) ([]models.Company, int, error) {
	args := m.Called(ctx, authUserID, p)
	return args.Get(0).([]models.Company), args.Int(1), args.Error(2)
}

func (m *MockCompanyService) Update(ctx context.Context, id, authUserID string, req *models.UpdateCompanyRequest) (*models.Company, error) {
	args := m.Called(ctx, id, authUserID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Company), args.Error(1)
}

func (m *MockCompanyService) Delete(ctx context.Context, id, authUserID string) error {
	args := m.Called(ctx, id, authUserID)
	return args.Error(0)
}

// Helper to create test request with auth context
func createTestRequest(t *testing.T, method, path, body string, authUserID string) *http.Request {
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, path, bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserIDKey, authUserID))
	return req
}

// Tests for CompanyHandler.Create
func TestCompanyHandler_Create(t *testing.T) {
	mockSvc := new(MockCompanyService)
	handler := NewCompanyHandler(mockSvc)

	t.Run("success - create company", func(t *testing.T) {
		mockSvc.On("Create", mock.Anything, "user_123", mock.Anything).Return(&models.Company{
			ID:         "company-123",
			AuthUserID: "user_123",
			Name:       "Test Company",
			CategoryID: 1,
		}, nil)

		body := `{"name":"Test Company","category_id":1}`
		req := createTestRequest(t, "POST", "/companies", body, "user_123")
		w := httptest.NewRecorder()

		handler.Create(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("error - invalid body", func(t *testing.T) {
		body := `{"invalid`
		req := createTestRequest(t, "POST", "/companies", body, "user_123")
		w := httptest.NewRecorder()

		handler.Create(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// Tests for CompanyHandler.List
func TestCompanyHandler_List(t *testing.T) {
	mockSvc := new(MockCompanyService)
	handler := NewCompanyHandler(mockSvc)

	t.Run("success - list companies", func(t *testing.T) {
		companies := []models.Company{
			{ID: "company-1", Name: "Company 1"},
			{ID: "company-2", Name: "Company 2"},
		}
		mockSvc.On("GetAll", mock.Anything, "user_123", mock.Anything).Return(companies, 2, nil)

		req := createTestRequest(t, "GET", "/companies", "", "user_123")
		w := httptest.NewRecorder()

		handler.List(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response, "data")
	})

	t.Run("success - empty list", func(t *testing.T) {
		mockSvc.On("GetAll", mock.Anything, "user_123", mock.Anything).Return([]models.Company{}, 0, nil)

		req := createTestRequest(t, "GET", "/companies", "", "user_123")
		w := httptest.NewRecorder()

		handler.List(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// Tests for helpers
func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()

	writeJSON(w, http.StatusOK, map[string]string{"test": "value"})

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
}

func TestWriteError(t *testing.T) {
	w := httptest.NewRecorder()

	writeError(w, http.StatusBadRequest, "test error")

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "test error")
}

// Tests for decode helper
func TestDecode(t *testing.T) {
	t.Run("valid body", func(t *testing.T) {
		body := `{"name":"test","category_id":1}`
		req := httptest.NewRequest("POST", "/", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")

		var result struct {
			Name       string `json:"name"`
			CategoryID int    `json:"category_id"`
		}

		err := decode(req, &result)

		assert.NoError(t, err)
		assert.Equal(t, "test", result.Name)
		assert.Equal(t, 1, result.CategoryID)
	})

	t.Run("invalid body", func(t *testing.T) {
		body := `{"invalid`
		req := httptest.NewRequest("POST", "/", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")

		var result struct {
			Name string `json:"name"`
		}

		err := decode(req, &result)

		assert.Error(t, err)
	})

	t.Run("wrong content type", func(t *testing.T) {
		body := "name=test"
		req := httptest.NewRequest("POST", "/", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		var result struct {
			Name string `json:"name"`
		}

		err := decode(req, &result)

		assert.Error(t, err)
	})
}

// Tests for parsePagination
func TestParsePagination(t *testing.T) {
	tests := []struct {
		name          string
		query         string
		expectedPage  int
		expectedLimit int
	}{
		{
			name:          "default values",
			query:         "/?page=1&limit=20",
			expectedPage:  1,
			expectedLimit: 20,
		},
		{
			name:          "custom page",
			query:         "/?page=3",
			expectedPage:  3,
			expectedLimit: 20,
		},
		{
			name:          "custom limit",
			query:         "/?limit=50",
			expectedPage:  1,
			expectedLimit: 50,
		},
		{
			name:          "custom page and limit",
			query:         "/?page=2&limit=50",
			expectedPage:  2,
			expectedLimit: 50,
		},
		{
			name:          "page below 1",
			query:         "/?page=0",
			expectedPage:  1,
			expectedLimit: 20,
		},
		{
			name:          "limit above 100",
			query:         "/?limit=200",
			expectedPage:  1,
			expectedLimit: 20, // Default when out of range
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.query, nil)
			p := parsePagination(req)

			assert.Equal(t, tt.expectedPage, p.Page)
			assert.Equal(t, tt.expectedLimit, p.Limit)
		})
	}
}
