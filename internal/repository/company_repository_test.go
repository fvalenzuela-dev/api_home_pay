package repository

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/fernandovalenzuela/api-home-pay/internal/models"
	"github.com/stretchr/testify/assert"
)

func setupCompanyMockDB(t *testing.T) (*companyRepository, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}

	repo := &companyRepository{
		db:  db,
		ctx: context.Background(),
	}

	cleanup := func() {
		db.Close()
	}

	return repo, mock, cleanup
}

func TestCompanyRepository_Create_Success(t *testing.T) {
	repo, mock, cleanup := setupCompanyMockDB(t)
	defer cleanup()

	company := &models.Company{
		Name:       "Test Company",
		WebsiteURL: "https://test.com",
	}

	mock.ExpectQuery(`INSERT INTO companies`).
		WithArgs("user123", "Test Company", "https://test.com").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).
			AddRow(1, "2024-01-01T00:00:00Z"))

	err := repo.Create("user123", company)

	assert.NoError(t, err)
	assert.Equal(t, 1, company.ID)
	assert.Equal(t, "2024-01-01T00:00:00Z", company.CreatedAt)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCompanyRepository_Create_WithoutWebsite(t *testing.T) {
	repo, mock, cleanup := setupCompanyMockDB(t)
	defer cleanup()

	company := &models.Company{
		Name: "Test Company",
	}

	mock.ExpectQuery(`INSERT INTO companies`).
		WithArgs("user123", "Test Company", "").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).
			AddRow(1, "2024-01-01T00:00:00Z"))

	err := repo.Create("user123", company)

	assert.NoError(t, err)
	assert.Equal(t, 1, company.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCompanyRepository_GetByID_Success(t *testing.T) {
	repo, mock, cleanup := setupCompanyMockDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT id, name, website_url, created_at FROM companies`).
		WithArgs(1, "user123").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "website_url", "created_at"}).
			AddRow(1, "Acme Corp", "https://acme.com", "2024-01-01T00:00:00Z"))

	company, err := repo.GetByID("user123", 1)

	assert.NoError(t, err)
	assert.NotNil(t, company)
	assert.Equal(t, 1, company.ID)
	assert.Equal(t, "Acme Corp", company.Name)
	assert.Equal(t, "https://acme.com", company.WebsiteURL)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCompanyRepository_GetByID_NotFound(t *testing.T) {
	repo, mock, cleanup := setupCompanyMockDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT id, name, website_url, created_at FROM companies`).
		WithArgs(999, "user123").
		WillReturnError(sql.ErrNoRows)

	company, err := repo.GetByID("user123", 999)

	assert.NoError(t, err)
	assert.Nil(t, company)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCompanyRepository_GetAll_Success(t *testing.T) {
	repo, mock, cleanup := setupCompanyMockDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT id, name, website_url, created_at FROM companies`).
		WithArgs("user123").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "website_url", "created_at"}).
			AddRow(1, "Acme Corp", "https://acme.com", "2024-01-01T00:00:00Z").
			AddRow(2, "Tech Inc", "https://tech.com", "2024-01-02T00:00:00Z"))

	companies, err := repo.GetAll("user123")

	assert.NoError(t, err)
	assert.Len(t, companies, 2)
	assert.Equal(t, "Acme Corp", companies[0].Name)
	assert.Equal(t, "Tech Inc", companies[1].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCompanyRepository_Update_Success(t *testing.T) {
	repo, mock, cleanup := setupCompanyMockDB(t)
	defer cleanup()

	company := &models.Company{
		ID:         1,
		Name:       "Updated Company",
		WebsiteURL: "https://updated.com",
	}

	mock.ExpectQuery(`UPDATE companies`).
		WithArgs("Updated Company", "https://updated.com", 1, "user123").
		WillReturnRows(sqlmock.NewRows([]string{"created_at"}).
			AddRow("2024-01-01T00:00:00Z"))

	err := repo.Update("user123", company)

	assert.NoError(t, err)
	assert.Equal(t, "2024-01-01T00:00:00Z", company.CreatedAt)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCompanyRepository_Delete_Success(t *testing.T) {
	repo, mock, cleanup := setupCompanyMockDB(t)
	defer cleanup()

	mock.ExpectExec(`DELETE FROM companies`).
		WithArgs(1, "user123").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Delete("user123", 1)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCompanyRepository_ExistsByName_True(t *testing.T) {
	repo, mock, cleanup := setupCompanyMockDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs("Acme Corp", "user123").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	exists, err := repo.ExistsByName("user123", "Acme Corp")

	assert.NoError(t, err)
	assert.True(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCompanyRepository_ExistsByName_False(t *testing.T) {
	repo, mock, cleanup := setupCompanyMockDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs("NonExistent", "user123").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	exists, err := repo.ExistsByName("user123", "NonExistent")

	assert.NoError(t, err)
	assert.False(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCompanyRepository_HasServiceAccounts_True(t *testing.T) {
	repo, mock, cleanup := setupCompanyMockDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	exists, err := repo.HasServiceAccounts(1)

	assert.NoError(t, err)
	assert.True(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCompanyRepository_HasServiceAccounts_False(t *testing.T) {
	repo, mock, cleanup := setupCompanyMockDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	exists, err := repo.HasServiceAccounts(1)

	assert.NoError(t, err)
	assert.False(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}
