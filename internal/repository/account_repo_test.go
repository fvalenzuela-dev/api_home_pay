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

// Ensure fmt is used
var _ = fmt.Sprintf

var (
	accountTestPool    *pgxpool.Pool
	accountTestUserID  = "test-user-integration"
	accountTestRepo    AccountRepository
	companyTestRepo    CompanyRepository
)

func setupAccountTestDB(t *testing.T) *postgres.PostgresContainer {
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

func initAccountTestRepo(t *testing.T, pgContainer *postgres.PostgresContainer) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Wait for container to be fully ready
	time.Sleep(2 * time.Second)

	// Get connection string from container settings
	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err, "failed to get connection string")

	// Replace database name if needed
	if connStr == "" {
		host, err := pgContainer.Host(ctx)
		require.NoError(t, err)
		port, err := pgContainer.MappedPort(ctx, "5432")
		require.NoError(t, err)
		connStr = fmt.Sprintf("postgres://***REMOVED***%s:%s/homepay_test?sslmode=disable", host, port.String())
	}

	// Retry connection
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

	accountTestPool = pool
	accountTestRepo = NewAccountRepository(pool)
	companyTestRepo = NewCompanyRepository(pool)

	// Create schema
	_, err = pool.Exec(ctx, `
		CREATE SCHEMA IF NOT EXISTS homepay;
		
		CREATE TABLE IF NOT EXISTS homepay.categories (
			id SMALLINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
			name VARCHAR(255) NOT NULL
		);
		
		CREATE TABLE IF NOT EXISTS homepay.companies (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			auth_user_id VARCHAR(255) NOT NULL,
			category_id SMALLINT,
			name VARCHAR(255) NOT NULL,
			website VARCHAR(255),
			phone VARCHAR(50),
			is_active BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			deleted_at TIMESTAMP WITH TIME ZONE,
			FOREIGN KEY (category_id) REFERENCES homepay.categories(id)
		);
		
		CREATE TABLE IF NOT EXISTS homepay.accounts (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			company_id UUID NOT NULL,
			group_id UUID,
			account_number VARCHAR(50),
			name VARCHAR(255) NOT NULL,
			billing_day SMALLINT NOT NULL,
			auto_accumulate BOOLEAN DEFAULT FALSE,
			is_active BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			deleted_at TIMESTAMP WITH TIME ZONE,
			FOREIGN KEY (company_id) REFERENCES homepay.companies(id)
		);
		
		CREATE INDEX IF NOT EXISTS idx_companies_user_id ON homepay.companies(auth_user_id);
		CREATE INDEX IF NOT EXISTS idx_accounts_company_id ON homepay.accounts(company_id);
	`)
	require.NoError(t, err, "failed to create schema")
	
	// Insert a category for testing
	_, err = pool.Exec(ctx, `INSERT INTO homepay.categories (name) VALUES ('Test Category')`)
	require.NoError(t, err, "failed to insert category")
}

func teardownAccountTestDB(t *testing.T, pgContainer *postgres.PostgresContainer) {
	t.Helper()

	if accountTestPool != nil {
		accountTestPool.Close()
	}
	pgContainer.Terminate(context.Background())
}

// TDD: Test first - these tests define the expected behavior

func TestAccountRepo_Create_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	pgContainer := setupAccountTestDB(t)
	initAccountTestRepo(t, pgContainer)
	defer teardownAccountTestDB(t, pgContainer)

	t.Run("creates account successfully", func(t *testing.T) {
		// First create a company
		company, err := companyTestRepo.Create(ctx, accountTestUserID, &models.CreateCompanyRequest{
			Name: "Test Company",
		})
		require.NoError(t, err)

		req := &models.CreateAccountRequest{
			Name:            "Test Account",
			BillingDay:      15,
			AutoAccumulate:  true,
		}

		account, err := accountTestRepo.Create(ctx, company.ID, accountTestUserID, req)

		require.NoError(t, err)
		assert.NotNil(t, account)
		assert.NotEmpty(t, account.ID)
		assert.Equal(t, company.ID, account.CompanyID)
		assert.Equal(t, "Test Account", account.Name)
		assert.Equal(t, 15, account.BillingDay)
		assert.True(t, account.IsActive)
	})
}

func TestAccountRepo_GetByID_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	pgContainer := setupAccountTestDB(t)
	initAccountTestRepo(t, pgContainer)
	defer teardownAccountTestDB(t, pgContainer)

	// Create test data
	company, err := companyTestRepo.Create(ctx, accountTestUserID, &models.CreateCompanyRequest{
		Name: "Test Company",
	})
	require.NoError(t, err)

	created, err := accountTestRepo.Create(ctx, company.ID, accountTestUserID, &models.CreateAccountRequest{
		Name:            "GetByID Test",
		BillingDay:      10,
		AutoAccumulate:  false,
	})
	require.NoError(t, err)

	t.Run("finds existing account", func(t *testing.T) {
		account, err := accountTestRepo.GetByID(ctx, created.ID, accountTestUserID)

		require.NoError(t, err)
		assert.NotNil(t, account)
		assert.Equal(t, created.ID, account.ID)
		assert.Equal(t, "GetByID Test", account.Name)
	})

	t.Run("returns nil for non-existent", func(t *testing.T) {
		account, err := accountTestRepo.GetByID(ctx, "00000000-0000-0000-0000-000000000000", accountTestUserID)

		require.NoError(t, err)
		assert.Nil(t, account)
	})
}

func TestAccountRepo_GetAllByCompany_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	pgContainer := setupAccountTestDB(t)
	initAccountTestRepo(t, pgContainer)
	defer teardownAccountTestDB(t, pgContainer)

	// Create company and accounts
	company, err := companyTestRepo.Create(ctx, accountTestUserID, &models.CreateCompanyRequest{
		Name: "Test Company",
	})
	require.NoError(t, err)

	for i := 1; i <= 3; i++ {
		_, err := accountTestRepo.Create(ctx, company.ID, accountTestUserID, &models.CreateAccountRequest{
			Name:       fmt.Sprintf("Account %d", i),
			BillingDay: i * 5,
		})
		require.NoError(t, err)
	}

	t.Run("returns all accounts for company", func(t *testing.T) {
		result, total, err := accountTestRepo.GetAllByCompany(ctx, company.ID, accountTestUserID, models.PaginationParams{Page: 1, Limit: 10})

		require.NoError(t, err)
		assert.Len(t, result, 3)
		assert.Equal(t, 3, total)
	})

	t.Run("respects pagination", func(t *testing.T) {
		result, total, err := accountTestRepo.GetAllByCompany(ctx, company.ID, accountTestUserID, models.PaginationParams{Page: 1, Limit: 2})

		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, 3, total)
	})
}

func TestAccountRepo_Update_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	pgContainer := setupAccountTestDB(t)
	initAccountTestRepo(t, pgContainer)
	defer teardownAccountTestDB(t, pgContainer)

	// Create test data
	company, err := companyTestRepo.Create(ctx, accountTestUserID, &models.CreateCompanyRequest{
		Name: "Test Company",
	})
	require.NoError(t, err)

	created, err := accountTestRepo.Create(ctx, company.ID, accountTestUserID, &models.CreateAccountRequest{
		Name:       "Original Name",
		BillingDay: 1,
	})
	require.NoError(t, err)

	t.Run("updates account name", func(t *testing.T) {
		newName := "Updated Name"
		req := &models.UpdateAccountRequest{Name: &newName}

		account, err := accountTestRepo.Update(ctx, created.ID, accountTestUserID, req)

		require.NoError(t, err)
		assert.NotNil(t, account)
		assert.Equal(t, "Updated Name", account.Name)
	})
}

func TestAccountRepo_SoftDelete_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	pgContainer := setupAccountTestDB(t)
	initAccountTestRepo(t, pgContainer)
	defer teardownAccountTestDB(t, pgContainer)

	// Create test data
	company, err := companyTestRepo.Create(ctx, accountTestUserID, &models.CreateCompanyRequest{
		Name: "Test Company",
	})
	require.NoError(t, err)

	created, err := accountTestRepo.Create(ctx, company.ID, accountTestUserID, &models.CreateAccountRequest{
		Name: "To Delete",
	})
	require.NoError(t, err)

	t.Run("soft deletes account", func(t *testing.T) {
		err := accountTestRepo.SoftDelete(ctx, created.ID, accountTestUserID)

		require.NoError(t, err)

		// Verify it's not found anymore
		account, err := accountTestRepo.GetByID(ctx, created.ID, accountTestUserID)
		require.NoError(t, err)
		assert.Nil(t, account)
	})
}

// Unit tests for account repository

func TestAccountRepo_Interfaces(t *testing.T) {
	t.Run("AccountRepository interface is satisfied by accountRepo", func(t *testing.T) {
		var _ AccountRepository = (*accountRepo)(nil)
	})
}

func TestScanAccount(t *testing.T) {
	t.Run("scanAccount function exists", func(t *testing.T) {
		assert.NotNil(t, scanAccount)
	})
}

func TestAccountRepo_Create_Model(t *testing.T) {
	t.Run("CreateAccountRequest validation", func(t *testing.T) {
		req := models.CreateAccountRequest{
			Name:            "Test Account",
			BillingDay:      15,
			AutoAccumulate:  true,
		}
		assert.Equal(t, "Test Account", req.Name)
		assert.Equal(t, 15, req.BillingDay)
		assert.True(t, req.AutoAccumulate)
	})

	t.Run("CreateAccountRequest with optional fields", func(t *testing.T) {
		groupID := "group-123"
		accountNum := "123456789"
		req := models.CreateAccountRequest{
			GroupID:        &groupID,
			AccountNumber:  &accountNum,
			Name:           "Test Account",
			BillingDay:     1,
			AutoAccumulate: false,
		}
		assert.NotNil(t, req.GroupID)
		assert.NotNil(t, req.AccountNumber)
		assert.Equal(t, "group-123", *req.GroupID)
	})
}

func TestAccountRepo_Update_Model(t *testing.T) {
	t.Run("UpdateAccountRequest with pointer fields", func(t *testing.T) {
		name := "Updated Account"
		billingDay := 20
		autoAcc := true

		req := models.UpdateAccountRequest{
			Name:           &name,
			BillingDay:     &billingDay,
			AutoAccumulate: &autoAcc,
		}
		assert.Equal(t, "Updated Account", *req.Name)
		assert.Equal(t, 20, *req.BillingDay)
		assert.True(t, *req.AutoAccumulate)
	})

	t.Run("UpdateAccountRequest with nil fields", func(t *testing.T) {
		req := models.UpdateAccountRequest{}
		assert.Nil(t, req.Name)
		assert.Nil(t, req.BillingDay)
		assert.Nil(t, req.AutoAccumulate)
	})
}
