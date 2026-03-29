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

const expenseCols = `id, auth_user_id, category, description, amount, expense_date, created_at, deleted_at`

func scanExpense(rows pgx.Row, e *models.Expense) error {
	return rows.Scan(&e.ID, &e.AuthUserID, &e.Category, &e.Description, &e.Amount, &e.ExpenseDate, &e.CreatedAt, &e.DeletedAt)
}

func (r *expenseRepo) Create(ctx context.Context, authUserID string, req *models.CreateExpenseRequest) (*models.Expense, error) {
	expDate, err := time.Parse("2006-01-02", req.ExpenseDate)
	if err != nil {
		return nil, fmt.Errorf("invalid expense_date format, expected YYYY-MM-DD")
	}

	var e models.Expense
	err = scanExpense(r.db.QueryRow(ctx, `
		INSERT INTO homepay.variable_expenses (auth_user_id, category, description, amount, expense_date)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING `+expenseCols,
		authUserID, req.Category, req.Description, req.Amount, expDate), &e)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *expenseRepo) GetByID(ctx context.Context, id, authUserID string) (*models.Expense, error) {
	var e models.Expense
	err := scanExpense(r.db.QueryRow(ctx, `
		SELECT `+expenseCols+`
		FROM homepay.variable_expenses
		WHERE id = $1 AND auth_user_id = $2 AND deleted_at IS NULL
	`, id, authUserID), &e)
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
	conds := []string{"auth_user_id = $1", "deleted_at IS NULL"}
	n := 2

	if filters.Month != nil && filters.Year != nil {
		conds = append(conds, fmt.Sprintf("EXTRACT(MONTH FROM expense_date) = $%d", n))
		args = append(args, *filters.Month)
		n++
		conds = append(conds, fmt.Sprintf("EXTRACT(YEAR FROM expense_date) = $%d", n))
		args = append(args, *filters.Year)
		n++
	}
	if filters.Category != nil {
		conds = append(conds, fmt.Sprintf("category = $%d", n))
		args = append(args, *filters.Category)
	}

	query := fmt.Sprintf(`
		SELECT `+expenseCols+`
		FROM homepay.variable_expenses
		WHERE %s
		ORDER BY expense_date DESC
	`, strings.Join(conds, " AND "))

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expenses []models.Expense
	for rows.Next() {
		var e models.Expense
		if err := rows.Scan(&e.ID, &e.AuthUserID, &e.Category, &e.Description, &e.Amount, &e.ExpenseDate, &e.CreatedAt, &e.DeletedAt); err != nil {
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
	err := scanExpense(r.db.QueryRow(ctx, `
		UPDATE homepay.variable_expenses
		SET category    = COALESCE($3, category),
		    description = COALESCE($4, description),
		    amount      = COALESCE($5, amount),
		    expense_date = COALESCE($6, expense_date)
		WHERE id = $1 AND auth_user_id = $2 AND deleted_at IS NULL
		RETURNING `+expenseCols,
		id, authUserID, req.Category, req.Description, req.Amount, expDate), &e)
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
		UPDATE homepay.variable_expenses
		SET deleted_at = NOW()
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
