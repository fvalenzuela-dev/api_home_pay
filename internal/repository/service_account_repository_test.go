package repository

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/fernandovalenzuela/api-home-pay/internal/models"
	"github.com/stretchr/testify/assert"
)

func setupServiceAccountMockDB(t *testing.T) (*serviceAccountRepository, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}

	repo := &serviceAccountRepository{
		db:  db,
		ctx: context.Background(),
	}

	cleanup := func() {
		db.Close()
	}

	return repo, mock, cleanup
}

func TestServiceAccountRepository_Create_Success(t *testing.T) {
	repo, mock, cleanup := setupServiceAccountMockDB(t)
	defer cleanup()

	account := &models.ServiceAccount{
		CompanyID:         1,
		AccountIdentifier: "ACC123456",
		Alias:             "My Account",
	}

	mock.ExpectQuery(`INSERT INTO service_accounts`).
		WithArgs(1, "ACC123456", "My Account", "user123").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	err := repo.Create("user123", account)

	assert.NoError(t, err)
	assert.Equal(t, 1, account.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestServiceAccountRepository_GetByID_Success(t *testing.T) {
	repo, mock, cleanup := setupServiceAccountMockDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT sa.id, sa.company_id, sa.account_identifier, sa.alias, c.id, c.name, c.website_url FROM service_accounts sa LEFT JOIN companies c ON sa.company_id = c.id`).
		WithArgs(1, "user123").
		WillReturnRows(sqlmock.NewRows([]string{
			"sa.id", "sa.company_id", "sa.account_identifier", "sa.alias",
			"c.id", "c.name", "c.website_url",
		}).AddRow(1, 1, "ACC123456", "My Account", 1, "Acme Corp", "https://acme.com"))

	account, err := repo.GetByID("user123", 1)

	assert.NoError(t, err)
	assert.NotNil(t, account)
	assert.Equal(t, 1, account.ID)
	assert.Equal(t, "ACC123456", account.AccountIdentifier)
	assert.Equal(t, "My Account", account.Alias)
	assert.NotNil(t, account.Company)
	assert.Equal(t, "Acme Corp", account.Company.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestServiceAccountRepository_GetByID_NotFound(t *testing.T) {
	repo, mock, cleanup := setupServiceAccountMockDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT sa.id, sa.company_id, sa.account_identifier, sa.alias`).
		WithArgs(999, "user123").
		WillReturnError(sql.ErrNoRows)

	account, err := repo.GetByID("user123", 999)

	assert.NoError(t, err)
	assert.Nil(t, account)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestServiceAccountRepository_GetAll_Success(t *testing.T) {
	repo, mock, cleanup := setupServiceAccountMockDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT sa.id, sa.company_id, sa.account_identifier, sa.alias, c.id, c.name, c.website_url FROM service_accounts sa LEFT JOIN companies c ON sa.company_id = c.id`).
		WithArgs("user123").
		WillReturnRows(sqlmock.NewRows([]string{
			"sa.id", "sa.company_id", "sa.account_identifier", "sa.alias",
			"c.id", "c.name", "c.website_url",
		}).
			AddRow(1, 1, "ACC123456", "Account 1", 1, "Acme Corp", "https://acme.com").
			AddRow(2, 1, "ACC789012", "Account 2", 1, "Acme Corp", "https://acme.com"))

	accounts, err := repo.GetAll("user123", nil)

	assert.NoError(t, err)
	assert.Len(t, accounts, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestServiceAccountRepository_GetAll_WithCompanyFilter(t *testing.T) {
	repo, mock, cleanup := setupServiceAccountMockDB(t)
	defer cleanup()

	companyID := 1

	mock.ExpectQuery(`SELECT sa.id, sa.company_id, sa.account_identifier, sa.alias, c.id, c.name, c.website_url FROM service_accounts sa LEFT JOIN companies c ON sa.company_id = c.id`).
		WithArgs("user123", 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"sa.id", "sa.company_id", "sa.account_identifier", "sa.alias",
			"c.id", "c.name", "c.website_url",
		}).AddRow(1, 1, "ACC123456", "Account 1", 1, "Acme Corp", "https://acme.com"))

	accounts, err := repo.GetAll("user123", &companyID)

	assert.NoError(t, err)
	assert.Len(t, accounts, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestServiceAccountRepository_Update_Success(t *testing.T) {
	repo, mock, cleanup := setupServiceAccountMockDB(t)
	defer cleanup()

	account := &models.ServiceAccount{
		ID:                1,
		CompanyID:         1,
		AccountIdentifier: "UPDATED123",
		Alias:             "Updated Account",
	}

	mock.ExpectQuery(`UPDATE service_accounts`).
		WithArgs(1, "UPDATED123", "Updated Account", 1, "user123").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	err := repo.Update("user123", account)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestServiceAccountRepository_Delete_Success(t *testing.T) {
	repo, mock, cleanup := setupServiceAccountMockDB(t)
	defer cleanup()

	mock.ExpectExec(`DELETE FROM service_accounts`).
		WithArgs(1, "user123").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Delete("user123", 1)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestServiceAccountRepository_ExistsByIdentifier_True(t *testing.T) {
	repo, mock, cleanup := setupServiceAccountMockDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs("ACC123456", 1, "user123").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	exists, err := repo.ExistsByIdentifier("user123", 1, "ACC123456")

	assert.NoError(t, err)
	assert.True(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestServiceAccountRepository_ExistsByIdentifier_False(t *testing.T) {
	repo, mock, cleanup := setupServiceAccountMockDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs("NONEXISTENT", 1, "user123").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	exists, err := repo.ExistsByIdentifier("user123", 1, "NONEXISTENT")

	assert.NoError(t, err)
	assert.False(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestServiceAccountRepository_HasExpenses_True(t *testing.T) {
	repo, mock, cleanup := setupServiceAccountMockDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	exists, err := repo.HasExpenses(1)

	assert.NoError(t, err)
	assert.True(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestServiceAccountRepository_HasExpenses_False(t *testing.T) {
	repo, mock, cleanup := setupServiceAccountMockDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	exists, err := repo.HasExpenses(1)

	assert.NoError(t, err)
	assert.False(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestServiceAccountRepository_CompanyExistsAndBelongsToUser_True(t *testing.T) {
	repo, mock, cleanup := setupServiceAccountMockDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs(1, "user123").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	exists, err := repo.CompanyExistsAndBelongsToUser("user123", 1)

	assert.NoError(t, err)
	assert.True(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestServiceAccountRepository_CompanyExistsAndBelongsToUser_False(t *testing.T) {
	repo, mock, cleanup := setupServiceAccountMockDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs(999, "user123").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	exists, err := repo.CompanyExistsAndBelongsToUser("user123", 999)

	assert.NoError(t, err)
	assert.False(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}
