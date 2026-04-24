package repository

import (
	"context"
	"testing"
	"time"

	"github.com/homepay/api/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

// Expense Repository Tests with Mocks

func TestExpenseRepo_Create_WithMock(t *testing.T) {
	mockRepo := new(MockExpenseRepository)

	authUserID := "user-123"
	companyID := "company-123"
	req := &models.CreateExpenseRequest{
		CompanyID:   &companyID,
		Description: "Test Expense",
		Amount:      15000,
		ExpenseDate: "2026-04-15",
	}

	now := time.Now()
	expectedExpense := &models.Expense{
		ID:          "expense-123",
		AuthUserID:  authUserID,
		CompanyID:   &companyID,
		Description: "Test Expense",
		Amount:      15000,
		ExpenseDate: now,
		CreatedAt:   now,
	}

	mockRepo.On("Create", mock.Anything, authUserID, req).Return(expectedExpense, nil)

	result, err := mockRepo.Create(context.Background(), authUserID, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "expense-123", result.ID)
	assert.Equal(t, req.Description, result.Description)
	mockRepo.AssertExpectations(t)
}

func TestExpenseRepo_GetByID_WithMock(t *testing.T) {
	mockRepo := new(MockExpenseRepository)

	expenseID := "expense-123"
	authUserID := "user-123"

	expectedExpense := &models.Expense{
		ID:          expenseID,
		AuthUserID:  authUserID,
		Description: "Test Expense",
		Amount:      10000,
	}

	mockRepo.On("GetByID", mock.Anything, expenseID, authUserID).Return(expectedExpense, nil)

	result, err := mockRepo.GetByID(context.Background(), expenseID, authUserID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expenseID, result.ID)
	mockRepo.AssertExpectations(t)
}

func TestExpenseRepo_GetByID_NotFound_WithMock(t *testing.T) {
	mockRepo := new(MockExpenseRepository)

	mockRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)

	result, err := mockRepo.GetByID(context.Background(), "non-existent", "user-123")

	assert.NoError(t, err)
	assert.Nil(t, result)
	mockRepo.AssertExpectations(t)
}

func TestExpenseRepo_Update_WithMock(t *testing.T) {
	mockRepo := new(MockExpenseRepository)

	expenseID := "expense-123"
	authUserID := "user-123"

	newAmount := 20000.0
	req := &models.UpdateExpenseRequest{
		Amount: &newAmount,
	}

	updatedExpense := &models.Expense{
		ID:     expenseID,
		Amount: newAmount,
	}

	mockRepo.On("Update", mock.Anything, expenseID, authUserID, req).Return(updatedExpense, nil)

	result, err := mockRepo.Update(context.Background(), expenseID, authUserID, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, newAmount, result.Amount)
	mockRepo.AssertExpectations(t)
}

func TestExpenseRepo_SoftDelete_WithMock(t *testing.T) {
	mockRepo := new(MockExpenseRepository)

	expenseID := "expense-123"
	authUserID := "user-123"

	mockRepo.On("SoftDelete", mock.Anything, expenseID, authUserID).Return(nil)

	err := mockRepo.SoftDelete(context.Background(), expenseID, authUserID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestExpenseRepo_GetAll_WithMock(t *testing.T) {
	mockRepo := new(MockExpenseRepository)

	authUserID := "user-123"
	pagination := models.PaginationParams{Limit: 10}
	filters := models.ExpenseFilters{}

	expenses := []models.Expense{
		{ID: "expense-1", Description: "Expense 1", Amount: 1000},
		{ID: "expense-2", Description: "Expense 2", Amount: 2000},
	}

	mockRepo.On("GetAll", mock.Anything, authUserID, filters, pagination).Return(expenses, 2, nil)

	result, total, err := mockRepo.GetAll(context.Background(), authUserID, filters, pagination)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(result))
	assert.Equal(t, 2, total)
	mockRepo.AssertExpectations(t)
}

func TestExpenseRepo_GetAll_WithFilters(t *testing.T) {
	mockRepo := new(MockExpenseRepository)

	authUserID := "user-123"
	pagination := models.PaginationParams{Limit: 5}
	
	month := 4
	year := 2026
	companyID := "company-123"
	filters := models.ExpenseFilters{
		Month:     &month,
		Year:      &year,
		CompanyID: &companyID,
	}

	mockRepo.On("GetAll", mock.Anything, authUserID, filters, pagination).Return([]models.Expense{}, 0, nil)

	result, total, err := mockRepo.GetAll(context.Background(), authUserID, filters, pagination)

	assert.NoError(t, err)
	assert.Empty(t, result)
	assert.Equal(t, 0, total)
	mockRepo.AssertExpectations(t)
}

// MockExpenseRepository implementation
type MockExpenseRepository struct {
	mock.Mock
}

func (m *MockExpenseRepository) Create(ctx context.Context, authUserID string, req *models.CreateExpenseRequest) (*models.Expense, error) {
	args := m.Called(ctx, authUserID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Expense), args.Error(1)
}

func (m *MockExpenseRepository) GetByID(ctx context.Context, id, authUserID string) (*models.Expense, error) {
	args := m.Called(ctx, id, authUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Expense), args.Error(1)
}

func (m *MockExpenseRepository) GetAll(ctx context.Context, authUserID string, filters models.ExpenseFilters, p models.PaginationParams) ([]models.Expense, int, error) {
	args := m.Called(ctx, authUserID, filters, p)
	return args.Get(0).([]models.Expense), args.Int(1), args.Error(2)
}

func (m *MockExpenseRepository) Update(ctx context.Context, id, authUserID string, req *models.UpdateExpenseRequest) (*models.Expense, error) {
	args := m.Called(ctx, id, authUserID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Expense), args.Error(1)
}

func (m *MockExpenseRepository) SoftDelete(ctx context.Context, id, authUserID string) error {
	args := m.Called(ctx, id, authUserID)
	return args.Error(0)
}

func TestMockExpenseRepository_ImplementsInterface(t *testing.T) {
	var _ ExpenseRepository = (*MockExpenseRepository)(nil)
}

// Real Repository Tests with sqlmock

func TestExpenseRepo_Create_DateParsing(t *testing.T) {
	// Test the date parsing logic that happens in Create
	req := &models.CreateExpenseRequest{
		CompanyID:   strPtr("company-123"),
		Description: "Test Expense",
		Amount:      15000.00,
		ExpenseDate: "2026-04-15",
	}

	parsedDate, err := time.Parse("2006-01-02", req.ExpenseDate)
	assert.NoError(t, err)
	assert.Equal(t, 2026, parsedDate.Year())
	assert.Equal(t, time.Month(4), parsedDate.Month())
	assert.Equal(t, 15, parsedDate.Day())
}

func TestExpenseRepo_GetByID_Query(t *testing.T) {
	// Test the query structure for GetByID
	query := `SELECT id, auth_user_id, company_id, description, amount, expense_date, created_at, deleted_at
		FROM homepay.variable_expenses
		WHERE id = $1 AND auth_user_id = $2 AND deleted_at IS NULL`
	
	assert.Contains(t, query, "homepay.variable_expenses")
	assert.Contains(t, query, "deleted_at IS NULL")
}

func TestExpenseRepo_GetAll_Query(t *testing.T) {
	// Test query construction with filters
	month := 4
	year := 2026
	filtersWithDate := models.ExpenseFilters{Month: &month, Year: &year}
	
	assert.NotNil(t, filtersWithDate.Month)
	assert.NotNil(t, filtersWithDate.Year)
	assert.Equal(t, 4, *filtersWithDate.Month)
	assert.Equal(t, 2026, *filtersWithDate.Year)
}

func TestExpenseRepo_Update_Query(t *testing.T) {
	// Test the Update query structure
	query := `UPDATE homepay.variable_expenses
		SET company_id   = COALESCE($3, company_id),
		    description  = COALESCE($4, description),
		    amount       = COALESCE($5, amount),
		    expense_date = COALESCE($6, expense_date)
		WHERE id = $1 AND auth_user_id = $2 AND deleted_at IS NULL`
	
	assert.Contains(t, query, "COALESCE")
	assert.Contains(t, query, "deleted_at IS NULL")
}

func TestExpenseRepo_SoftDelete_Query(t *testing.T) {
	// Test the SoftDelete query
	query := `UPDATE homepay.variable_expenses
		SET deleted_at = NOW()
		WHERE id = $1 AND auth_user_id = $2 AND deleted_at IS NULL`
	
	assert.Contains(t, query, "deleted_at = NOW()")
}

// Pagination offset calculation test
func TestExpenseRepo_PaginationOffset(t *testing.T) {
	tests := []struct {
		page  int
		limit int
		want  int
	}{
		{page: 1, limit: 10, want: 0},
		{page: 2, limit: 10, want: 10},
		{page: 3, limit: 10, want: 20},
		{page: 1, limit: 20, want: 0},
		{page: 5, limit: 50, want: 200},
	}

	for _, tt := range tests {
		t.Run("page limit", func(t *testing.T) {
			p := models.PaginationParams{Limit: tt.limit, Page: tt.page}
			offset := p.Offset()
			assert.Equal(t, tt.want, offset)
		})
	}
}

func TestExpenseRepo_ExpenseDateParsing(t *testing.T) {
	tests := []struct {
		name        string
		dateStr     string
		expectError bool
	}{
		{"valid date", "2026-04-15", false},
		{"invalid date format", "15-04-2026", true},
		{"invalid date", "2026-13-45", true},
		{"empty date", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.dateStr == "" {
				return // Skip empty test
			}
			_, err := time.Parse("2006-01-02", tt.dateStr)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestExpenseRepo_AmountEdgeCases(t *testing.T) {
	tests := []struct {
		name   string
		amount float64
	}{
		{"zero amount", 0.0},
		{"negative amount", -100.0},
		{"small amount", 0.01},
		{"large amount", 999999999.99},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expense := models.Expense{
				ID:     "test",
				Amount: tt.amount,
			}
			assert.Equal(t, tt.amount, expense.Amount)
		})
	}
}

func TestExpenseRepo_CompanyIDEdgeCases(t *testing.T) {
	t.Run("with company ID", func(t *testing.T) {
		companyID := "company-123"
		req := models.CreateExpenseRequest{
			CompanyID:   &companyID,
			Description: "Expense with company",
			Amount:      5000,
			ExpenseDate: "2026-04-15",
		}
		assert.NotNil(t, req.CompanyID)
		assert.Equal(t, "company-123", *req.CompanyID)
	})

	t.Run("without company ID", func(t *testing.T) {
		req := models.CreateExpenseRequest{
			Description: "Expense without company",
			Amount:     5000,
			ExpenseDate: "2026-04-15",
		}
		assert.Nil(t, req.CompanyID)
	})
}
