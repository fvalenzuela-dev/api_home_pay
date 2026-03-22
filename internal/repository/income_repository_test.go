package repository

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/fernandovalenzuela/api-home-pay/internal/models"
	"github.com/stretchr/testify/assert"
)

func setupIncomeMockDB(t *testing.T) (*incomeRepository, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}

	repo := &incomeRepository{
		db:  db,
		ctx: context.Background(),
	}

	cleanup := func() {
		db.Close()
	}

	return repo, mock, cleanup
}

func TestIncomeRepository_Create_Success(t *testing.T) {
	repo, mock, cleanup := setupIncomeMockDB(t)
	defer cleanup()

	income := &models.Income{
		PeriodID:    1,
		Description: "Salary",
		Amount:      5000.00,
		IsRecurring: true,
		ReceivedAt:  "2024-06-01",
	}

	mock.ExpectQuery(`INSERT INTO incomes`).
		WithArgs(1, "Salary", 5000.00, true, "2024-06-01", "user123").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).
			AddRow(1, "2024-01-01T00:00:00Z"))

	err := repo.Create("user123", income)

	assert.NoError(t, err)
	assert.Equal(t, 1, income.ID)
	assert.Equal(t, "2024-01-01T00:00:00Z", income.CreatedAt)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestIncomeRepository_Create_WithoutReceivedAt(t *testing.T) {
	repo, mock, cleanup := setupIncomeMockDB(t)
	defer cleanup()

	income := &models.Income{
		PeriodID:    1,
		Description: "Bonus",
		Amount:      1000.00,
		IsRecurring: false,
	}

	mock.ExpectQuery(`INSERT INTO incomes`).
		WithArgs(1, "Bonus", 1000.00, false, nil, "user123").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).
			AddRow(1, "2024-01-01T00:00:00Z"))

	err := repo.Create("user123", income)

	assert.NoError(t, err)
	assert.Equal(t, 1, income.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestIncomeRepository_Create_EmptyUserID(t *testing.T) {
	repo, mock, cleanup := setupIncomeMockDB(t)
	defer cleanup()

	income := &models.Income{Description: "Test"}

	err := repo.Create("", income)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user_id is required")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestIncomeRepository_GetByID_Success(t *testing.T) {
	repo, mock, cleanup := setupIncomeMockDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT i.id, i.period_id, i.description, i.amount, i.is_recurring, i.received_at, i.created_at, p.id, p.month_number, p.year_number FROM incomes i LEFT JOIN periods p ON i.period_id = p.id`).
		WithArgs(1, "user123").
		WillReturnRows(sqlmock.NewRows([]string{
			"i.id", "i.period_id", "i.description", "i.amount", "i.is_recurring",
			"i.received_at", "i.created_at", "p.id", "p.month_number", "p.year_number",
		}).AddRow(
			1, 1, "Salary", 5000.00, true,
			sql.NullString{String: "2024-06-01", Valid: true}, "2024-01-01T00:00:00Z",
			1, sql.NullInt32{Int32: 6, Valid: true}, 2024,
		))

	income, err := repo.GetByID("user123", 1)

	assert.NoError(t, err)
	assert.NotNil(t, income)
	assert.Equal(t, 1, income.ID)
	assert.Equal(t, "Salary", income.Description)
	assert.Equal(t, 5000.00, income.Amount)
	assert.True(t, income.IsRecurring)
	assert.NotNil(t, income.Period)
	assert.Equal(t, 6, income.Period.MonthNumber)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestIncomeRepository_GetByID_NotFound(t *testing.T) {
	repo, mock, cleanup := setupIncomeMockDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT i.id, i.period_id`).
		WithArgs(999, "user123").
		WillReturnError(sql.ErrNoRows)

	income, err := repo.GetByID("user123", 999)

	assert.NoError(t, err)
	assert.Nil(t, income)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestIncomeRepository_GetAll_Success(t *testing.T) {
	repo, mock, cleanup := setupIncomeMockDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT i.id, i.period_id, i.description, i.amount, i.is_recurring, i.received_at, i.created_at, p.id, p.month_number, p.year_number FROM incomes i LEFT JOIN periods p ON i.period_id = p.id`).
		WithArgs("user123").
		WillReturnRows(sqlmock.NewRows([]string{
			"i.id", "i.period_id", "i.description", "i.amount", "i.is_recurring",
			"i.received_at", "i.created_at", "p.id", "p.month_number", "p.year_number",
		}).AddRow(
			1, 1, "Salary", 5000.00, true, sql.NullString{String: "2024-06-01", Valid: true}, "2024-01-01T00:00:00Z",
			1, sql.NullInt32{Int32: 6, Valid: true}, 2024,
		).AddRow(
			2, 1, "Bonus", 1000.00, false, sql.NullString{}, "2024-01-02T00:00:00Z",
			1, sql.NullInt32{Int32: 6, Valid: true}, 2024,
		))

	incomes, err := repo.GetAll("user123", nil)

	assert.NoError(t, err)
	assert.Len(t, incomes, 2)
	assert.Equal(t, "Salary", incomes[0].Description)
	assert.Equal(t, "Bonus", incomes[1].Description)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestIncomeRepository_GetAll_WithPeriodFilter(t *testing.T) {
	repo, mock, cleanup := setupIncomeMockDB(t)
	defer cleanup()

	periodID := 1

	mock.ExpectQuery(`SELECT i.id, i.period_id, i.description, i.amount, i.is_recurring, i.received_at, i.created_at, p.id, p.month_number, p.year_number FROM incomes i LEFT JOIN periods p ON i.period_id = p.id`).
		WithArgs("user123", 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"i.id", "i.period_id", "i.description", "i.amount", "i.is_recurring",
			"i.received_at", "i.created_at", "p.id", "p.month_number", "p.year_number",
		}).AddRow(
			1, 1, "Salary", 5000.00, true, sql.NullString{String: "2024-06-01", Valid: true}, "2024-01-01T00:00:00Z",
			1, sql.NullInt32{Int32: 6, Valid: true}, 2024,
		))

	incomes, err := repo.GetAll("user123", &periodID)

	assert.NoError(t, err)
	assert.Len(t, incomes, 1)
	assert.Equal(t, "Salary", incomes[0].Description)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestIncomeRepository_Update_Success(t *testing.T) {
	repo, mock, cleanup := setupIncomeMockDB(t)
	defer cleanup()

	income := &models.Income{
		ID:          1,
		PeriodID:    1,
		Description: "Updated Income",
		Amount:      5500.00,
		IsRecurring: true,
		ReceivedAt:  "2024-06-15",
	}

	mock.ExpectQuery(`UPDATE incomes`).
		WithArgs(1, "Updated Income", 5500.00, true, "2024-06-15", 1, "user123").
		WillReturnRows(sqlmock.NewRows([]string{"created_at"}).AddRow("2024-01-01T00:00:00Z"))

	err := repo.Update("user123", income)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestIncomeRepository_Delete_Success(t *testing.T) {
	repo, mock, cleanup := setupIncomeMockDB(t)
	defer cleanup()

	mock.ExpectExec(`DELETE FROM incomes`).
		WithArgs(1, "user123").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Delete("user123", 1)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestIncomeRepository_PeriodExistsAndBelongsToUser_True(t *testing.T) {
	repo, mock, cleanup := setupIncomeMockDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs(1, "user123").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	exists, err := repo.PeriodExistsAndBelongsToUser("user123", 1)

	assert.NoError(t, err)
	assert.True(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestIncomeRepository_PeriodExistsAndBelongsToUser_False(t *testing.T) {
	repo, mock, cleanup := setupIncomeMockDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs(999, "user123").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	exists, err := repo.PeriodExistsAndBelongsToUser("user123", 999)

	assert.NoError(t, err)
	assert.False(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestIncomeRepository_GetTotalByPeriod_Success(t *testing.T) {
	repo, mock, cleanup := setupIncomeMockDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT COALESCE\(SUM\(amount\), 0\) as total_amount, COUNT\(\*\) as income_count FROM incomes`).
		WithArgs("user123", 1).
		WillReturnRows(sqlmock.NewRows([]string{"total_amount", "income_count"}).
			AddRow(6000.00, 2))

	total, count, err := repo.GetTotalByPeriod("user123", 1)

	assert.NoError(t, err)
	assert.Equal(t, 6000.00, total)
	assert.Equal(t, 2, count)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestIncomeRepository_GetTotalByPeriod_NoIncomes(t *testing.T) {
	repo, mock, cleanup := setupIncomeMockDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT COALESCE\(SUM\(amount\), 0\) as total_amount, COUNT\(\*\) as income_count FROM incomes`).
		WithArgs("user123", 1).
		WillReturnRows(sqlmock.NewRows([]string{"total_amount", "income_count"}).
			AddRow(0.00, 0))

	total, count, err := repo.GetTotalByPeriod("user123", 1)

	assert.NoError(t, err)
	assert.Equal(t, 0.00, total)
	assert.Equal(t, 0, count)
	assert.NoError(t, mock.ExpectationsWereMet())
}
