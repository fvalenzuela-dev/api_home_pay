package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/fernandovalenzuela/api-home-pay/internal/models"
)

type expenseRepository struct {
	db  *sql.DB
	ctx context.Context
}

func NewExpenseRepository(db *sql.DB) ExpenseRepository {
	return &expenseRepository{
		db:  db,
		ctx: context.Background(),
	}
}

func (r *expenseRepository) Create(userID string, expense *models.Expense) error {
	query := `
		INSERT INTO expenses (
			category_id, period_id, account_id, description, due_date,
			current_amount, amount_paid, current_installment, total_installments,
			installment_group_id, is_recurring, notes, user_id
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, created_at, updated_at
	`

	if userID == "" {
		return fmt.Errorf("user_id is required")
	}

	var accountID interface{}
	if expense.AccountID != nil && *expense.AccountID > 0 {
		accountID = *expense.AccountID
	} else {
		accountID = nil
	}

	var dueDate interface{}
	if expense.DueDate != nil && *expense.DueDate != "" {
		dueDate = *expense.DueDate
	} else {
		dueDate = nil
	}

	var installmentGroupID interface{}
	if expense.InstallmentGroupID != nil && *expense.InstallmentGroupID != "" {
		installmentGroupID = *expense.InstallmentGroupID
	} else {
		installmentGroupID = nil
	}

	err := r.db.QueryRowContext(r.ctx, query,
		expense.CategoryID,
		expense.PeriodID,
		accountID,
		expense.Description,
		dueDate,
		expense.CurrentAmount,
		expense.AmountPaid,
		expense.CurrentInstallment,
		expense.TotalInstallments,
		installmentGroupID,
		expense.IsRecurring,
		expense.Notes,
		userID,
	).Scan(&expense.ID, &expense.CreatedAt, &expense.UpdatedAt)

	if err != nil {
		slog.Error("db error: failed to create expense", "error", err)
		return fmt.Errorf("failed to create expense: %w", err)
	}

	return nil
}

func (r *expenseRepository) GetByID(userID string, id int) (*models.Expense, error) {
	query := `
		SELECT
			e.id, e.category_id, e.period_id, e.account_id, e.description,
			e.due_date, e.current_amount, e.amount_paid, e.current_installment,
			e.total_installments, e.installment_group_id, e.is_recurring, e.notes,
			e.created_at, e.updated_at,
			c.id, c.name,
			p.id, p.month_number, p.year_number,
			sa.id, sa.company_id, sa.account_identifier, sa.alias
		FROM expenses e
		LEFT JOIN categories c ON e.category_id = c.id
		LEFT JOIN periods p ON e.period_id = p.id
		LEFT JOIN service_accounts sa ON e.account_id = sa.id
		WHERE e.id = $1 AND e.user_id = $2
	`

	if userID == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	expense := &models.Expense{}
	var category models.Category
	var period models.Period
	var serviceAccount models.ServiceAccount
	var accountID, companyID sql.NullInt64
	var dueDate, installmentGroupID, notes, alias sql.NullString
	var monthNumber sql.NullInt32

	err := r.db.QueryRowContext(r.ctx, query, id, userID).Scan(
		&expense.ID,
		&expense.CategoryID,
		&expense.PeriodID,
		&accountID,
		&expense.Description,
		&dueDate,
		&expense.CurrentAmount,
		&expense.AmountPaid,
		&expense.CurrentInstallment,
		&expense.TotalInstallments,
		&installmentGroupID,
		&expense.IsRecurring,
		&notes,
		&expense.CreatedAt,
		&expense.UpdatedAt,
		&category.ID,
		&category.Name,
		&period.ID,
		&monthNumber,
		&period.YearNumber,
		&serviceAccount.ID,
		&companyID,
		&serviceAccount.AccountIdentifier,
		&alias,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		slog.Error("db error: failed to get expense by ID", "error", err)
		return nil, fmt.Errorf("failed to get expense by ID: %w", err)
	}

	if accountID.Valid {
		accountIDInt := int(accountID.Int64)
		expense.AccountID = &accountIDInt
		if companyID.Valid {
			serviceAccount.CompanyID = int(companyID.Int64)
		}
		if alias.Valid {
			serviceAccount.Alias = alias.String
		}
		expense.ServiceAccount = &serviceAccount
	}

	if dueDate.Valid {
		expense.DueDate = &dueDate.String
	}

	if installmentGroupID.Valid {
		idStr := installmentGroupID.String
		expense.InstallmentGroupID = &idStr
	}

	if notes.Valid {
		expense.Notes = notes.String
	}

	if monthNumber.Valid {
		period.MonthNumber = int(monthNumber.Int32)
	}

	expense.Category = &category
	expense.Period = &period

	return expense, nil
}

func (r *expenseRepository) GetAll(userID string, filters ExpenseFilters) ([]models.Expense, error) {
	if userID == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	query := `
		SELECT
			e.id, e.category_id, e.period_id, e.account_id, e.description,
			e.due_date, e.current_amount, e.amount_paid, e.current_installment,
			e.total_installments, e.installment_group_id, e.is_recurring, e.notes,
			e.created_at, e.updated_at,
			c.id, c.name,
			p.id, p.month_number, p.year_number,
			sa.id, sa.company_id, sa.account_identifier, sa.alias
		FROM expenses e
		LEFT JOIN categories c ON e.category_id = c.id
		LEFT JOIN periods p ON e.period_id = p.id
		LEFT JOIN service_accounts sa ON e.account_id = sa.id
		WHERE e.user_id = $1
	`

	args := []interface{}{userID}
	argCount := 1

	if filters.PeriodID != nil && *filters.PeriodID > 0 {
		argCount++
		query += fmt.Sprintf(" AND e.period_id = $%d", argCount)
		args = append(args, *filters.PeriodID)
	}

	if filters.CategoryID != nil && *filters.CategoryID > 0 {
		argCount++
		query += fmt.Sprintf(" AND e.category_id = $%d", argCount)
		args = append(args, *filters.CategoryID)
	}

	if filters.AccountID != nil && *filters.AccountID > 0 {
		argCount++
		query += fmt.Sprintf(" AND e.account_id = $%d", argCount)
		args = append(args, *filters.AccountID)
	}

	if filters.PaymentStatus != nil && *filters.PaymentStatus != "" {
		switch *filters.PaymentStatus {
		case "paid":
			query += " AND e.amount_paid >= e.current_amount"
		case "partial":
			query += " AND e.amount_paid > 0 AND e.amount_paid < e.current_amount"
		case "pending":
			query += " AND e.amount_paid = 0"
		}
	}

	query += " ORDER BY e.due_date ASC NULLS LAST, e.created_at DESC"

	rows, err := r.db.QueryContext(r.ctx, query, args...)
	if err != nil {
		slog.Error("db error: failed to get expenses", "error", err)
		return nil, fmt.Errorf("failed to get expenses: %w", err)
	}
	defer rows.Close()

	var expenses []models.Expense
	for rows.Next() {
		var expense models.Expense
		var category models.Category
		var period models.Period
		var serviceAccount models.ServiceAccount
		var accountID, companyID sql.NullInt64
		var dueDate, installmentGroupID, notes, alias sql.NullString
		var monthNumber sql.NullInt32

		err := rows.Scan(
			&expense.ID,
			&expense.CategoryID,
			&expense.PeriodID,
			&accountID,
			&expense.Description,
			&dueDate,
			&expense.CurrentAmount,
			&expense.AmountPaid,
			&expense.CurrentInstallment,
			&expense.TotalInstallments,
			&installmentGroupID,
			&expense.IsRecurring,
			&notes,
			&expense.CreatedAt,
			&expense.UpdatedAt,
			&category.ID,
			&category.Name,
			&period.ID,
			&monthNumber,
			&period.YearNumber,
			&serviceAccount.ID,
			&companyID,
			&serviceAccount.AccountIdentifier,
			&alias,
		)
		if err != nil {
			slog.Error("db error: failed to scan expense", "error", err)
			return nil, fmt.Errorf("failed to scan expense: %w", err)
		}

		if accountID.Valid {
			accountIDInt := int(accountID.Int64)
			expense.AccountID = &accountIDInt
			if companyID.Valid {
				serviceAccount.CompanyID = int(companyID.Int64)
			}
			if alias.Valid {
				serviceAccount.Alias = alias.String
			}
			expense.ServiceAccount = &serviceAccount
		}

		if dueDate.Valid {
			expense.DueDate = &dueDate.String
		}

		if installmentGroupID.Valid {
			idStr := installmentGroupID.String
			expense.InstallmentGroupID = &idStr
		}

		if notes.Valid {
			expense.Notes = notes.String
		}

		if monthNumber.Valid {
			period.MonthNumber = int(monthNumber.Int32)
		}

		expense.Category = &category
		expense.Period = &period

		expenses = append(expenses, expense)
	}

	if err = rows.Err(); err != nil {
		slog.Error("db error: error iterating expenses", "error", err)
		return nil, fmt.Errorf("error iterating expenses: %w", err)
	}

	return expenses, nil
}

func (r *expenseRepository) Update(userID string, expense *models.Expense) error {
	query := `
		UPDATE expenses
		SET category_id = $1, period_id = $2, account_id = $3, description = $4,
		    due_date = $5, current_amount = $6, amount_paid = $7,
		    current_installment = $8, total_installments = $9,
		    installment_group_id = $10, is_recurring = $11, notes = $12,
		    updated_at = NOW()
		WHERE id = $13 AND user_id = $14
		RETURNING updated_at
	`

	if userID == "" {
		return fmt.Errorf("user_id is required")
	}

	var accountID interface{}
	if expense.AccountID != nil && *expense.AccountID > 0 {
		accountID = *expense.AccountID
	} else {
		accountID = nil
	}

	var dueDate interface{}
	if expense.DueDate != nil && *expense.DueDate != "" {
		dueDate = *expense.DueDate
	} else {
		dueDate = nil
	}

	var installmentGroupID interface{}
	if expense.InstallmentGroupID != nil && *expense.InstallmentGroupID != "" {
		installmentGroupID = *expense.InstallmentGroupID
	} else {
		installmentGroupID = nil
	}

	var updatedAt string
	err := r.db.QueryRowContext(r.ctx, query,
		expense.CategoryID,
		expense.PeriodID,
		accountID,
		expense.Description,
		dueDate,
		expense.CurrentAmount,
		expense.AmountPaid,
		expense.CurrentInstallment,
		expense.TotalInstallments,
		installmentGroupID,
		expense.IsRecurring,
		expense.Notes,
		expense.ID,
		userID,
	).Scan(&updatedAt)

	if err == sql.ErrNoRows {
		return fmt.Errorf("expense not found or access denied")
	}
	if err != nil {
		slog.Error("db error: failed to update expense", "error", err)
		return fmt.Errorf("failed to update expense: %w", err)
	}

	expense.UpdatedAt = updatedAt
	return nil
}

func (r *expenseRepository) Delete(userID string, id int) error {
	query := `
		DELETE FROM expenses
		WHERE id = $1 AND user_id = $2
	`

	if userID == "" {
		return fmt.Errorf("user_id is required")
	}

	result, err := r.db.ExecContext(r.ctx, query, id, userID)
	if err != nil {
		slog.Error("db error: failed to delete expense", "error", err)
		return fmt.Errorf("failed to delete expense: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		slog.Error("db error: failed to get rows affected", "error", err)
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("expense not found or access denied")
	}

	return nil
}

func (r *expenseRepository) MarkAsPaid(userID string, id int) error {
	query := `
		UPDATE expenses
		SET amount_paid = current_amount, updated_at = NOW()
		WHERE id = $1 AND user_id = $2
		RETURNING updated_at
	`

	if userID == "" {
		return fmt.Errorf("user_id is required")
	}

	var updatedAt string
	err := r.db.QueryRowContext(r.ctx, query, id, userID).Scan(&updatedAt)
	if err == sql.ErrNoRows {
		return fmt.Errorf("expense not found or access denied")
	}
	if err != nil {
		slog.Error("db error: failed to mark expense as paid", "error", err)
		return fmt.Errorf("failed to mark expense as paid: %w", err)
	}

	return nil
}

func (r *expenseRepository) UpdateAmountPaid(userID string, id int, amount float64) error {
	query := `
		UPDATE expenses
		SET amount_paid = $1, updated_at = NOW()
		WHERE id = $2 AND user_id = $3
		RETURNING updated_at
	`

	if userID == "" {
		return fmt.Errorf("user_id is required")
	}

	var updatedAt string
	err := r.db.QueryRowContext(r.ctx, query, amount, id, userID).Scan(&updatedAt)
	if err == sql.ErrNoRows {
		return fmt.Errorf("expense not found or access denied")
	}
	if err != nil {
		slog.Error("db error: failed to update amount paid", "error", err)
		return fmt.Errorf("failed to update amount paid: %w", err)
	}

	return nil
}

func (r *expenseRepository) CategoryExistsAndBelongsToUser(userID string, categoryID int) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM categories
			WHERE id = $1 AND user_id = $2
		)
	`

	if userID == "" {
		return false, fmt.Errorf("user_id is required")
	}

	var exists bool
	err := r.db.QueryRowContext(r.ctx, query, categoryID, userID).Scan(&exists)
	if err != nil {
		slog.Error("db error: failed to check category existence", "error", err)
		return false, fmt.Errorf("failed to check category existence: %w", err)
	}

	return exists, nil
}

func (r *expenseRepository) PeriodExistsAndBelongsToUser(userID string, periodID int) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM periods
			WHERE id = $1 AND user_id = $2
		)
	`

	if userID == "" {
		return false, fmt.Errorf("user_id is required")
	}

	var exists bool
	err := r.db.QueryRowContext(r.ctx, query, periodID, userID).Scan(&exists)
	if err != nil {
		slog.Error("db error: failed to check period existence", "error", err)
		return false, fmt.Errorf("failed to check period existence: %w", err)
	}

	return exists, nil
}

func (r *expenseRepository) ServiceAccountExistsAndBelongsToUser(userID string, accountID int) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM service_accounts
			WHERE id = $1 AND user_id = $2
		)
	`

	if userID == "" {
		return false, fmt.Errorf("user_id is required")
	}

	var exists bool
	err := r.db.QueryRowContext(r.ctx, query, accountID, userID).Scan(&exists)
	if err != nil {
		slog.Error("db error: failed to check service account existence", "error", err)
		return false, fmt.Errorf("failed to check service account existence: %w", err)
	}

	return exists, nil
}

func (r *expenseRepository) GetPendingExpenses(userID string, daysAhead int, overdueOnly bool) ([]models.Expense, error) {
	if userID == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	query := `
		SELECT
			e.id, e.category_id, e.period_id, e.account_id, e.description,
			e.due_date, e.current_amount, e.amount_paid, e.current_installment,
			e.total_installments, e.installment_group_id, e.is_recurring, e.notes,
			e.created_at, e.updated_at,
			c.id, c.name,
			p.id, p.month_number, p.year_number,
			sa.id, sa.company_id, sa.account_identifier, sa.alias
		FROM expenses e
		LEFT JOIN categories c ON e.category_id = c.id
		LEFT JOIN periods p ON e.period_id = p.id
		LEFT JOIN service_accounts sa ON e.account_id = sa.id
		WHERE e.user_id = $1
	`

	args := []interface{}{userID}
	argCount := 1

	if overdueOnly {
		// Only show overdue expenses (due_date < today AND amount_paid < current_amount)
		argCount++
		query += fmt.Sprintf(" AND e.due_date < CURRENT_DATE AND e.amount_paid < e.current_amount")
	} else {
		// Show expenses due within daysAhead days
		argCount++
		query += fmt.Sprintf(" AND e.due_date <= CURRENT_DATE + INTERVAL '%d days' AND e.amount_paid < e.current_amount", daysAhead)
	}

	query += " ORDER BY e.due_date ASC NULLS LAST"

	rows, err := r.db.QueryContext(r.ctx, query, args...)
	if err != nil {
		slog.Error("db error: failed to get pending expenses", "error", err)
		return nil, fmt.Errorf("failed to get pending expenses: %w", err)
	}
	defer rows.Close()

	var expenses []models.Expense
	for rows.Next() {
		var expense models.Expense
		var category models.Category
		var period models.Period
		var serviceAccount models.ServiceAccount
		var accountID, companyID sql.NullInt64
		var dueDate, installmentGroupID, notes, alias sql.NullString
		var monthNumber sql.NullInt32

		err := rows.Scan(
			&expense.ID,
			&expense.CategoryID,
			&expense.PeriodID,
			&accountID,
			&expense.Description,
			&dueDate,
			&expense.CurrentAmount,
			&expense.AmountPaid,
			&expense.CurrentInstallment,
			&expense.TotalInstallments,
			&installmentGroupID,
			&expense.IsRecurring,
			&notes,
			&expense.CreatedAt,
			&expense.UpdatedAt,
			&category.ID,
			&category.Name,
			&period.ID,
			&monthNumber,
			&period.YearNumber,
			&serviceAccount.ID,
			&companyID,
			&serviceAccount.AccountIdentifier,
			&alias,
		)
		if err != nil {
			slog.Error("db error: failed to scan expense", "error", err)
			return nil, fmt.Errorf("failed to scan expense: %w", err)
		}

		if accountID.Valid {
			accountIDInt := int(accountID.Int64)
			expense.AccountID = &accountIDInt
			if companyID.Valid {
				serviceAccount.CompanyID = int(companyID.Int64)
			}
			if alias.Valid {
				serviceAccount.Alias = alias.String
			}
			expense.ServiceAccount = &serviceAccount
		}

		if dueDate.Valid {
			expense.DueDate = &dueDate.String
		}

		if installmentGroupID.Valid {
			idStr := installmentGroupID.String
			expense.InstallmentGroupID = &idStr
		}

		if notes.Valid {
			expense.Notes = notes.String
		}

		if monthNumber.Valid {
			period.MonthNumber = int(monthNumber.Int32)
		}

		expense.Category = &category
		expense.Period = &period

		expenses = append(expenses, expense)
	}

	if err = rows.Err(); err != nil {
		slog.Error("db error: error iterating pending expenses", "error", err)
		return nil, fmt.Errorf("error iterating pending expenses: %w", err)
	}

	return expenses, nil
}

func (r *expenseRepository) GetSummaryByPeriod(userID string, periodID int) (*ExpenseSummary, error) {
	query := `
		SELECT
			COALESCE(SUM(current_amount), 0) as total_amount,
			COALESCE(SUM(amount_paid), 0) as paid_amount,
			COALESCE(SUM(CASE WHEN amount_paid < current_amount THEN current_amount - amount_paid ELSE 0 END), 0) as pending_amount,
			COUNT(*) as expense_count
		FROM expenses
		WHERE user_id = $1 AND period_id = $2
	`

	if userID == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	summary := &ExpenseSummary{}
	err := r.db.QueryRowContext(r.ctx, query, userID, periodID).Scan(
		&summary.TotalAmount,
		&summary.PaidAmount,
		&summary.PendingAmount,
		&summary.ExpenseCount,
	)
	if err != nil {
		slog.Error("db error: failed to get expense summary", "error", err)
		return nil, fmt.Errorf("failed to get expense summary: %w", err)
	}

	return summary, nil
}
