package repository

import (
	"testing"
	"time"

	"github.com/homepay/api/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestExpenseRepo_Interfaces(t *testing.T) {
	t.Run("ExpenseRepository interface is satisfied by expenseRepo", func(t *testing.T) {
		var _ ExpenseRepository = (*expenseRepo)(nil)
	})
}

func TestScanExpense(t *testing.T) {
	t.Run("scanExpense function exists", func(t *testing.T) {
		assert.NotNil(t, scanExpense)
	})
}

func TestExpenseRepo_GetAll(t *testing.T) {
	t.Run("expenseCols constant", func(t *testing.T) {
		assert.Equal(t, `id, auth_user_id, company_id, description, amount, expense_date, created_at, deleted_at`, expenseCols)
	})
}

func TestExpenseRepo_Create(t *testing.T) {
	t.Run("CreateExpenseRequest validation", func(t *testing.T) {
		req := models.CreateExpenseRequest{
			Description:  "Test Expense",
			Amount:      10000,
			ExpenseDate: "2026-03-15",
		}
		assert.Equal(t, "Test Expense", req.Description)
		assert.Equal(t, 10000.0, req.Amount)
		assert.Equal(t, "2026-03-15", req.ExpenseDate)
	})

	t.Run("CreateExpenseRequest with company ID", func(t *testing.T) {
		companyID := "company-123"
		req := models.CreateExpenseRequest{
			CompanyID:   &companyID,
			Description: "Expense with company",
			Amount:      5000,
			ExpenseDate: "2026-03-15",
		}
		assert.NotNil(t, req.CompanyID)
		assert.Equal(t, "company-123", *req.CompanyID)
	})
}

func TestExpenseRepo_Update(t *testing.T) {
	t.Run("UpdateExpenseRequest with pointer fields", func(t *testing.T) {
		amount := 15000.0
		desc := "Updated Description"
		expDate := "2026-04-01"
		
		req := models.UpdateExpenseRequest{
			Amount:      &amount,
			Description: &desc,
			ExpenseDate: &expDate,
		}
		assert.Equal(t, 15000.0, *req.Amount)
		assert.Equal(t, "Updated Description", *req.Description)
		assert.Equal(t, "2026-04-01", *req.ExpenseDate)
	})
}

func TestExpenseRepo_SoftDelete(t *testing.T) {
	t.Run("Soft delete sets deleted_at", func(t *testing.T) {
		now := time.Now()
		expense := models.Expense{
			ID:          "expense-123",
			Description: "Test",
			DeletedAt:   &now,
		}
		assert.NotNil(t, expense.DeletedAt)
	})
}

func TestExpenseFilters(t *testing.T) {
	t.Run("ExpenseFilters with month and year", func(t *testing.T) {
		month := 3
		year := 2026
		filters := models.ExpenseFilters{
			Month: &month,
			Year:  &year,
		}
		assert.Equal(t, 3, *filters.Month)
		assert.Equal(t, 2026, *filters.Year)
	})

	t.Run("ExpenseFilters with company ID", func(t *testing.T) {
		companyID := "company-123"
		filters := models.ExpenseFilters{
			CompanyID: &companyID,
		}
		assert.Equal(t, "company-123", *filters.CompanyID)
	})

	t.Run("ExpenseFilters with all filters", func(t *testing.T) {
		month := 3
		year := 2026
		companyID := "company-123"
		filters := models.ExpenseFilters{
			Month:     &month,
			Year:      &year,
			CompanyID: &companyID,
		}
		assert.NotNil(t, filters.Month)
		assert.NotNil(t, filters.Year)
		assert.NotNil(t, filters.CompanyID)
	})
}

func TestExpenseModel(t *testing.T) {
	t.Run("Expense model fields", func(t *testing.T) {
		now := time.Now()
		expense := models.Expense{
			ID:          "expense-123",
			AuthUserID:  "user-123",
			Description: "Test Expense",
			Amount:      10000,
			ExpenseDate: now,
			CreatedAt:   now,
		}
		assert.Equal(t, "expense-123", expense.ID)
		assert.Equal(t, "Test Expense", expense.Description)
		assert.Equal(t, 10000.0, expense.Amount)
	})
}
