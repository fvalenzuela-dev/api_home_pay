package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/homepay/api/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ExpenseRepository interface {
	Create(ctx context.Context, authUserID string, req *models.CreateExpenseRequest) (*models.Expense, error)
	GetByID(ctx context.Context, id, authUserID string) (*models.Expense, error)
	GetAll(ctx context.Context, authUserID string, filters models.ExpenseFilters) ([]models.Expense, error)
	Update(ctx context.Context, id, authUserID string, req *models.UpdateExpenseRequest) (*models.Expense, error)
	SoftDelete(ctx context.Context, id, authUserID string) error
}

type expenseRepo struct {
	db *pgxpool.Pool
}

func NewExpenseRepository(db *pgxpool.Pool) ExpenseRepository {
	return &expenseRepo{db: db}
}

func (r *expenseRepo) Create(ctx context.Context, authUserID string, req *models.CreateExpenseRequest) (*models.Expense, error) {
	expDate, err := time.Parse("2006-01-02", req.ExpenseDate)
	if err != nil {
		return nil, fmt.Errorf("invalid expense_date format, expected YYYY-MM-DD")
	}

	var e models.Expense
	err = r.db.QueryRow(ctx, `
		INSERT INTO homepay.expenses (user_id, description, amount, category, expense_date)
		SELECT id, $2, $3, $4, $5 FROM homepay.users WHERE auth_user_id = $1 AND deleted_at IS NULL
		RETURNING id, user_id, description, amount, category, expense_date, created_at, updated_at, deleted_at
	`, authUserID, req.Description, req.Amount, req.Category, expDate).Scan(
		&e.ID, &e.UserID, &e.Description, &e.Amount, &e.Category, &e.ExpenseDate,
		&e.CreatedAt, &e.UpdatedAt, &e.DeletedAt,
	)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *expenseRepo) GetByID(ctx context.Context, id, authUserID string) (*models.Expense, error) {
	var e models.Expense
	err := r.db.QueryRow(ctx, `
		SELECT e.id, e.user_id, e.description, e.amount, e.category, e.expense_date,
		       e.created_at, e.updated_at, e.deleted_at
		FROM homepay.expenses e
		JOIN homepay.users u ON u.id = e.user_id
		WHERE e.id = $1 AND u.auth_user_id = $2 AND e.deleted_at IS NULL
	`, id, authUserID).Scan(
		&e.ID, &e.UserID, &e.Description, &e.Amount, &e.Category, &e.ExpenseDate,
		&e.CreatedAt, &e.UpdatedAt, &e.DeletedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *expenseRepo) GetAll(ctx context.Context, authUserID string, filters models.ExpenseFilters) ([]models.Expense, error) {
	args := []any{authUserID}
	conds := []string{"u.auth_user_id = $1", "e.deleted_at IS NULL"}
	n := 2

	if filters.Month != nil && filters.Year != nil {
		conds = append(conds, fmt.Sprintf("EXTRACT(MONTH FROM e.expense_date) = $%d", n))
		args = append(args, *filters.Month)
		n++
		conds = append(conds, fmt.Sprintf("EXTRACT(YEAR FROM e.expense_date) = $%d", n))
		args = append(args, *filters.Year)
		n++
	}
	if filters.Category != nil {
		conds = append(conds, fmt.Sprintf("e.category = $%d", n))
		args = append(args, *filters.Category)
	}

	query := fmt.Sprintf(`
		SELECT e.id, e.user_id, e.description, e.amount, e.category, e.expense_date,
		       e.created_at, e.updated_at, e.deleted_at
		FROM homepay.expenses e
		JOIN homepay.users u ON u.id = e.user_id
		WHERE %s
		ORDER BY e.expense_date DESC
	`, strings.Join(conds, " AND "))

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expenses []models.Expense
	for rows.Next() {
		var e models.Expense
		if err := rows.Scan(&e.ID, &e.UserID, &e.Description, &e.Amount, &e.Category, &e.ExpenseDate,
			&e.CreatedAt, &e.UpdatedAt, &e.DeletedAt); err != nil {
			return nil, err
		}
		expenses = append(expenses, e)
	}
	return expenses, rows.Err()
}

func (r *expenseRepo) Update(ctx context.Context, id, authUserID string, req *models.UpdateExpenseRequest) (*models.Expense, error) {
	var expDate *time.Time
	if req.ExpenseDate != nil {
		t, err := time.Parse("2006-01-02", *req.ExpenseDate)
		if err != nil {
			return nil, fmt.Errorf("invalid expense_date format, expected YYYY-MM-DD")
		}
		expDate = &t
	}

	var e models.Expense
	err := r.db.QueryRow(ctx, `
		UPDATE homepay.expenses e
		SET description = COALESCE($3, e.description),
		    amount = COALESCE($4, e.amount),
		    category = COALESCE($5, e.category),
		    expense_date = COALESCE($6, e.expense_date),
		    updated_at = NOW()
		FROM homepay.users u
		WHERE e.id = $1 AND u.id = e.user_id AND u.auth_user_id = $2 AND e.deleted_at IS NULL
		RETURNING e.id, e.user_id, e.description, e.amount, e.category, e.expense_date,
		          e.created_at, e.updated_at, e.deleted_at
	`, id, authUserID, req.Description, req.Amount, req.Category, expDate).Scan(
		&e.ID, &e.UserID, &e.Description, &e.Amount, &e.Category, &e.ExpenseDate,
		&e.CreatedAt, &e.UpdatedAt, &e.DeletedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *expenseRepo) SoftDelete(ctx context.Context, id, authUserID string) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE homepay.expenses e
		SET deleted_at = NOW()
		FROM homepay.users u
		WHERE e.id = $1 AND u.id = e.user_id AND u.auth_user_id = $2 AND e.deleted_at IS NULL
	`, id, authUserID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
