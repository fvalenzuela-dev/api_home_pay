package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/fernandovalenzuela/api-home-pay/internal/models"
)

type companyRepository struct {
	db  *sql.DB
	ctx context.Context
}

func NewCompanyRepository(db *sql.DB) CompanyRepository {
	return &companyRepository{
		db:  db,
		ctx: context.Background(),
	}
}

func (r *companyRepository) Create(userID string, company *models.Company) error {
	query := `
		INSERT INTO companies (user_id, name, website_url, created_at)
		VALUES ($1, $2, $3, NOW())
		RETURNING id, created_at
	`

	if userID == "" {
		return fmt.Errorf("user_id is required")
	}

	err := r.db.QueryRowContext(r.ctx, query, userID, company.Name, company.WebsiteURL).Scan(
		&company.ID,
		&company.CreatedAt,
	)
	if err != nil {
		slog.Error("db error: failed to create company", "error", err)
		return fmt.Errorf("failed to create company: %w", err)
	}

	return nil
}

func (r *companyRepository) GetByID(userID string, id int) (*models.Company, error) {
	query := `
		SELECT id, name, website_url, created_at
		FROM companies
		WHERE id = $1 AND user_id = $2
	`

	if userID == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	company := &models.Company{}
	err := r.db.QueryRowContext(r.ctx, query, id, userID).Scan(
		&company.ID,
		&company.Name,
		&company.WebsiteURL,
		&company.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		slog.Error("db error: failed to get company by ID", "error", err)
		return nil, fmt.Errorf("failed to get company by ID: %w", err)
	}

	return company, nil
}

func (r *companyRepository) GetAll(userID string) ([]models.Company, error) {
	query := `
		SELECT id, name, website_url, created_at
		FROM companies
		WHERE user_id = $1
		ORDER BY name ASC
	`

	if userID == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	rows, err := r.db.QueryContext(r.ctx, query, userID)
	if err != nil {
		slog.Error("db error: failed to get companies", "error", err)
		return nil, fmt.Errorf("failed to get companies: %w", err)
	}
	defer rows.Close()

	var companies []models.Company
	for rows.Next() {
		var company models.Company
		err := rows.Scan(
			&company.ID,
			&company.Name,
			&company.WebsiteURL,
			&company.CreatedAt,
		)
		if err != nil {
			slog.Error("db error: failed to scan company", "error", err)
			return nil, fmt.Errorf("failed to scan company: %w", err)
		}
		companies = append(companies, company)
	}

	if err = rows.Err(); err != nil {
		slog.Error("db error: error iterating companies", "error", err)
		return nil, fmt.Errorf("error iterating companies: %w", err)
	}

	return companies, nil
}

func (r *companyRepository) Update(userID string, company *models.Company) error {
	query := `
		UPDATE companies
		SET name = $1, website_url = $2
		WHERE id = $3 AND user_id = $4
		RETURNING created_at
	`

	if userID == "" {
		return fmt.Errorf("user_id is required")
	}

	var createdAt string
	err := r.db.QueryRowContext(r.ctx, query, company.Name, company.WebsiteURL, company.ID, userID).Scan(&createdAt)
	if err == sql.ErrNoRows {
		return fmt.Errorf("company not found or access denied")
	}
	if err != nil {
		slog.Error("db error: failed to update company", "error", err)
		return fmt.Errorf("failed to update company: %w", err)
	}

	company.CreatedAt = createdAt
	return nil
}

func (r *companyRepository) Delete(userID string, id int) error {
	query := `
		DELETE FROM companies
		WHERE id = $1 AND user_id = $2
	`

	if userID == "" {
		return fmt.Errorf("user_id is required")
	}

	result, err := r.db.ExecContext(r.ctx, query, id, userID)
	if err != nil {
		slog.Error("db error: failed to delete company", "error", err)
		return fmt.Errorf("failed to delete company: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		slog.Error("db error: failed to get rows affected", "error", err)
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("company not found or access denied")
	}

	return nil
}

func (r *companyRepository) ExistsByName(userID string, name string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM companies
			WHERE name = $1 AND user_id = $2
		)
	`

	if userID == "" {
		return false, fmt.Errorf("user_id is required")
	}

	var exists bool
	err := r.db.QueryRowContext(r.ctx, query, name, userID).Scan(&exists)
	if err != nil {
		slog.Error("db error: failed to check company existence", "error", err)
		return false, fmt.Errorf("failed to check company existence: %w", err)
	}

	return exists, nil
}

func (r *companyRepository) HasServiceAccounts(id int) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM service_accounts
			WHERE company_id = $1
		)
	`

	var exists bool
	err := r.db.QueryRowContext(r.ctx, query, id).Scan(&exists)
	if err != nil {
		slog.Error("db error: failed to check company service accounts", "error", err)
		return false, fmt.Errorf("failed to check company service accounts: %w", err)
	}

	return exists, nil
}
