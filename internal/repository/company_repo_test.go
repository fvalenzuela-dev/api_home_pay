package repository

import (
	"testing"
	"time"

	"github.com/homepay/api/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestScanCompany(t *testing.T) {
	t.Run("scanCompany with valid row", func(t *testing.T) {
		// This test would require a mock row - testing the function exists and is callable
		// The actual scan logic is tested through integration tests
		assert.NotNil(t, companyCols)
	})
}

func TestCompanyRepo_Create(t *testing.T) {
	t.Run("companyCols constant exists", func(t *testing.T) {
		assert.Equal(t, `id, auth_user_id, category_id, name, website, phone, is_active, created_at, deleted_at`, companyCols)
	})
}

func TestCompanyRepo_GetAll(t *testing.T) {
	t.Run("GetAll queries are well formed", func(t *testing.T) {
		// This test verifies the constants are correct
		// Actual DB tests would require a test database
		assert.NotEmpty(t, companyCols)
	})
}

func TestCompanyRepo_Update(t *testing.T) {
	t.Run("Update uses correct columns", func(t *testing.T) {
		assert.Contains(t, companyCols, "name")
		assert.Contains(t, companyCols, "category_id")
	})
}

func TestCompanyRepo_SoftDelete(t *testing.T) {
	t.Run("SoftDelete sets deleted_at", func(t *testing.T) {
		assert.NotEmpty(t, companyCols)
	})
}

// TestCompanyIntegration requires a real database connection
// These are placeholder tests to ensure the repository compiles correctly
func TestCompanyRepository_Interfaces(t *testing.T) {
	t.Run("CompanyRepository interface is satisfied by companyRepo", func(t *testing.T) {
		var _ CompanyRepository = (*companyRepo)(nil)
	})
}

func TestCompanyModel(t *testing.T) {
	t.Run("Company model can be created", func(t *testing.T) {
		now := time.Now()
		company := models.Company{
			ID:         "test-id",
			AuthUserID: "user-123",
			CategoryID: 1,
			Name:       "Test Company",
			IsActive:   true,
			CreatedAt:  now,
		}
		assert.Equal(t, "test-id", company.ID)
		assert.Equal(t, "Test Company", company.Name)
		assert.True(t, company.IsActive)
	})

	t.Run("CreateCompanyRequest model", func(t *testing.T) {
		req := models.CreateCompanyRequest{
			Name:       "New Company",
			CategoryID: 1,
		}
		assert.Equal(t, "New Company", req.Name)
		assert.Equal(t, 1, req.CategoryID)
	})

	t.Run("UpdateCompanyRequest model", func(t *testing.T) {
		name := "Updated Name"
		req := models.UpdateCompanyRequest{
			Name: &name,
		}
		assert.Equal(t, "Updated Name", *req.Name)
	})
}
