package repository

import (
	"database/sql"
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/fernandovalenzuela/api-home-pay/internal/models"
	"github.com/stretchr/testify/assert"
)

func setupCategoryMockDB(t *testing.T) (*categoryRepository, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}

	repo := &categoryRepository{
		db:  db,
		ctx: context.Background(),
	}

	cleanup := func() {
		db.Close()
	}

	return repo, mock, cleanup
}

func TestCategoryRepository_Create_Success(t *testing.T) {
	repo, mock, cleanup := setupCategoryMockDB(t)
	defer cleanup()

	category := &models.Category{
		Name: "Test Category",
	}

	mock.ExpectQuery(`INSERT INTO categories`).
		WithArgs("user123", "Test Category").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).
			AddRow(1, "2024-01-01T00:00:00Z"))

	err := repo.Create("user123", category)

	assert.NoError(t, err)
	assert.Equal(t, 1, category.ID)
	assert.Equal(t, "2024-01-01T00:00:00Z", category.CreatedAt)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCategoryRepository_Create_EmptyUserID(t *testing.T) {
	repo, mock, cleanup := setupCategoryMockDB(t)
	defer cleanup()

	category := &models.Category{Name: "Test"}

	err := repo.Create("", category)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user_id is required")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCategoryRepository_Create_DatabaseError(t *testing.T) {
	repo, mock, cleanup := setupCategoryMockDB(t)
	defer cleanup()

	category := &models.Category{Name: "Test"}

	mock.ExpectQuery(`INSERT INTO categories`).
		WithArgs("user123", "Test").
		WillReturnError(errors.New("connection failed"))

	err := repo.Create("user123", category)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create category")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCategoryRepository_GetByID_Success(t *testing.T) {
	repo, mock, cleanup := setupCategoryMockDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT id, name, created_at FROM categories`).
		WithArgs(1, "user123").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "created_at"}).
			AddRow(1, "Groceries", "2024-01-01T00:00:00Z"))

	category, err := repo.GetByID("user123", 1)

	assert.NoError(t, err)
	assert.NotNil(t, category)
	assert.Equal(t, 1, category.ID)
	assert.Equal(t, "Groceries", category.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCategoryRepository_GetByID_NotFound(t *testing.T) {
	repo, mock, cleanup := setupCategoryMockDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT id, name, created_at FROM categories`).
		WithArgs(999, "user123").
		WillReturnError(sql.ErrNoRows)

	category, err := repo.GetByID("user123", 999)

	assert.NoError(t, err)
	assert.Nil(t, category)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCategoryRepository_GetByID_EmptyUserID(t *testing.T) {
	repo, mock, cleanup := setupCategoryMockDB(t)
	defer cleanup()

	category, err := repo.GetByID("", 1)

	assert.Error(t, err)
	assert.Nil(t, category)
	assert.Contains(t, err.Error(), "user_id is required")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCategoryRepository_GetAll_Success(t *testing.T) {
	repo, mock, cleanup := setupCategoryMockDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT id, name, created_at FROM categories`).
		WithArgs("user123").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "created_at"}).
			AddRow(1, "Groceries", "2024-01-01T00:00:00Z").
			AddRow(2, "Utilities", "2024-01-02T00:00:00Z"))

	categories, err := repo.GetAll("user123")

	assert.NoError(t, err)
	assert.Len(t, categories, 2)
	assert.Equal(t, "Groceries", categories[0].Name)
	assert.Equal(t, "Utilities", categories[1].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCategoryRepository_GetAll_EmptyResult(t *testing.T) {
	repo, mock, cleanup := setupCategoryMockDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT id, name, created_at FROM categories`).
		WithArgs("user123").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "created_at"}))

	categories, err := repo.GetAll("user123")

	assert.NoError(t, err)
	assert.Empty(t, categories)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCategoryRepository_Update_Success(t *testing.T) {
	repo, mock, cleanup := setupCategoryMockDB(t)
	defer cleanup()

	category := &models.Category{
		ID:   1,
		Name: "Updated Category",
	}

	mock.ExpectQuery(`UPDATE categories`).
		WithArgs("Updated Category", 1, "user123").
		WillReturnRows(sqlmock.NewRows([]string{"created_at"}).
			AddRow("2024-01-01T00:00:00Z"))

	err := repo.Update("user123", category)

	assert.NoError(t, err)
	assert.Equal(t, "2024-01-01T00:00:00Z", category.CreatedAt)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCategoryRepository_Update_NotFound(t *testing.T) {
	repo, mock, cleanup := setupCategoryMockDB(t)
	defer cleanup()

	category := &models.Category{
		ID:   999,
		Name: "Updated",
	}

	mock.ExpectQuery(`UPDATE categories`).
		WithArgs("Updated", 999, "user123").
		WillReturnError(sql.ErrNoRows)

	err := repo.Update("user123", category)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found or access denied")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCategoryRepository_Delete_Success(t *testing.T) {
	repo, mock, cleanup := setupCategoryMockDB(t)
	defer cleanup()

	mock.ExpectExec(`DELETE FROM categories`).
		WithArgs(1, "user123").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Delete("user123", 1)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCategoryRepository_Delete_NotFound(t *testing.T) {
	repo, mock, cleanup := setupCategoryMockDB(t)
	defer cleanup()

	mock.ExpectExec(`DELETE FROM categories`).
		WithArgs(999, "user123").
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.Delete("user123", 999)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found or access denied")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCategoryRepository_ExistsByName_True(t *testing.T) {
	repo, mock, cleanup := setupCategoryMockDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs("Groceries", "user123").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	exists, err := repo.ExistsByName("user123", "Groceries")

	assert.NoError(t, err)
	assert.True(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCategoryRepository_ExistsByName_False(t *testing.T) {
	repo, mock, cleanup := setupCategoryMockDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs("NonExistent", "user123").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	exists, err := repo.ExistsByName("user123", "NonExistent")

	assert.NoError(t, err)
	assert.False(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCategoryRepository_HasExpenses_True(t *testing.T) {
	repo, mock, cleanup := setupCategoryMockDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	exists, err := repo.HasExpenses(1)

	assert.NoError(t, err)
	assert.True(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCategoryRepository_HasExpenses_False(t *testing.T) {
	repo, mock, cleanup := setupCategoryMockDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	exists, err := repo.HasExpenses(1)

	assert.NoError(t, err)
	assert.False(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}
