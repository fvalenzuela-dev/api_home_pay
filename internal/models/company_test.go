package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCompany_Struct(t *testing.T) {
	now := time.Now()
	website := "https://example.com"
	phone := "+1234567890"

	company := Company{
		ID:         "company-123",
		AuthUserID: "user_123",
		CategoryID: 1,
		Name:       "Test Company",
		Website:    &website,
		Phone:      &phone,
		IsActive:   true,
		CreatedAt:  now,
	}

	assert.Equal(t, "company-123", company.ID)
	assert.Equal(t, "user_123", company.AuthUserID)
	assert.Equal(t, 1, company.CategoryID)
	assert.Equal(t, "Test Company", company.Name)
	assert.True(t, company.IsActive)
	assert.NotNil(t, company.Website)
	assert.NotNil(t, company.Phone)
	assert.Equal(t, "https://example.com", *company.Website)
	assert.Equal(t, "+1234567890", *company.Phone)
}

func TestCompany_OptionalFields(t *testing.T) {
	company := Company{
		ID:         "company-123",
		AuthUserID: "user_123",
		CategoryID: 1,
		Name:       "Test Company",
		Website:    nil,
		Phone:      nil,
		IsActive:   true,
	}

	// Optional fields should be nil
	assert.Nil(t, company.Website)
	assert.Nil(t, company.Phone)
}

func TestCompany_SoftDelete(t *testing.T) {
	now := time.Now()
	company := Company{
		ID:         "company-123",
		AuthUserID: "user_123",
		Name:       "Test Company",
		IsActive:   true,
	}

	// Soft delete
	company.DeletedAt = &now

	assert.NotNil(t, company.DeletedAt)
	assert.True(t, company.DeletedAt.After(company.CreatedAt) || company.DeletedAt.Equal(company.CreatedAt))
}

func TestCreateCompanyRequest_Validation(t *testing.T) {
	website := "https://valid.com"

	tests := []struct {
		name    string
		req     CreateCompanyRequest
		isValid bool
	}{
		{
			name: "valid request with all fields",
			req: CreateCompanyRequest{
				Name:       "Company Name",
				CategoryID: 1,
				Website:    &website,
			},
			isValid: true,
		},
		{
			name: "valid request with only required fields",
			req: CreateCompanyRequest{
				Name:       "Company Name",
				CategoryID: 1,
			},
			isValid: true,
		},
		{
			name: "empty name",
			req: CreateCompanyRequest{
				Name:       "",
				CategoryID: 1,
			},
			isValid: false,
		},
		{
			name: "zero category ID",
			req: CreateCompanyRequest{
				Name:       "Company Name",
				CategoryID: 0,
			},
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.isValid {
				assert.NotEmpty(t, tt.req.Name)
				assert.Greater(t, tt.req.CategoryID, 0)
			} else {
				if tt.req.Name == "" {
					assert.Empty(t, tt.req.Name)
				}
				if tt.req.CategoryID == 0 {
					assert.Equal(t, 0, tt.req.CategoryID)
				}
			}
		})
	}
}

func TestUpdateCompanyRequest_PartialUpdate(t *testing.T) {
	// Test partial update with pointers
	newName := "New Company Name"
	newCategoryID := 5

	req := UpdateCompanyRequest{
		Name:       &newName,
		CategoryID: &newCategoryID,
	}

	assert.NotNil(t, req.Name)
	assert.NotNil(t, req.CategoryID)
	assert.Equal(t, "New Company Name", *req.Name)
	assert.Equal(t, 5, *req.CategoryID)
}

func TestUpdateCompanyRequest_NilMeansNoUpdate(t *testing.T) {
	req := UpdateCompanyRequest{
		Name:       nil,
		CategoryID: nil,
		Website:    nil,
		Phone:      nil,
	}

	// All nil means don't update any field
	assert.Nil(t, req.Name)
	assert.Nil(t, req.CategoryID)
	assert.Nil(t, req.Website)
	assert.Nil(t, req.Phone)
}
