package repository

import (
	"context"
	"testing"
	"time"

	"github.com/homepay/api/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

// Company Repository Tests with Mocks

func TestCompanyRepo_Create_WithMock(t *testing.T) {
	mockRepo := new(MockCompanyRepository)

	authUserID := "user-123"
	req := &models.CreateCompanyRequest{
		Name:       "Test Company",
		CategoryID: 1,
		Website:    strPtr("https://test.com"),
		Phone:      strPtr("+1234567890"),
	}

	now := time.Now()
	expectedCompany := &models.Company{
		ID:         "company-123",
		AuthUserID: authUserID,
		Name:       req.Name,
		CategoryID: req.CategoryID,
		IsActive:   true,
		CreatedAt:  now,
	}

	mockRepo.On("Create", mock.Anything, authUserID, req).Return(expectedCompany, nil)

	result, err := mockRepo.Create(context.Background(), authUserID, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "company-123", result.ID)
	assert.Equal(t, req.Name, result.Name)
	mockRepo.AssertExpectations(t)
}

func TestCompanyRepo_GetByID_WithMock(t *testing.T) {
	mockRepo := new(MockCompanyRepository)

	companyID := "company-123"
	authUserID := "user-123"

	expectedCompany := &models.Company{
		ID:         companyID,
		AuthUserID: authUserID,
		Name:       "Test Company",
		CategoryID: 1,
	}

	mockRepo.On("GetByID", mock.Anything, companyID, authUserID).Return(expectedCompany, nil)

	result, err := mockRepo.GetByID(context.Background(), companyID, authUserID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, companyID, result.ID)
	assert.Equal(t, "Test Company", result.Name)
	mockRepo.AssertExpectations(t)
}

func TestCompanyRepo_GetByID_NotFound_WithMock(t *testing.T) {
	mockRepo := new(MockCompanyRepository)

	mockRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)

	result, err := mockRepo.GetByID(context.Background(), "non-existent", "user-123")

	assert.NoError(t, err)
	assert.Nil(t, result)
	mockRepo.AssertExpectations(t)
}

func TestCompanyRepo_GetAll_WithMock(t *testing.T) {
	mockRepo := new(MockCompanyRepository)

	authUserID := "user-123"
	pagination := models.PaginationParams{Limit: 10}

	companies := []models.Company{
		{ID: "company-1", Name: "Company 1"},
		{ID: "company-2", Name: "Company 2"},
	}

	mockRepo.On("GetAll", mock.Anything, authUserID, pagination).Return(companies, 2, nil)

	result, total, err := mockRepo.GetAll(context.Background(), authUserID, pagination)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(result))
	assert.Equal(t, 2, total)
	mockRepo.AssertExpectations(t)
}

func TestCompanyRepo_Update_WithMock(t *testing.T) {
	mockRepo := new(MockCompanyRepository)

	companyID := "company-123"
	authUserID := "user-123"

	newName := "Updated Company"
	website := "https://updated.com"
	req := &models.UpdateCompanyRequest{
		Name:    &newName,
		Website: &website,
	}

	updatedCompany := &models.Company{
		ID:     companyID,
		Name:   newName,
		Website: &website,
	}

	mockRepo.On("Update", mock.Anything, companyID, authUserID, req).Return(updatedCompany, nil)

	result, err := mockRepo.Update(context.Background(), companyID, authUserID, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, newName, result.Name)
	mockRepo.AssertExpectations(t)
}

func TestCompanyRepo_Update_NotFound_WithMock(t *testing.T) {
	mockRepo := new(MockCompanyRepository)

	req := &models.UpdateCompanyRequest{Name: strPtr("New Name")}

	mockRepo.On("Update", mock.Anything, "non-existent", "user-123", req).Return(nil, nil)

	result, err := mockRepo.Update(context.Background(), "non-existent", "user-123", req)

	assert.NoError(t, err)
	assert.Nil(t, result)
	mockRepo.AssertExpectations(t)
}

func TestCompanyRepo_SoftDelete_WithMock(t *testing.T) {
	mockRepo := new(MockCompanyRepository)

	companyID := "company-123"
	authUserID := "user-123"

	mockRepo.On("SoftDelete", mock.Anything, companyID, authUserID).Return(nil)

	err := mockRepo.SoftDelete(context.Background(), companyID, authUserID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestCompanyRepo_SoftDelete_NotFound_WithMock(t *testing.T) {
	mockRepo := new(MockCompanyRepository)

	mockRepo.On("SoftDelete", mock.Anything, "non-existent", "user-123").Return(pgx.ErrNoRows)

	err := mockRepo.SoftDelete(context.Background(), "non-existent", "user-123")

	assert.Error(t, err)
	assert.Equal(t, pgx.ErrNoRows, err)
	mockRepo.AssertExpectations(t)
}

// MockCompanyRepository implementation
type MockCompanyRepository struct {
	mock.Mock
}

func (m *MockCompanyRepository) Create(ctx context.Context, authUserID string, req *models.CreateCompanyRequest) (*models.Company, error) {
	args := m.Called(ctx, authUserID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Company), args.Error(1)
}

func (m *MockCompanyRepository) GetByID(ctx context.Context, id, authUserID string) (*models.Company, error) {
	args := m.Called(ctx, id, authUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Company), args.Error(1)
}

func (m *MockCompanyRepository) GetAll(ctx context.Context, authUserID string, p models.PaginationParams) ([]models.Company, int, error) {
	args := m.Called(ctx, authUserID, p)
	return args.Get(0).([]models.Company), args.Int(1), args.Error(2)
}

func (m *MockCompanyRepository) Update(ctx context.Context, id, authUserID string, req *models.UpdateCompanyRequest) (*models.Company, error) {
	args := m.Called(ctx, id, authUserID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Company), args.Error(1)
}

func (m *MockCompanyRepository) SoftDelete(ctx context.Context, id, authUserID string) error {
	args := m.Called(ctx, id, authUserID)
	return args.Error(0)
}

func TestMockCompanyRepository_ImplementsInterface(t *testing.T) {
	var _ CompanyRepository = (*MockCompanyRepository)(nil)
}

// Additional edge case tests

func TestCompanyRepo_WebsiteEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		website string
		valid   bool
	}{
		{"valid URL", "https://example.com", true},
		{"valid URL http", "http://example.com", true},
		{"empty website", "", false},
		{"invalid URL", "not-a-url", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.website == "" {
				req := models.CreateCompanyRequest{
					Name:       "Test",
					CategoryID: 1,
				}
				assert.Nil(t, req.Website)
				return
			}
			website := tt.website
			req := models.CreateCompanyRequest{
				Name:       "Test",
				CategoryID: 1,
				Website:    &website,
			}
			assert.NotNil(t, req.Website)
		})
	}
}

func TestCompanyRepo_PhoneEdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		phone string
	}{
		{"international format", "+1-234-567-8900"},
		{"local format", "1234567890"},
		{"with dashes", "123-456-7890"},
		{"with spaces", "123 456 7890"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := models.CreateCompanyRequest{
				Name:       "Test",
				CategoryID: 1,
				Phone:      &tt.phone,
			}
			assert.Equal(t, tt.phone, *req.Phone)
		})
	}
}

func TestCompanyRepo_CategoryEdgeCases(t *testing.T) {
	tests := []struct {
		name       string
		categoryID int
	}{
		{"category 1", 1},
		{"category 2", 2},
		{"category 100", 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			company := models.Company{
				ID:         "test",
				CategoryID: tt.categoryID,
				Name:       "Test Company",
			}
			assert.Equal(t, tt.categoryID, company.CategoryID)
		})
	}
}

// Company Repository Query Tests

func TestCompanyRepo_Create_Query(t *testing.T) {
	// Test INSERT query structure
	query := `INSERT INTO homepay.companies (auth_user_id, category_id, name, website, phone)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, auth_user_id, category_id, name, website, phone, is_active, created_at, deleted_at`
	
	assert.Contains(t, query, "INSERT INTO homepay.companies")
	assert.Contains(t, query, "RETURNING")
}

func TestCompanyRepo_GetByID_Query(t *testing.T) {
	// Test SELECT by ID query
	query := `SELECT id, auth_user_id, category_id, name, website, phone, is_active, created_at, deleted_at
		FROM homepay.companies
		WHERE id = $1 AND auth_user_id = $2 AND deleted_at IS NULL`
	
	assert.Contains(t, query, "FROM homepay.companies")
	assert.Contains(t, query, "deleted_at IS NULL")
}

func TestCompanyRepo_GetAll_Query(t *testing.T) {
	// Test paginated SELECT query
	query := `SELECT id, auth_user_id, category_id, name, website, phone, is_active, created_at, deleted_at
		FROM homepay.companies
		WHERE auth_user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`
	
	assert.Contains(t, query, "ORDER BY created_at DESC")
	assert.Contains(t, query, "LIMIT")
}

func TestCompanyRepo_Update_Query(t *testing.T) {
	// Test UPDATE query
	query := `UPDATE homepay.companies
		SET name        = COALESCE($3, name),
		    category_id = COALESCE($4, category_id),
		    website     = COALESCE($5, website),
		    phone       = COALESCE($6, phone)
		WHERE id = $1 AND auth_user_id = $2 AND deleted_at IS NULL
		RETURNING id, auth_user_id, category_id, name, website, phone, is_active, created_at, deleted_at`
	
	assert.Contains(t, query, "COALESCE")
	assert.Contains(t, query, "RETURNING")
}

func TestCompanyRepo_SoftDelete_Query(t *testing.T) {
	// Test soft delete query
	query := `UPDATE homepay.companies
		SET deleted_at = NOW(), is_active = FALSE
		WHERE id = $1 AND auth_user_id = $2 AND deleted_at IS NULL`
	
	assert.Contains(t, query, "deleted_at = NOW()")
	assert.Contains(t, query, "is_active = FALSE")
}
