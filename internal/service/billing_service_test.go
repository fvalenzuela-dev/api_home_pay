package service

import (
	"context"
	"testing"

	"github.com/homepay/api/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockBillingRepoForTest struct {
	mock.Mock
}

func (m *MockBillingRepoForTest) Create(ctx context.Context, accountID string, req *models.CreateBillingRequest) (*models.AccountBilling, error) {
	args := m.Called(ctx, accountID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AccountBilling), args.Error(1)
}

func (m *MockBillingRepoForTest) CreateCarryOver(ctx context.Context, accountID string, period int, amount float64, carriedFrom string) (*models.AccountBilling, error) {
	args := m.Called(ctx, accountID, period, amount, carriedFrom)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AccountBilling), args.Error(1)
}

func (m *MockBillingRepoForTest) GetByID(ctx context.Context, id, authUserID string) (*models.AccountBilling, error) {
	args := m.Called(ctx, id, authUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AccountBilling), args.Error(1)
}

func (m *MockBillingRepoForTest) GetByAccountAndPeriod(ctx context.Context, accountID string, period int) (*models.AccountBilling, error) {
	args := m.Called(ctx, accountID, period)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AccountBilling), args.Error(1)
}

func (m *MockBillingRepoForTest) GetAllByAccount(ctx context.Context, accountID, authUserID string, p models.PaginationParams) ([]models.AccountBilling, int, error) {
	args := m.Called(ctx, accountID, authUserID, p)
	return args.Get(0).([]models.AccountBilling), args.Int(1), args.Error(2)
}

func (m *MockBillingRepoForTest) GetUnpaidByAccount(ctx context.Context, accountID string) (*models.AccountBilling, error) {
	args := m.Called(ctx, accountID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AccountBilling), args.Error(1)
}

func (m *MockBillingRepoForTest) GetAllByPeriod(ctx context.Context, authUserID string, period int, isPaid *bool, p models.PaginationParams) ([]models.AccountBillingWithDetails, int, error) {
	args := m.Called(ctx, authUserID, period, isPaid, p)
	return args.Get(0).([]models.AccountBillingWithDetails), args.Int(1), args.Error(2)
}

func (m *MockBillingRepoForTest) BulkInsertForPeriod(ctx context.Context, period int, inserts []models.PeriodBillingInsert) error {
	args := m.Called(ctx, period, inserts)
	return args.Error(0)
}

func (m *MockBillingRepoForTest) Update(ctx context.Context, id, authUserID string, req *models.UpdateBillingRequest) (*models.AccountBilling, error) {
	args := m.Called(ctx, id, authUserID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AccountBilling), args.Error(1)
}

func (m *MockBillingRepoForTest) MarkPaid(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockBillingRepoForTest) SoftDeleteByAccount(ctx context.Context, accountID string) error {
	args := m.Called(ctx, accountID)
	return args.Error(0)
}

type MockAccountRepoForTest struct {
	mock.Mock
}

func (m *MockAccountRepoForTest) Create(ctx context.Context, companyID, authUserID string, req *models.CreateAccountRequest) (*models.Account, error) {
	args := m.Called(ctx, companyID, authUserID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Account), args.Error(1)
}

func (m *MockAccountRepoForTest) GetByID(ctx context.Context, id, authUserID string) (*models.Account, error) {
	args := m.Called(ctx, id, authUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Account), args.Error(1)
}

func (m *MockAccountRepoForTest) GetAllByCompany(ctx context.Context, companyID, authUserID string, p models.PaginationParams) ([]models.Account, int, error) {
	args := m.Called(ctx, companyID, authUserID, p)
	return args.Get(0).([]models.Account), args.Int(1), args.Error(2)
}

func (m *MockAccountRepoForTest) GetAllActiveByUser(ctx context.Context, authUserID string) ([]models.Account, error) {
	args := m.Called(ctx, authUserID)
	return args.Get(0).([]models.Account), args.Error(1)
}

func (m *MockAccountRepoForTest) GetActiveIDsByCompany(ctx context.Context, companyID string) ([]string, error) {
	args := m.Called(ctx, companyID)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockAccountRepoForTest) Update(ctx context.Context, id, authUserID string, req *models.UpdateAccountRequest) (*models.Account, error) {
	args := m.Called(ctx, id, authUserID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Account), args.Error(1)
}

func (m *MockAccountRepoForTest) SoftDelete(ctx context.Context, id, authUserID string) error {
	args := m.Called(ctx, id, authUserID)
	return args.Error(0)
}

func (m *MockAccountRepoForTest) SoftDeleteByCompany(ctx context.Context, companyID string) error {
	args := m.Called(ctx, companyID)
	return args.Error(0)
}

func ptrFloat64(f float64) *float64 {
	return &f
}

func TestBillingService_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockBilling := new(MockBillingRepoForTest)
		mockAccounts := new(MockAccountRepoForTest)
		svc := NewBillingService(mockBilling, mockAccounts)
		req := &models.CreateBillingRequest{Period: 202603, AmountBilled: 50000}
		mockAccounts.On("GetByID", mock.Anything, "acc1", "user_123").Return(&models.Account{ID: "acc1"}, nil)
		mockBilling.On("Create", mock.Anything, "acc1", req).Return(&models.AccountBilling{ID: "b1", AmountBilled: 50000, AmountPaid: 0, IsPaid: false}, nil)

		result, err := svc.Create(context.Background(), "acc1", "user_123", req)

		assert.NoError(t, err)
		assert.Equal(t, "b1", result.ID)
	})

	t.Run("account not found", func(t *testing.T) {
		mockBilling := new(MockBillingRepoForTest)
		mockAccounts := new(MockAccountRepoForTest)
		svc := NewBillingService(mockBilling, mockAccounts)
		req := &models.CreateBillingRequest{Period: 202603, AmountBilled: 50000}
		mockAccounts.On("GetByID", mock.Anything, "acc1", "user_123").Return(nil, nil)

		result, err := svc.Create(context.Background(), "acc1", "user_123", req)

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("error - invalid period month", func(t *testing.T) {
		mockBilling := new(MockBillingRepoForTest)
		mockAccounts := new(MockAccountRepoForTest)
		svc := NewBillingService(mockBilling, mockAccounts)
		req := &models.CreateBillingRequest{Period: 202613, AmountBilled: 50000}

		result, err := svc.Create(context.Background(), "account-123", "user_123", req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "mes debe estar entre 01 y 12")
	})

	t.Run("error - year too low", func(t *testing.T) {
		mockBilling := new(MockBillingRepoForTest)
		mockAccounts := new(MockAccountRepoForTest)
		svc := NewBillingService(mockBilling, mockAccounts)
		req := &models.CreateBillingRequest{Period: 199912, AmountBilled: 50000}

		result, err := svc.Create(context.Background(), "account-123", "user_123", req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "año mínimo 2000")
	})

	t.Run("error - amount_billed must be greater than 0", func(t *testing.T) {
		mockBilling := new(MockBillingRepoForTest)
		mockAccounts := new(MockAccountRepoForTest)
		svc := NewBillingService(mockBilling, mockAccounts)
		req := &models.CreateBillingRequest{Period: 202603, AmountBilled: 0}

		result, err := svc.Create(context.Background(), "account-123", "user_123", req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "amount_billed debe ser mayor a 0")
	})
}

func TestBillingService_GetAllByPeriod(t *testing.T) {
	mockBilling := new(MockBillingRepoForTest)
	mockAccounts := new(MockAccountRepoForTest)
	svc := NewBillingService(mockBilling, mockAccounts)

	t.Run("error - invalid period", func(t *testing.T) {
		result, total, err := svc.GetAllByPeriod(context.Background(), "user_123", 202613, nil, models.PaginationParams{})

		assert.Error(t, err)
		assert.Equal(t, 0, total)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "period inválido")
	})
}

func TestBillingService_OpenPeriod(t *testing.T) {
	mockBilling := new(MockBillingRepoForTest)
	mockAccounts := new(MockAccountRepoForTest)
	svc := NewBillingService(mockBilling, mockAccounts)

	t.Run("error - invalid period", func(t *testing.T) {
		result, err := svc.OpenPeriod(context.Background(), "user_123", 202613)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "period inválido")
	})
}

func TestBillingService_GetAllByAccount(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockBilling := new(MockBillingRepoForTest)
		mockAccounts := new(MockAccountRepoForTest)
		svc := NewBillingService(mockBilling, mockAccounts)
		mockBilling.On("GetAllByAccount", mock.Anything, "acc1", "user_123", mock.Anything).Return([]models.AccountBilling{{ID: "b1"}}, 1, nil)

		result, total, err := svc.GetAllByAccount(context.Background(), "acc1", "user_123", models.PaginationParams{})

		assert.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, result, 1)
	})
}

func TestBillingService_GetByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockBilling := new(MockBillingRepoForTest)
		mockAccounts := new(MockAccountRepoForTest)
		svc := NewBillingService(mockBilling, mockAccounts)
		mockBilling.On("GetByID", mock.Anything, "b1", "user_123").Return(&models.AccountBilling{ID: "b1"}, nil)

		result, err := svc.GetByID(context.Background(), "b1", "user_123")

		assert.NoError(t, err)
		assert.Equal(t, "b1", result.ID)
	})

	t.Run("not found", func(t *testing.T) {
		mockBilling := new(MockBillingRepoForTest)
		mockAccounts := new(MockAccountRepoForTest)
		svc := NewBillingService(mockBilling, mockAccounts)
		mockBilling.On("GetByID", mock.Anything, "notfound", "user_123").Return(nil, nil)

		result, err := svc.GetByID(context.Background(), "notfound", "user_123")

		assert.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestBillingService_Update(t *testing.T) {
	t.Run("success - mark paid", func(t *testing.T) {
		mockBilling := new(MockBillingRepoForTest)
		mockAccounts := new(MockAccountRepoForTest)
		svc := NewBillingService(mockBilling, mockAccounts)
		amount := 100.0
		req := &models.UpdateBillingRequest{AmountPaid: &amount}
		mockBilling.On("Update", mock.Anything, "b1", "user_123", req).Return(&models.AccountBilling{ID: "b1", AmountBilled: 100, AmountPaid: 100, IsPaid: false}, nil)
		mockBilling.On("MarkPaid", mock.Anything, "b1").Return(nil)

		result, err := svc.Update(context.Background(), "b1", "user_123", req)

		assert.NoError(t, err)
		assert.True(t, result.IsPaid)
	})

	t.Run("not found", func(t *testing.T) {
		mockBilling := new(MockBillingRepoForTest)
		mockAccounts := new(MockAccountRepoForTest)
		svc := NewBillingService(mockBilling, mockAccounts)
		req := &models.UpdateBillingRequest{}
		mockBilling.On("Update", mock.Anything, "notfound", "user_123", req).Return(nil, nil)

		result, err := svc.Update(context.Background(), "notfound", "user_123", req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "not found")
	})
}
