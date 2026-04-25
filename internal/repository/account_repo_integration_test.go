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
	testPoolAccount   *pgxpool.Pool
	testRepoAccount   AccountRepository
	testUserIDAccount = "test-user-integration-account"
)

func setupTestDBAccount(t *testing.T) *postgres.PostgresContainer {
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

func initTestRepoAccount(t *testing.T, pgContainer *postgres.PostgresContainer) {
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
		connStr = fmt.Sprintf("postgres://user:pass@%s:%s/homepay_test?sslmode=disable", host, port.String())
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

	testPoolAccount = pool
	testRepoAccount = NewAccountRepository(pool)

	_, err = pool.Exec(ctx, `
		CREATE SCHEMA IF NOT EXISTS homepay;

		CREATE TABLE IF NOT EXISTS homepay.companies (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			auth_user_id VARCHAR(255) NOT NULL,
			name VARCHAR(255) NOT NULL,
			is_active BOOLEAN NOT NULL DEFAULT TRUE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			deleted_at TIMESTAMPTZ
		);

		CREATE TABLE IF NOT EXISTS homepay.account_groups (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			auth_user_id VARCHAR(255) NOT NULL,
			name VARCHAR(100) NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			deleted_at TIMESTAMPTZ
		);

		CREATE TABLE IF NOT EXISTS homepay.accounts (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			company_id UUID NOT NULL REFERENCES homepay.companies(id),
			group_id UUID REFERENCES homepay.account_groups(id),
			account_number VARCHAR(50),
			name VARCHAR(255) NOT NULL,
			billing_day SMALLINT NOT NULL DEFAULT 1,
			auto_accumulate BOOLEAN NOT NULL DEFAULT FALSE,
			is_active BOOLEAN NOT NULL DEFAULT TRUE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			deleted_at TIMESTAMPTZ
		);

		CREATE INDEX IF NOT EXISTS idx_companies_user ON homepay.companies(auth_user_id);
		CREATE INDEX IF NOT EXISTS idx_accounts_company ON homepay.accounts(company_id);
		CREATE INDEX IF NOT EXISTS idx_account_groups_user ON homepay.account_groups(auth_user_id);
	`)
	require.NoError(t, err, "failed to create schema")
}

func teardownTestDBAccount(t *testing.T, pgContainer *postgres.PostgresContainer) {
	t.Helper()

	if testPoolAccount != nil {
		testPoolAccount.Close()
	}
	pgContainer.Terminate(context.Background())
}

func createTestCompanyAccount(t *testing.T) string {
	ctx := context.Background()
	var companyID string
	err := testPoolAccount.QueryRow(ctx, `
		INSERT INTO homepay.companies (auth_user_id, name)
		VALUES ($1, 'Test Company')
		RETURNING id
	`, testUserIDAccount).Scan(&companyID)
	require.NoError(t, err)
	return companyID
}

func createTestAccountGroupAccount(t *testing.T) string {
	ctx := context.Background()
	var groupID string
	err := testPoolAccount.QueryRow(ctx, `
		INSERT INTO homepay.account_groups (auth_user_id, name)
		VALUES ($1, 'Test Group')
		RETURNING id
	`, testUserIDAccount).Scan(&groupID)
	require.NoError(t, err)
	return groupID
}

func TestAccountRepo_Create_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	pgContainer := setupTestDBAccount(t)
	initTestRepoAccount(t, pgContainer)
	defer teardownTestDBAccount(t, pgContainer)

	t.Run("creates account successfully", func(t *testing.T) {
		companyID := createTestCompanyAccount(t)

		req := &models.CreateAccountRequest{
			Name:           "New Account",
			BillingDay:     20,
			AutoAccumulate: true,
		}

		account, err := testRepoAccount.Create(ctx, companyID, testUserIDAccount, req)

		require.NoError(t, err)
		require.NotNil(t, account)
		assert.Equal(t, "New Account", account.Name)
		assert.Equal(t, 20, account.BillingDay)
		assert.True(t, account.AutoAccumulate)
		assert.True(t, account.IsActive)
		assert.NotEmpty(t, account.ID)
		assert.Equal(t, companyID, account.CompanyID)
	})

	t.Run("creates account with optional fields", func(t *testing.T) {
		companyID := createTestCompanyAccount(t)
		groupID := createTestAccountGroupAccount(t)
		accountNum := "123456789"

		req := &models.CreateAccountRequest{
			Name:           "Account with extras",
			BillingDay:     25,
			AutoAccumulate: false,
			GroupID:        &groupID,
			AccountNumber:  &accountNum,
		}

		account, err := testRepoAccount.Create(ctx, companyID, testUserIDAccount, req)

		require.NoError(t, err)
		require.NotNil(t, account)
		assert.Equal(t, "Account with extras", account.Name)
		assert.NotNil(t, account.GroupID)
		assert.Equal(t, groupID, *account.GroupID)
		assert.NotNil(t, account.AccountNumber)
		assert.Equal(t, accountNum, *account.AccountNumber)
	})

	t.Run("fails with invalid company", func(t *testing.T) {
		req := &models.CreateAccountRequest{
			Name:       "Invalid Account",
			BillingDay: 1,
		}

		account, err := testRepoAccount.Create(ctx, "invalid-uuid", testUserIDAccount, req)

		assert.Nil(t, account)
		assert.Error(t, err)
	})
}

func TestAccountRepo_GetByID_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	pgContainer := setupTestDBAccount(t)
	initTestRepoAccount(t, pgContainer)
	defer teardownTestDBAccount(t, pgContainer)

	companyID := createTestCompanyAccount(t)

	var createdAccountID string
	err := testPoolAccount.QueryRow(ctx, `
		INSERT INTO homepay.accounts (company_id, name, billing_day, auto_accumulate)
		VALUES ($1, 'Test Account', 15, true)
		RETURNING id
	`, companyID).Scan(&createdAccountID)
	require.NoError(t, err)

	t.Run("gets account by id", func(t *testing.T) {
		account, err := testRepoAccount.GetByID(ctx, createdAccountID, testUserIDAccount)

		require.NoError(t, err)
		require.NotNil(t, account)
		assert.Equal(t, createdAccountID, account.ID)
		assert.Equal(t, "Test Account", account.Name)
		assert.Equal(t, companyID, account.CompanyID)
	})

	t.Run("returns nil for non-existent account", func(t *testing.T) {
		account, err := testRepoAccount.GetByID(ctx, "00000000-0000-0000-0000-000000000001", testUserIDAccount)

		assert.NoError(t, err)
		assert.Nil(t, account)
	})
}

func TestAccountRepo_GetAllByCompany_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	pgContainer := setupTestDBAccount(t)
	initTestRepoAccount(t, pgContainer)
	defer teardownTestDBAccount(t, pgContainer)

	companyID := createTestCompanyAccount(t)

	for i := 0; i < 5; i++ {
		req := &models.CreateAccountRequest{
			Name:       fmt.Sprintf("Account %d", i),
			BillingDay: 1 + i,
		}
		_, err := testRepoAccount.Create(ctx, companyID, testUserIDAccount, req)
		require.NoError(t, err)
	}

	t.Run("gets all accounts with pagination", func(t *testing.T) {
		pagination := models.PaginationParams{Page: 1, Limit: 2}

		accounts, total, err := testRepoAccount.GetAllByCompany(ctx, companyID, testUserIDAccount, pagination)

		require.NoError(t, err)
		assert.Len(t, accounts, 2)
		assert.Equal(t, 5, total)
	})

	t.Run("gets all accounts without pagination limits", func(t *testing.T) {
		pagination := models.PaginationParams{Page: 1, Limit: 100}

		accounts, total, err := testRepoAccount.GetAllByCompany(ctx, companyID, testUserIDAccount, pagination)

		require.NoError(t, err)
		assert.Len(t, accounts, 5)
		assert.Equal(t, 5, total)
	})
}

func TestAccountRepo_GetAllActiveByUser_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	pgContainer := setupTestDBAccount(t)
	initTestRepoAccount(t, pgContainer)
	defer teardownTestDBAccount(t, pgContainer)

	companyID := createTestCompanyAccount(t)

	for i := 0; i < 3; i++ {
		req := &models.CreateAccountRequest{
			Name:       fmt.Sprintf("Active Account %d", i),
			BillingDay: 1,
		}
		_, err := testRepoAccount.Create(ctx, companyID, testUserIDAccount, req)
		require.NoError(t, err)
	}

	t.Run("gets all active accounts for user", func(t *testing.T) {
		accounts, err := testRepoAccount.GetAllActiveByUser(ctx, testUserIDAccount)

		require.NoError(t, err)
		assert.Len(t, accounts, 3)
		for _, acc := range accounts {
			assert.True(t, acc.IsActive)
		}
	})

	t.Run("returns empty for non-existent user", func(t *testing.T) {
		accounts, err := testRepoAccount.GetAllActiveByUser(ctx, "non-existent-user")

		require.NoError(t, err)
		assert.Len(t, accounts, 0)
	})
}

func TestAccountRepo_GetActiveIDsByCompany_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	pgContainer := setupTestDBAccount(t)
	initTestRepoAccount(t, pgContainer)
	defer teardownTestDBAccount(t, pgContainer)

	companyID := createTestCompanyAccount(t)

	for i := 0; i < 3; i++ {
		req := &models.CreateAccountRequest{
			Name:       fmt.Sprintf("Account %d", i),
			BillingDay: 1,
		}
		_, err := testRepoAccount.Create(ctx, companyID, testUserIDAccount, req)
		require.NoError(t, err)
	}

	t.Run("gets active account IDs", func(t *testing.T) {
		ids, err := testRepoAccount.GetActiveIDsByCompany(ctx, companyID)

		require.NoError(t, err)
		assert.Len(t, ids, 3)
	})

	t.Run("returns empty for company without accounts", func(t *testing.T) {
		otherCompanyID := createTestCompanyAccount(t)

		ids, err := testRepoAccount.GetActiveIDsByCompany(ctx, otherCompanyID)

		require.NoError(t, err)
		assert.Len(t, ids, 0)
	})
}

func TestAccountRepo_Update_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	pgContainer := setupTestDBAccount(t)
	initTestRepoAccount(t, pgContainer)
	defer teardownTestDBAccount(t, pgContainer)

	companyID := createTestCompanyAccount(t)

	var accountID string
	err := testPoolAccount.QueryRow(ctx, `
		INSERT INTO homepay.accounts (company_id, name, billing_day, auto_accumulate)
		VALUES ($1, 'Original Account', 10, false)
		RETURNING id
	`, companyID).Scan(&accountID)
	require.NoError(t, err)

	t.Run("updates account successfully", func(t *testing.T) {
		newName := "Updated Account Name"
		newBillingDay := 30
		autoAcc := false

		req := &models.UpdateAccountRequest{
			Name:           &newName,
			BillingDay:     &newBillingDay,
			AutoAccumulate: &autoAcc,
		}

		account, err := testRepoAccount.Update(ctx, accountID, testUserIDAccount, req)

		require.NoError(t, err)
		require.NotNil(t, account)
		assert.Equal(t, newName, account.Name)
		assert.Equal(t, newBillingDay, account.BillingDay)
		assert.False(t, account.AutoAccumulate)
	})

	t.Run("returns nil for non-existent account", func(t *testing.T) {
		req := &models.UpdateAccountRequest{
			Name: strPtr("New Name"),
		}

		account, err := testRepoAccount.Update(ctx, "00000000-0000-0000-0000-000000000001", testUserIDAccount, req)

		assert.NoError(t, err)
		assert.Nil(t, account)
	})
}

func TestAccountRepo_SoftDelete_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	pgContainer := setupTestDBAccount(t)
	initTestRepoAccount(t, pgContainer)
	defer teardownTestDBAccount(t, pgContainer)

	companyID := createTestCompanyAccount(t)

	var accountID string
	err := testPoolAccount.QueryRow(ctx, `
		INSERT INTO homepay.accounts (company_id, name, billing_day, auto_accumulate)
		VALUES ($1, 'To Delete', 10, true)
		RETURNING id
	`, companyID).Scan(&accountID)
	require.NoError(t, err)

	t.Run("soft deletes account", func(t *testing.T) {
		err := testRepoAccount.SoftDelete(ctx, accountID, testUserIDAccount)

		require.NoError(t, err)

		account, err := testRepoAccount.GetByID(ctx, accountID, testUserIDAccount)
		require.NoError(t, err)
		assert.Nil(t, account)
	})

	t.Run("returns error for non-existent account", func(t *testing.T) {
		err := testRepoAccount.SoftDelete(ctx, "00000000-0000-0000-0000-000000000001", testUserIDAccount)

		assert.Error(t, err)
	})
}

func TestAccountRepo_SoftDeleteByCompany_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	pgContainer := setupTestDBAccount(t)
	initTestRepoAccount(t, pgContainer)
	defer teardownTestDBAccount(t, pgContainer)

	companyID := createTestCompanyAccount(t)

	for i := 0; i < 3; i++ {
		req := &models.CreateAccountRequest{
			Name:       fmt.Sprintf("Account %d", i),
			BillingDay: 1,
		}
		_, err := testRepoAccount.Create(ctx, companyID, testUserIDAccount, req)
		require.NoError(t, err)
	}

	t.Run("soft deletes all accounts by company", func(t *testing.T) {
		err := testRepoAccount.SoftDeleteByCompany(ctx, companyID)

		require.NoError(t, err)

		ids, err := testRepoAccount.GetActiveIDsByCompany(ctx, companyID)
		require.NoError(t, err)
		assert.Len(t, ids, 0)
	})
}
