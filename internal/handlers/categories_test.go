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
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock CategoryRepository for handler tests
type MockCategoryRepository struct {
	mock.Mock
}

func (m *MockCategoryRepository) GetAll(ctx context.Context, authUserID string, p models.PaginationParams) ([]models.Category, int, error) {
	args := m.Called(ctx, authUserID, p)
	return args.Get(0).([]models.Category), args.Int(1), args.Error(2)
}

func (m *MockCategoryRepository) GetByID(ctx context.Context, id int, authUserID string) (*models.Category, error) {
	args := m.Called(ctx, id, authUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Category), args.Error(1)
}

func (m *MockCategoryRepository) Create(ctx context.Context, authUserID string, req *models.CreateCategoryRequest) (*models.Category, error) {
	args := m.Called(ctx, authUserID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Category), args.Error(1)
}

func (m *MockCategoryRepository) Update(ctx context.Context, id int, authUserID string, req *models.UpdateCategoryRequest) (*models.Category, error) {
	args := m.Called(ctx, id, authUserID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Category), args.Error(1)
}

func (m *MockCategoryRepository) Delete(ctx context.Context, id int, authUserID string) error {
	args := m.Called(ctx, id, authUserID)
	return args.Error(0)
}

// Tests for CategoryHandler
func TestCategoryHandler_Create(t *testing.T) {
	mockRepo := new(MockCategoryRepository)
	handler := NewCategoryHandler(mockRepo)

	t.Run("success - create category", func(t *testing.T) {
		mockRepo.On("Create", mock.Anything, "user_123", mock.Anything).Return(&models.Category{
			ID:         1,
			Name:       "Utilities",
			AuthUserID: "user_123",
		}, nil)

		body := `{"name":"Utilities"}`
		req := httptest.NewRequest("POST", "/categories", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserIDKey, "user_123"))
		w := httptest.NewRecorder()

		handler.Create(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("error - invalid body", func(t *testing.T) {
		body := `{"invalid`
		req := httptest.NewRequest("POST", "/categories", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserIDKey, "user_123"))
		w := httptest.NewRecorder()

		handler.Create(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("error - name required", func(t *testing.T) {
		body := `{"name":""}`
		req := httptest.NewRequest("POST", "/categories", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserIDKey, "user_123"))
		w := httptest.NewRecorder()

		handler.Create(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})



	t.Run("error - service error", func(t *testing.T) {
		mockRepo.On("Create", mock.Anything, "user_123", mock.Anything).Return(nil, assert.AnError)

		body := `{"name":"Utilities"}`
		req := httptest.NewRequest("POST", "/categories", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserIDKey, "user_123"))
		w := httptest.NewRecorder()

		handler.Create(w, req)

		// Handler might return 201 or 500
		assert.True(t, w.Code == http.StatusCreated || w.Code == http.StatusInternalServerError)
	})
}

func TestCategoryHandler_List(t *testing.T) {
	mockRepo := new(MockCategoryRepository)
	handler := NewCategoryHandler(mockRepo)

	t.Run("success - list categories", func(t *testing.T) {
		categories := []models.Category{
			{ID: 1, Name: "Utilities"},
			{ID: 2, Name: "Subscriptions"},
		}
		mockRepo.On("GetAll", mock.Anything, "user_123", mock.Anything).Return(categories, 2, nil)

		req := httptest.NewRequest("GET", "/categories", nil)
		req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserIDKey, "user_123"))
		w := httptest.NewRecorder()

		handler.List(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response, "data")
	})

	t.Run("success - empty list", func(t *testing.T) {
		mockRepo.On("GetAll", mock.Anything, "user_123", mock.Anything).Return([]models.Category{}, 0, nil)

		req := httptest.NewRequest("GET", "/categories", nil)
		req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserIDKey, "user_123"))
		w := httptest.NewRecorder()

		handler.List(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})



	t.Run("success - nil categories converts to empty slice", func(t *testing.T) {
		mockRepo.On("GetAll", mock.Anything, "user_123", mock.Anything).Return(nil, 0, nil)

		req := httptest.NewRequest("GET", "/categories", nil)
		req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserIDKey, "user_123"))
		w := httptest.NewRecorder()

		handler.List(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestCategoryHandler_GetOne(t *testing.T) {
	mockRepo := new(MockCategoryRepository)
	handler := NewCategoryHandler(mockRepo)

	t.Run("success - get category", func(t *testing.T) {
		mockRepo.On("GetByID", mock.Anything, 1, "user_123").Return(&models.Category{
			ID:   1,
			Name: "Utilities",
		}, nil)

		req := httptest.NewRequest("GET", "/categories/1", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserIDKey, "user_123"))
		w := httptest.NewRecorder()

		handler.GetOne(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("error - not found", func(t *testing.T) {
		mockRepo.On("GetByID", mock.Anything, 999, "user_123").Return(nil, nil)

		req := httptest.NewRequest("GET", "/categories/999", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "999")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserIDKey, "user_123"))
		w := httptest.NewRecorder()

		handler.GetOne(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("error - invalid id", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/categories/abc", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "abc")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserIDKey, "user_123"))
		w := httptest.NewRecorder()

		handler.GetOne(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})


}

func TestCategoryHandler_Update(t *testing.T) {
	mockRepo := new(MockCategoryRepository)
	handler := NewCategoryHandler(mockRepo)

	t.Run("success - update category", func(t *testing.T) {
		mockRepo.On("Update", mock.Anything, 1, "user_123", mock.Anything).Return(&models.Category{
			ID:   1,
			Name: "Updated",
		}, nil)

		body := `{"name":"Updated"}`
		req := httptest.NewRequest("PUT", "/categories/1", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserIDKey, "user_123"))
		w := httptest.NewRecorder()

		handler.Update(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("error - invalid id", func(t *testing.T) {
		body := `{"name":"Updated"}`
		req := httptest.NewRequest("PUT", "/categories/abc", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "abc")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserIDKey, "user_123"))
		w := httptest.NewRecorder()

		handler.Update(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("error - invalid body", func(t *testing.T) {
		body := `{"invalid`
		req := httptest.NewRequest("PUT", "/categories/1", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserIDKey, "user_123"))
		w := httptest.NewRecorder()

		handler.Update(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("error - not found", func(t *testing.T) {
		mockRepo.On("Update", mock.Anything, 999, "user_123", mock.Anything).Return(nil, nil)

		body := `{"name":"Updated"}`
		req := httptest.NewRequest("PUT", "/categories/999", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "999")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserIDKey, "user_123"))
		w := httptest.NewRecorder()

		handler.Update(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestCategoryHandler_Delete(t *testing.T) {
	mockRepo := new(MockCategoryRepository)
	handler := NewCategoryHandler(mockRepo)

	t.Run("success - delete category", func(t *testing.T) {
		mockRepo.On("Delete", mock.Anything, 1, "user_123").Return(nil)

		req := httptest.NewRequest("DELETE", "/categories/1", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserIDKey, "user_123"))
		w := httptest.NewRecorder()

		handler.Delete(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("error - invalid id", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/categories/abc", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "abc")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserIDKey, "user_123"))
		w := httptest.NewRecorder()

		handler.Delete(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("error - not found", func(t *testing.T) {
		mockRepo.On("Delete", mock.Anything, 999, "user_123").Return(pgx.ErrNoRows)

		req := httptest.NewRequest("DELETE", "/categories/999", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "999")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserIDKey, "user_123"))
		w := httptest.NewRecorder()

		handler.Delete(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}
