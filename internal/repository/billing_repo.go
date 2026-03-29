package repository

import (
	"context"

	"github.com/homepay/api/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BillingRepository interface {
	Create(ctx context.Context, accountID string, req *models.CreateBillingRequest) (*models.AccountBilling, error)
	CreateCarryOver(ctx context.Context, accountID string, month, year int, amount float64, carriedFrom string) (*models.AccountBilling, error)
	GetByID(ctx context.Context, id, authUserID string) (*models.AccountBilling, error)
	GetAllByAccount(ctx context.Context, accountID, authUserID string) ([]models.AccountBilling, error)
	GetUnpaidByAccount(ctx context.Context, accountID string) (*models.AccountBilling, error)
	GetAllByMonth(ctx context.Context, authUserID string, month, year int) ([]models.AccountBilling, error)
	Update(ctx context.Context, id, authUserID string, req *models.UpdateBillingRequest) (*models.AccountBilling, error)
	MarkPaid(ctx context.Context, id string) error
	SoftDeleteByAccount(ctx context.Context, accountID string) error
}

type billingRepo struct {
	db *pgxpool.Pool
}

func NewBillingRepository(db *pgxpool.Pool) BillingRepository {
	return &billingRepo{db: db}
}

func scanBilling(row pgx.Row, b *models.AccountBilling) error {
	return row.Scan(
		&b.ID, &b.AccountID, &b.Month, &b.Year,
		&b.AmountBilled, &b.AmountPaid, &b.IsPaid, &b.PaidAt,
		&b.CarriedFrom, &b.CreatedAt, &b.UpdatedAt, &b.DeletedAt,
	)
}

func (r *billingRepo) Create(ctx context.Context, accountID string, req *models.CreateBillingRequest) (*models.AccountBilling, error) {
	var b models.AccountBilling
	err := scanBilling(r.db.QueryRow(ctx, `
		INSERT INTO homepay.account_billings (account_id, month, year, amount_billed)
		VALUES ($1, $2, $3, $4)
		RETURNING id, account_id, month, year, amount_billed, amount_paid, is_paid, paid_at,
		          carried_from, created_at, updated_at, deleted_at
	`, accountID, req.Month, req.Year, req.AmountBilled), &b)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *billingRepo) CreateCarryOver(ctx context.Context, accountID string, month, year int, amount float64, carriedFrom string) (*models.AccountBilling, error) {
	var b models.AccountBilling
	err := scanBilling(r.db.QueryRow(ctx, `
		INSERT INTO homepay.account_billings (account_id, month, year, amount_billed, carried_from)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, account_id, month, year, amount_billed, amount_paid, is_paid, paid_at,
		          carried_from, created_at, updated_at, deleted_at
	`, accountID, month, year, amount, carriedFrom), &b)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *billingRepo) GetByID(ctx context.Context, id, authUserID string) (*models.AccountBilling, error) {
	var b models.AccountBilling
	err := scanBilling(r.db.QueryRow(ctx, `
		SELECT ab.id, ab.account_id, ab.month, ab.year, ab.amount_billed, ab.amount_paid,
		       ab.is_paid, ab.paid_at, ab.carried_from, ab.created_at, ab.updated_at, ab.deleted_at
		FROM homepay.account_billings ab
		JOIN homepay.accounts a ON a.id = ab.account_id
		JOIN homepay.companies c ON c.id = a.company_id
		JOIN homepay.users u ON u.id = c.user_id
		WHERE ab.id = $1 AND u.auth_user_id = $2 AND ab.deleted_at IS NULL
	`, id, authUserID), &b)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *billingRepo) GetAllByAccount(ctx context.Context, accountID, authUserID string) ([]models.AccountBilling, error) {
	rows, err := r.db.Query(ctx, `
		SELECT ab.id, ab.account_id, ab.month, ab.year, ab.amount_billed, ab.amount_paid,
		       ab.is_paid, ab.paid_at, ab.carried_from, ab.created_at, ab.updated_at, ab.deleted_at
		FROM homepay.account_billings ab
		JOIN homepay.accounts a ON a.id = ab.account_id
		JOIN homepay.companies c ON c.id = a.company_id
		JOIN homepay.users u ON u.id = c.user_id
		WHERE ab.account_id = $1 AND u.auth_user_id = $2 AND ab.deleted_at IS NULL
		ORDER BY ab.year DESC, ab.month DESC
	`, accountID, authUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var billings []models.AccountBilling
	for rows.Next() {
		var b models.AccountBilling
		if err := rows.Scan(
			&b.ID, &b.AccountID, &b.Month, &b.Year,
			&b.AmountBilled, &b.AmountPaid, &b.IsPaid, &b.PaidAt,
			&b.CarriedFrom, &b.CreatedAt, &b.UpdatedAt, &b.DeletedAt,
		); err != nil {
			return nil, err
		}
		billings = append(billings, b)
	}
	return billings, rows.Err()
}

func (r *billingRepo) GetUnpaidByAccount(ctx context.Context, accountID string) (*models.AccountBilling, error) {
	var b models.AccountBilling
	err := scanBilling(r.db.QueryRow(ctx, `
		SELECT id, account_id, month, year, amount_billed, amount_paid, is_paid, paid_at,
		       carried_from, created_at, updated_at, deleted_at
		FROM homepay.account_billings
		WHERE account_id = $1 AND is_paid = FALSE AND deleted_at IS NULL
		ORDER BY year DESC, month DESC
		LIMIT 1
	`, accountID), &b)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *billingRepo) GetAllByMonth(ctx context.Context, authUserID string, month, year int) ([]models.AccountBilling, error) {
	rows, err := r.db.Query(ctx, `
		SELECT ab.id, ab.account_id, ab.month, ab.year, ab.amount_billed, ab.amount_paid,
		       ab.is_paid, ab.paid_at, ab.carried_from, ab.created_at, ab.updated_at, ab.deleted_at
		FROM homepay.account_billings ab
		JOIN homepay.accounts a ON a.id = ab.account_id
		JOIN homepay.companies c ON c.id = a.company_id
		JOIN homepay.users u ON u.id = c.user_id
		WHERE u.auth_user_id = $1 AND ab.month = $2 AND ab.year = $3 AND ab.deleted_at IS NULL
	`, authUserID, month, year)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var billings []models.AccountBilling
	for rows.Next() {
		var b models.AccountBilling
		if err := rows.Scan(
			&b.ID, &b.AccountID, &b.Month, &b.Year,
			&b.AmountBilled, &b.AmountPaid, &b.IsPaid, &b.PaidAt,
			&b.CarriedFrom, &b.CreatedAt, &b.UpdatedAt, &b.DeletedAt,
		); err != nil {
			return nil, err
		}
		billings = append(billings, b)
	}
	return billings, rows.Err()
}

func (r *billingRepo) Update(ctx context.Context, id, authUserID string, req *models.UpdateBillingRequest) (*models.AccountBilling, error) {
	var b models.AccountBilling
	err := scanBilling(r.db.QueryRow(ctx, `
		UPDATE homepay.account_billings ab
		SET amount_paid = COALESCE($3, ab.amount_paid),
		    is_paid = COALESCE($4, ab.is_paid),
		    updated_at = NOW()
		FROM homepay.accounts a
		JOIN homepay.companies c ON c.id = a.company_id
		JOIN homepay.users u ON u.id = c.user_id
		WHERE ab.id = $1 AND ab.account_id = a.id AND u.auth_user_id = $2 AND ab.deleted_at IS NULL
		RETURNING ab.id, ab.account_id, ab.month, ab.year, ab.amount_billed, ab.amount_paid,
		          ab.is_paid, ab.paid_at, ab.carried_from, ab.created_at, ab.updated_at, ab.deleted_at
	`, id, authUserID, req.AmountPaid, req.IsPaid), &b)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *billingRepo) MarkPaid(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE homepay.account_billings
		SET is_paid = TRUE, paid_at = NOW(), updated_at = NOW()
		WHERE id = $1
	`, id)
	return err
}

func (r *billingRepo) SoftDeleteByAccount(ctx context.Context, accountID string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE homepay.account_billings SET deleted_at = NOW()
		WHERE account_id = $1 AND deleted_at IS NULL
	`, accountID)
	return err
}
