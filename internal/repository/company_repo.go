package repository

import (
	"context"
	"fmt"

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

func (r *companyRepo) Create(ctx context.Context, authUserID string, req *models.CreateCompanyRequest) (*models.Company, error) {
	var c models.Company
	err := r.db.QueryRow(ctx, `
		INSERT INTO homepay.companies (user_id, name, category)
		SELECT id, $2, $3 FROM homepay.users WHERE auth_user_id = $1 AND deleted_at IS NULL
		RETURNING id, user_id, name, category, created_at, updated_at, deleted_at
	`, authUserID, req.Name, req.Category).Scan(
		&c.ID, &c.UserID, &c.Name, &c.Category, &c.CreatedAt, &c.UpdatedAt, &c.DeletedAt,
	)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *companyRepo) GetByID(ctx context.Context, id, authUserID string) (*models.Company, error) {
	var c models.Company
	err := r.db.QueryRow(ctx, `
		SELECT c.id, c.user_id, c.name, c.category, c.created_at, c.updated_at, c.deleted_at
		FROM homepay.companies c
		JOIN homepay.users u ON u.id = c.user_id
		WHERE c.id = $1 AND u.auth_user_id = $2 AND c.deleted_at IS NULL
	`, id, authUserID).Scan(
		&c.ID, &c.UserID, &c.Name, &c.Category, &c.CreatedAt, &c.UpdatedAt, &c.DeletedAt,
	)
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
		SELECT c.id, c.user_id, c.name, c.category, c.created_at, c.updated_at, c.deleted_at
		FROM homepay.companies c
		JOIN homepay.users u ON u.id = c.user_id
		WHERE u.auth_user_id = $1 AND c.deleted_at IS NULL
		ORDER BY c.created_at DESC
	`, authUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var companies []models.Company
	for rows.Next() {
		var c models.Company
		if err := rows.Scan(&c.ID, &c.UserID, &c.Name, &c.Category, &c.CreatedAt, &c.UpdatedAt, &c.DeletedAt); err != nil {
			return nil, err
		}
		companies = append(companies, c)
	}
	return companies, rows.Err()
}

func (r *companyRepo) Update(ctx context.Context, id, authUserID string, req *models.UpdateCompanyRequest) (*models.Company, error) {
	var c models.Company
	err := r.db.QueryRow(ctx, `
		UPDATE homepay.companies c
		SET name = COALESCE($3, c.name),
		    category = COALESCE($4, c.category),
		    updated_at = NOW()
		FROM homepay.users u
		WHERE c.id = $1 AND u.id = c.user_id AND u.auth_user_id = $2 AND c.deleted_at IS NULL
		RETURNING c.id, c.user_id, c.name, c.category, c.created_at, c.updated_at, c.deleted_at
	`, id, authUserID, req.Name, req.Category).Scan(
		&c.ID, &c.UserID, &c.Name, &c.Category, &c.CreatedAt, &c.UpdatedAt, &c.DeletedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("update company: %w", err)
	}
	return &c, nil
}

func (r *companyRepo) SoftDelete(ctx context.Context, id, authUserID string) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE homepay.companies c
		SET deleted_at = NOW()
		FROM homepay.users u
		WHERE c.id = $1 AND u.id = c.user_id AND u.auth_user_id = $2 AND c.deleted_at IS NULL
	`, id, authUserID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
