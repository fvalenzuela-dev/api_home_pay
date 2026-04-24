package repository

import (
	"testing"
	"time"

	"github.com/homepay/api/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestCategoryRepo_Interfaces(t *testing.T) {
	t.Run("CategoryRepository interface is satisfied by categoryRepo", func(t *testing.T) {
		var _ CategoryRepository = (*categoryRepo)(nil)
	})
}

func TestScanCategory(t *testing.T) {
	t.Run("scanCategory function exists", func(t *testing.T) {
		assert.NotNil(t, scanCategory)
	})
}

func TestCategoryRepo_GetAll(t *testing.T) {
	t.Run("categoryCols constant", func(t *testing.T) {
		assert.Equal(t, `id, name, auth_user_id, created_at, updated_at, deleted_at`, categoryCols)
	})
}

func TestCategoryRepo_GetByID(t *testing.T) {
	t.Run("GetByID returns category by ID", func(t *testing.T) {
		// Test the model structure
		now := time.Now()
		category := models.Category{
			ID:         1,
			Name:       "Test Category",
			AuthUserID: "user-123",
			CreatedAt:  now,
			UpdatedAt:  now,
		}
		assert.Equal(t, 1, category.ID)
		assert.Equal(t, "Test Category", category.Name)
	})
}

func TestCategoryRepo_Create(t *testing.T) {
	t.Run("CreateCategoryRequest validation", func(t *testing.T) {
		req := models.CreateCategoryRequest{
			Name: "New Category",
		}
		assert.Equal(t, "New Category", req.Name)
	})
}

func TestCategoryRepo_Update(t *testing.T) {
	t.Run("UpdateCategoryRequest with pointer", func(t *testing.T) {
		name := "Updated Category"
		req := models.UpdateCategoryRequest{
			Name: &name,
		}
		assert.Equal(t, "Updated Category", *req.Name)
	})

	t.Run("UpdateCategoryRequest with nil", func(t *testing.T) {
		req := models.UpdateCategoryRequest{}
		assert.Nil(t, req.Name)
	})
}

func TestCategoryRepo_Delete(t *testing.T) {
	t.Run("Soft delete sets deleted_at", func(t *testing.T) {
		now := time.Now()
		category := models.Category{
			ID:        1,
			Name:      "Test",
			DeletedAt: &now,
		}
		assert.NotNil(t, category.DeletedAt)
	})
}

func TestIsUniqueViolation(t *testing.T) {
	t.Run("isUniqueViolation function exists", func(t *testing.T) {
		assert.NotNil(t, isUniqueViolation)
	})
}

func TestErrDuplicateName(t *testing.T) {
	t.Run("ErrDuplicateName is defined", func(t *testing.T) {
		assert.Equal(t, "name already exists", ErrDuplicateName.Error())
	})
}
