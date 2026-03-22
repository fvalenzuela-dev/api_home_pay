package repository

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/fernandovalenzuela/api-home-pay/internal/models"
	"github.com/stretchr/testify/assert"
)

func setupPeriodMockDB(t *testing.T) (*periodRepository, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}

	repo := &periodRepository{
		db:  db,
		ctx: context.Background(),
	}

	cleanup := func() {
		db.Close()
	}

	return repo, mock, cleanup
}

func TestPeriodRepository_Create_Success(t *testing.T) {
	repo, mock, cleanup := setupPeriodMockDB(t)
	defer cleanup()

	period := &models.Period{
		MonthNumber: 6,
		YearNumber:  2024,
	}

	mock.ExpectQuery(`INSERT INTO periods`).
		WithArgs(6, 2024).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	err := repo.Create("user123", period)

	assert.NoError(t, err)
	assert.Equal(t, 1, period.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPeriodRepository_GetByID_Success(t *testing.T) {
	repo, mock, cleanup := setupPeriodMockDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT id, month_number, year_number FROM periods`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "month_number", "year_number"}).
			AddRow(1, 6, 2024))

	period, err := repo.GetByID("user123", 1)

	assert.NoError(t, err)
	assert.NotNil(t, period)
	assert.Equal(t, 1, period.ID)
	assert.Equal(t, 6, period.MonthNumber)
	assert.Equal(t, 2024, period.YearNumber)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPeriodRepository_GetByID_NotFound(t *testing.T) {
	repo, mock, cleanup := setupPeriodMockDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT id, month_number, year_number FROM periods`).
		WithArgs(999).
		WillReturnError(sql.ErrNoRows)

	period, err := repo.GetByID("user123", 999)

	assert.NoError(t, err)
	assert.Nil(t, period)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPeriodRepository_GetAll_Success(t *testing.T) {
	repo, mock, cleanup := setupPeriodMockDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT id, month_number, year_number FROM periods`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "month_number", "year_number"}).
			AddRow(1, 1, 2024).
			AddRow(2, 2, 2024).
			AddRow(3, 3, 2024))

	periods, err := repo.GetAll("user123")

	assert.NoError(t, err)
	assert.Len(t, periods, 3)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPeriodRepository_Update_Success(t *testing.T) {
	repo, mock, cleanup := setupPeriodMockDB(t)
	defer cleanup()

	period := &models.Period{
		ID:          1,
		MonthNumber: 7,
		YearNumber:  2024,
	}

	mock.ExpectExec(`UPDATE periods`).
		WithArgs(7, 2024, 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Update("user123", period)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPeriodRepository_Update_NotFound(t *testing.T) {
	repo, mock, cleanup := setupPeriodMockDB(t)
	defer cleanup()

	period := &models.Period{
		ID:          999,
		MonthNumber: 7,
		YearNumber:  2024,
	}

	mock.ExpectExec(`UPDATE periods`).
		WithArgs(7, 2024, 999).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.Update("user123", period)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found or access denied")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPeriodRepository_Delete_Success(t *testing.T) {
	repo, mock, cleanup := setupPeriodMockDB(t)
	defer cleanup()

	mock.ExpectExec(`DELETE FROM periods`).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Delete("user123", 1)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPeriodRepository_Delete_NotFound(t *testing.T) {
	repo, mock, cleanup := setupPeriodMockDB(t)
	defer cleanup()

	mock.ExpectExec(`DELETE FROM periods`).
		WithArgs(999).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.Delete("user123", 999)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found or access denied")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPeriodRepository_ExistsByMonthYear_True(t *testing.T) {
	repo, mock, cleanup := setupPeriodMockDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs(6, 2024).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	exists, err := repo.ExistsByMonthYear("user123", 6, 2024)

	assert.NoError(t, err)
	assert.True(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPeriodRepository_ExistsByMonthYear_False(t *testing.T) {
	repo, mock, cleanup := setupPeriodMockDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs(13, 2024).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	exists, err := repo.ExistsByMonthYear("user123", 13, 2024)

	assert.NoError(t, err)
	assert.False(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPeriodRepository_HasExpensesOrIncomes_True(t *testing.T) {
	repo, mock, cleanup := setupPeriodMockDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	exists, err := repo.HasExpensesOrIncomes(1)

	assert.NoError(t, err)
	assert.True(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPeriodRepository_HasExpensesOrIncomes_False(t *testing.T) {
	repo, mock, cleanup := setupPeriodMockDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	exists, err := repo.HasExpensesOrIncomes(1)

	assert.NoError(t, err)
	assert.False(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}
