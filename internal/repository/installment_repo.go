package repository

import (
	"context"

	"github.com/homepay/api/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type InstallmentRepository interface {
	CreatePlan(ctx context.Context, authUserID string, plan *models.InstallmentPlan) (*models.InstallmentPlan, error)
	CreatePayments(ctx context.Context, payments []models.InstallmentPayment) error
	GetPlan(ctx context.Context, id, authUserID string) (*models.InstallmentPlan, error)
	GetAllPlans(ctx context.Context, authUserID string, p models.PaginationParams) ([]models.InstallmentPlan, int, error)
	GetPaymentsByPlan(ctx context.Context, planID string, p models.PaginationParams) ([]models.InstallmentPayment, int, error)
	GetActivePaymentsByMonth(ctx context.Context, authUserID string, month, year int) ([]models.InstallmentPayment, error)
	UpdatePayment(ctx context.Context, planID, paymentID, authUserID string) (*models.InstallmentPayment, error)
	IncrementPaid(ctx context.Context, planID string, total int) error
	SoftDeletePlan(ctx context.Context, id, authUserID string) error
}

type installmentRepo struct {
	db *pgxpool.Pool
}

func NewInstallmentRepository(db *pgxpool.Pool) InstallmentRepository {
	return &installmentRepo{db: db}
}

const planCols = `id, auth_user_id, description, total_amount, total_installments, installments_paid, start_date, is_completed, created_at, deleted_at`
const paymentCols = `id, plan_id, installment_number, amount, due_date, is_paid, paid_at, created_at, deleted_at`

func scanPlan(row pgx.Row, p *models.InstallmentPlan) error {
	return row.Scan(&p.ID, &p.AuthUserID, &p.Description, &p.TotalAmount, &p.TotalInstallments,
		&p.InstallmentsPaid, &p.StartDate, &p.IsCompleted, &p.CreatedAt, &p.DeletedAt)
}

func scanPayment(row pgx.Row, p *models.InstallmentPayment) error {
	return row.Scan(&p.ID, &p.PlanID, &p.InstallmentNumber, &p.Amount, &p.DueDate,
		&p.IsPaid, &p.PaidAt, &p.CreatedAt, &p.DeletedAt)
}

func (r *installmentRepo) CreatePlan(ctx context.Context, authUserID string, plan *models.InstallmentPlan) (*models.InstallmentPlan, error) {
	var p models.InstallmentPlan
	err := scanPlan(r.db.QueryRow(ctx, `
		INSERT INTO homepay.installment_plans (auth_user_id, description, total_amount, total_installments, start_date)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING `+planCols,
		authUserID, plan.Description, plan.TotalAmount, plan.TotalInstallments, plan.StartDate), &p)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *installmentRepo) CreatePayments(ctx context.Context, payments []models.InstallmentPayment) error {
	for _, p := range payments {
		_, err := r.db.Exec(ctx, `
			INSERT INTO homepay.installment_payments (plan_id, installment_number, amount, due_date)
			VALUES ($1, $2, $3, $4)
		`, p.PlanID, p.InstallmentNumber, p.Amount, p.DueDate)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *installmentRepo) GetPlan(ctx context.Context, id, authUserID string) (*models.InstallmentPlan, error) {
	var p models.InstallmentPlan
	err := scanPlan(r.db.QueryRow(ctx, `
		SELECT `+planCols+`
		FROM homepay.installment_plans
		WHERE id = $1 AND auth_user_id = $2 AND deleted_at IS NULL
	`, id, authUserID), &p)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *installmentRepo) GetAllPlans(ctx context.Context, authUserID string, p models.PaginationParams) ([]models.InstallmentPlan, int, error) {
	var total int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM homepay.installment_plans
		WHERE auth_user_id = $1 AND deleted_at IS NULL
	`, authUserID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(ctx, `
		SELECT `+planCols+`
		FROM homepay.installment_plans
		WHERE auth_user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, authUserID, p.Limit, p.Offset())
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var plans []models.InstallmentPlan
	for rows.Next() {
		var pl models.InstallmentPlan
		if err := rows.Scan(&pl.ID, &pl.AuthUserID, &pl.Description, &pl.TotalAmount, &pl.TotalInstallments,
			&pl.InstallmentsPaid, &pl.StartDate, &pl.IsCompleted, &pl.CreatedAt, &pl.DeletedAt); err != nil {
			return nil, 0, err
		}
		plans = append(plans, pl)
	}
	return plans, total, rows.Err()
}

func (r *installmentRepo) GetPaymentsByPlan(ctx context.Context, planID string, p models.PaginationParams) ([]models.InstallmentPayment, int, error) {
	var total int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM homepay.installment_payments
		WHERE plan_id = $1 AND deleted_at IS NULL
	`, planID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(ctx, `
		SELECT `+paymentCols+`
		FROM homepay.installment_payments
		WHERE plan_id = $1 AND deleted_at IS NULL
		ORDER BY installment_number
		LIMIT $2 OFFSET $3
	`, planID, p.Limit, p.Offset())
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var payments []models.InstallmentPayment
	for rows.Next() {
		var pay models.InstallmentPayment
		if err := rows.Scan(&pay.ID, &pay.PlanID, &pay.InstallmentNumber, &pay.Amount, &pay.DueDate,
			&pay.IsPaid, &pay.PaidAt, &pay.CreatedAt, &pay.DeletedAt); err != nil {
			return nil, 0, err
		}
		payments = append(payments, pay)
	}
	return payments, total, rows.Err()
}

func (r *installmentRepo) GetActivePaymentsByMonth(ctx context.Context, authUserID string, month, year int) ([]models.InstallmentPayment, error) {
	rows, err := r.db.Query(ctx, `
		SELECT ip.`+paymentCols+`
		FROM homepay.installment_payments ip
		JOIN homepay.installment_plans pl ON pl.id = ip.plan_id
		WHERE pl.auth_user_id = $1
		  AND EXTRACT(MONTH FROM ip.due_date) = $2
		  AND EXTRACT(YEAR  FROM ip.due_date) = $3
		  AND pl.deleted_at IS NULL
		  AND ip.deleted_at IS NULL
		ORDER BY ip.due_date
	`, authUserID, month, year)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []models.InstallmentPayment
	for rows.Next() {
		var p models.InstallmentPayment
		if err := rows.Scan(&p.ID, &p.PlanID, &p.InstallmentNumber, &p.Amount, &p.DueDate,
			&p.IsPaid, &p.PaidAt, &p.CreatedAt, &p.DeletedAt); err != nil {
			return nil, err
		}
		payments = append(payments, p)
	}
	return payments, rows.Err()
}

func (r *installmentRepo) UpdatePayment(ctx context.Context, planID, paymentID, authUserID string) (*models.InstallmentPayment, error) {
	var p models.InstallmentPayment
	err := scanPayment(r.db.QueryRow(ctx, `
		UPDATE homepay.installment_payments ip
		SET is_paid = TRUE, paid_at = CURRENT_DATE
		FROM homepay.installment_plans pl
		WHERE ip.id = $1 AND ip.plan_id = $2 AND pl.id = ip.plan_id AND pl.auth_user_id = $3
		  AND ip.is_paid = FALSE AND pl.deleted_at IS NULL AND ip.deleted_at IS NULL
		RETURNING ip.`+paymentCols,
		paymentID, planID, authUserID), &p)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *installmentRepo) IncrementPaid(ctx context.Context, planID string, total int) error {
	_, err := r.db.Exec(ctx, `
		UPDATE homepay.installment_plans
		SET installments_paid = installments_paid + 1,
		    is_completed = (installments_paid + 1 >= $2)
		WHERE id = $1
	`, planID, total)
	return err
}

func (r *installmentRepo) SoftDeletePlan(ctx context.Context, id, authUserID string) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE homepay.installment_plans
		SET deleted_at = NOW()
		WHERE id = $1 AND auth_user_id = $2 AND deleted_at IS NULL
	`, id, authUserID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	// También elimina los pagos del plan
	_, err = r.db.Exec(ctx, `
		UPDATE homepay.installment_payments SET deleted_at = NOW()
		WHERE plan_id = $1 AND deleted_at IS NULL
	`, id)
	return err
}
