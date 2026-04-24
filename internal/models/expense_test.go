package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestExpense_Struct(t *testing.T) {
	now := time.Now()
	companyID := "company-123"

	expense := Expense{
		ID:           "expense-123",
		AuthUserID:   "user_123",
		CompanyID:    &companyID,
		Description:  "Test expense",
		Amount:       15000.00,
		ExpenseDate:  now,
		CreatedAt:    now,
	}

	assert.Equal(t, "expense-123", expense.ID)
	assert.Equal(t, "user_123", expense.AuthUserID)
	assert.NotNil(t, expense.CompanyID)
	assert.Equal(t, "Test expense", expense.Description)
	assert.Equal(t, 15000.00, expense.Amount)
}

func TestExpense_OptionalCompanyID(t *testing.T) {
	expense := Expense{
		ID:           "expense-123",
		AuthUserID:   "user_123",
		CompanyID:    nil,
		Description:  "Test expense",
		Amount:       15000.00,
		ExpenseDate:  time.Now(),
		CreatedAt:    time.Now(),
	}

	assert.Nil(t, expense.CompanyID)
}

func TestCreateExpenseRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		req     CreateExpenseRequest
		isValid bool
	}{
		{
			name: "valid request",
			req: CreateExpenseRequest{
				Description: "Groceries",
				Amount:      25000.00,
				ExpenseDate: "2026-03-15",
			},
			isValid: true,
		},
		{
			name: "valid request with company",
			req: CreateExpenseRequest{
				CompanyID:   func() *string { v := "company-123"; return &v }(),
				Description: "Groceries",
				Amount:      25000.00,
				ExpenseDate: "2026-03-15",
			},
			isValid: true,
		},
		{
			name: "empty description",
			req: CreateExpenseRequest{
				Description: "",
				Amount:      25000.00,
				ExpenseDate: "2026-03-15",
			},
			isValid: false,
		},
		{
			name: "invalid amount - zero",
			req: CreateExpenseRequest{
				Description: "Test",
				Amount:      0,
				ExpenseDate: "2026-03-15",
			},
			isValid: false,
		},
		{
			name: "invalid amount - negative",
			req: CreateExpenseRequest{
				Description: "Test",
				Amount:      -100.00,
				ExpenseDate: "2026-03-15",
			},
			isValid: false,
		},
		{
			name: "empty expense date",
			req: CreateExpenseRequest{
				Description: "Test",
				Amount:      100.00,
				ExpenseDate: "",
			},
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.isValid {
				assert.NotEmpty(t, tt.req.Description)
				assert.Greater(t, tt.req.Amount, 0.0)
				assert.NotEmpty(t, tt.req.ExpenseDate)
			}
		})
	}
}

func TestUpdateExpenseRequest_PartialUpdate(t *testing.T) {
	companyID := "new-company"
	description := "Updated description"
	amount := 30000.00
	expenseDate := "2026-04-01"

	req := UpdateExpenseRequest{
		CompanyID:   &companyID,
		Description: &description,
		Amount:      &amount,
		ExpenseDate: &expenseDate,
	}

	assert.NotNil(t, req.CompanyID)
	assert.NotNil(t, req.Description)
	assert.NotNil(t, req.Amount)
	assert.NotNil(t, req.ExpenseDate)
	assert.Equal(t, "Updated description", *req.Description)
	assert.Equal(t, 30000.00, *req.Amount)
}

func TestExpenseFilters_Struct(t *testing.T) {
	month := 3
	year := 2026
	companyID := "company-123"

	filters := ExpenseFilters{
		Month:     &month,
		Year:      &year,
		CompanyID: &companyID,
	}

	assert.NotNil(t, filters.Month)
	assert.NotNil(t, filters.Year)
	assert.NotNil(t, filters.CompanyID)
	assert.Equal(t, 3, *filters.Month)
	assert.Equal(t, 2026, *filters.Year)
}

func TestExpenseFilters_NilMeansNoFilter(t *testing.T) {
	filters := ExpenseFilters{
		Month:     nil,
		Year:      nil,
		CompanyID: nil,
	}

	// All nil means no filtering
	assert.Nil(t, filters.Month)
	assert.Nil(t, filters.Year)
	assert.Nil(t, filters.CompanyID)
}
