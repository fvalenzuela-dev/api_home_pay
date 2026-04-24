package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/homepay/api/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

var (
	testPoolCompany   *pgxpool.Pool
	testRepoCompany   CompanyRepository
	testUserIDCompany = "test-user-integration-company"
	testCategoryID    int
)

func setupTestDBCompany(t *testing.T) *postgres.PostgresContainer {
	t.Helper()

	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("homepay_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
	)
	require.NoError(t, err, "failed to start postgres container")

	return pgContainer
}

func initTestRepoCompany(t *testing.T, pgContainer *postgres.PostgresContainer) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	time.Sleep(2 * time.Second)

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err, "failed to get connection string")

	if connStr == "" {
		host, err := pgContainer.Host(ctx)
		require.NoError(t, err)
		port, err := pgContainer.MappedPort(ctx, "5432")
		require.NoError(t, err)
		connStr = fmt.Sprintf("postgres://test:test@%s:%s/homepay_test?sslmode=disable", host, port.String())
	}

	var pool *pgxpool.Pool
	for i := 0; i < 10; i++ {
		pool, err = pgxpool.New(ctx, connStr)
		if err == nil {
			err = pool.Ping(ctx)
			if err == nil {
				break
			}
			pool.Close()
		}
		time.Sleep(500 * time.Millisecond)
	}
	require.NoError(t, err, "failed to connect to database after retries")

	testPoolCompany = pool
	testRepoCompany = NewCompanyRepository(pool)

	_, err = pool.Exec(ctx, `
		CREATE SCHEMA IF NOT EXISTS homepay;

		CREATE TABLE IF NOT EXISTS homepay.categories (
			id SMALLINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
			auth_user_id VARCHAR(255) NOT NULL,
			name VARCHAR(100) NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			deleted_at TIMESTAMPTZ
		);

		CREATE TABLE IF NOT EXISTS homepay.companies (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			auth_user_id VARCHAR(255) NOT NULL,
			category_id SMALLINT NOT NULL REFERENCES homepay.categories(id),
			name VARCHAR(255) NOT NULL,
			website VARCHAR(255),
			phone VARCHAR(50),
			is_active BOOLEAN NOT NULL DEFAULT TRUE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			deleted_at TIMESTAMPTZ
		);

		CREATE INDEX IF NOT EXISTS idx_categories_user ON homepay.categories(auth_user_id);
		CREATE INDEX IF NOT EXISTS idx_companies_user ON homepay.companies(auth_user_id);
		CREATE INDEX IF NOT EXISTS idx_companies_category ON homepay.companies(category_id);
	`)
	require.NoError(t, err, "failed to create schema")

	var catID int
	err = pool.QueryRow(ctx, `
		INSERT INTO homepay.categories (auth_user_id, name)
		VALUES ($1, 'Test Category')
		RETURNING id
	`, testUserIDCompany).Scan(&catID)
	require.NoError(t, err, "failed to create test category")
	testCategoryID = catID
}

func teardownTestDBCompany(t *testing.T, pgContainer *postgres.PostgresContainer) {
	t.Helper()

	if testPoolCompany != nil {
		testPoolCompany.Close()
	}
	pgContainer.Terminate(context.Background())
}

func createTestCompanyDirect(t *testing.T) string {
	ctx := context.Background()
	var companyID string
	err := testPoolCompany.QueryRow(ctx, `
		INSERT INTO homepay.companies (auth_user_id, category_id, name)
		VALUES ($1, $2, 'Test Company')
		RETURNING id
	`, testUserIDCompany, testCategoryID).Scan(&companyID)
	require.NoError(t, err)
	return companyID
}

func TestCompanyRepo_Create_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	pgContainer := setupTestDBCompany(t)
	initTestRepoCompany(t, pgContainer)
	defer teardownTestDBCompany(t, pgContainer)

	t.Run("creates company successfully", func(t *testing.T) {
		req := &models.CreateCompanyRequest{
			Name:       "New Company",
			CategoryID: testCategoryID,
		}

		company, err := testRepoCompany.Create(ctx, testUserIDCompany, req)

		require.NoError(t, err)
		require.NotNil(t, company)
		assert.Equal(t, "New Company", company.Name)
		assert.Equal(t, testCategoryID, company.CategoryID)
		assert.True(t, company.IsActive)
		assert.NotEmpty(t, company.ID)
		assert.Equal(t, testUserIDCompany, company.AuthUserID)
	})

	t.Run("creates company with optional fields", func(t *testing.T) {
		website := "https://example.com"
		phone := "+1234567890"

		req := &models.CreateCompanyRequest{
			Name:       "Company with extras",
			CategoryID: testCategoryID,
			Website:    &website,
			Phone:      &phone,
		}

		company, err := testRepoCompany.Create(ctx, testUserIDCompany, req)

		require.NoError(t, err)
		require.NotNil(t, company)
		assert.Equal(t, "Company with extras", company.Name)
		assert.NotNil(t, company.Website)
		assert.Equal(t, website, *company.Website)
		assert.NotNil(t, company.Phone)
		assert.Equal(t, phone, *company.Phone)
	})
}

func TestCompanyRepo_GetByID_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	pgContainer := setupTestDBCompany(t)
	initTestRepoCompany(t, pgContainer)
	defer teardownTestDBCompany(t, pgContainer)

	companyID := createTestCompanyDirect(t)

	t.Run("gets company by id", func(t *testing.T) {
		company, err := testRepoCompany.GetByID(ctx, companyID, testUserIDCompany)

		require.NoError(t, err)
		require.NotNil(t, company)
		assert.Equal(t, companyID, company.ID)
		assert.Equal(t, "Test Company", company.Name)
		assert.Equal(t, testCategoryID, company.CategoryID)
	})

	t.Run("returns nil for non-existent company", func(t *testing.T) {
		company, err := testRepoCompany.GetByID(ctx, "00000000-0000-0000-0000-000000000001", testUserIDCompany)

		assert.NoError(t, err)
		assert.Nil(t, company)
	})

	t.Run("returns nil for other user's company", func(t *testing.T) {
		company, err := testRepoCompany.GetByID(ctx, companyID, "other-user-id")

		assert.NoError(t, err)
		assert.Nil(t, company)
	})
}

func TestCompanyRepo_GetAll_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	pgContainer := setupTestDBCompany(t)
	initTestRepoCompany(t, pgContainer)
	defer teardownTestDBCompany(t, pgContainer)

	for i := 0; i < 5; i++ {
		req := &models.CreateCompanyRequest{
			Name:       fmt.Sprintf("Company %d", i),
			CategoryID: testCategoryID,
		}
		_, err := testRepoCompany.Create(ctx, testUserIDCompany, req)
		require.NoError(t, err)
	}

	t.Run("gets all companies with pagination", func(t *testing.T) {
		pagination := models.PaginationParams{Page: 1, Limit: 2}

		companies, total, err := testRepoCompany.GetAll(ctx, testUserIDCompany, pagination)

		require.NoError(t, err)
		assert.Len(t, companies, 2)
		assert.Equal(t, 5, total)
	})

	t.Run("gets all companies without pagination limits", func(t *testing.T) {
		pagination := models.PaginationParams{Page: 1, Limit: 100}

		companies, total, err := testRepoCompany.GetAll(ctx, testUserIDCompany, pagination)

		require.NoError(t, err)
		assert.Len(t, companies, 5)
		assert.Equal(t, 5, total)
	})

	t.Run("returns empty for non-existent user", func(t *testing.T) {
		pagination := models.PaginationParams{Page: 1, Limit: 100}

		companies, total, err := testRepoCompany.GetAll(ctx, "non-existent-user", pagination)

		require.NoError(t, err)
		assert.Len(t, companies, 0)
		assert.Equal(t, 0, total)
	})
}

func TestCompanyRepo_Update_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	pgContainer := setupTestDBCompany(t)
	initTestRepoCompany(t, pgContainer)
	defer teardownTestDBCompany(t, pgContainer)

	companyID := createTestCompanyDirect(t)

	t.Run("updates company successfully", func(t *testing.T) {
		newName := "Updated Company Name"

		req := &models.UpdateCompanyRequest{
			Name: &newName,
		}

		company, err := testRepoCompany.Update(ctx, companyID, testUserIDCompany, req)

		require.NoError(t, err)
		require.NotNil(t, company)
		assert.Equal(t, newName, company.Name)
	})

	t.Run("updates company with optional fields", func(t *testing.T) {
		website := "https://updated.com"

		req := &models.UpdateCompanyRequest{
			Website: &website,
		}

		company, err := testRepoCompany.Update(ctx, companyID, testUserIDCompany, req)

		require.NoError(t, err)
		require.NotNil(t, company)
		assert.NotNil(t, company.Website)
		assert.Equal(t, website, *company.Website)
	})

	t.Run("returns nil for non-existent company", func(t *testing.T) {
		name := "New Name"
		req := &models.UpdateCompanyRequest{
			Name: &name,
		}

		company, err := testRepoCompany.Update(ctx, "00000000-0000-0000-0000-000000000001", testUserIDCompany, req)

		assert.NoError(t, err)
		assert.Nil(t, company)
	})
}

func TestCompanyRepo_SoftDelete_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	pgContainer := setupTestDBCompany(t)
	initTestRepoCompany(t, pgContainer)
	defer teardownTestDBCompany(t, pgContainer)

	companyID := createTestCompanyDirect(t)

	t.Run("soft deletes company", func(t *testing.T) {
		err := testRepoCompany.SoftDelete(ctx, companyID, testUserIDCompany)

		require.NoError(t, err)

		company, err := testRepoCompany.GetByID(ctx, companyID, testUserIDCompany)
		require.NoError(t, err)
		assert.Nil(t, company)
	})

	t.Run("returns error for non-existent company", func(t *testing.T) {
		err := testRepoCompany.SoftDelete(ctx, "00000000-0000-0000-0000-000000000001", testUserIDCompany)

		assert.Error(t, err)
	})
}
