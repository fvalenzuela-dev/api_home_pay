package repository

import (
	"context"

	"github.com/homepay/api/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AccountRepository interface {
	Create(ctx context.Context, companyID, authUserID string, req *models.CreateAccountRequest) (*models.Account, error)
	GetByID(ctx context.Context, id, authUserID string) (*models.Account, error)
	GetAllByCompany(ctx context.Context, companyID, authUserID string) ([]models.Account, error)
	GetActiveIDsByCompany(ctx context.Context, companyID string) ([]string, error)
	Update(ctx context.Context, id, authUserID string, req *models.UpdateAccountRequest) (*models.Account, error)
	SoftDelete(ctx context.Context, id, authUserID string) error
	SoftDeleteByCompany(ctx context.Context, companyID string) error
}

type accountRepo struct {
	db *pgxpool.Pool
}

func NewAccountRepository(db *pgxpool.Pool) AccountRepository {
	return &accountRepo{db: db}
}

func (r *accountRepo) Create(ctx context.Context, companyID, authUserID string, req *models.CreateAccountRequest) (*models.Account, error) {
	var a models.Account
	err := r.db.QueryRow(ctx, `
		INSERT INTO homepay.accounts (company_id, name, billing_day, auto_accumulate)
		SELECT c.id, $3, $4, $5
		FROM homepay.companies c
		JOIN homepay.users u ON u.id = c.user_id
		WHERE c.id = $1 AND u.auth_user_id = $2 AND c.deleted_at IS NULL
		RETURNING id, company_id, name, billing_day, auto_accumulate, created_at, updated_at, deleted_at
	`, companyID, authUserID, req.Name, req.BillingDay, req.AutoAccumulate).Scan(
		&a.ID, &a.CompanyID, &a.Name, &a.BillingDay, &a.AutoAccumulate, &a.CreatedAt, &a.UpdatedAt, &a.DeletedAt,
	)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *accountRepo) GetByID(ctx context.Context, id, authUserID string) (*models.Account, error) {
	var a models.Account
	err := r.db.QueryRow(ctx, `
		SELECT a.id, a.company_id, a.name, a.billing_day, a.auto_accumulate, a.created_at, a.updated_at, a.deleted_at
		FROM homepay.accounts a
		JOIN homepay.companies c ON c.id = a.company_id
		JOIN homepay.users u ON u.id = c.user_id
		WHERE a.id = $1 AND u.auth_user_id = $2 AND a.deleted_at IS NULL
	`, id, authUserID).Scan(
		&a.ID, &a.CompanyID, &a.Name, &a.BillingDay, &a.AutoAccumulate, &a.CreatedAt, &a.UpdatedAt, &a.DeletedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *accountRepo) GetAllByCompany(ctx context.Context, companyID, authUserID string) ([]models.Account, error) {
	rows, err := r.db.Query(ctx, `
		SELECT a.id, a.company_id, a.name, a.billing_day, a.auto_accumulate, a.created_at, a.updated_at, a.deleted_at
		FROM homepay.accounts a
		JOIN homepay.companies c ON c.id = a.company_id
		JOIN homepay.users u ON u.id = c.user_id
		WHERE a.company_id = $1 AND u.auth_user_id = $2 AND a.deleted_at IS NULL
		ORDER BY a.created_at DESC
	`, companyID, authUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []models.Account
	for rows.Next() {
		var a models.Account
		if err := rows.Scan(&a.ID, &a.CompanyID, &a.Name, &a.BillingDay, &a.AutoAccumulate, &a.CreatedAt, &a.UpdatedAt, &a.DeletedAt); err != nil {
			return nil, err
		}
		accounts = append(accounts, a)
	}
	return accounts, rows.Err()
}

func (r *accountRepo) GetActiveIDsByCompany(ctx context.Context, companyID string) ([]string, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id FROM homepay.accounts WHERE company_id = $1 AND deleted_at IS NULL
	`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (r *accountRepo) Update(ctx context.Context, id, authUserID string, req *models.UpdateAccountRequest) (*models.Account, error) {
	var a models.Account
	err := r.db.QueryRow(ctx, `
		UPDATE homepay.accounts a
		SET name = COALESCE($3, a.name),
		    billing_day = COALESCE($4, a.billing_day),
		    auto_accumulate = COALESCE($5, a.auto_accumulate),
		    updated_at = NOW()
		FROM homepay.companies c
		JOIN homepay.users u ON u.id = c.user_id
		WHERE a.id = $1 AND a.company_id = c.id AND u.auth_user_id = $2 AND a.deleted_at IS NULL
		RETURNING a.id, a.company_id, a.name, a.billing_day, a.auto_accumulate, a.created_at, a.updated_at, a.deleted_at
	`, id, authUserID, req.Name, req.BillingDay, req.AutoAccumulate).Scan(
		&a.ID, &a.CompanyID, &a.Name, &a.BillingDay, &a.AutoAccumulate, &a.CreatedAt, &a.UpdatedAt, &a.DeletedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *accountRepo) SoftDelete(ctx context.Context, id, authUserID string) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE homepay.accounts a
		SET deleted_at = NOW()
		FROM homepay.companies c
		JOIN homepay.users u ON u.id = c.user_id
		WHERE a.id = $1 AND a.company_id = c.id AND u.auth_user_id = $2 AND a.deleted_at IS NULL
	`, id, authUserID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *accountRepo) SoftDeleteByCompany(ctx context.Context, companyID string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE homepay.accounts SET deleted_at = NOW()
		WHERE company_id = $1 AND deleted_at IS NULL
	`, companyID)
	return err
}
