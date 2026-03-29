package repository

import (
	"context"

	"github.com/homepay/api/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CompanyRepository interface {
	Create(ctx context.Context, authUserID string, req *models.CreateCompanyRequest) (*models.Company, error)
	GetByID(ctx context.Context, id, authUserID string) (*models.Company, error)
	GetAll(ctx context.Context, authUserID string) ([]models.Company, error)
	Update(ctx context.Context, id, authUserID string, req *models.UpdateCompanyRequest) (*models.Company, error)
	SoftDelete(ctx context.Context, id, authUserID string) error
}

type companyRepo struct {
	db *pgxpool.Pool
}

func NewCompanyRepository(db *pgxpool.Pool) CompanyRepository {
	return &companyRepo{db: db}
}

func scanCompany(row pgx.Row, c *models.Company) error {
	return row.Scan(&c.ID, &c.AuthUserID, &c.Name, &c.Category, &c.IsActive, &c.CreatedAt, &c.DeletedAt)
}

func (r *companyRepo) Create(ctx context.Context, authUserID string, req *models.CreateCompanyRequest) (*models.Company, error) {
	var c models.Company
	err := scanCompany(r.db.QueryRow(ctx, `
		INSERT INTO homepay.companies (auth_user_id, name, category)
		VALUES ($1, $2, $3)
		RETURNING id, auth_user_id, name, category, is_active, created_at, deleted_at
	`, authUserID, req.Name, req.Category), &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *companyRepo) GetByID(ctx context.Context, id, authUserID string) (*models.Company, error) {
	var c models.Company
	err := scanCompany(r.db.QueryRow(ctx, `
		SELECT id, auth_user_id, name, category, is_active, created_at, deleted_at
		FROM homepay.companies
		WHERE id = $1 AND auth_user_id = $2 AND deleted_at IS NULL
	`, id, authUserID), &c)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *companyRepo) GetAll(ctx context.Context, authUserID string) ([]models.Company, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, auth_user_id, name, category, is_active, created_at, deleted_at
		FROM homepay.companies
		WHERE auth_user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`, authUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var companies []models.Company
	for rows.Next() {
		var c models.Company
		if err := rows.Scan(&c.ID, &c.AuthUserID, &c.Name, &c.Category, &c.IsActive, &c.CreatedAt, &c.DeletedAt); err != nil {
			return nil, err
		}
		companies = append(companies, c)
	}
	return companies, rows.Err()
}

func (r *companyRepo) Update(ctx context.Context, id, authUserID string, req *models.UpdateCompanyRequest) (*models.Company, error) {
	var c models.Company
	err := scanCompany(r.db.QueryRow(ctx, `
		UPDATE homepay.companies
		SET name     = COALESCE($3, name),
		    category = COALESCE($4, category)
		WHERE id = $1 AND auth_user_id = $2 AND deleted_at IS NULL
		RETURNING id, auth_user_id, name, category, is_active, created_at, deleted_at
	`, id, authUserID, req.Name, req.Category), &c)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *companyRepo) SoftDelete(ctx context.Context, id, authUserID string) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE homepay.companies
		SET deleted_at = NOW(), is_active = FALSE
		WHERE id = $1 AND auth_user_id = $2 AND deleted_at IS NULL
	`, id, authUserID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
