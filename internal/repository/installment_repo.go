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
	GetAllPlans(ctx context.Context, authUserID string) ([]models.InstallmentPlan, error)
	GetPaymentsByPlan(ctx context.Context, planID string) ([]models.InstallmentPayment, error)
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

func (r *installmentRepo) CreatePlan(ctx context.Context, authUserID string, plan *models.InstallmentPlan) (*models.InstallmentPlan, error) {
	var p models.InstallmentPlan
	err := r.db.QueryRow(ctx, `
		INSERT INTO homepay.installment_plans (user_id, description, total_amount, total_installments, start_date)
		SELECT id, $2, $3, $4, $5 FROM homepay.users WHERE auth_user_id = $1 AND deleted_at IS NULL
		RETURNING id, user_id, description, total_amount, total_installments, installments_paid,
		          start_date, is_completed, created_at, updated_at, deleted_at
	`, authUserID, plan.Description, plan.TotalAmount, plan.TotalInstallments, plan.StartDate).Scan(
		&p.ID, &p.UserID, &p.Description, &p.TotalAmount, &p.TotalInstallments,
		&p.InstallmentsPaid, &p.StartDate, &p.IsCompleted, &p.CreatedAt, &p.UpdatedAt, &p.DeletedAt,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *installmentRepo) CreatePayments(ctx context.Context, payments []models.InstallmentPayment) error {
	for _, p := range payments {
		_, err := r.db.Exec(ctx, `
			INSERT INTO homepay.installment_payments (plan_id, number, amount, due_date)
			VALUES ($1, $2, $3, $4)
		`, p.PlanID, p.Number, p.Amount, p.DueDate)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *installmentRepo) GetPlan(ctx context.Context, id, authUserID string) (*models.InstallmentPlan, error) {
	var p models.InstallmentPlan
	err := r.db.QueryRow(ctx, `
		SELECT ip.id, ip.user_id, ip.description, ip.total_amount, ip.total_installments,
		       ip.installments_paid, ip.start_date, ip.is_completed, ip.created_at, ip.updated_at, ip.deleted_at
		FROM homepay.installment_plans ip
		JOIN homepay.users u ON u.id = ip.user_id
		WHERE ip.id = $1 AND u.auth_user_id = $2 AND ip.deleted_at IS NULL
	`, id, authUserID).Scan(
		&p.ID, &p.UserID, &p.Description, &p.TotalAmount, &p.TotalInstallments,
		&p.InstallmentsPaid, &p.StartDate, &p.IsCompleted, &p.CreatedAt, &p.UpdatedAt, &p.DeletedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *installmentRepo) GetAllPlans(ctx context.Context, authUserID string) ([]models.InstallmentPlan, error) {
	rows, err := r.db.Query(ctx, `
		SELECT ip.id, ip.user_id, ip.description, ip.total_amount, ip.total_installments,
		       ip.installments_paid, ip.start_date, ip.is_completed, ip.created_at, ip.updated_at, ip.deleted_at
		FROM homepay.installment_plans ip
		JOIN homepay.users u ON u.id = ip.user_id
		WHERE u.auth_user_id = $1 AND ip.deleted_at IS NULL
		ORDER BY ip.created_at DESC
	`, authUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []models.InstallmentPlan
	for rows.Next() {
		var p models.InstallmentPlan
		if err := rows.Scan(
			&p.ID, &p.UserID, &p.Description, &p.TotalAmount, &p.TotalInstallments,
			&p.InstallmentsPaid, &p.StartDate, &p.IsCompleted, &p.CreatedAt, &p.UpdatedAt, &p.DeletedAt,
		); err != nil {
			return nil, err
		}
		plans = append(plans, p)
	}
	return plans, rows.Err()
}

func (r *installmentRepo) GetPaymentsByPlan(ctx context.Context, planID string) ([]models.InstallmentPayment, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, plan_id, number, amount, due_date, is_paid, paid_at, created_at, updated_at
		FROM homepay.installment_payments
		WHERE plan_id = $1
		ORDER BY number
	`, planID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []models.InstallmentPayment
	for rows.Next() {
		var p models.InstallmentPayment
		if err := rows.Scan(&p.ID, &p.PlanID, &p.Number, &p.Amount, &p.DueDate,
			&p.IsPaid, &p.PaidAt, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		payments = append(payments, p)
	}
	return payments, rows.Err()
}

func (r *installmentRepo) GetActivePaymentsByMonth(ctx context.Context, authUserID string, month, year int) ([]models.InstallmentPayment, error) {
	rows, err := r.db.Query(ctx, `
		SELECT ip.id, ip.plan_id, ip.number, ip.amount, ip.due_date, ip.is_paid, ip.paid_at,
		       ip.created_at, ip.updated_at
		FROM homepay.installment_payments ip
		JOIN homepay.installment_plans pl ON pl.id = ip.plan_id
		JOIN homepay.users u ON u.id = pl.user_id
		WHERE u.auth_user_id = $1
		  AND EXTRACT(MONTH FROM ip.due_date) = $2
		  AND EXTRACT(YEAR FROM ip.due_date) = $3
		  AND pl.deleted_at IS NULL
		ORDER BY ip.due_date
	`, authUserID, month, year)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []models.InstallmentPayment
	for rows.Next() {
		var p models.InstallmentPayment
		if err := rows.Scan(&p.ID, &p.PlanID, &p.Number, &p.Amount, &p.DueDate,
			&p.IsPaid, &p.PaidAt, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		payments = append(payments, p)
	}
	return payments, rows.Err()
}

func (r *installmentRepo) UpdatePayment(ctx context.Context, planID, paymentID, authUserID string) (*models.InstallmentPayment, error) {
	var p models.InstallmentPayment
	err := r.db.QueryRow(ctx, `
		UPDATE homepay.installment_payments ip
		SET is_paid = TRUE, paid_at = NOW(), updated_at = NOW()
		FROM homepay.installment_plans pl
		JOIN homepay.users u ON u.id = pl.user_id
		WHERE ip.id = $1 AND ip.plan_id = $2 AND pl.id = ip.plan_id AND u.auth_user_id = $3
		  AND ip.is_paid = FALSE AND pl.deleted_at IS NULL
		RETURNING ip.id, ip.plan_id, ip.number, ip.amount, ip.due_date, ip.is_paid, ip.paid_at,
		          ip.created_at, ip.updated_at
	`, paymentID, planID, authUserID).Scan(
		&p.ID, &p.PlanID, &p.Number, &p.Amount, &p.DueDate,
		&p.IsPaid, &p.PaidAt, &p.CreatedAt, &p.UpdatedAt,
	)
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
		    is_completed = (installments_paid + 1 >= $2),
		    updated_at = NOW()
		WHERE id = $1
	`, planID, total)
	return err
}

func (r *installmentRepo) SoftDeletePlan(ctx context.Context, id, authUserID string) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE homepay.installment_plans ip
		SET deleted_at = NOW()
		FROM homepay.users u
		WHERE ip.id = $1 AND u.id = ip.user_id AND u.auth_user_id = $2 AND ip.deleted_at IS NULL
	`, id, authUserID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
