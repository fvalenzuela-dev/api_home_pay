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
	testPoolExpense   *pgxpool.Pool
	testRepoExpense   ExpenseRepository
	testUserIDExpense = "test-user-integration-expense"
)

func setupTestDBExpense(t *testing.T) *postgres.PostgresContainer {
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

func initTestRepoExpense(t *testing.T, pgContainer *postgres.PostgresContainer) {
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

	testPoolExpense = pool
	testRepoExpense = NewExpenseRepository(pool)

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

		CREATE TABLE IF NOT EXISTS homepay.variable_expenses (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			auth_user_id VARCHAR(255) NOT NULL,
			company_id UUID REFERENCES homepay.companies(id),
			description VARCHAR(255) NOT NULL,
			amount DECIMAL(12,2) NOT NULL,
			expense_date DATE NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			deleted_at TIMESTAMPTZ
		);

		CREATE INDEX IF NOT EXISTS idx_companies_user ON homepay.companies(auth_user_id);
		CREATE INDEX IF NOT EXISTS idx_variable_expenses_user ON homepay.variable_expenses(auth_user_id);
		CREATE INDEX IF NOT EXISTS idx_variable_expenses_date ON homepay.variable_expenses(expense_date);
		CREATE INDEX IF NOT EXISTS idx_variable_expenses_company ON homepay.variable_expenses(company_id);
	`)
	require.NoError(t, err, "failed to create schema")
}

func teardownTestDBExpense(t *testing.T, pgContainer *postgres.PostgresContainer) {
	t.Helper()

	if testPoolExpense != nil {
		testPoolExpense.Close()
	}
	pgContainer.Terminate(context.Background())
}

func createTestCompanyExpense(t *testing.T) string {
	ctx := context.Background()
	var companyID string
	err := testPoolExpense.QueryRow(ctx, `
		INSERT INTO homepay.companies (auth_user_id, name)
		VALUES ($1, 'Test Company')
		RETURNING id
	`, testUserIDExpense).Scan(&companyID)
	require.NoError(t, err)
	return companyID
}

func TestExpenseRepo_Create_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	pgContainer := setupTestDBExpense(t)
	initTestRepoExpense(t, pgContainer)
	defer teardownTestDBExpense(t, pgContainer)

	t.Run("creates expense successfully", func(t *testing.T) {
		req := &models.CreateExpenseRequest{
			Description: "Test Expense",
			Amount:      100.50,
			ExpenseDate: "2024-03-15",
		}

		expense, err := testRepoExpense.Create(ctx, testUserIDExpense, req)

		require.NoError(t, err)
		require.NotNil(t, expense)
		assert.Equal(t, "Test Expense", expense.Description)
		assert.Equal(t, 100.50, expense.Amount)
		assert.Equal(t, testUserIDExpense, expense.AuthUserID)
		assert.NotEmpty(t, expense.ID)
	})

	t.Run("creates expense with company", func(t *testing.T) {
		companyID := createTestCompanyExpense(t)

		req := &models.CreateExpenseRequest{
			CompanyID:   &companyID,
			Description: "Expense with company",
			Amount:      200.00,
			ExpenseDate: "2024-03-20",
		}

		expense, err := testRepoExpense.Create(ctx, testUserIDExpense, req)

		require.NoError(t, err)
		require.NotNil(t, expense)
		assert.NotNil(t, expense.CompanyID)
		assert.Equal(t, companyID, *expense.CompanyID)
	})

	t.Run("fails with invalid date format", func(t *testing.T) {
		req := &models.CreateExpenseRequest{
			Description: "Invalid date",
			Amount:      50.00,
			ExpenseDate: "15/03/2024",
		}

		expense, err := testRepoExpense.Create(ctx, testUserIDExpense, req)

		assert.Nil(t, expense)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid expense_date format")
	})
}

func TestExpenseRepo_GetByID_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	pgContainer := setupTestDBExpense(t)
	initTestRepoExpense(t, pgContainer)
	defer teardownTestDBExpense(t, pgContainer)

	var createdExpenseID string
	err := testPoolExpense.QueryRow(ctx, `
		INSERT INTO homepay.variable_expenses (auth_user_id, description, amount, expense_date)
		VALUES ($1, 'Test Expense', 150.75, '2024-03-10')
		RETURNING id
	`, testUserIDExpense).Scan(&createdExpenseID)
	require.NoError(t, err)

	t.Run("gets expense by id", func(t *testing.T) {
		expense, err := testRepoExpense.GetByID(ctx, createdExpenseID, testUserIDExpense)

		require.NoError(t, err)
		require.NotNil(t, expense)
		assert.Equal(t, createdExpenseID, expense.ID)
		assert.Equal(t, "Test Expense", expense.Description)
		assert.Equal(t, 150.75, expense.Amount)
	})

	t.Run("returns nil for non-existent expense", func(t *testing.T) {
		expense, err := testRepoExpense.GetByID(ctx, "00000000-0000-0000-0000-000000000001", testUserIDExpense)

		assert.NoError(t, err)
		assert.Nil(t, expense)
	})

	t.Run("returns nil for expense from different user", func(t *testing.T) {
		expense, err := testRepoExpense.GetByID(ctx, createdExpenseID, "different-user")

		assert.NoError(t, err)
		assert.Nil(t, expense)
	})
}

func TestExpenseRepo_GetAll_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	pgContainer := setupTestDBExpense(t)
	initTestRepoExpense(t, pgContainer)
	defer teardownTestDBExpense(t, pgContainer)

	companyID := createTestCompanyExpense(t)

	for i := 0; i < 5; i++ {
		_, err := testPoolExpense.Exec(ctx, `
			INSERT INTO homepay.variable_expenses (auth_user_id, company_id, description, amount, expense_date)
			VALUES ($1, $2, $3, $4, $5)
		`, testUserIDExpense, companyID, fmt.Sprintf("Expense %d", i), 100.0+float64(i)*10, fmt.Sprintf("2024-03-%02d", 10+i))
		require.NoError(t, err)
	}

	t.Run("gets all expenses with pagination", func(t *testing.T) {
		pagination := models.PaginationParams{Page: 1, Limit: 2}

		expenses, total, err := testRepoExpense.GetAll(ctx, testUserIDExpense, models.ExpenseFilters{}, pagination)

		require.NoError(t, err)
		assert.Len(t, expenses, 2)
		assert.Equal(t, 5, total)
	})

	t.Run("gets all expenses without pagination limits", func(t *testing.T) {
		pagination := models.PaginationParams{Page: 1, Limit: 100}

		expenses, total, err := testRepoExpense.GetAll(ctx, testUserIDExpense, models.ExpenseFilters{}, pagination)

		require.NoError(t, err)
		assert.Len(t, expenses, 5)
		assert.Equal(t, 5, total)
	})

	t.Run("filters by month and year", func(t *testing.T) {
		_, err := testPoolExpense.Exec(ctx, `
			INSERT INTO homepay.variable_expenses (auth_user_id, description, amount, expense_date)
			VALUES ($1, 'April Expense', 500.00, '2024-04-15')
		`, testUserIDExpense)
		require.NoError(t, err)

		month := 3
		year := 2024
		pagination := models.PaginationParams{Page: 1, Limit: 100}
		filters := models.ExpenseFilters{Month: &month, Year: &year}

		expenses, total, err := testRepoExpense.GetAll(ctx, testUserIDExpense, filters, pagination)

		require.NoError(t, err)
		assert.Len(t, expenses, 5)
		assert.Equal(t, 5, total)
	})

	t.Run("filters by company", func(t *testing.T) {
		otherCompanyID := createTestCompanyExpense(t)
		_, err := testPoolExpense.Exec(ctx, `
			INSERT INTO homepay.variable_expenses (auth_user_id, company_id, description, amount, expense_date)
			VALUES ($1, $2, 'Other Company Expense', 999.00, '2024-03-05')
		`, testUserIDExpense, otherCompanyID)
		require.NoError(t, err)

		pagination := models.PaginationParams{Page: 1, Limit: 100}
		filters := models.ExpenseFilters{CompanyID: &companyID}

		expenses, total, err := testRepoExpense.GetAll(ctx, testUserIDExpense, filters, pagination)

		require.NoError(t, err)
		assert.Len(t, expenses, 5)
		assert.Equal(t, 5, total)
	})

	t.Run("returns empty for non-existent user", func(t *testing.T) {
		pagination := models.PaginationParams{Page: 1, Limit: 100}

		expenses, total, err := testRepoExpense.GetAll(ctx, "non-existent-user", models.ExpenseFilters{}, pagination)

		require.NoError(t, err)
		assert.Len(t, expenses, 0)
		assert.Equal(t, 0, total)
	})
}

func TestExpenseRepo_Update_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	pgContainer := setupTestDBExpense(t)
	initTestRepoExpense(t, pgContainer)
	defer teardownTestDBExpense(t, pgContainer)

	var expenseID string
	err := testPoolExpense.QueryRow(ctx, `
		INSERT INTO homepay.variable_expenses (auth_user_id, description, amount, expense_date)
		VALUES ($1, 'Original Expense', 100.00, '2024-03-01')
		RETURNING id
	`, testUserIDExpense).Scan(&expenseID)
	require.NoError(t, err)

	t.Run("updates expense successfully", func(t *testing.T) {
		newDescription := "Updated Expense"
		newAmount := 250.75

		req := &models.UpdateExpenseRequest{
			Description: &newDescription,
			Amount:      &newAmount,
		}

		expense, err := testRepoExpense.Update(ctx, expenseID, testUserIDExpense, req)

		require.NoError(t, err)
		require.NotNil(t, expense)
		assert.Equal(t, newDescription, expense.Description)
		assert.Equal(t, newAmount, expense.Amount)
	})

	t.Run("updates expense date", func(t *testing.T) {
		newDate := "2024-06-15"

		req := &models.UpdateExpenseRequest{
			ExpenseDate: &newDate,
		}

		expense, err := testRepoExpense.Update(ctx, expenseID, testUserIDExpense, req)

		require.NoError(t, err)
		require.NotNil(t, expense)
		assert.Equal(t, "2024-06-15", expense.ExpenseDate.Format("2006-01-02"))
	})

	t.Run("returns nil for non-existent expense", func(t *testing.T) {
		req := &models.UpdateExpenseRequest{
			Description: strPtr("New Name"),
		}

		expense, err := testRepoExpense.Update(ctx, "00000000-0000-0000-0000-000000000001", testUserIDExpense, req)

		assert.NoError(t, err)
		assert.Nil(t, expense)
	})

	t.Run("fails with invalid date format", func(t *testing.T) {
		invalidDate := "15/06/2024"

		req := &models.UpdateExpenseRequest{
			ExpenseDate: &invalidDate,
		}

		expense, err := testRepoExpense.Update(ctx, expenseID, testUserIDExpense, req)

		assert.Nil(t, expense)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid expense_date format")
	})
}

func TestExpenseRepo_SoftDelete_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	pgContainer := setupTestDBExpense(t)
	initTestRepoExpense(t, pgContainer)
	defer teardownTestDBExpense(t, pgContainer)

	var expenseID string
	err := testPoolExpense.QueryRow(ctx, `
		INSERT INTO homepay.variable_expenses (auth_user_id, description, amount, expense_date)
		VALUES ($1, 'To Delete', 100.00, '2024-03-01')
		RETURNING id
	`, testUserIDExpense).Scan(&expenseID)
	require.NoError(t, err)

	t.Run("soft deletes expense", func(t *testing.T) {
		err := testRepoExpense.SoftDelete(ctx, expenseID, testUserIDExpense)

		require.NoError(t, err)

		expense, err := testRepoExpense.GetByID(ctx, expenseID, testUserIDExpense)
		require.NoError(t, err)
		assert.Nil(t, expense)
	})

	t.Run("returns error for non-existent expense", func(t *testing.T) {
		err := testRepoExpense.SoftDelete(ctx, "00000000-0000-0000-0000-000000000001", testUserIDExpense)

		assert.Error(t, err)
	})

	t.Run("returns error when expense already deleted", func(t *testing.T) {
		err := testRepoExpense.SoftDelete(ctx, expenseID, testUserIDExpense)

		assert.Error(t, err)
	})
}
