package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/homepay/api/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockBillingRepoForDashboardTest struct {
	mock.Mock
}

func (m *MockBillingRepoForDashboardTest) Create(ctx context.Context, accountID string, req *models.CreateBillingRequest) (*models.AccountBilling, error) {
	args := m.Called(ctx, accountID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AccountBilling), args.Error(1)
}

func (m *MockBillingRepoForDashboardTest) CreateCarryOver(ctx context.Context, accountID string, period int, amount float64, carriedFrom string) (*models.AccountBilling, error) {
	args := m.Called(ctx, accountID, period, amount, carriedFrom)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AccountBilling), args.Error(1)
}

func (m *MockBillingRepoForDashboardTest) GetByID(ctx context.Context, id, authUserID string) (*models.AccountBilling, error) {
	args := m.Called(ctx, id, authUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AccountBilling), args.Error(1)
}

func (m *MockBillingRepoForDashboardTest) GetByAccountAndPeriod(ctx context.Context, accountID string, period int) (*models.AccountBilling, error) {
	args := m.Called(ctx, accountID, period)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AccountBilling), args.Error(1)
}

func (m *MockBillingRepoForDashboardTest) GetAllByAccount(ctx context.Context, accountID, authUserID string, p models.PaginationParams) ([]models.AccountBilling, int, error) {
	args := m.Called(ctx, accountID, authUserID, p)
	return args.Get(0).([]models.AccountBilling), args.Int(1), args.Error(2)
}

func (m *MockBillingRepoForDashboardTest) GetUnpaidByAccount(ctx context.Context, accountID string) (*models.AccountBilling, error) {
	args := m.Called(ctx, accountID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AccountBilling), args.Error(1)
}

func (m *MockBillingRepoForDashboardTest) GetAllByPeriod(ctx context.Context, authUserID string, period int, isPaid *bool, p models.PaginationParams) ([]models.AccountBillingWithDetails, int, error) {
	args := m.Called(ctx, authUserID, period, isPaid, p)
	return args.Get(0).([]models.AccountBillingWithDetails), args.Int(1), args.Error(2)
}

func (m *MockBillingRepoForDashboardTest) BulkInsertForPeriod(ctx context.Context, period int, inserts []models.PeriodBillingInsert) error {
	args := m.Called(ctx, period, inserts)
	return args.Error(0)
}

func (m *MockBillingRepoForDashboardTest) Update(ctx context.Context, id, authUserID string, req *models.UpdateBillingRequest) (*models.AccountBilling, error) {
	args := m.Called(ctx, id, authUserID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AccountBilling), args.Error(1)
}

func (m *MockBillingRepoForDashboardTest) MarkPaid(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockBillingRepoForDashboardTest) SoftDeleteByAccount(ctx context.Context, accountID string) error {
	args := m.Called(ctx, accountID)
	return args.Error(0)
}

type MockExpenseRepoForDashboardTest struct {
	mock.Mock
}

func (m *MockExpenseRepoForDashboardTest) Create(ctx context.Context, authUserID string, req *models.CreateExpenseRequest) (*models.Expense, error) {
	args := m.Called(ctx, authUserID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Expense), args.Error(1)
}

func (m *MockExpenseRepoForDashboardTest) GetByID(ctx context.Context, id, authUserID string) (*models.Expense, error) {
	args := m.Called(ctx, id, authUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Expense), args.Error(1)
}

func (m *MockExpenseRepoForDashboardTest) GetAll(ctx context.Context, authUserID string, filters models.ExpenseFilters, p models.PaginationParams) ([]models.Expense, int, error) {
	args := m.Called(ctx, authUserID, filters, p)
	return args.Get(0).([]models.Expense), args.Int(1), args.Error(2)
}

func (m *MockExpenseRepoForDashboardTest) Update(ctx context.Context, id, authUserID string, req *models.UpdateExpenseRequest) (*models.Expense, error) {
	args := m.Called(ctx, id, authUserID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Expense), args.Error(1)
}

func (m *MockExpenseRepoForDashboardTest) SoftDelete(ctx context.Context, id, authUserID string) error {
	args := m.Called(ctx, id, authUserID)
	return args.Error(0)
}

type MockInstallmentRepoForDashboardTest struct {
	mock.Mock
}

func (m *MockInstallmentRepoForDashboardTest) CreatePlan(ctx context.Context, authUserID string, plan *models.InstallmentPlan) (*models.InstallmentPlan, error) {
	args := m.Called(ctx, authUserID, plan)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.InstallmentPlan), args.Error(1)
}

func (m *MockInstallmentRepoForDashboardTest) CreatePayments(ctx context.Context, payments []models.InstallmentPayment) error {
	args := m.Called(ctx, payments)
	return args.Error(0)
}

func (m *MockInstallmentRepoForDashboardTest) GetPlan(ctx context.Context, id, authUserID string) (*models.InstallmentPlan, error) {
	args := m.Called(ctx, id, authUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.InstallmentPlan), args.Error(1)
}

func (m *MockInstallmentRepoForDashboardTest) GetAllPlans(ctx context.Context, authUserID string, p models.PaginationParams) ([]models.InstallmentPlan, int, error) {
	args := m.Called(ctx, authUserID, p)
	return args.Get(0).([]models.InstallmentPlan), args.Int(1), args.Error(2)
}

func (m *MockInstallmentRepoForDashboardTest) GetPaymentsByPlan(ctx context.Context, planID string, p models.PaginationParams) ([]models.InstallmentPayment, int, error) {
	args := m.Called(ctx, planID, p)
	return args.Get(0).([]models.InstallmentPayment), args.Int(1), args.Error(2)
}

func (m *MockInstallmentRepoForDashboardTest) GetActivePaymentsByMonth(ctx context.Context, authUserID string, month, year int) ([]models.InstallmentPayment, error) {
	args := m.Called(ctx, authUserID, month, year)
	return args.Get(0).([]models.InstallmentPayment), args.Error(1)
}

func (m *MockInstallmentRepoForDashboardTest) UpdatePayment(ctx context.Context, planID, paymentID, authUserID string) (*models.InstallmentPayment, error) {
	args := m.Called(ctx, planID, paymentID, authUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.InstallmentPayment), args.Error(1)
}

func (m *MockInstallmentRepoForDashboardTest) IncrementPaid(ctx context.Context, planID string, total int) error {
	args := m.Called(ctx, planID, total)
	return args.Error(0)
}

func (m *MockInstallmentRepoForDashboardTest) SoftDeletePlan(ctx context.Context, id, authUserID string) error {
	args := m.Called(ctx, id, authUserID)
	return args.Error(0)
}

func TestDashboardService_EmptySummary(t *testing.T) {
	mockBilling := new(MockBillingRepoForDashboardTest)
	mockExpense := new(MockExpenseRepoForDashboardTest)
	mockInstallment := new(MockInstallmentRepoForDashboardTest)
	svc := NewDashboardService(mockBilling, mockExpense, mockInstallment)

	t.Run("empty summary with no data", func(t *testing.T) {
		mockBilling.On("GetAllByPeriod", mock.Anything, "user_123", 202603, (*bool)(nil), mock.Anything).Return([]models.AccountBillingWithDetails{}, 0, nil)
		mockExpense.On("GetAll", mock.Anything, "user_123", mock.Anything, mock.Anything).Return([]models.Expense{}, 0, nil)
		mockInstallment.On("GetActivePaymentsByMonth", mock.Anything, "user_123", 3, 2026).Return([]models.InstallmentPayment{}, nil)

		result, err := svc.GetSummary(context.Background(), "user_123", 3, 2026)

		assert.NoError(t, err)
		assert.Equal(t, 0.0, result.TotalBilled)
		assert.Equal(t, 0.0, result.TotalPaid)
	})
}

func TestDashboardService_GetSummary(t *testing.T) {
	t.Run("success with data", func(t *testing.T) {
		mockBilling := new(MockBillingRepoForDashboardTest)
		mockExpense := new(MockExpenseRepoForDashboardTest)
		mockInstallment := new(MockInstallmentRepoForDashboardTest)
		svc := NewDashboardService(mockBilling, mockExpense, mockInstallment)

		billings := []models.AccountBillingWithDetails{
			{AccountBilling: models.AccountBilling{ID: "b1", AccountID: "acc1", AmountBilled: 50000, AmountPaid: 30000, IsPaid: false}},
		}
		mockBilling.On("GetAllByPeriod", mock.Anything, "user_123", 202603, (*bool)(nil), mock.Anything).Return(billings, 1, nil)
		mockExpense.On("GetAll", mock.Anything, "user_123", mock.Anything, mock.Anything).Return([]models.Expense{}, 0, nil)
		mockInstallment.On("GetActivePaymentsByMonth", mock.Anything, "user_123", 3, 2026).Return([]models.InstallmentPayment{}, nil)

		result, err := svc.GetSummary(context.Background(), "user_123", 3, 2026)

		assert.NoError(t, err)
		assert.Equal(t, 50000.0, result.TotalBilled)
		assert.Equal(t, 20000.0, result.TotalPending)
	})
}

func TestDashboardService_GetSummary_ExpensesByCompany(t *testing.T) {
	t.Run("expenses grouped by company", func(t *testing.T) {
		mockBilling := new(MockBillingRepoForDashboardTest)
		mockExpense := new(MockExpenseRepoForDashboardTest)
		mockInstallment := new(MockInstallmentRepoForDashboardTest)
		svc := NewDashboardService(mockBilling, mockExpense, mockInstallment)

		company1 := "company-1"
		company2 := "company-2"

		billings := []models.AccountBillingWithDetails{}
		mockBilling.On("GetAllByPeriod", mock.Anything, "user_123", 202603, (*bool)(nil), mock.Anything).Return(billings, 0, nil)

		expenses := []models.Expense{
			{ID: "e1", CompanyID: &company1, Amount: 10000},
			{ID: "e2", CompanyID: &company1, Amount: 5000},
			{ID: "e3", CompanyID: &company2, Amount: 15000},
		}
		mockExpense.On("GetAll", mock.Anything, "user_123", mock.Anything, mock.Anything).Return(expenses, 3, nil)
		mockInstallment.On("GetActivePaymentsByMonth", mock.Anything, "user_123", 3, 2026).Return([]models.InstallmentPayment{}, nil)

		result, err := svc.GetSummary(context.Background(), "user_123", 3, 2026)

		assert.NoError(t, err)
		assert.Equal(t, 30000.0, result.TotalExpenses)
		assert.Len(t, result.ExpensesByCompany, 2)
		assert.Equal(t, 15000.0, result.ExpensesByCompany[company1])
		assert.Equal(t, 15000.0, result.ExpensesByCompany[company2])
	})
}

func TestDashboardService_GetSummary_NullCompanyID(t *testing.T) {
	t.Run("expenses with null company_id included in total but not in map", func(t *testing.T) {
		mockBilling := new(MockBillingRepoForDashboardTest)
		mockExpense := new(MockExpenseRepoForDashboardTest)
		mockInstallment := new(MockInstallmentRepoForDashboardTest)
		svc := NewDashboardService(mockBilling, mockExpense, mockInstallment)

		billings := []models.AccountBillingWithDetails{}
		mockBilling.On("GetAllByPeriod", mock.Anything, "user_123", 202603, (*bool)(nil), mock.Anything).Return(billings, 0, nil)

		// Expense with no company (nil CompanyID)
		expenses := []models.Expense{
			{ID: "e1", CompanyID: nil, Amount: 8000},
			{ID: "e2", CompanyID: nil, Amount: 2000},
		}
		mockExpense.On("GetAll", mock.Anything, "user_123", mock.Anything, mock.Anything).Return(expenses, 2, nil)
		mockInstallment.On("GetActivePaymentsByMonth", mock.Anything, "user_123", 3, 2026).Return([]models.InstallmentPayment{}, nil)

		result, err := svc.GetSummary(context.Background(), "user_123", 3, 2026)

		assert.NoError(t, err)
		assert.Equal(t, 10000.0, result.TotalExpenses)
		assert.Len(t, result.ExpensesByCompany, 0) // No company ID, so no entries
	})
}

func TestDashboardService_GetSummary_EmptyResult(t *testing.T) {
	t.Run("empty summary with no data", func(t *testing.T) {
		mockBilling := new(MockBillingRepoForDashboardTest)
		mockExpense := new(MockExpenseRepoForDashboardTest)
		mockInstallment := new(MockInstallmentRepoForDashboardTest)
		svc := NewDashboardService(mockBilling, mockExpense, mockInstallment)

		mockBilling.On("GetAllByPeriod", mock.Anything, "user_123", 202603, (*bool)(nil), mock.Anything).Return([]models.AccountBillingWithDetails{}, 0, nil)
		mockExpense.On("GetAll", mock.Anything, "user_123", mock.Anything, mock.Anything).Return([]models.Expense{}, 0, nil)
		mockInstallment.On("GetActivePaymentsByMonth", mock.Anything, "user_123", 3, 2026).Return([]models.InstallmentPayment{}, nil)

		result, err := svc.GetSummary(context.Background(), "user_123", 3, 2026)

		assert.NoError(t, err)
		assert.Equal(t, 0.0, result.TotalBilled)
		assert.Equal(t, 0.0, result.TotalPaid)
		assert.Equal(t, 0.0, result.TotalPending)
		assert.Equal(t, 0.0, result.TotalExpenses)
		assert.Len(t, result.ExpensesByCompany, 0)
	})
}

func TestDashboardService_GetSummary_RepoError(t *testing.T) {
	t.Run("error from installment repo propagates", func(t *testing.T) {
		mockBilling := new(MockBillingRepoForDashboardTest)
		mockExpense := new(MockExpenseRepoForDashboardTest)
		mockInstallment := new(MockInstallmentRepoForDashboardTest)
		svc := NewDashboardService(mockBilling, mockExpense, mockInstallment)

		billings := []models.AccountBillingWithDetails{}
		mockBilling.On("GetAllByPeriod", mock.Anything, "user_123", 202603, (*bool)(nil), mock.Anything).Return(billings, 0, nil)
		mockExpense.On("GetAll", mock.Anything, "user_123", mock.Anything, mock.Anything).Return([]models.Expense{}, 0, nil)
		mockInstallment.On("GetActivePaymentsByMonth", mock.Anything, "user_123", 3, 2026).Return([]models.InstallmentPayment{}, fmt.Errorf("installment repo error"))

		result, err := svc.GetSummary(context.Background(), "user_123", 3, 2026)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "installment repo error")
	})
}

func TestDashboardService_GetSummary_DateRangeBoundary(t *testing.T) {
	t.Run("expenses on from/to dates are included", func(t *testing.T) {
		mockBilling := new(MockBillingRepoForDashboardTest)
		mockExpense := new(MockExpenseRepoForDashboardTest)
		mockInstallment := new(MockInstallmentRepoForDashboardTest)
		svc := NewDashboardService(mockBilling, mockExpense, mockInstallment)

		billings := []models.AccountBillingWithDetails{}
		mockBilling.On("GetAllByPeriod", mock.Anything, "user_123", 202603, (*bool)(nil), mock.Anything).Return(billings, 0, nil)

		company1 := "company-1"
		expenses := []models.Expense{
			{ID: "e1", CompanyID: &company1, Amount: 5000}, // Boundary expense
		}
		mockExpense.On("GetAll", mock.Anything, "user_123", mock.Anything, mock.Anything).Return(expenses, 1, nil)
		mockInstallment.On("GetActivePaymentsByMonth", mock.Anything, "user_123", 3, 2026).Return([]models.InstallmentPayment{}, nil)

		result, err := svc.GetSummary(context.Background(), "user_123", 3, 2026)

		assert.NoError(t, err)
		assert.Equal(t, 5000.0, result.TotalExpenses)
	})
}

func TestDashboardService_GetSummary_BillingError(t *testing.T) {
	t.Run("error from billing repo propagates", func(t *testing.T) {
		mockBilling := new(MockBillingRepoForDashboardTest)
		mockExpense := new(MockExpenseRepoForDashboardTest)
		mockInstallment := new(MockInstallmentRepoForDashboardTest)
		svc := NewDashboardService(mockBilling, mockExpense, mockInstallment)

		mockBilling.On("GetAllByPeriod", mock.Anything, "user_123", 202603, (*bool)(nil), mock.Anything).Return([]models.AccountBillingWithDetails{}, 0, fmt.Errorf("billing repo error"))

		result, err := svc.GetSummary(context.Background(), "user_123", 3, 2026)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "billing repo error")
	})
}

func TestDashboardService_GetSummary_ExpenseError(t *testing.T) {
	t.Run("error from expense repo propagates", func(t *testing.T) {
		mockBilling := new(MockBillingRepoForDashboardTest)
		mockExpense := new(MockExpenseRepoForDashboardTest)
		mockInstallment := new(MockInstallmentRepoForDashboardTest)
		svc := NewDashboardService(mockBilling, mockExpense, mockInstallment)

		billings := []models.AccountBillingWithDetails{}
		mockBilling.On("GetAllByPeriod", mock.Anything, "user_123", 202603, (*bool)(nil), mock.Anything).Return(billings, 0, nil)
		mockExpense.On("GetAll", mock.Anything, "user_123", mock.Anything, mock.Anything).Return([]models.Expense{}, 0, fmt.Errorf("expense repo error"))

		result, err := svc.GetSummary(context.Background(), "user_123", 3, 2026)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "expense repo error")
	})
}

func TestDashboardService_GetSummary_WithInstallments(t *testing.T) {
	t.Run("installments included in summary", func(t *testing.T) {
		mockBilling := new(MockBillingRepoForDashboardTest)
		mockExpense := new(MockExpenseRepoForDashboardTest)
		mockInstallment := new(MockInstallmentRepoForDashboardTest)
		svc := NewDashboardService(mockBilling, mockExpense, mockInstallment)

		billings := []models.AccountBillingWithDetails{}
		mockBilling.On("GetAllByPeriod", mock.Anything, "user_123", 202603, (*bool)(nil), mock.Anything).Return(billings, 0, nil)
		mockExpense.On("GetAll", mock.Anything, "user_123", mock.Anything, mock.Anything).Return([]models.Expense{}, 0, nil)

		installments := []models.InstallmentPayment{
			{ID: "pmt1", PlanID: "plan1", InstallmentNumber: 1, Amount: 10000, IsPaid: true},
			{ID: "pmt2", PlanID: "plan1", InstallmentNumber: 2, Amount: 10000, IsPaid: false},
		}
		mockInstallment.On("GetActivePaymentsByMonth", mock.Anything, "user_123", 3, 2026).Return(installments, nil)

		result, err := svc.GetSummary(context.Background(), "user_123", 3, 2026)

		assert.NoError(t, err)
		assert.Equal(t, 20000.0, result.TotalInstallments)
		assert.Len(t, result.PendingCommitments, 1)
		assert.Equal(t, "installment", result.PendingCommitments[0].Type)
	})
}
