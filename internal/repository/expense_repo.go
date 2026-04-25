package repository

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/homepay/api/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ExpenseRepository interface {
	Create(ctx context.Context, authUserID string, req *models.CreateExpenseRequest) (*models.Expense, error)
	GetByID(ctx context.Context, id, authUserID string) (*models.Expense, error)
	GetAll(ctx context.Context, authUserID string, filters models.ExpenseFilters, p models.PaginationParams) ([]models.Expense, int, error)
	Update(ctx context.Context, id, authUserID string, req *models.UpdateExpenseRequest) (*models.Expense, error)
	SoftDelete(ctx context.Context, id, authUserID string) error
}

type expenseRepo struct {
	db *pgxpool.Pool
}

func NewExpenseRepository(db *pgxpool.Pool) ExpenseRepository {
	return &expenseRepo{db: db}
}

const expenseCols = `id, auth_user_id, company_id, description, amount, expense_date, created_at, deleted_at`

func scanExpense(row pgx.Row, e *models.Expense) error {
	return row.Scan(&e.ID, &e.AuthUserID, &e.CompanyID, &e.Description, &e.Amount, &e.ExpenseDate, &e.CreatedAt, &e.DeletedAt)
}

func (r *expenseRepo) Create(ctx context.Context, authUserID string, req *models.CreateExpenseRequest) (*models.Expense, error) {
	expDate, err := time.Parse("2006-01-02", req.ExpenseDate)
	if err != nil {
		return nil, fmt.Errorf("invalid expense_date format, expected YYYY-MM-DD")
	}

	var e models.Expense
	err = scanExpense(r.db.QueryRow(ctx, `
		INSERT INTO homepay.variable_expenses (auth_user_id, company_id, description, amount, expense_date)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING `+expenseCols,
		authUserID, req.CompanyID, req.Description, req.Amount, expDate), &e)
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

func (r *expenseRepo) GetAll(ctx context.Context, authUserID string, filters models.ExpenseFilters, p models.PaginationParams) ([]models.Expense, int, error) {
	// Build parameterized query safely - placeholders ($1, $2, etc.) are numbers only,
	// user values are passed separately in args slice, preventing SQL injection
	args := []any{authUserID}
	argNum := 1

	conds := []string{"auth_user_id = $1", "deleted_at IS NULL"}

	if filters.Month != nil && filters.Year != nil {
		argNum++
		conds = append(conds, fmt.Sprintf("EXTRACT(MONTH FROM expense_date) = $%d", argNum))
		args = append(args, *filters.Month)
		argNum++
		conds = append(conds, fmt.Sprintf("EXTRACT(YEAR FROM expense_date) = $%d", argNum))
		args = append(args, *filters.Year)
	}
	if filters.CompanyID != nil {
		argNum++
		conds = append(conds, fmt.Sprintf("company_id = $%d", argNum))
		args = append(args, *filters.CompanyID)
	}

	where := strings.Join(conds, " AND ")

	var total int
	err := r.db.QueryRow(ctx,
		"SELECT COUNT(*) FROM homepay.variable_expenses WHERE "+where,
		args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	argNum++
	args = append(args, p.Limit)
	argNum++
	args = append(args, p.Offset())

	query := "SELECT " + expenseCols + " FROM homepay.variable_expenses WHERE " + where +
		" ORDER BY expense_date DESC LIMIT $" + strconv.Itoa(argNum-1) + " OFFSET $" + strconv.Itoa(argNum)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var expenses []models.Expense
	for rows.Next() {
		var e models.Expense
		if err := rows.Scan(&e.ID, &e.AuthUserID, &e.CompanyID, &e.Description, &e.Amount, &e.ExpenseDate, &e.CreatedAt, &e.DeletedAt); err != nil {
			return nil, 0, err
		}
		expenses = append(expenses, e)
	}
	return expenses, total, rows.Err()
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
		SET company_id   = COALESCE($3, company_id),
		    description  = COALESCE($4, description),
		    amount       = COALESCE($5, amount),
		    expense_date = COALESCE($6, expense_date)
		WHERE id = $1 AND auth_user_id = $2 AND deleted_at IS NULL
		RETURNING `+expenseCols,
		id, authUserID, req.CompanyID, req.Description, req.Amount, expDate), &e)
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
