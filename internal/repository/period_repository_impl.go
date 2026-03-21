package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/fernandovalenzuela/api-home-pay/internal/models"
)

type periodRepository struct {
	db  *sql.DB
	ctx context.Context
}

func NewPeriodRepository(db *sql.DB) PeriodRepository {
	return &periodRepository{
		db:  db,
		ctx: context.Background(),
	}
}

func (r *periodRepository) Create(userID string, period *models.Period) error {
	query := `
		INSERT INTO periods (user_id, month_number, year_number)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	if userID == "" {
		return fmt.Errorf("user_id is required")
	}

	err := r.db.QueryRowContext(r.ctx, query, userID, period.MonthNumber, period.YearNumber).Scan(&period.ID)
	if err != nil {
		return fmt.Errorf("failed to create period: %w", err)
	}

	return nil
}

func (r *periodRepository) GetByID(userID string, id int) (*models.Period, error) {
	query := `
		SELECT id, month_number, year_number
		FROM periods
		WHERE id = $1 AND user_id = $2
	`

	if userID == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	period := &models.Period{}
	err := r.db.QueryRowContext(r.ctx, query, id, userID).Scan(
		&period.ID,
		&period.MonthNumber,
		&period.YearNumber,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get period by ID: %w", err)
	}

	return period, nil
}

func (r *periodRepository) GetAll(userID string) ([]models.Period, error) {
	query := `
		SELECT id, month_number, year_number
		FROM periods
		WHERE user_id = $1
		ORDER BY year_number DESC, month_number DESC
	`

	if userID == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	rows, err := r.db.QueryContext(r.ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get periods: %w", err)
	}
	defer rows.Close()

	var periods []models.Period
	for rows.Next() {
		var period models.Period
		err := rows.Scan(
			&period.ID,
			&period.MonthNumber,
			&period.YearNumber,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan period: %w", err)
		}
		periods = append(periods, period)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating periods: %w", err)
	}

	return periods, nil
}

func (r *periodRepository) Update(userID string, period *models.Period) error {
	query := `
		UPDATE periods
		SET month_number = $1, year_number = $2
		WHERE id = $3 AND user_id = $4
	`

	if userID == "" {
		return fmt.Errorf("user_id is required")
	}

	result, err := r.db.ExecContext(r.ctx, query, period.MonthNumber, period.YearNumber, period.ID, userID)
	if err != nil {
		return fmt.Errorf("failed to update period: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("period not found or access denied")
	}

	return nil
}

func (r *periodRepository) Delete(userID string, id int) error {
	query := `
		DELETE FROM periods
		WHERE id = $1 AND user_id = $2
	`

	if userID == "" {
		return fmt.Errorf("user_id is required")
	}

	result, err := r.db.ExecContext(r.ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete period: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("period not found or access denied")
	}

	return nil
}

func (r *periodRepository) ExistsByMonthYear(userID string, monthNumber, yearNumber int) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM periods
			WHERE month_number = $1 AND year_number = $2 AND user_id = $3
		)
	`

	if userID == "" {
		return false, fmt.Errorf("user_id is required")
	}

	var exists bool
	err := r.db.QueryRowContext(r.ctx, query, monthNumber, yearNumber, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check period existence: %w", err)
	}

	return exists, nil
}

func (r *periodRepository) HasExpensesOrIncomes(id int) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM expenses WHERE period_id = $1
			UNION
			SELECT 1 FROM incomes WHERE period_id = $1
		)
	`

	var exists bool
	err := r.db.QueryRowContext(r.ctx, query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check period dependencies: %w", err)
	}

	return exists, nil
}
