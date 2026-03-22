package services

import (
	"testing"

	"github.com/fernandovalenzuela/api-home-pay/internal/models"
	"github.com/fernandovalenzuela/api-home-pay/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockExpenseRepository is a mock implementation of ExpenseRepository
type MockExpenseRepository struct {
	mock.Mock
}

func (m *MockExpenseRepository) Create(userID string, expense *models.Expense) error {
	args := m.Called(userID, expense)
	return args.Error(0)
}

func (m *MockExpenseRepository) GetByID(userID string, id int) (*models.Expense, error) {
	args := m.Called(userID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Expense), args.Error(1)
}

func (m *MockExpenseRepository) GetAll(userID string, filters repository.ExpenseFilters) ([]models.Expense, error) {
	args := m.Called(userID, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Expense), args.Error(1)
}

func (m *MockExpenseRepository) Update(userID string, expense *models.Expense) error {
	args := m.Called(userID, expense)
	return args.Error(0)
}

func (m *MockExpenseRepository) Delete(userID string, id int) error {
	args := m.Called(userID, id)
	return args.Error(0)
}

func (m *MockExpenseRepository) MarkAsPaid(userID string, id int) error {
	args := m.Called(userID, id)
	return args.Error(0)
}

func (m *MockExpenseRepository) UpdateAmountPaid(userID string, id int, amount float64) error {
	args := m.Called(userID, id, amount)
	return args.Error(0)
}

func (m *MockExpenseRepository) CategoryExistsAndBelongsToUser(userID string, categoryID int) (bool, error) {
	args := m.Called(userID, categoryID)
	return args.Bool(0), args.Error(1)
}

func (m *MockExpenseRepository) PeriodExistsAndBelongsToUser(userID string, periodID int) (bool, error) {
	args := m.Called(userID, periodID)
	return args.Bool(0), args.Error(1)
}

func (m *MockExpenseRepository) ServiceAccountExistsAndBelongsToUser(userID string, accountID int) (bool, error) {
	args := m.Called(userID, accountID)
	return args.Bool(0), args.Error(1)
}

func (m *MockExpenseRepository) GetPendingExpenses(userID string, daysAhead int, overdueOnly bool) ([]models.Expense, error) {
	args := m.Called(userID, daysAhead, overdueOnly)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Expense), args.Error(1)
}

func (m *MockExpenseRepository) GetSummaryByPeriod(userID string, periodID int) (*repository.ExpenseSummary, error) {
	args := m.Called(userID, periodID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.ExpenseSummary), args.Error(1)
}

func TestExpenseService_Create_Success(t *testing.T) {
	mockRepo := new(MockExpenseRepository)
	service := NewExpenseService(mockRepo)

	accountID := 1
	dueDate := "2024-06-15"
	expense := &models.Expense{
		CategoryID:        1,
		PeriodID:          1,
		AccountID:         &accountID,
		Description:       "Test Expense",
		DueDate:           &dueDate,
		CurrentAmount:     100.00,
		AmountPaid:        0,
		TotalInstallments: 1,
	}

	mockRepo.On("CategoryExistsAndBelongsToUser", "user123", 1).Return(true, nil)
	mockRepo.On("PeriodExistsAndBelongsToUser", "user123", 1).Return(true, nil)
	mockRepo.On("ServiceAccountExistsAndBelongsToUser", "user123", 1).Return(true, nil)
	mockRepo.On("Create", "user123", expense).Return(nil)

	err := service.Create("user123", expense)

	assert.NoError(t, err)
	assert.Equal(t, 1, expense.CurrentInstallment) // Should be defaulted to 1
	mockRepo.AssertExpectations(t)
}

func TestExpenseService_Create_WithoutOptionalFields(t *testing.T) {
	mockRepo := new(MockExpenseRepository)
	service := NewExpenseService(mockRepo)

	expense := &models.Expense{
		CategoryID:    1,
		PeriodID:      1,
		Description:   "Test Expense",
		CurrentAmount: 100.00,
		AmountPaid:    0,
	}

	mockRepo.On("CategoryExistsAndBelongsToUser", "user123", 1).Return(true, nil)
	mockRepo.On("PeriodExistsAndBelongsToUser", "user123", 1).Return(true, nil)
	mockRepo.On("Create", "user123", expense).Return(nil)

	err := service.Create("user123", expense)

	assert.NoError(t, err)
	assert.Equal(t, 1, expense.CurrentInstallment) // Defaulted
	assert.Equal(t, 1, expense.TotalInstallments)  // Defaulted
	mockRepo.AssertExpectations(t)
}

func TestExpenseService_Create_EmptyDescription(t *testing.T) {
	mockRepo := new(MockExpenseRepository)
	service := NewExpenseService(mockRepo)

	expense := &models.Expense{
		Description:   "",
		CurrentAmount: 100.00,
	}

	err := service.Create("user123", expense)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "description cannot be empty")
}

func TestExpenseService_Create_WhitespaceDescription(t *testing.T) {
	mockRepo := new(MockExpenseRepository)
	service := NewExpenseService(mockRepo)

	expense := &models.Expense{
		Description:   "   ",
		CurrentAmount: 100.00,
	}

	err := service.Create("user123", expense)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "description cannot be empty")
}

func TestExpenseService_Create_ZeroAmount(t *testing.T) {
	mockRepo := new(MockExpenseRepository)
	service := NewExpenseService(mockRepo)

	expense := &models.Expense{
		Description:   "Test",
		CurrentAmount: 0,
	}

	err := service.Create("user123", expense)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be greater than zero")
}

func TestExpenseService_Create_NegativeAmount(t *testing.T) {
	mockRepo := new(MockExpenseRepository)
	service := NewExpenseService(mockRepo)

	expense := &models.Expense{
		Description:   "Test",
		CurrentAmount: -50.00,
	}

	err := service.Create("user123", expense)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be greater than zero")
}

func TestExpenseService_Create_NegativeAmountPaid(t *testing.T) {
	mockRepo := new(MockExpenseRepository)
	service := NewExpenseService(mockRepo)

	expense := &models.Expense{
		Description:   "Test",
		CurrentAmount: 100.00,
		AmountPaid:    -10.00,
	}

	err := service.Create("user123", expense)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "amount paid cannot be negative")
}

func TestExpenseService_Create_InvalidDueDate(t *testing.T) {
	mockRepo := new(MockExpenseRepository)
	service := NewExpenseService(mockRepo)

	dueDate := "invalid-date"
	expense := &models.Expense{
		Description:   "Test",
		CurrentAmount: 100.00,
		DueDate:       &dueDate,
	}

	err := service.Create("user123", expense)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid due date format")
}

func TestExpenseService_Create_InvalidDueDateWrongFormat(t *testing.T) {
	mockRepo := new(MockExpenseRepository)
	service := NewExpenseService(mockRepo)

	dueDate := "15-06-2024" // Wrong format (DD-MM-YYYY instead of YYYY-MM-DD)
	expense := &models.Expense{
		Description:   "Test",
		CurrentAmount: 100.00,
		DueDate:       &dueDate,
	}

	err := service.Create("user123", expense)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid due date format")
}

func TestExpenseService_Create_InvalidInstallmentRange(t *testing.T) {
	mockRepo := new(MockExpenseRepository)
	service := NewExpenseService(mockRepo)

	expense := &models.Expense{
		Description:        "Test",
		CurrentAmount:      100.00,
		CurrentInstallment: 5,
		TotalInstallments:  3,
	}

	err := service.Create("user123", expense)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "current installment cannot be greater than total installments")
}

func TestExpenseService_Create_CategoryNotFound(t *testing.T) {
	mockRepo := new(MockExpenseRepository)
	service := NewExpenseService(mockRepo)

	expense := &models.Expense{
		Description:       "Test",
		CurrentAmount:     100.00,
		CategoryID:        999,
		PeriodID:          1,
		TotalInstallments: 1,
	}

	mockRepo.On("CategoryExistsAndBelongsToUser", "user123", 999).Return(false, nil)

	err := service.Create("user123", expense)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "category not found or access denied")
	mockRepo.AssertExpectations(t)
}

func TestExpenseService_Create_PeriodNotFound(t *testing.T) {
	mockRepo := new(MockExpenseRepository)
	service := NewExpenseService(mockRepo)

	expense := &models.Expense{
		Description:       "Test",
		CurrentAmount:     100.00,
		CategoryID:        1,
		PeriodID:          999,
		TotalInstallments: 1,
	}

	mockRepo.On("CategoryExistsAndBelongsToUser", "user123", 1).Return(true, nil)
	mockRepo.On("PeriodExistsAndBelongsToUser", "user123", 999).Return(false, nil)

	err := service.Create("user123", expense)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "period not found or access denied")
	mockRepo.AssertExpectations(t)
}

func TestExpenseService_Create_ServiceAccountNotFound(t *testing.T) {
	mockRepo := new(MockExpenseRepository)
	service := NewExpenseService(mockRepo)

	accountID := 999
	expense := &models.Expense{
		Description:       "Test",
		CurrentAmount:     100.00,
		CategoryID:        1,
		PeriodID:          1,
		AccountID:         &accountID,
		TotalInstallments: 1,
	}

	mockRepo.On("CategoryExistsAndBelongsToUser", "user123", 1).Return(true, nil)
	mockRepo.On("PeriodExistsAndBelongsToUser", "user123", 1).Return(true, nil)
	mockRepo.On("ServiceAccountExistsAndBelongsToUser", "user123", 999).Return(false, nil)

	err := service.Create("user123", expense)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "service account not found or access denied")
	mockRepo.AssertExpectations(t)
}

func TestExpenseService_GetByID_Success(t *testing.T) {
	mockRepo := new(MockExpenseRepository)
	service := NewExpenseService(mockRepo)

	expectedExpense := &models.Expense{
		ID:            1,
		Description:   "Test Expense",
		CurrentAmount: 100.00,
	}

	mockRepo.On("GetByID", "user123", 1).Return(expectedExpense, nil)

	expense, err := service.GetByID("user123", 1)

	assert.NoError(t, err)
	assert.NotNil(t, expense)
	assert.Equal(t, "Test Expense", expense.Description)
	mockRepo.AssertExpectations(t)
}

func TestExpenseService_GetByID_InvalidID(t *testing.T) {
	mockRepo := new(MockExpenseRepository)
	service := NewExpenseService(mockRepo)

	expense, err := service.GetByID("user123", 0)

	assert.Error(t, err)
	assert.Nil(t, expense)
	assert.Contains(t, err.Error(), "invalid expense ID")
}

func TestExpenseService_GetByID_NotFound(t *testing.T) {
	mockRepo := new(MockExpenseRepository)
	service := NewExpenseService(mockRepo)

	mockRepo.On("GetByID", "user123", 999).Return(nil, nil)

	expense, err := service.GetByID("user123", 999)

	assert.Error(t, err)
	assert.Nil(t, expense)
	assert.Contains(t, err.Error(), "not found")
	mockRepo.AssertExpectations(t)
}

func TestExpenseService_GetAll_Success(t *testing.T) {
	mockRepo := new(MockExpenseRepository)
	service := NewExpenseService(mockRepo)

	filters := repository.ExpenseFilters{}
	expectedExpenses := []models.Expense{
		{ID: 1, Description: "Expense 1"},
		{ID: 2, Description: "Expense 2"},
	}

	mockRepo.On("GetAll", "user123", filters).Return(expectedExpenses, nil)

	expenses, err := service.GetAll("user123", filters)

	assert.NoError(t, err)
	assert.Len(t, expenses, 2)
	mockRepo.AssertExpectations(t)
}

func TestExpenseService_GetAll_WithPeriodFilter(t *testing.T) {
	mockRepo := new(MockExpenseRepository)
	service := NewExpenseService(mockRepo)

	periodID := 1
	filters := repository.ExpenseFilters{PeriodID: &periodID}

	mockRepo.On("PeriodExistsAndBelongsToUser", "user123", 1).Return(true, nil)
	mockRepo.On("GetAll", "user123", filters).Return([]models.Expense{}, nil)

	_, err := service.GetAll("user123", filters)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestExpenseService_GetAll_WithPeriodFilterNotFound(t *testing.T) {
	mockRepo := new(MockExpenseRepository)
	service := NewExpenseService(mockRepo)

	periodID := 999
	filters := repository.ExpenseFilters{PeriodID: &periodID}

	mockRepo.On("PeriodExistsAndBelongsToUser", "user123", 999).Return(false, nil)

	expenses, err := service.GetAll("user123", filters)

	assert.Error(t, err)
	assert.Nil(t, expenses)
	assert.Contains(t, err.Error(), "period not found or access denied")
	mockRepo.AssertExpectations(t)
}

func TestExpenseService_GetAll_WithCategoryFilter(t *testing.T) {
	mockRepo := new(MockExpenseRepository)
	service := NewExpenseService(mockRepo)

	categoryID := 1
	filters := repository.ExpenseFilters{CategoryID: &categoryID}

	mockRepo.On("CategoryExistsAndBelongsToUser", "user123", 1).Return(true, nil)
	mockRepo.On("GetAll", "user123", filters).Return([]models.Expense{}, nil)

	_, err := service.GetAll("user123", filters)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestExpenseService_GetAll_WithAccountFilter(t *testing.T) {
	mockRepo := new(MockExpenseRepository)
	service := NewExpenseService(mockRepo)

	accountID := 1
	filters := repository.ExpenseFilters{AccountID: &accountID}

	mockRepo.On("ServiceAccountExistsAndBelongsToUser", "user123", 1).Return(true, nil)
	mockRepo.On("GetAll", "user123", filters).Return([]models.Expense{}, nil)

	_, err := service.GetAll("user123", filters)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestExpenseService_GetAll_WithInvalidPaymentStatus(t *testing.T) {
	mockRepo := new(MockExpenseRepository)
	service := NewExpenseService(mockRepo)

	paymentStatus := "invalid"
	filters := repository.ExpenseFilters{PaymentStatus: &paymentStatus}

	expenses, err := service.GetAll("user123", filters)

	assert.Error(t, err)
	assert.Nil(t, expenses)
	assert.Contains(t, err.Error(), "invalid payment status")
}

func TestExpenseService_GetAll_WithValidPaymentStatuses(t *testing.T) {
	validStatuses := []string{"paid", "partial", "pending"}

	for _, status := range validStatuses {
		t.Run("status_"+status, func(t *testing.T) {
			mockRepo := new(MockExpenseRepository)
			service := NewExpenseService(mockRepo)

			filters := repository.ExpenseFilters{PaymentStatus: &status}

			mockRepo.On("GetAll", "user123", filters).Return([]models.Expense{}, nil)

			_, err := service.GetAll("user123", filters)

			assert.NoError(t, err)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestExpenseService_Update_Success(t *testing.T) {
	mockRepo := new(MockExpenseRepository)
	service := NewExpenseService(mockRepo)

	existingExpense := &models.Expense{
		ID:                1,
		CategoryID:        1,
		PeriodID:          1,
		Description:       "Old Description",
		CurrentAmount:     100.00,
		TotalInstallments: 1,
	}
	updatedExpense := &models.Expense{
		ID:                1,
		CategoryID:        1,
		PeriodID:          1,
		Description:       "New Description",
		CurrentAmount:     150.00,
		TotalInstallments: 1,
	}

	mockRepo.On("GetByID", "user123", 1).Return(existingExpense, nil)
	mockRepo.On("CategoryExistsAndBelongsToUser", "user123", 1).Return(true, nil)
	mockRepo.On("PeriodExistsAndBelongsToUser", "user123", 1).Return(true, nil)
	mockRepo.On("Update", "user123", updatedExpense).Return(nil)

	err := service.Update("user123", updatedExpense)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestExpenseService_Update_InvalidID(t *testing.T) {
	mockRepo := new(MockExpenseRepository)
	service := NewExpenseService(mockRepo)

	expense := &models.Expense{
		ID:            0,
		Description:   "Test",
		CurrentAmount: 100.00,
	}

	err := service.Update("user123", expense)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid expense ID")
}

func TestExpenseService_Update_NotFound(t *testing.T) {
	mockRepo := new(MockExpenseRepository)
	service := NewExpenseService(mockRepo)

	expense := &models.Expense{
		ID:            999,
		Description:   "Test",
		CurrentAmount: 100.00,
	}

	mockRepo.On("GetByID", "user123", 999).Return(nil, nil)

	err := service.Update("user123", expense)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	mockRepo.AssertExpectations(t)
}

func TestExpenseService_Delete_Success(t *testing.T) {
	mockRepo := new(MockExpenseRepository)
	service := NewExpenseService(mockRepo)

	existingExpense := &models.Expense{
		ID:          1,
		Description: "Test",
	}

	mockRepo.On("GetByID", "user123", 1).Return(existingExpense, nil)
	mockRepo.On("Delete", "user123", 1).Return(nil)

	err := service.Delete("user123", 1)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestExpenseService_Delete_InvalidID(t *testing.T) {
	mockRepo := new(MockExpenseRepository)
	service := NewExpenseService(mockRepo)

	err := service.Delete("user123", 0)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid expense ID")
}

func TestExpenseService_Delete_NotFound(t *testing.T) {
	mockRepo := new(MockExpenseRepository)
	service := NewExpenseService(mockRepo)

	mockRepo.On("GetByID", "user123", 999).Return(nil, nil)

	err := service.Delete("user123", 999)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	mockRepo.AssertExpectations(t)
}

func TestExpenseService_MarkAsPaid_Success(t *testing.T) {
	mockRepo := new(MockExpenseRepository)
	service := NewExpenseService(mockRepo)

	existingExpense := &models.Expense{
		ID:          1,
		Description: "Test",
	}

	mockRepo.On("GetByID", "user123", 1).Return(existingExpense, nil)
	mockRepo.On("MarkAsPaid", "user123", 1).Return(nil)

	err := service.MarkAsPaid("user123", 1)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestExpenseService_MarkAsPaid_InvalidID(t *testing.T) {
	mockRepo := new(MockExpenseRepository)
	service := NewExpenseService(mockRepo)

	err := service.MarkAsPaid("user123", 0)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid expense ID")
}

func TestExpenseService_GetPendingExpenses_Success(t *testing.T) {
	mockRepo := new(MockExpenseRepository)
	service := NewExpenseService(mockRepo)

	expectedExpenses := []models.Expense{
		{ID: 1, Description: "Pending 1"},
		{ID: 2, Description: "Pending 2"},
	}

	mockRepo.On("GetPendingExpenses", "user123", 7, false).Return(expectedExpenses, nil)

	expenses, err := service.GetPendingExpenses("user123", 7, false)

	assert.NoError(t, err)
	assert.Len(t, expenses, 2)
	mockRepo.AssertExpectations(t)
}

func TestExpenseService_GetPendingExpenses_EmptyUserID(t *testing.T) {
	mockRepo := new(MockExpenseRepository)
	service := NewExpenseService(mockRepo)

	expenses, err := service.GetPendingExpenses("", 7, false)

	assert.Error(t, err)
	assert.Nil(t, expenses)
	assert.Contains(t, err.Error(), "user_id is required")
}

func TestExpenseService_GetPendingExpenses_NegativeDaysAhead(t *testing.T) {
	mockRepo := new(MockExpenseRepository)
	service := NewExpenseService(mockRepo)

	expectedExpenses := []models.Expense{}

	mockRepo.On("GetPendingExpenses", "user123", 0, false).Return(expectedExpenses, nil)

	expenses, err := service.GetPendingExpenses("user123", -5, false)

	assert.NoError(t, err)
	assert.Empty(t, expenses)
	mockRepo.AssertExpectations(t)
}

func TestExpenseService_GetPendingExpenses_OverdueOnly(t *testing.T) {
	mockRepo := new(MockExpenseRepository)
	service := NewExpenseService(mockRepo)

	expectedExpenses := []models.Expense{
		{ID: 1, Description: "Overdue 1"},
	}

	mockRepo.On("GetPendingExpenses", "user123", 7, true).Return(expectedExpenses, nil)

	expenses, err := service.GetPendingExpenses("user123", 7, true)

	assert.NoError(t, err)
	assert.Len(t, expenses, 1)
	mockRepo.AssertExpectations(t)
}
