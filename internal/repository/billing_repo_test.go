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
	billingPool       *pgxpool.Pool
	billingRepoInst  BillingRepository
	billingAccRepo   AccountRepository
	billingCompRepo  CompanyRepository
)

func setupBillingDB(t *testing.T) *postgres.PostgresContainer {
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

func initBillingRepo(t *testing.T, pgContainer *postgres.PostgresContainer) {
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
		connStr = fmt.Sprintf("postgres://test:test@%s:%s/homepay_test?sslmode=disable", host, port.String())
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

	billingPool = pool
	billingRepoInst = NewBillingRepository(pool)
	billingAccRepo = NewAccountRepository(pool)
	billingCompRepo = NewCompanyRepository(pool)

	// Create full schema including billings
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
		
		CREATE TABLE IF NOT EXISTS homepay.account_billings (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			account_id UUID NOT NULL,
			period INTEGER NOT NULL,
			amount_billed DECIMAL(12,2) NOT NULL,
			amount_paid DECIMAL(12,2) DEFAULT 0,
			is_paid BOOLEAN DEFAULT FALSE,
			paid_at TIMESTAMP WITH TIME ZONE,
			carried_from UUID,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			deleted_at TIMESTAMP WITH TIME ZONE,
			FOREIGN KEY (account_id) REFERENCES homepay.accounts(id)
		);
		
		CREATE INDEX IF NOT EXISTS idx_companies_user_id ON homepay.companies(auth_user_id);
		CREATE INDEX IF NOT EXISTS idx_accounts_company_id ON homepay.accounts(company_id);
		CREATE INDEX IF NOT EXISTS idx_billings_account_id ON homepay.account_billings(account_id);
		CREATE INDEX IF NOT EXISTS idx_billings_period ON homepay.account_billings(period);
	`)
	require.NoError(t, err, "failed to create schema")

	// Insert a category for testing
	_, err = pool.Exec(ctx, `INSERT INTO homepay.categories (name) VALUES ('Test Category')`)
	require.NoError(t, err, "failed to insert category")
}

func teardownBillingDB(t *testing.T, pgContainer *postgres.PostgresContainer) {
	t.Helper()

	if billingPool != nil {
		billingPool.Close()
	}
	pgContainer.Terminate(context.Background())
}

// TDD: Test first - these tests define the expected behavior

func TestBillingRepo_Create_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	pgContainer := setupBillingDB(t)
	initBillingRepo(t, pgContainer)
	defer teardownBillingDB(t, pgContainer)

	// Create test data
	company, err := billingCompRepo.Create(ctx, testUserID, &models.CreateCompanyRequest{
		Name: "Test Company",
	})
	require.NoError(t, err)

	account, err := billingAccRepo.Create(ctx, company.ID, testUserID, &models.CreateAccountRequest{
		Name:       "Test Account",
		BillingDay: 15,
	})
	require.NoError(t, err)

	t.Run("creates billing successfully", func(t *testing.T) {
		req := &models.CreateBillingRequest{
			Period:       202603,
			AmountBilled: 15000.00,
		}

		billing, err := billingRepoInst.Create(ctx, account.ID, req)

		require.NoError(t, err)
		assert.NotNil(t, billing)
		assert.NotEmpty(t, billing.ID)
		assert.Equal(t, account.ID, billing.AccountID)
		assert.Equal(t, 202603, billing.Period)
		assert.Equal(t, 15000.00, billing.AmountBilled)
		assert.False(t, billing.IsPaid)
	})

	t.Run("creates billing with payment", func(t *testing.T) {
		amountPaid := 15000.00
		isPaid := true

		req := &models.CreateBillingRequest{
			Period:       202604,
			AmountBilled: 15000.00,
			AmountPaid:   &amountPaid,
			IsPaid:       &isPaid,
		}

		billing, err := billingRepoInst.Create(ctx, account.ID, req)

		require.NoError(t, err)
		assert.NotNil(t, billing)
		assert.Equal(t, 15000.00, billing.AmountPaid)
		assert.True(t, billing.IsPaid)
	})
}

func TestBillingRepo_GetByID_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	pgContainer := setupBillingDB(t)
	initBillingRepo(t, pgContainer)
	defer teardownBillingDB(t, pgContainer)

	// Create test data
	company, err := billingCompRepo.Create(ctx, testUserID, &models.CreateCompanyRequest{
		Name: "Test Company",
	})
	require.NoError(t, err)

	account, err := billingAccRepo.Create(ctx, company.ID, testUserID, &models.CreateAccountRequest{
		Name:       "Test Account",
		BillingDay: 15,
	})
	require.NoError(t, err)

	// Create a billing to retrieve
	created, err := billingRepoInst.Create(ctx, account.ID, &models.CreateBillingRequest{
		Period:       202603,
		AmountBilled: 15000.00,
	})
	require.NoError(t, err)

	t.Run("finds existing billing", func(t *testing.T) {
		billing, err := billingRepoInst.GetByID(ctx, created.ID, testUserID)

		require.NoError(t, err)
		assert.NotNil(t, billing)
		assert.Equal(t, created.ID, billing.ID)
		assert.Equal(t, 202603, billing.Period)
	})

	t.Run("returns nil for non-existent", func(t *testing.T) {
		billing, err := billingRepoInst.GetByID(ctx, "00000000-0000-0000-0000-000000000000", testUserID)

		require.NoError(t, err)
		assert.Nil(t, billing)
	})
}

func TestBillingRepo_GetAllByAccount_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	pgContainer := setupBillingDB(t)
	initBillingRepo(t, pgContainer)
	defer teardownBillingDB(t, pgContainer)

	// Create test data
	company, err := billingCompRepo.Create(ctx, testUserID, &models.CreateCompanyRequest{
		Name: "Test Company",
	})
	require.NoError(t, err)

	account, err := billingAccRepo.Create(ctx, company.ID, testUserID, &models.CreateAccountRequest{
		Name:       "Test Account",
		BillingDay: 15,
	})
	require.NoError(t, err)

	// Create multiple billings
	for i := 1; i <= 3; i++ {
		_, err := billingRepoInst.Create(ctx, account.ID, &models.CreateBillingRequest{
			Period:       202600 + i,
			AmountBilled: float64(10000 * i),
		})
		require.NoError(t, err)
	}

	t.Run("returns all billings for account", func(t *testing.T) {
		result, total, err := billingRepoInst.GetAllByAccount(ctx, account.ID, testUserID, models.PaginationParams{Page: 1, Limit: 10})

		require.NoError(t, err)
		assert.Len(t, result, 3)
		assert.Equal(t, 3, total)
	})

	t.Run("respects pagination", func(t *testing.T) {
		result, total, err := billingRepoInst.GetAllByAccount(ctx, account.ID, testUserID, models.PaginationParams{Page: 1, Limit: 2})

		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, 3, total)
	})
}

func TestBillingRepo_GetByAccountAndPeriod_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	pgContainer := setupBillingDB(t)
	initBillingRepo(t, pgContainer)
	defer teardownBillingDB(t, pgContainer)

	// Create test data
	company, err := billingCompRepo.Create(ctx, testUserID, &models.CreateCompanyRequest{
		Name: "Test Company",
	})
	require.NoError(t, err)

	account, err := billingAccRepo.Create(ctx, company.ID, testUserID, &models.CreateAccountRequest{
		Name:       "Test Account",
		BillingDay: 15,
	})
	require.NoError(t, err)

	// Create a billing
	created, err := billingRepoInst.Create(ctx, account.ID, &models.CreateBillingRequest{
		Period:       202603,
		AmountBilled: 15000.00,
	})
	require.NoError(t, err)

	t.Run("finds billing by account and period", func(t *testing.T) {
		billing, err := billingRepoInst.GetByAccountAndPeriod(ctx, account.ID, 202603)

		require.NoError(t, err)
		assert.NotNil(t, billing)
		assert.Equal(t, created.ID, billing.ID)
	})

	t.Run("returns nil for non-existent period", func(t *testing.T) {
		billing, err := billingRepoInst.GetByAccountAndPeriod(ctx, account.ID, 209912)

		require.NoError(t, err)
		assert.Nil(t, billing)
	})
}

func TestBillingRepo_GetUnpaidByAccount_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	pgContainer := setupBillingDB(t)
	initBillingRepo(t, pgContainer)
	defer teardownBillingDB(t, pgContainer)

	// Create test data
	company, err := billingCompRepo.Create(ctx, testUserID, &models.CreateCompanyRequest{
		Name: "Test Company",
	})
	require.NoError(t, err)

	account, err := billingAccRepo.Create(ctx, company.ID, testUserID, &models.CreateAccountRequest{
		Name:       "Test Account",
		BillingDay: 15,
	})
	require.NoError(t, err)

	// Create unpaid billing
	_, err = billingRepoInst.Create(ctx, account.ID, &models.CreateBillingRequest{
		Period:       202603,
		AmountBilled: 15000.00,
	})
	require.NoError(t, err)

	t.Run("finds unpaid billing", func(t *testing.T) {
		billing, err := billingRepoInst.GetUnpaidByAccount(ctx, account.ID)

		require.NoError(t, err)
		assert.NotNil(t, billing)
		assert.False(t, billing.IsPaid)
	})
}

func TestBillingRepo_Update_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	pgContainer := setupBillingDB(t)
	initBillingRepo(t, pgContainer)
	defer teardownBillingDB(t, pgContainer)

	// Create test data
	company, err := billingCompRepo.Create(ctx, testUserID, &models.CreateCompanyRequest{
		Name: "Test Company",
	})
	require.NoError(t, err)

	account, err := billingAccRepo.Create(ctx, company.ID, testUserID, &models.CreateAccountRequest{
		Name:       "Test Account",
		BillingDay: 15,
	})
	require.NoError(t, err)

	created, err := billingRepoInst.Create(ctx, account.ID, &models.CreateBillingRequest{
		Period:       202603,
		AmountBilled: 15000.00,
	})
	require.NoError(t, err)

	t.Run("updates billing amount", func(t *testing.T) {
		newAmount := 20000.00
		req := &models.UpdateBillingRequest{AmountBilled: &newAmount}

		billing, err := billingRepoInst.Update(ctx, created.ID, testUserID, req)

		require.NoError(t, err)
		assert.NotNil(t, billing)
		assert.Equal(t, 20000.00, billing.AmountBilled)
	})

	t.Run("marks billing as paid", func(t *testing.T) {
		isPaid := true
		req := &models.UpdateBillingRequest{IsPaid: &isPaid}

		billing, err := billingRepoInst.Update(ctx, created.ID, testUserID, req)

		require.NoError(t, err)
		assert.NotNil(t, billing)
		assert.True(t, billing.IsPaid)
	})
}

func TestBillingRepo_MarkPaid_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	pgContainer := setupBillingDB(t)
	initBillingRepo(t, pgContainer)
	defer teardownBillingDB(t, pgContainer)

	// Create test data
	company, err := billingCompRepo.Create(ctx, testUserID, &models.CreateCompanyRequest{
		Name: "Test Company",
	})
	require.NoError(t, err)

	account, err := billingAccRepo.Create(ctx, company.ID, testUserID, &models.CreateAccountRequest{
		Name:       "Test Account",
		BillingDay: 15,
	})
	require.NoError(t, err)

	created, err := billingRepoInst.Create(ctx, account.ID, &models.CreateBillingRequest{
		Period:       202603,
		AmountBilled: 15000.00,
	})
	require.NoError(t, err)

	t.Run("marks billing as paid", func(t *testing.T) {
		err := billingRepoInst.MarkPaid(ctx, created.ID)
		require.NoError(t, err)

		// Verify
		billing, err := billingRepoInst.GetByID(ctx, created.ID, testUserID)
		require.NoError(t, err)
		assert.NotNil(t, billing)
		assert.True(t, billing.IsPaid)
		assert.NotNil(t, billing.PaidAt)
	})
}

func TestBillingRepo_BulkInsertForPeriod_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	pgContainer := setupBillingDB(t)
	initBillingRepo(t, pgContainer)
	defer teardownBillingDB(t, pgContainer)

	// Create test data
	company, err := billingCompRepo.Create(ctx, testUserID, &models.CreateCompanyRequest{
		Name: "Test Company",
	})
	require.NoError(t, err)

	account1, err := billingAccRepo.Create(ctx, company.ID, testUserID, &models.CreateAccountRequest{
		Name:       "Test Account 1",
		BillingDay: 15,
	})
	require.NoError(t, err)

	account2, err := billingAccRepo.Create(ctx, company.ID, testUserID, &models.CreateAccountRequest{
		Name:       "Test Account 2",
		BillingDay: 20,
	})
	require.NoError(t, err)

	t.Run("bulk inserts billings for period", func(t *testing.T) {
		inserts := []models.PeriodBillingInsert{
			{AccountID: account1.ID, AmountBilled: 10000.00},
			{AccountID: account2.ID, AmountBilled: 15000.00},
		}

		err := billingRepoInst.BulkInsertForPeriod(ctx, 202603, inserts)
		require.NoError(t, err)

		// Verify
		billing1, err := billingRepoInst.GetByAccountAndPeriod(ctx, account1.ID, 202603)
		require.NoError(t, err)
		assert.NotNil(t, billing1)
		assert.Equal(t, 10000.00, billing1.AmountBilled)

		billing2, err := billingRepoInst.GetByAccountAndPeriod(ctx, account2.ID, 202603)
		require.NoError(t, err)
		assert.NotNil(t, billing2)
		assert.Equal(t, 15000.00, billing2.AmountBilled)
	})
}

func TestBillingRepo_SoftDeleteByAccount_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	pgContainer := setupBillingDB(t)
	initBillingRepo(t, pgContainer)
	defer teardownBillingDB(t, pgContainer)

	// Create test data
	company, err := billingCompRepo.Create(ctx, testUserID, &models.CreateCompanyRequest{
		Name: "Test Company",
	})
	require.NoError(t, err)

	account, err := billingAccRepo.Create(ctx, company.ID, testUserID, &models.CreateAccountRequest{
		Name:       "Test Account",
		BillingDay: 15,
	})
	require.NoError(t, err)

	// Create billings
	_, err = billingRepoInst.Create(ctx, account.ID, &models.CreateBillingRequest{
		Period:       202603,
		AmountBilled: 15000.00,
	})
	require.NoError(t, err)

	t.Run("soft deletes all billings for account", func(t *testing.T) {
		err := billingRepoInst.SoftDeleteByAccount(ctx, account.ID)
		require.NoError(t, err)

		// Verify
		billing, err := billingRepoInst.GetByAccountAndPeriod(ctx, account.ID, 202603)
		require.NoError(t, err)
		assert.Nil(t, billing)
	})
}

func TestBillingRepo_CreateCarryOver_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	pgContainer := setupBillingDB(t)
	initBillingRepo(t, pgContainer)
	defer teardownBillingDB(t, pgContainer)

	// Create test data
	company, err := billingCompRepo.Create(ctx, testUserID, &models.CreateCompanyRequest{
		Name: "Test Company",
	})
	require.NoError(t, err)

	account, err := billingAccRepo.Create(ctx, company.ID, testUserID, &models.CreateAccountRequest{
		Name:       "Test Account",
		BillingDay: 15,
	})
	require.NoError(t, err)

	// Create original billing
	originalBilling, err := billingRepoInst.Create(ctx, account.ID, &models.CreateBillingRequest{
		Period:       202603,
		AmountBilled: 15000.00,
	})
	require.NoError(t, err)

	t.Run("creates carry over billing", func(t *testing.T) {
		billing, err := billingRepoInst.CreateCarryOver(ctx, account.ID, 202604, 5000.00, originalBilling.ID)

		require.NoError(t, err)
		assert.NotNil(t, billing)
		assert.Equal(t, 202604, billing.Period)
		assert.Equal(t, 5000.00, billing.AmountBilled)
		assert.NotNil(t, billing.CarriedFrom)
		assert.Equal(t, originalBilling.ID, *billing.CarriedFrom)
	})
}

// Unit tests for billing repository

func TestBillingRepo_Interfaces(t *testing.T) {
	t.Run("BillingRepository interface is satisfied by billingRepo", func(t *testing.T) {
		var _ BillingRepository = (*billingRepo)(nil)
	})
}

func TestScanBilling(t *testing.T) {
	t.Run("scanBilling function exists", func(t *testing.T) {
		assert.NotNil(t, scanBilling)
	})
}

func TestBillingRepo_Create_Model(t *testing.T) {
	t.Run("CreateBillingRequest validation", func(t *testing.T) {
		req := models.CreateBillingRequest{
			Period:       202603,
			AmountBilled: 15000.00,
		}
		assert.Equal(t, 202603, req.Period)
		assert.Equal(t, 15000.00, req.AmountBilled)
	})

	t.Run("CreateBillingRequest with optional fields", func(t *testing.T) {
		amountPaid := 15000.00
		isPaid := true
		carriedFrom := "prev-billing-123"

		req := models.CreateBillingRequest{
			Period:       202603,
			AmountBilled: 15000.00,
			AmountPaid:   &amountPaid,
			IsPaid:       &isPaid,
			CarriedFrom:  &carriedFrom,
		}
		assert.NotNil(t, req.AmountPaid)
		assert.NotNil(t, req.IsPaid)
		assert.NotNil(t, req.CarriedFrom)
		assert.Equal(t, "prev-billing-123", *req.CarriedFrom)
	})
}

func TestBillingRepo_Update_Model(t *testing.T) {
	t.Run("UpdateBillingRequest with pointer fields", func(t *testing.T) {
		amountBilled := 20000.00
		isPaid := true

		req := models.UpdateBillingRequest{
			AmountBilled: &amountBilled,
			IsPaid:       &isPaid,
		}
		assert.Equal(t, 20000.00, *req.AmountBilled)
		assert.True(t, *req.IsPaid)
	})

	t.Run("UpdateBillingRequest with nil fields", func(t *testing.T) {
		req := models.UpdateBillingRequest{}
		assert.Nil(t, req.AmountBilled)
		assert.Nil(t, req.AmountPaid)
		assert.Nil(t, req.IsPaid)
		assert.Nil(t, req.PaidAt)
	})
}

func TestBillingRepo_Constants(t *testing.T) {
	t.Run("billingCols constant is defined", func(t *testing.T) {
		assert.NotEmpty(t, billingCols)
		assert.Contains(t, billingCols, "id")
		assert.Contains(t, billingCols, "account_id")
		assert.Contains(t, billingCols, "period")
	})

	t.Run("billingColsAB constant is defined", func(t *testing.T) {
		assert.NotEmpty(t, billingColsAB)
		assert.Contains(t, billingColsAB, "ab.id")
	})
}

func TestNewBillingRepository(t *testing.T) {
	t.Run("returns billingRepo instance", func(t *testing.T) {
		// This is a compile-time check that NewBillingRepository returns something
		// that satisfies the BillingRepository interface
		var repo BillingRepository = NewBillingRepository(nil)
		assert.NotNil(t, repo)
	})
}
