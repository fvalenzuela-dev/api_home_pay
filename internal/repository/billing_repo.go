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
GetByAccountAndPeriod(ctx context.Context, accountID, authUserID string, period int) (*models.AccountBilling, error)
	GetAllByAccount(ctx context.Context, accountID, authUserID string, p models.PaginationParams) ([]models.AccountBilling, int, error)
GetUnpaidByAccount(ctx context.Context, accountID, authUserID string) (*models.AccountBilling, error)
	GetAllByPeriod(ctx context.Context, authUserID string, period int, isPaid *bool, p models.PaginationParams) ([]models.AccountBillingWithDetails, int, error)
	BulkInsertForPeriod(ctx context.Context, period int, inserts []models.PeriodBillingInsert) error
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

// billingColsAB — columnas con prefijo ab. para queries con JOIN (evita ambigüedad en created_at / deleted_at)
const billingColsAB = `ab.id, ab.account_id, ab.period, ab.amount_billed, ab.amount_paid, ab.is_paid, ab.paid_at, ab.carried_from, ab.created_at, ab.deleted_at`

func scanBilling(row pgx.Row, b *models.AccountBilling) error {
	return row.Scan(&b.ID, &b.AccountID, &b.Period, &b.AmountBilled, &b.AmountPaid,
		&b.IsPaid, &b.PaidAt, &b.CarriedFrom, &b.CreatedAt, &b.DeletedAt)
}

func (r *billingRepo) Create(ctx context.Context, accountID string, req *models.CreateBillingRequest) (*models.AccountBilling, error) {
	amountPaid := 0.0
	if req.AmountPaid != nil {
		amountPaid = *req.AmountPaid
	}
	isPaid := false
	if req.IsPaid != nil {
		isPaid = *req.IsPaid
	}
	var b models.AccountBilling
	err := scanBilling(r.db.QueryRow(ctx, `
		INSERT INTO homepay.account_billings (account_id, period, amount_billed, amount_paid, is_paid, paid_at, carried_from)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING `+billingCols,
		accountID, req.Period, req.AmountBilled, amountPaid, isPaid, req.PaidAt, req.CarriedFrom), &b)
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
		SELECT `+billingColsAB+`
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
		SELECT `+billingColsAB+`
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

func (r *billingRepo) GetByAccountAndPeriod(ctx context.Context, accountID, authUserID string, period int) (*models.AccountBilling, error) {
	var b models.AccountBilling
	err := scanBilling(r.db.QueryRow(ctx, `
		SELECT `+billingColsAB+`
		FROM homepay.account_billings ab
		JOIN homepay.accounts a ON a.id = ab.account_id
		JOIN homepay.companies c ON c.id = a.company_id
		WHERE ab.account_id = $1 AND c.auth_user_id = $2 AND ab.period = $3 AND ab.deleted_at IS NULL
	`, accountID, authUserID, period), &b)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *billingRepo) GetUnpaidByAccount(ctx context.Context, accountID, authUserID string) (*models.AccountBilling, error) {
	var b models.AccountBilling
	err := scanBilling(r.db.QueryRow(ctx, `
		SELECT `+billingColsAB+`
		FROM homepay.account_billings ab
		JOIN homepay.accounts a ON a.id = ab.account_id
		JOIN homepay.companies c ON c.id = a.company_id
		WHERE ab.account_id = $1 AND c.auth_user_id = $2 AND ab.is_paid = FALSE AND ab.deleted_at IS NULL
		ORDER BY ab.period DESC
		LIMIT 1
	`, accountID, authUserID), &b)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *billingRepo) GetAllByPeriod(ctx context.Context, authUserID string, period int, isPaid *bool, p models.PaginationParams) ([]models.AccountBillingWithDetails, int, error) {
	paidFilter := ""
	if isPaid != nil {
		if *isPaid {
			paidFilter = " AND ab.is_paid = TRUE"
		} else {
			paidFilter = " AND ab.is_paid = FALSE"
		}
	}

	var total int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM homepay.account_billings ab
		JOIN homepay.accounts a ON a.id = ab.account_id
		JOIN homepay.companies c ON c.id = a.company_id
		WHERE c.auth_user_id = $1 AND ab.period = $2 AND ab.deleted_at IS NULL`+paidFilter,
		authUserID, period).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(ctx, `
		SELECT `+billingColsAB+`, cat.name, c.name, a.name
		FROM homepay.account_billings ab
		JOIN homepay.accounts a ON a.id = ab.account_id
		JOIN homepay.companies c ON c.id = a.company_id
		JOIN homepay.categories cat ON cat.id = c.category_id
		WHERE c.auth_user_id = $1 AND ab.period = $2 AND ab.deleted_at IS NULL`+paidFilter+`
		ORDER BY ab.created_at DESC
		LIMIT $3 OFFSET $4`,
		authUserID, period, p.Limit, p.Offset())
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var billings []models.AccountBillingWithDetails
	for rows.Next() {
		var b models.AccountBillingWithDetails
		if err := rows.Scan(
			&b.ID, &b.AccountID, &b.Period, &b.AmountBilled, &b.AmountPaid,
			&b.IsPaid, &b.PaidAt, &b.CarriedFrom, &b.CreatedAt, &b.DeletedAt,
			&b.CategoryName, &b.CompanyName, &b.AccountName,
		); err != nil {
			return nil, 0, err
		}
		billings = append(billings, b)
	}
	return billings, total, rows.Err()
}

func (r *billingRepo) BulkInsertForPeriod(ctx context.Context, period int, inserts []models.PeriodBillingInsert) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, ins := range inserts {
		_, err := tx.Exec(ctx, `
			INSERT INTO homepay.account_billings (account_id, period, amount_billed, carried_from)
			VALUES ($1, $2, $3, $4)
		`, ins.AccountID, period, ins.AmountBilled, ins.CarriedFrom)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *billingRepo) Update(ctx context.Context, id, authUserID string, req *models.UpdateBillingRequest) (*models.AccountBilling, error) {
	var b models.AccountBilling
	err := scanBilling(r.db.QueryRow(ctx, `
		UPDATE homepay.account_billings ab
		SET amount_billed = COALESCE($3, ab.amount_billed),
		    amount_paid   = COALESCE($4, ab.amount_paid),
		    is_paid       = COALESCE($5, ab.is_paid),
		    paid_at       = COALESCE($6, ab.paid_at)
		FROM homepay.accounts a
		JOIN homepay.companies c ON c.id = a.company_id
		WHERE ab.id = $1 AND ab.account_id = a.id AND c.auth_user_id = $2 AND ab.deleted_at IS NULL
		RETURNING `+billingColsAB,
		id, authUserID, req.AmountBilled, req.AmountPaid, req.IsPaid, req.PaidAt), &b)
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
