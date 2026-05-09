package repository

import (
	"context"
	"strconv"

	"github.com/homepay/api/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AccountRepository interface {
	Create(ctx context.Context, companyID, authUserID string, req *models.CreateAccountRequest) (*models.Account, error)
	GetByID(ctx context.Context, id, authUserID string) (*models.Account, error)
	GetAllFiltered(ctx context.Context, authUserID string, companyID *string, sort, order string, p models.PaginationParams) ([]models.Account, int, error)
	GetAllActiveByUser(ctx context.Context, authUserID string) ([]models.Account, error)
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

const accountCols = `a.id, a.company_id, c.name as company_name, a.group_id, a.account_number, a.name, a.billing_day, a.auto_accumulate, a.is_active, a.created_at, a.deleted_at`

func scanAccount(row pgx.Row, a *models.Account) error {
	return row.Scan(&a.ID, &a.CompanyID, &a.GroupID, &a.AccountNumber, &a.Name, &a.BillingDay, &a.AutoAccumulate, &a.IsActive, &a.CreatedAt, &a.DeletedAt)
}

func scanAccountWithCompanyName(row pgx.Row, a *models.Account) error {
	return row.Scan(&a.ID, &a.CompanyID, &a.CompanyName, &a.GroupID, &a.AccountNumber, &a.Name, &a.BillingDay, &a.AutoAccumulate, &a.IsActive, &a.CreatedAt, &a.DeletedAt)
}

func (r *accountRepo) Create(ctx context.Context, companyID, authUserID string, req *models.CreateAccountRequest) (*models.Account, error) {
	var a models.Account
	err := r.db.QueryRow(ctx, `
		INSERT INTO homepay.accounts (company_id, group_id, account_number, name, billing_day, auto_accumulate)
		SELECT id, $3, $4, $5, $6, $7
		FROM homepay.companies
		WHERE id = $1 AND auth_user_id = $2 AND deleted_at IS NULL
		RETURNING id, company_id, group_id, account_number, name, billing_day, auto_accumulate, is_active, created_at, deleted_at
	`, companyID, authUserID, req.GroupID, req.AccountNumber, req.Name, req.BillingDay, req.AutoAccumulate).Scan(&a.ID, &a.CompanyID, &a.GroupID, &a.AccountNumber, &a.Name, &a.BillingDay, &a.AutoAccumulate, &a.IsActive, &a.CreatedAt, &a.DeletedAt)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *accountRepo) GetByID(ctx context.Context, id, authUserID string) (*models.Account, error) {
	var a models.Account
	err := scanAccountWithCompanyName(r.db.QueryRow(ctx, `
		SELECT a.id, a.company_id, c.name as company_name, a.group_id, a.account_number, a.name, a.billing_day, a.auto_accumulate, a.is_active, a.created_at, a.deleted_at
		FROM homepay.accounts a
		JOIN homepay.companies c ON c.id = a.company_id
		WHERE a.id = $1 AND c.auth_user_id = $2 AND a.deleted_at IS NULL
	`, id, authUserID), &a)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *accountRepo) GetAllFiltered(ctx context.Context, authUserID string, companyID *string, sort, order string, p models.PaginationParams) ([]models.Account, int, error) {
	if sort == "" {
		sort = "created_at"
	}
	if order == "" {
		order = "desc"
	}
	validSorts := map[string]string{
		"created_at":   "a.created_at",
		"name":         "a.name",
		"billing_day":  "a.billing_day",
		"company_name": "c.name",
	}
	orderBy := validSorts[sort]
	if orderBy == "" {
		orderBy = "a.created_at"
	}
	if order != "asc" && order != "desc" {
		order = "desc"
	}

	var total int
	countQuery := `
		SELECT COUNT(*) FROM homepay.accounts a
		JOIN homepay.companies c ON c.id = a.company_id
		WHERE c.auth_user_id = $1 AND a.deleted_at IS NULL`
	args := []interface{}{authUserID}
	if companyID != nil && *companyID != "" {
		countQuery += ` AND a.company_id = $2`
		args = append(args, *companyID)
	}
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT a.id, a.company_id, c.name as company_name, a.group_id, a.account_number, a.name, a.billing_day, a.auto_accumulate, a.is_active, a.created_at, a.deleted_at
		FROM homepay.accounts a
		JOIN homepay.companies c ON c.id = a.company_id
		WHERE c.auth_user_id = $1 AND a.deleted_at IS NULL`
	queryArgs := []interface{}{authUserID}
	argIdx := 2
	if companyID != nil && *companyID != "" {
		query += ` AND a.company_id = $` + strconv.Itoa(argIdx)
		queryArgs = append(queryArgs, *companyID)
		argIdx++
	}
	query += ` ORDER BY ` + orderBy + ` ` + order
	query += ` LIMIT $` + strconv.Itoa(argIdx) + ` OFFSET $` + strconv.Itoa(argIdx+1)
	queryArgs = append(queryArgs, p.Limit, p.Offset())

	rows, err := r.db.Query(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var accounts []models.Account
	for rows.Next() {
		var a models.Account
		if err := rows.Scan(&a.ID, &a.CompanyID, &a.CompanyName, &a.GroupID, &a.AccountNumber, &a.Name, &a.BillingDay, &a.AutoAccumulate, &a.IsActive, &a.CreatedAt, &a.DeletedAt); err != nil {
			return nil, 0, err
		}
		accounts = append(accounts, a)
	}
	return accounts, total, rows.Err()
}

func (r *accountRepo) GetAllActiveByUser(ctx context.Context, authUserID string) ([]models.Account, error) {
	rows, err := r.db.Query(ctx, `
		SELECT a.id, a.company_id, c.name as company_name, a.group_id, a.account_number, a.name, a.billing_day, a.auto_accumulate, a.is_active, a.created_at, a.deleted_at
		FROM homepay.accounts a
		JOIN homepay.companies c ON c.id = a.company_id
		WHERE c.auth_user_id = $1 AND a.deleted_at IS NULL AND c.deleted_at IS NULL
		ORDER BY a.created_at
	`, authUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []models.Account
	for rows.Next() {
		var a models.Account
		if err := rows.Scan(&a.ID, &a.CompanyID, &a.CompanyName, &a.GroupID, &a.AccountNumber, &a.Name, &a.BillingDay, &a.AutoAccumulate, &a.IsActive, &a.CreatedAt, &a.DeletedAt); err != nil {
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
	err := scanAccountWithCompanyName(r.db.QueryRow(ctx, `
		UPDATE homepay.accounts a
		SET group_id        = COALESCE($3, a.group_id),
		    account_number  = COALESCE($4, a.account_number),
		    name            = COALESCE($5, a.name),
		    billing_day     = COALESCE($6, a.billing_day),
		    auto_accumulate = COALESCE($7, a.auto_accumulate)
		FROM homepay.companies c
		WHERE a.id = $1 AND a.company_id = c.id AND c.auth_user_id = $2 AND a.deleted_at IS NULL
		RETURNING a.id, a.company_id, c.name as company_name, a.group_id, a.account_number, a.name, a.billing_day, a.auto_accumulate, a.is_active, a.created_at, a.deleted_at
	`, id, authUserID, req.GroupID, req.AccountNumber, req.Name, req.BillingDay, req.AutoAccumulate), &a)
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
		SET deleted_at = NOW(), is_active = FALSE
		FROM homepay.companies c
		WHERE a.id = $1 AND a.company_id = c.id AND c.auth_user_id = $2 AND a.deleted_at IS NULL
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
		UPDATE homepay.accounts SET deleted_at = NOW(), is_active = FALSE
		WHERE company_id = $1 AND deleted_at IS NULL
	`, companyID)
	return err
}
