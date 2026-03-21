package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/fernandovalenzuela/api-home-pay/internal/models"
)

type incomeRepository struct {
	db  *sql.DB
	ctx context.Context
}

func NewIncomeRepository(db *sql.DB) IncomeRepository {
	return &incomeRepository{
		db:  db,
		ctx: context.Background(),
	}
}

func (r *incomeRepository) Create(userID string, income *models.Income) error {
	query := `
		INSERT INTO incomes (period_id, description, amount, is_recurring, received_at, user_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`

	if userID == "" {
		return fmt.Errorf("user_id is required")
	}

	var receivedAt interface{}
	if income.ReceivedAt != "" {
		receivedAt = income.ReceivedAt
	} else {
		receivedAt = nil
	}

	err := r.db.QueryRowContext(r.ctx, query,
		income.PeriodID,
		income.Description,
		income.Amount,
		income.IsRecurring,
		receivedAt,
		userID,
	).Scan(&income.ID, &income.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create income: %w", err)
	}

	return nil
}

func (r *incomeRepository) GetByID(userID string, id int) (*models.Income, error) {
	query := `
		SELECT i.id, i.period_id, i.description, i.amount, i.is_recurring, i.received_at, i.created_at,
		       p.id, p.month_number, p.year_number
		FROM incomes i
		LEFT JOIN periods p ON i.period_id = p.id
		WHERE i.id = $1 AND i.user_id = $2
	`

	if userID == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	income := &models.Income{}
	var period models.Period
	var monthNumber sql.NullInt32
	var receivedAt sql.NullString

	err := r.db.QueryRowContext(r.ctx, query, id, userID).Scan(
		&income.ID,
		&income.PeriodID,
		&income.Description,
		&income.Amount,
		&income.IsRecurring,
		&receivedAt,
		&income.CreatedAt,
		&period.ID,
		&monthNumber,
		&period.YearNumber,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get income by ID: %w", err)
	}

	if monthNumber.Valid {
		period.MonthNumber = int(monthNumber.Int32)
	}
	income.Period = &period

	if receivedAt.Valid {
		income.ReceivedAt = receivedAt.String
	}

	return income, nil
}

func (r *incomeRepository) GetAll(userID string, periodID *int) ([]models.Income, error) {
	var query string
	var rows *sql.Rows
	var err error

	if userID == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	if periodID != nil && *periodID > 0 {
		query = `
			SELECT i.id, i.period_id, i.description, i.amount, i.is_recurring, i.received_at, i.created_at,
			       p.id, p.month_number, p.year_number
			FROM incomes i
			LEFT JOIN periods p ON i.period_id = p.id
			WHERE i.user_id = $1 AND i.period_id = $2
			ORDER BY i.received_at DESC NULLS LAST, i.created_at DESC
		`
		rows, err = r.db.QueryContext(r.ctx, query, userID, *periodID)
	} else {
		query = `
			SELECT i.id, i.period_id, i.description, i.amount, i.is_recurring, i.received_at, i.created_at,
			       p.id, p.month_number, p.year_number
			FROM incomes i
			LEFT JOIN periods p ON i.period_id = p.id
			WHERE i.user_id = $1
			ORDER BY i.received_at DESC NULLS LAST, i.created_at DESC
		`
		rows, err = r.db.QueryContext(r.ctx, query, userID)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get incomes: %w", err)
	}
	defer rows.Close()

	var incomes []models.Income
	for rows.Next() {
		var income models.Income
		var period models.Period
		var monthNumber sql.NullInt32
		var receivedAt sql.NullString

		err := rows.Scan(
			&income.ID,
			&income.PeriodID,
			&income.Description,
			&income.Amount,
			&income.IsRecurring,
			&receivedAt,
			&income.CreatedAt,
			&period.ID,
			&monthNumber,
			&period.YearNumber,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan income: %w", err)
		}

		if monthNumber.Valid {
			period.MonthNumber = int(monthNumber.Int32)
		}
		income.Period = &period

		if receivedAt.Valid {
			income.ReceivedAt = receivedAt.String
		}

		incomes = append(incomes, income)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating incomes: %w", err)
	}

	return incomes, nil
}

func (r *incomeRepository) Update(userID string, income *models.Income) error {
	query := `
		UPDATE incomes
		SET period_id = $1, description = $2, amount = $3, is_recurring = $4, received_at = $5
		WHERE id = $6 AND user_id = $7
		RETURNING created_at
	`

	if userID == "" {
		return fmt.Errorf("user_id is required")
	}

	var receivedAt interface{}
	if income.ReceivedAt != "" {
		receivedAt = income.ReceivedAt
	} else {
		receivedAt = nil
	}

	var createdAt string
	err := r.db.QueryRowContext(r.ctx, query,
		income.PeriodID,
		income.Description,
		income.Amount,
		income.IsRecurring,
		receivedAt,
		income.ID,
		userID,
	).Scan(&createdAt)

	if err == sql.ErrNoRows {
		return fmt.Errorf("income not found or access denied")
	}
	if err != nil {
		return fmt.Errorf("failed to update income: %w", err)
	}

	income.CreatedAt = createdAt
	return nil
}

func (r *incomeRepository) Delete(userID string, id int) error {
	query := `
		DELETE FROM incomes
		WHERE id = $1 AND user_id = $2
	`

	if userID == "" {
		return fmt.Errorf("user_id is required")
	}

	result, err := r.db.ExecContext(r.ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete income: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("income not found or access denied")
	}

	return nil
}

func (r *incomeRepository) PeriodExistsAndBelongsToUser(userID string, periodID int) (bool, error) {
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
		return false, fmt.Errorf("failed to check period existence: %w", err)
	}

	return exists, nil
}

func (r *incomeRepository) GetTotalByPeriod(userID string, periodID int) (float64, int, error) {
	query := `
		SELECT 
			COALESCE(SUM(amount), 0) as total_amount,
			COUNT(*) as income_count
		FROM incomes
		WHERE user_id = $1 AND period_id = $2
	`

	if userID == "" {
		return 0, 0, fmt.Errorf("user_id is required")
	}

	var totalAmount float64
	var incomeCount int
	err := r.db.QueryRowContext(r.ctx, query, userID, periodID).Scan(&totalAmount, &incomeCount)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get income total: %w", err)
	}

	return totalAmount, incomeCount, nil
}
