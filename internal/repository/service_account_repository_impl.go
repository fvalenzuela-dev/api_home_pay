package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/fernandovalenzuela/api-home-pay/internal/models"
)

type serviceAccountRepository struct {
	db  *sql.DB
	ctx context.Context
}

func NewServiceAccountRepository(db *sql.DB) ServiceAccountRepository {
	return &serviceAccountRepository{
		db:  db,
		ctx: context.Background(),
	}
}

func (r *serviceAccountRepository) Create(userID string, account *models.ServiceAccount) error {
	query := `
		INSERT INTO service_accounts (company_id, account_identifier, alias, user_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	if userID == "" {
		return fmt.Errorf("user_id is required")
	}

	var companyID interface{}
	if account.CompanyID > 0 {
		companyID = account.CompanyID
	} else {
		companyID = nil
	}

	err := r.db.QueryRowContext(r.ctx, query, companyID, account.AccountIdentifier, account.Alias, userID).Scan(&account.ID)
	if err != nil {
		return fmt.Errorf("failed to create service account: %w", err)
	}

	return nil
}

func (r *serviceAccountRepository) GetByID(userID string, id int) (*models.ServiceAccount, error) {
	query := `
		SELECT sa.id, sa.company_id, sa.account_identifier, sa.alias,
		       c.id, c.name, c.website_url
		FROM service_accounts sa
		LEFT JOIN companies c ON sa.company_id = c.id
		WHERE sa.id = $1 AND sa.user_id = $2
	`

	if userID == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	account := &models.ServiceAccount{}
	var company models.Company
	var companyID sql.NullInt64
	var companyName, companyWebsite sql.NullString

	err := r.db.QueryRowContext(r.ctx, query, id, userID).Scan(
		&account.ID,
		&companyID,
		&account.AccountIdentifier,
		&account.Alias,
		&company.ID,
		&companyName,
		&companyWebsite,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get service account by ID: %w", err)
	}

	if companyID.Valid {
		account.CompanyID = int(companyID.Int64)
		if companyName.Valid {
			company.Name = companyName.String
		}
		if companyWebsite.Valid {
			company.WebsiteURL = companyWebsite.String
		}
		account.Company = &company
	}

	return account, nil
}

func (r *serviceAccountRepository) GetAll(userID string, companyID *int) ([]models.ServiceAccount, error) {
	var query string
	var rows *sql.Rows
	var err error

	if userID == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	if companyID != nil && *companyID > 0 {
		query = `
			SELECT sa.id, sa.company_id, sa.account_identifier, sa.alias,
			       c.id, c.name, c.website_url
			FROM service_accounts sa
			LEFT JOIN companies c ON sa.company_id = c.id
			WHERE sa.user_id = $1 AND sa.company_id = $2
			ORDER BY sa.account_identifier ASC
		`
		rows, err = r.db.QueryContext(r.ctx, query, userID, *companyID)
	} else {
		query = `
			SELECT sa.id, sa.company_id, sa.account_identifier, sa.alias,
			       c.id, c.name, c.website_url
			FROM service_accounts sa
			LEFT JOIN companies c ON sa.company_id = c.id
			WHERE sa.user_id = $1
			ORDER BY sa.account_identifier ASC
		`
		rows, err = r.db.QueryContext(r.ctx, query, userID)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get service accounts: %w", err)
	}
	defer rows.Close()

	var accounts []models.ServiceAccount
	for rows.Next() {
		var account models.ServiceAccount
		var company models.Company
		var companyID sql.NullInt64
		var companyName, companyWebsite sql.NullString

		err := rows.Scan(
			&account.ID,
			&companyID,
			&account.AccountIdentifier,
			&account.Alias,
			&company.ID,
			&companyName,
			&companyWebsite,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan service account: %w", err)
		}

		if companyID.Valid {
			account.CompanyID = int(companyID.Int64)
			if companyName.Valid {
				company.Name = companyName.String
			}
			if companyWebsite.Valid {
				company.WebsiteURL = companyWebsite.String
			}
			account.Company = &company
		}

		accounts = append(accounts, account)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating service accounts: %w", err)
	}

	return accounts, nil
}

func (r *serviceAccountRepository) Update(userID string, account *models.ServiceAccount) error {
	query := `
		UPDATE service_accounts
		SET company_id = $1, account_identifier = $2, alias = $3
		WHERE id = $4 AND user_id = $5
		RETURNING id
	`

	if userID == "" {
		return fmt.Errorf("user_id is required")
	}

	var companyID interface{}
	if account.CompanyID > 0 {
		companyID = account.CompanyID
	} else {
		companyID = nil
	}

	var returnedID int
	err := r.db.QueryRowContext(r.ctx, query, companyID, account.AccountIdentifier, account.Alias, account.ID, userID).Scan(&returnedID)
	if err == sql.ErrNoRows {
		return fmt.Errorf("service account not found or access denied")
	}
	if err != nil {
		return fmt.Errorf("failed to update service account: %w", err)
	}

	return nil
}

func (r *serviceAccountRepository) Delete(userID string, id int) error {
	query := `
		DELETE FROM service_accounts
		WHERE id = $1 AND user_id = $2
	`

	if userID == "" {
		return fmt.Errorf("user_id is required")
	}

	result, err := r.db.ExecContext(r.ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete service account: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("service account not found or access denied")
	}

	return nil
}

func (r *serviceAccountRepository) ExistsByIdentifier(userID string, companyID int, identifier string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM service_accounts
			WHERE account_identifier = $1 AND company_id = $2 AND user_id = $3
		)
	`

	if userID == "" {
		return false, fmt.Errorf("user_id is required")
	}

	var exists bool
	err := r.db.QueryRowContext(r.ctx, query, identifier, companyID, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check service account existence: %w", err)
	}

	return exists, nil
}

func (r *serviceAccountRepository) HasExpenses(id int) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM expenses
			WHERE account_id = $1
		)
	`

	var exists bool
	err := r.db.QueryRowContext(r.ctx, query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check service account expenses: %w", err)
	}

	return exists, nil
}

func (r *serviceAccountRepository) CompanyExistsAndBelongsToUser(userID string, companyID int) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM companies
			WHERE id = $1 AND user_id = $2
		)
	`

	if userID == "" {
		return false, fmt.Errorf("user_id is required")
	}

	var exists bool
	err := r.db.QueryRowContext(r.ctx, query, companyID, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check company existence: %w", err)
	}

	return exists, nil
}
