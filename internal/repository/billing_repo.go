package repository

import (
	"context"

	"github.com/homepay/api/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BillingRepository interface {
	Create(ctx context.Context, accountID string, req *models.CreateBillingRequest) (*models.AccountBilling, error)
	CreateCarryOver(ctx context.Context, accountID string, period int, amount float64, carriedFrom string) (*models.AccountBilling, error)
	GetByID(ctx context.Context, id, authUserID string) (*models.AccountBilling, error)
	GetAllByAccount(ctx context.Context, accountID, authUserID string, p models.PaginationParams) ([]models.AccountBilling, int, error)
	GetUnpaidByAccount(ctx context.Context, accountID string) (*models.AccountBilling, error)
	GetAllByPeriod(ctx context.Context, authUserID string, period int, p models.PaginationParams) ([]models.AccountBilling, int, error)
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

const billingCols = `id, account_id, period, amount_billed, amount_paid, is_paid, paid_at, carried_from, created_at, deleted_at`

func scanBilling(row pgx.Row, b *models.AccountBilling) error {
	return row.Scan(&b.ID, &b.AccountID, &b.Period, &b.AmountBilled, &b.AmountPaid,
		&b.IsPaid, &b.PaidAt, &b.CarriedFrom, &b.CreatedAt, &b.DeletedAt)
}

func (r *billingRepo) Create(ctx context.Context, accountID string, req *models.CreateBillingRequest) (*models.AccountBilling, error) {
	var b models.AccountBilling
	err := scanBilling(r.db.QueryRow(ctx, `
		INSERT INTO homepay.account_billings (account_id, period, amount_billed)
		VALUES ($1, $2, $3)
		RETURNING `+billingCols,
		accountID, req.Period, req.AmountBilled), &b)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *billingRepo) CreateCarryOver(ctx context.Context, accountID string, period int, amount float64, carriedFrom string) (*models.AccountBilling, error) {
	var b models.AccountBilling
	err := scanBilling(r.db.QueryRow(ctx, `
		INSERT INTO homepay.account_billings (account_id, period, amount_billed, carried_from)
		VALUES ($1, $2, $3, $4)
		RETURNING `+billingCols,
		accountID, period, amount, carriedFrom), &b)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *billingRepo) GetByID(ctx context.Context, id, authUserID string) (*models.AccountBilling, error) {
	var b models.AccountBilling
	err := scanBilling(r.db.QueryRow(ctx, `
		SELECT ab.`+billingCols+`
		FROM homepay.account_billings ab
		JOIN homepay.accounts a ON a.id = ab.account_id
		JOIN homepay.companies c ON c.id = a.company_id
		WHERE ab.id = $1 AND c.auth_user_id = $2 AND ab.deleted_at IS NULL
	`, id, authUserID), &b)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *billingRepo) GetAllByAccount(ctx context.Context, accountID, authUserID string, p models.PaginationParams) ([]models.AccountBilling, int, error) {
	var total int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM homepay.account_billings ab
		JOIN homepay.accounts a ON a.id = ab.account_id
		JOIN homepay.companies c ON c.id = a.company_id
		WHERE ab.account_id = $1 AND c.auth_user_id = $2 AND ab.deleted_at IS NULL
	`, accountID, authUserID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(ctx, `
		SELECT ab.`+billingCols+`
		FROM homepay.account_billings ab
		JOIN homepay.accounts a ON a.id = ab.account_id
		JOIN homepay.companies c ON c.id = a.company_id
		WHERE ab.account_id = $1 AND c.auth_user_id = $2 AND ab.deleted_at IS NULL
		ORDER BY ab.period DESC
		LIMIT $3 OFFSET $4
	`, accountID, authUserID, p.Limit, p.Offset())
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var billings []models.AccountBilling
	for rows.Next() {
		var b models.AccountBilling
		if err := rows.Scan(&b.ID, &b.AccountID, &b.Period, &b.AmountBilled, &b.AmountPaid,
			&b.IsPaid, &b.PaidAt, &b.CarriedFrom, &b.CreatedAt, &b.DeletedAt); err != nil {
			return nil, 0, err
		}
		billings = append(billings, b)
	}
	return billings, total, rows.Err()
}

func (r *billingRepo) GetUnpaidByAccount(ctx context.Context, accountID string) (*models.AccountBilling, error) {
	var b models.AccountBilling
	err := scanBilling(r.db.QueryRow(ctx, `
		SELECT `+billingCols+`
		FROM homepay.account_billings
		WHERE account_id = $1 AND is_paid = FALSE AND deleted_at IS NULL
		ORDER BY period DESC
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

func (r *billingRepo) GetAllByPeriod(ctx context.Context, authUserID string, period int, p models.PaginationParams) ([]models.AccountBilling, int, error) {
	var total int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM homepay.account_billings ab
		JOIN homepay.accounts a ON a.id = ab.account_id
		JOIN homepay.companies c ON c.id = a.company_id
		WHERE c.auth_user_id = $1 AND ab.period = $2 AND ab.deleted_at IS NULL
	`, authUserID, period).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(ctx, `
		SELECT ab.`+billingCols+`
		FROM homepay.account_billings ab
		JOIN homepay.accounts a ON a.id = ab.account_id
		JOIN homepay.companies c ON c.id = a.company_id
		WHERE c.auth_user_id = $1 AND ab.period = $2 AND ab.deleted_at IS NULL
		ORDER BY ab.period DESC
		LIMIT $3 OFFSET $4
	`, authUserID, period, p.Limit, p.Offset())
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var billings []models.AccountBilling
	for rows.Next() {
		var b models.AccountBilling
		if err := rows.Scan(&b.ID, &b.AccountID, &b.Period, &b.AmountBilled, &b.AmountPaid,
			&b.IsPaid, &b.PaidAt, &b.CarriedFrom, &b.CreatedAt, &b.DeletedAt); err != nil {
			return nil, 0, err
		}
		billings = append(billings, b)
	}
	return billings, total, rows.Err()
}

func (r *billingRepo) Update(ctx context.Context, id, authUserID string, req *models.UpdateBillingRequest) (*models.AccountBilling, error) {
	var b models.AccountBilling
	err := scanBilling(r.db.QueryRow(ctx, `
		UPDATE homepay.account_billings ab
		SET amount_paid = COALESCE($3, ab.amount_paid)
		FROM homepay.accounts a
		JOIN homepay.companies c ON c.id = a.company_id
		WHERE ab.id = $1 AND ab.account_id = a.id AND c.auth_user_id = $2 AND ab.deleted_at IS NULL
		RETURNING ab.`+billingCols,
		id, authUserID, req.AmountPaid), &b)
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
		SET is_paid = TRUE, paid_at = CURRENT_DATE
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
