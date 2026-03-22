package repository

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/fernandovalenzuela/api-home-pay/internal/models"
	"github.com/stretchr/testify/assert"
)

func setupExpenseMockDB(t *testing.T) (*expenseRepository, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}

	repo := &expenseRepository{
		db:  db,
		ctx: context.Background(),
	}

	cleanup := func() {
		db.Close()
	}

	return repo, mock, cleanup
}

func TestExpenseRepository_Create_Success(t *testing.T) {
	repo, mock, cleanup := setupExpenseMockDB(t)
	defer cleanup()

	accountID := 1
	dueDate := "2024-06-15"
	installmentGroupID := "group123"

	expense := &models.Expense{
		CategoryID:         1,
		PeriodID:           1,
		AccountID:          &accountID,
		Description:        "Test Expense",
		DueDate:            &dueDate,
		CurrentAmount:      100.50,
		AmountPaid:         0,
		CurrentInstallment: 1,
		TotalInstallments:  1,
		InstallmentGroupID: &installmentGroupID,
		IsRecurring:        false,
		Notes:              "Test notes",
	}

	mock.ExpectQuery(`INSERT INTO expenses`).
		WithArgs(1, 1, 1, "Test Expense", "2024-06-15", 100.50, 0.0, 1, 1, "group123", false, "Test notes", "user123").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow(1, "2024-01-01T00:00:00Z", "2024-01-01T00:00:00Z"))

	err := repo.Create("user123", expense)

	assert.NoError(t, err)
	assert.Equal(t, 1, expense.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestExpenseRepository_Create_WithoutOptionalFields(t *testing.T) {
	repo, mock, cleanup := setupExpenseMockDB(t)
	defer cleanup()

	expense := &models.Expense{
		CategoryID:        1,
		PeriodID:          1,
		Description:       "Test Expense",
		CurrentAmount:     50.00,
		AmountPaid:        0,
		TotalInstallments: 1,
	}

	mock.ExpectQuery(`INSERT INTO expenses`).
		WithArgs(1, 1, nil, "Test Expense", nil, 50.00, 0.0, 0, 1, nil, false, "", "user123").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow(1, "2024-01-01T00:00:00Z", "2024-01-01T00:00:00Z"))

	err := repo.Create("user123", expense)

	assert.NoError(t, err)
	assert.Equal(t, 1, expense.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestExpenseRepository_Create_EmptyUserID(t *testing.T) {
	repo, mock, cleanup := setupExpenseMockDB(t)
	defer cleanup()

	expense := &models.Expense{Description: "Test"}

	err := repo.Create("", expense)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user_id is required")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestExpenseRepository_GetByID_Success(t *testing.T) {
	repo, mock, cleanup := setupExpenseMockDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT e.id, e.category_id, e.period_id, e.account_id, e.description, e.due_date, e.current_amount, e.amount_paid, e.current_installment, e.total_installments, e.installment_group_id, e.is_recurring, e.notes, e.created_at, e.updated_at, c.id, c.name, p.id, p.month_number, p.year_number, sa.id, sa.company_id, sa.account_identifier, sa.alias FROM expenses e LEFT JOIN categories c ON e.category_id = c.id LEFT JOIN periods p ON e.period_id = p.id LEFT JOIN service_accounts sa ON e.account_id = sa.id`).
		WithArgs(1, "user123").
		WillReturnRows(sqlmock.NewRows([]string{
			"e.id", "e.category_id", "e.period_id", "e.account_id", "e.description",
			"e.due_date", "e.current_amount", "e.amount_paid", "e.current_installment",
			"e.total_installments", "e.installment_group_id", "e.is_recurring", "e.notes",
			"e.created_at", "e.updated_at", "c.id", "c.name", "p.id", "p.month_number",
			"p.year_number", "sa.id", "sa.company_id", "sa.account_identifier", "sa.alias",
		}).AddRow(
			1, 1, 1, sql.NullInt64{Int64: 1, Valid: true}, "Test Expense",
			sql.NullString{String: "2024-06-15", Valid: true}, 100.50, 50.00, 1,
			1, sql.NullString{String: "group123", Valid: true}, false, sql.NullString{String: "Notes", Valid: true},
			"2024-01-01T00:00:00Z", "2024-01-01T00:00:00Z", 1, "Groceries", 1, sql.NullInt32{Int32: 6, Valid: true},
			2024, 1, sql.NullInt64{Int64: 1, Valid: true}, "ACC123", sql.NullString{String: "My Account", Valid: true},
		))

	expense, err := repo.GetByID("user123", 1)

	assert.NoError(t, err)
	assert.NotNil(t, expense)
	assert.Equal(t, 1, expense.ID)
	assert.Equal(t, "Test Expense", expense.Description)
	assert.Equal(t, 100.50, expense.CurrentAmount)
	assert.Equal(t, 50.00, expense.AmountPaid)
	assert.NotNil(t, expense.Category)
	assert.NotNil(t, expense.Period)
	assert.NotNil(t, expense.ServiceAccount)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestExpenseRepository_GetByID_NotFound(t *testing.T) {
	repo, mock, cleanup := setupExpenseMockDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT e.id, e.category_id`).
		WithArgs(999, "user123").
		WillReturnError(sql.ErrNoRows)

	expense, err := repo.GetByID("user123", 999)

	assert.NoError(t, err)
	assert.Nil(t, expense)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestExpenseRepository_GetAll_Success(t *testing.T) {
	repo, mock, cleanup := setupExpenseMockDB(t)
	defer cleanup()

	filters := ExpenseFilters{}

	mock.ExpectQuery(`SELECT e.id, e.category_id, e.period_id, e.account_id, e.description`).
		WithArgs("user123").
		WillReturnRows(sqlmock.NewRows([]string{
			"e.id", "e.category_id", "e.period_id", "e.account_id", "e.description",
			"e.due_date", "e.current_amount", "e.amount_paid", "e.current_installment",
			"e.total_installments", "e.installment_group_id", "e.is_recurring", "e.notes",
			"e.created_at", "e.updated_at", "c.id", "c.name", "p.id", "p.month_number",
			"p.year_number", "sa.id", "sa.company_id", "sa.account_identifier", "sa.alias",
		}).AddRow(
			1, 1, 1, sql.NullInt64{Int64: 1, Valid: true}, "Expense 1",
			sql.NullString{String: "2024-06-15", Valid: true}, 100.00, 0.00, 1,
			1, sql.NullString{}, false, sql.NullString{},
			"2024-01-01T00:00:00Z", "2024-01-01T00:00:00Z", 1, "Category 1", 1, sql.NullInt32{Int32: 6, Valid: true},
			2024, 1, sql.NullInt64{Int64: 1, Valid: true}, "ACC123", sql.NullString{},
		).AddRow(
			2, 2, 1, sql.NullInt64{}, "Expense 2",
			sql.NullString{}, 200.00, 100.00, 1,
			2, sql.NullString{}, false, sql.NullString{},
			"2024-01-02T00:00:00Z", "2024-01-02T00:00:00Z", 2, "Category 2", 1, sql.NullInt32{Int32: 6, Valid: true},
			2024, 0, 0, "", "",
		))

	expenses, err := repo.GetAll("user123", filters)

	assert.NoError(t, err)
	assert.Len(t, expenses, 2)
	assert.Equal(t, "Expense 1", expenses[0].Description)
	assert.Equal(t, "Expense 2", expenses[1].Description)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestExpenseRepository_Update_Success(t *testing.T) {
	repo, mock, cleanup := setupExpenseMockDB(t)
	defer cleanup()

	accountID := 1
	dueDate := "2024-06-20"

	expense := &models.Expense{
		ID:                1,
		CategoryID:        1,
		PeriodID:          1,
		AccountID:         &accountID,
		Description:       "Updated Expense",
		DueDate:           &dueDate,
		CurrentAmount:     150.00,
		AmountPaid:        50.00,
		TotalInstallments: 1,
	}

	mock.ExpectQuery(`UPDATE expenses`).
		WithArgs(1, 1, 1, "Updated Expense", "2024-06-20", 150.00, 50.00, 0, 1, nil, false, "", 1, "user123").
		WillReturnRows(sqlmock.NewRows([]string{"updated_at"}).AddRow("2024-01-01T00:00:00Z"))

	err := repo.Update("user123", expense)

	assert.NoError(t, err)
	assert.Equal(t, "2024-01-01T00:00:00Z", expense.UpdatedAt)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestExpenseRepository_Delete_Success(t *testing.T) {
	repo, mock, cleanup := setupExpenseMockDB(t)
	defer cleanup()

	mock.ExpectExec(`DELETE FROM expenses`).
		WithArgs(1, "user123").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Delete("user123", 1)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestExpenseRepository_MarkAsPaid_Success(t *testing.T) {
	repo, mock, cleanup := setupExpenseMockDB(t)
	defer cleanup()

	mock.ExpectQuery(`UPDATE expenses SET amount_paid = current_amount`).
		WithArgs(1, "user123").
		WillReturnRows(sqlmock.NewRows([]string{"updated_at"}).AddRow("2024-01-01T00:00:00Z"))

	err := repo.MarkAsPaid("user123", 1)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestExpenseRepository_MarkAsPaid_NotFound(t *testing.T) {
	repo, mock, cleanup := setupExpenseMockDB(t)
	defer cleanup()

	mock.ExpectQuery(`UPDATE expenses SET amount_paid = current_amount`).
		WithArgs(999, "user123").
		WillReturnError(sql.ErrNoRows)

	err := repo.MarkAsPaid("user123", 999)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found or access denied")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestExpenseRepository_UpdateAmountPaid_Success(t *testing.T) {
	repo, mock, cleanup := setupExpenseMockDB(t)
	defer cleanup()

	mock.ExpectQuery(`UPDATE expenses SET amount_paid = \$1`).
		WithArgs(75.00, 1, "user123").
		WillReturnRows(sqlmock.NewRows([]string{"updated_at"}).AddRow("2024-01-01T00:00:00Z"))

	err := repo.UpdateAmountPaid("user123", 1, 75.00)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestExpenseRepository_CategoryExistsAndBelongsToUser_True(t *testing.T) {
	repo, mock, cleanup := setupExpenseMockDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs(1, "user123").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	exists, err := repo.CategoryExistsAndBelongsToUser("user123", 1)

	assert.NoError(t, err)
	assert.True(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestExpenseRepository_CategoryExistsAndBelongsToUser_False(t *testing.T) {
	repo, mock, cleanup := setupExpenseMockDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs(999, "user123").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	exists, err := repo.CategoryExistsAndBelongsToUser("user123", 999)

	assert.NoError(t, err)
	assert.False(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestExpenseRepository_PeriodExistsAndBelongsToUser_True(t *testing.T) {
	repo, mock, cleanup := setupExpenseMockDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs(1, "user123").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	exists, err := repo.PeriodExistsAndBelongsToUser("user123", 1)

	assert.NoError(t, err)
	assert.True(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestExpenseRepository_ServiceAccountExistsAndBelongsToUser_True(t *testing.T) {
	repo, mock, cleanup := setupExpenseMockDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs(1, "user123").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	exists, err := repo.ServiceAccountExistsAndBelongsToUser("user123", 1)

	assert.NoError(t, err)
	assert.True(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestExpenseRepository_GetSummaryByPeriod_Success(t *testing.T) {
	repo, mock, cleanup := setupExpenseMockDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT COALESCE\(SUM\(current_amount\), 0\) as total_amount`).
		WithArgs("user123", 1).
		WillReturnRows(sqlmock.NewRows([]string{"total_amount", "paid_amount", "pending_amount", "expense_count"}).
			AddRow(1000.00, 600.00, 400.00, 5))

	summary, err := repo.GetSummaryByPeriod("user123", 1)

	assert.NoError(t, err)
	assert.NotNil(t, summary)
	assert.Equal(t, 1000.00, summary.TotalAmount)
	assert.Equal(t, 600.00, summary.PaidAmount)
	assert.Equal(t, 400.00, summary.PendingAmount)
	assert.Equal(t, 5, summary.ExpenseCount)
	assert.NoError(t, mock.ExpectationsWereMet())
}
