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
	testPoolBilling   *pgxpool.Pool
	testRepoBilling   BillingRepository
	testUserIDBilling = "test-user-integration-billing"
)

func setupTestDBBilling(t *testing.T) *postgres.PostgresContainer {
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

func initTestRepoBilling(t *testing.T, pgContainer *postgres.PostgresContainer) {
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
		connStr = fmt.Sprintf("postgres://x:x@%s:%s/homepay_test?sslmode=disable", host, port.String())
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

	testPoolBilling = pool
	testRepoBilling = NewBillingRepository(pool)

	_, err = pool.Exec(ctx, `
		CREATE SCHEMA IF NOT EXISTS homepay;

		CREATE TABLE IF NOT EXISTS homepay.categories (
			id SMALLINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
			auth_user_id VARCHAR(255) NOT NULL,
			name VARCHAR(100) NOT NULL,
			icon VARCHAR(50),
			color VARCHAR(20),
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			deleted_at TIMESTAMPTZ
		);

		CREATE TABLE IF NOT EXISTS homepay.companies (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			auth_user_id VARCHAR(255) NOT NULL,
			category_id SMALLINT REFERENCES homepay.categories(id),
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

		CREATE TABLE IF NOT EXISTS homepay.account_billings (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			account_id UUID NOT NULL REFERENCES homepay.accounts(id),
			period INTEGER NOT NULL,
			amount_billed DECIMAL(12, 2) NOT NULL,
			amount_paid DECIMAL(12, 2) NOT NULL DEFAULT 0,
			is_paid BOOLEAN NOT NULL DEFAULT FALSE,
			paid_at TIMESTAMPTZ,
			carried_from UUID,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			deleted_at TIMESTAMPTZ
		);

		CREATE INDEX IF NOT EXISTS idx_categories_user ON homepay.categories(auth_user_id);
		CREATE INDEX IF NOT EXISTS idx_companies_user ON homepay.companies(auth_user_id);
		CREATE INDEX IF NOT EXISTS idx_accounts_company ON homepay.accounts(company_id);
		CREATE INDEX IF NOT EXISTS idx_account_billings_account ON homepay.account_billings(account_id);
		CREATE INDEX IF NOT EXISTS idx_account_billings_period ON homepay.account_billings(period);
	`)
	require.NoError(t, err, "failed to create schema")
}

func teardownTestDBBilling(t *testing.T, pgContainer *postgres.PostgresContainer) {
	t.Helper()

	if testPoolBilling != nil {
		testPoolBilling.Close()
	}
	pgContainer.Terminate(context.Background())
}

func createTestCategoryBilling(t *testing.T) string {
	ctx := context.Background()
	var categoryID int
	err := testPoolBilling.QueryRow(ctx, `
		INSERT INTO homepay.categories (auth_user_id, name, icon, color)
		VALUES ($1, 'Test Category', 'test', '#000000')
		RETURNING id
	`, testUserIDBilling).Scan(&categoryID)
	require.NoError(t, err)
	return fmt.Sprintf("%d", categoryID)
}

func createTestCompanyBilling(t *testing.T, categoryID string) string {
	ctx := context.Background()
	var companyID string
	err := testPoolBilling.QueryRow(ctx, `
		INSERT INTO homepay.companies (auth_user_id, category_id, name)
		VALUES ($1, $2, 'Test Company')
		RETURNING id
	`, testUserIDBilling, categoryID).Scan(&companyID)
	require.NoError(t, err)
	return companyID
}

func createTestAccountBilling(t *testing.T, companyID string) string {
	ctx := context.Background()
	var accountID string
	err := testPoolBilling.QueryRow(ctx, `
		INSERT INTO homepay.accounts (company_id, name, billing_day, auto_accumulate)
		VALUES ($1, 'Test Account', 15, true)
		RETURNING id
	`, companyID).Scan(&accountID)
	require.NoError(t, err)
	return accountID
}

func TestBillingRepo_Create_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	pgContainer := setupTestDBBilling(t)
	initTestRepoBilling(t, pgContainer)
	defer teardownTestDBBilling(t, pgContainer)

	categoryID := createTestCategoryBilling(t)
	companyID := createTestCompanyBilling(t, categoryID)
	accountID := createTestAccountBilling(t, companyID)

	t.Run("creates billing successfully", func(t *testing.T) {
		req := &models.CreateBillingRequest{
			Period:       202603,
			AmountBilled: 100.50,
		}

		billing, err := testRepoBilling.Create(ctx, accountID, req)

		require.NoError(t, err)
		require.NotNil(t, billing)
		assert.Equal(t, 202603, billing.Period)
		assert.Equal(t, 100.50, billing.AmountBilled)
		assert.Equal(t, 0.0, billing.AmountPaid)
		assert.False(t, billing.IsPaid)
		assert.NotEmpty(t, billing.ID)
		assert.Equal(t, accountID, billing.AccountID)
	})

	t.Run("creates billing with amount paid and is paid", func(t *testing.T) {
		amountPaid := 100.50
		isPaid := true
		paidAt := time.Now()

		req := &models.CreateBillingRequest{
			Period:       202604,
			AmountBilled: 100.50,
			AmountPaid:   &amountPaid,
			IsPaid:       &isPaid,
			PaidAt:       &paidAt,
		}

		billing, err := testRepoBilling.Create(ctx, accountID, req)

		require.NoError(t, err)
		require.NotNil(t, billing)
		assert.True(t, billing.IsPaid)
		assert.Equal(t, 100.50, billing.AmountPaid)
		assert.NotNil(t, billing.PaidAt)
	})
}

func TestBillingRepo_CreateCarryOver_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	pgContainer := setupTestDBBilling(t)
	initTestRepoBilling(t, pgContainer)
	defer teardownTestDBBilling(t, pgContainer)

	categoryID := createTestCategoryBilling(t)
	companyID := createTestCompanyBilling(t, categoryID)
	accountID := createTestAccountBilling(t, companyID)

	// Create original billing to carry over from
	originalBilling, err := testRepoBilling.Create(ctx, accountID, &models.CreateBillingRequest{
		Period:       202601,
		AmountBilled: 50.00,
	})
	require.NoError(t, err)

	t.Run("creates carry over billing", func(t *testing.T) {
		carryOver, err := testRepoBilling.CreateCarryOver(ctx, accountID, 202602, 50.00, originalBilling.ID)

		require.NoError(t, err)
		require.NotNil(t, carryOver)
		assert.Equal(t, 202602, carryOver.Period)
		assert.Equal(t, 50.00, carryOver.AmountBilled)
		assert.NotNil(t, carryOver.CarriedFrom)
		assert.Equal(t, originalBilling.ID, *carryOver.CarriedFrom)
	})
}

func TestBillingRepo_GetByID_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	pgContainer := setupTestDBBilling(t)
	initTestRepoBilling(t, pgContainer)
	defer teardownTestDBBilling(t, pgContainer)

	categoryID := createTestCategoryBilling(t)
	companyID := createTestCompanyBilling(t, categoryID)
	accountID := createTestAccountBilling(t, companyID)

	createdBilling, err := testRepoBilling.Create(ctx, accountID, &models.CreateBillingRequest{
		Period:       202603,
		AmountBilled: 100.00,
	})
	require.NoError(t, err)

	t.Run("gets billing by id", func(t *testing.T) {
		billing, err := testRepoBilling.GetByID(ctx, createdBilling.ID, testUserIDBilling)

		require.NoError(t, err)
		require.NotNil(t, billing)
		assert.Equal(t, createdBilling.ID, billing.ID)
		assert.Equal(t, 202603, billing.Period)
	})

	t.Run("returns nil for non-existent billing", func(t *testing.T) {
		billing, err := testRepoBilling.GetByID(ctx, "00000000-0000-0000-0000-000000000001", testUserIDBilling)

		assert.NoError(t, err)
		assert.Nil(t, billing)
	})
}

func TestBillingRepo_GetByAccountAndPeriod_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	pgContainer := setupTestDBBilling(t)
	initTestRepoBilling(t, pgContainer)
	defer teardownTestDBBilling(t, pgContainer)

	categoryID := createTestCategoryBilling(t)
	companyID := createTestCompanyBilling(t, categoryID)
	accountID := createTestAccountBilling(t, companyID)

	_, err := testRepoBilling.Create(ctx, accountID, &models.CreateBillingRequest{
		Period:       202603,
		AmountBilled: 100.00,
	})
	require.NoError(t, err)

	t.Run("gets billing by account and period", func(t *testing.T) {
		billing, err := testRepoBilling.GetByAccountAndPeriod(ctx, accountID, 202603)

		require.NoError(t, err)
		require.NotNil(t, billing)
		assert.Equal(t, 202603, billing.Period)
		assert.Equal(t, accountID, billing.AccountID)
	})

	t.Run("returns nil for non-existent period", func(t *testing.T) {
		billing, err := testRepoBilling.GetByAccountAndPeriod(ctx, accountID, 209912)

		assert.NoError(t, err)
		assert.Nil(t, billing)
	})
}

func TestBillingRepo_GetAllByAccount_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	pgContainer := setupTestDBBilling(t)
	initTestRepoBilling(t, pgContainer)
	defer teardownTestDBBilling(t, pgContainer)

	categoryID := createTestCategoryBilling(t)
	companyID := createTestCompanyBilling(t, categoryID)
	accountID := createTestAccountBilling(t, companyID)

	// Create multiple billings
	for i := 0; i < 5; i++ {
		_, err := testRepoBilling.Create(ctx, accountID, &models.CreateBillingRequest{
			Period:       202601 + i,
			AmountBilled: 100.00 + float64(i*10),
		})
		require.NoError(t, err)
	}

	t.Run("gets all billings with pagination", func(t *testing.T) {
		pagination := models.PaginationParams{Page: 1, Limit: 2}

		billings, total, err := testRepoBilling.GetAllByAccount(ctx, accountID, testUserIDBilling, pagination)

		require.NoError(t, err)
		assert.Len(t, billings, 2)
		assert.Equal(t, 5, total)
	})

	t.Run("gets all billings without pagination limits", func(t *testing.T) {
		pagination := models.PaginationParams{Page: 1, Limit: 100}

		billings, total, err := testRepoBilling.GetAllByAccount(ctx, accountID, testUserIDBilling, pagination)

		require.NoError(t, err)
		assert.Len(t, billings, 5)
		assert.Equal(t, 5, total)
	})
}

func TestBillingRepo_GetUnpaidByAccount_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	pgContainer := setupTestDBBilling(t)
	initTestRepoBilling(t, pgContainer)
	defer teardownTestDBBilling(t, pgContainer)

	categoryID := createTestCategoryBilling(t)
	companyID := createTestCompanyBilling(t, categoryID)
	accountID := createTestAccountBilling(t, companyID)

	// Create paid billing
	_, err := testRepoBilling.Create(ctx, accountID, &models.CreateBillingRequest{
		Period:       202601,
		AmountBilled: 100.00,
	})
	require.NoError(t, err)

	// Create unpaid billing
	_, err = testRepoBilling.Create(ctx, accountID, &models.CreateBillingRequest{
		Period:       202602,
		AmountBilled: 100.00,
	})
	require.NoError(t, err)

	t.Run("gets unpaid billing", func(t *testing.T) {
		billing, err := testRepoBilling.GetUnpaidByAccount(ctx, accountID)

		require.NoError(t, err)
		require.NotNil(t, billing)
		assert.False(t, billing.IsPaid)
		assert.Equal(t, 202602, billing.Period)
	})

	t.Run("returns nil when all paid", func(t *testing.T) {
		// Mark all as paid
		allBillings, _, err := testRepoBilling.GetAllByAccount(ctx, accountID, testUserIDBilling, models.PaginationParams{Page: 1, Limit: 100})
		require.NoError(t, err)

		for _, b := range allBillings {
			err := testRepoBilling.MarkPaid(ctx, b.ID)
			require.NoError(t, err)
		}

		billing, err := testRepoBilling.GetUnpaidByAccount(ctx, accountID)

		assert.NoError(t, err)
		assert.Nil(t, billing)
	})
}

func TestBillingRepo_GetAllByPeriod_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	pgContainer := setupTestDBBilling(t)
	initTestRepoBilling(t, pgContainer)
	defer teardownTestDBBilling(t, pgContainer)

	categoryID := createTestCategoryBilling(t)
	companyID := createTestCompanyBilling(t, categoryID)
	accountID := createTestAccountBilling(t, companyID)

	// Create billing for specific period
	_, err := testRepoBilling.Create(ctx, accountID, &models.CreateBillingRequest{
		Period:       202603,
		AmountBilled: 100.00,
	})
	require.NoError(t, err)

	t.Run("gets all billings by period", func(t *testing.T) {
		pagination := models.PaginationParams{Page: 1, Limit: 100}

		billings, total, err := testRepoBilling.GetAllByPeriod(ctx, testUserIDBilling, 202603, nil, pagination)

		require.NoError(t, err)
		assert.Len(t, billings, 1)
		assert.Equal(t, 1, total)
		assert.Equal(t, "Test Account", billings[0].AccountName)
		assert.Equal(t, "Test Company", billings[0].CompanyName)
	})

	t.Run("filters by paid status", func(t *testing.T) {
		isPaid := true
		pagination := models.PaginationParams{Page: 1, Limit: 100}

		billings, total, err := testRepoBilling.GetAllByPeriod(ctx, testUserIDBilling, 202603, &isPaid, pagination)

		require.NoError(t, err)
		assert.Len(t, billings, 0)
		assert.Equal(t, 0, total)
	})
}

func TestBillingRepo_Update_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	pgContainer := setupTestDBBilling(t)
	initTestRepoBilling(t, pgContainer)
	defer teardownTestDBBilling(t, pgContainer)

	categoryID := createTestCategoryBilling(t)
	companyID := createTestCompanyBilling(t, categoryID)
	accountID := createTestAccountBilling(t, companyID)

	createdBilling, err := testRepoBilling.Create(ctx, accountID, &models.CreateBillingRequest{
		Period:       202603,
		AmountBilled: 100.00,
	})
	require.NoError(t, err)

	t.Run("updates billing successfully", func(t *testing.T) {
		newAmount := 150.00

		req := &models.UpdateBillingRequest{
			AmountBilled: &newAmount,
		}

		billing, err := testRepoBilling.Update(ctx, createdBilling.ID, testUserIDBilling, req)

		require.NoError(t, err)
		require.NotNil(t, billing)
		assert.Equal(t, 150.00, billing.AmountBilled)
	})

	t.Run("updates paid status", func(t *testing.T) {
		isPaid := true

		req := &models.UpdateBillingRequest{
			IsPaid: &isPaid,
		}

		billing, err := testRepoBilling.Update(ctx, createdBilling.ID, testUserIDBilling, req)

		require.NoError(t, err)
		require.NotNil(t, billing)
		assert.True(t, billing.IsPaid)
		// Note: Update doesn't auto-set PaidAt, it only sets if provided in request
	})

	t.Run("returns nil for non-existent billing", func(t *testing.T) {
		req := &models.UpdateBillingRequest{}

		billing, err := testRepoBilling.Update(ctx, "00000000-0000-0000-0000-000000000001", testUserIDBilling, req)

		assert.NoError(t, err)
		assert.Nil(t, billing)
	})
}

func TestBillingRepo_MarkPaid_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	pgContainer := setupTestDBBilling(t)
	initTestRepoBilling(t, pgContainer)
	defer teardownTestDBBilling(t, pgContainer)

	categoryID := createTestCategoryBilling(t)
	companyID := createTestCompanyBilling(t, categoryID)
	accountID := createTestAccountBilling(t, companyID)

	createdBilling, err := testRepoBilling.Create(ctx, accountID, &models.CreateBillingRequest{
		Period:       202603,
		AmountBilled: 100.00,
	})
	require.NoError(t, err)

	t.Run("marks billing as paid", func(t *testing.T) {
		err := testRepoBilling.MarkPaid(ctx, createdBilling.ID)

		require.NoError(t, err)

		billing, err := testRepoBilling.GetByID(ctx, createdBilling.ID, testUserIDBilling)
		require.NoError(t, err)
		require.NotNil(t, billing)
		assert.True(t, billing.IsPaid)
		assert.NotNil(t, billing.PaidAt)
	})
}

func TestBillingRepo_BulkInsertForPeriod_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	pgContainer := setupTestDBBilling(t)
	initTestRepoBilling(t, pgContainer)
	defer teardownTestDBBilling(t, pgContainer)

	categoryID := createTestCategoryBilling(t)
	companyID := createTestCompanyBilling(t, categoryID)
	accountID1 := createTestAccountBilling(t, companyID)

	// Create second account for bulk insert
	var accountID2 string
	err := testPoolBilling.QueryRow(ctx, `
		INSERT INTO homepay.accounts (company_id, name, billing_day)
		VALUES ($1, 'Test Account 2', 15)
		RETURNING id
	`, companyID).Scan(&accountID2)
	require.NoError(t, err)

	t.Run("bulk inserts billings for period", func(t *testing.T) {
		inserts := []models.PeriodBillingInsert{
			{AccountID: accountID1, AmountBilled: 100.00},
			{AccountID: accountID2, AmountBilled: 150.00},
		}

		err := testRepoBilling.BulkInsertForPeriod(ctx, 202603, inserts)

		require.NoError(t, err)

		// Verify
		billing1, err := testRepoBilling.GetByAccountAndPeriod(ctx, accountID1, 202603)
		require.NoError(t, err)
		require.NotNil(t, billing1)
		assert.Equal(t, 100.00, billing1.AmountBilled)

		billing2, err := testRepoBilling.GetByAccountAndPeriod(ctx, accountID2, 202603)
		require.NoError(t, err)
		require.NotNil(t, billing2)
		assert.Equal(t, 150.00, billing2.AmountBilled)
	})
}

func TestBillingRepo_SoftDeleteByAccount_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	pgContainer := setupTestDBBilling(t)
	initTestRepoBilling(t, pgContainer)
	defer teardownTestDBBilling(t, pgContainer)

	categoryID := createTestCategoryBilling(t)
	companyID := createTestCompanyBilling(t, categoryID)
	accountID := createTestAccountBilling(t, companyID)

	// Create billings
	for i := 0; i < 3; i++ {
		_, err := testRepoBilling.Create(ctx, accountID, &models.CreateBillingRequest{
			Period:       202601 + i,
			AmountBilled: 100.00,
		})
		require.NoError(t, err)
	}

	t.Run("soft deletes all billings by account", func(t *testing.T) {
		err := testRepoBilling.SoftDeleteByAccount(ctx, accountID)

		require.NoError(t, err)

		billings, _, err := testRepoBilling.GetAllByAccount(ctx, accountID, testUserIDBilling, models.PaginationParams{Page: 1, Limit: 100})
		require.NoError(t, err)
		assert.Len(t, billings, 0)
	})
}
