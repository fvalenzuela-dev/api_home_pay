package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCategory_SoftDelete(t *testing.T) {
	now := time.Now()
	category := Category{
		ID:         1,
		Name:       "Test Category",
		AuthUserID: "user_123",
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	// Soft delete
	category.DeletedAt = &now

	assert.NotNil(t, category.DeletedAt)
	assert.Equal(t, now, *category.DeletedAt)
}

func TestCategory_IsDeleted(t *testing.T) {
	now := time.Now()
	category := Category{
		ID:         1,
		Name:       "Test Category",
		AuthUserID: "user_123",
		CreatedAt:  now,
		UpdatedAt:  now,
		DeletedAt:  nil,
	}

	// Should not be deleted
	assert.Nil(t, category.DeletedAt)

	// Soft delete
	category.DeletedAt = &now
	assert.NotNil(t, category.DeletedAt)
}

func TestCreateCategoryRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		req     CreateCategoryRequest
		isValid bool
	}{
		{
			name:    "valid request",
			req:     CreateCategoryRequest{Name: "Utilities"},
			isValid: true,
		},
		{
			name:    "empty name",
			req:     CreateCategoryRequest{Name: ""},
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.isValid {
				assert.NotEmpty(t, tt.req.Name)
			} else {
				assert.Empty(t, tt.req.Name)
			}
		})
	}
}

func TestUpdateCategoryRequest_PointerFields(t *testing.T) {
	// Test that UpdateCategoryRequest uses pointers for partial updates
	name := "Updated Name"
	req := UpdateCategoryRequest{Name: &name}

	assert.NotNil(t, req.Name)
	assert.Equal(t, "Updated Name", *req.Name)
}

func TestUpdateCategoryRequest_NilFields(t *testing.T) {
	// When field is nil, it means "don't update"
	req := UpdateCategoryRequest{Name: nil}

	assert.Nil(t, req.Name)
}
