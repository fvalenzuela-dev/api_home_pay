package service

import (
	"context"
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
