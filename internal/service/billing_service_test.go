package service

import (
	"context"
	"errors"
	"testing"

	"github.com/homepay/api/internal/models"
	"github.com/homepay/api/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockBillingRepository is a mock for testing BillingService
type MockBillingRepository struct {
	mock.Mock
}

func (m *MockBillingRepository) Create(ctx context.Context, accountID string, req *models.CreateBillingRequest) (*models.AccountBilling, error) {
	args := m.Called(ctx, accountID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AccountBilling), args.Error(1)
}

func (m *MockBillingRepository) CreateCarryOver(ctx context.Context, accountID string, period int, amount float64, carriedFrom string) (*models.AccountBilling, error) {
	args := m.Called(ctx, accountID, period, amount, carriedFrom)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AccountBilling), args.Error(1)
}

func (m *MockBillingRepository) GetByID(ctx context.Context, id, authUserID string) (*models.AccountBilling, error) {
	args := m.Called(ctx, id, authUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AccountBilling), args.Error(1)
}

func (m *MockBillingRepository) GetByAccountAndPeriod(ctx context.Context, accountID, authUserID string, period int) (*models.AccountBilling, error) {
	args := m.Called(ctx, accountID, authUserID, period)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AccountBilling), args.Error(1)
}

func (m *MockBillingRepository) GetAllByAccount(ctx context.Context, accountID, authUserID string, p models.PaginationParams) ([]models.AccountBilling, int, error) {
	args := m.Called(ctx, accountID, authUserID, p)
	return args.Get(0).([]models.AccountBilling), args.Int(1), args.Error(2)
}

func (m *MockBillingRepository) GetUnpaidByAccount(ctx context.Context, accountID, authUserID string) (*models.AccountBilling, error) {
	args := m.Called(ctx, accountID, authUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AccountBilling), args.Error(1)
}

func (m *MockBillingRepository) GetAllByPeriod(ctx context.Context, authUserID string, period int, isPaid *bool, p models.PaginationParams) ([]models.AccountBillingWithDetails, int, error) {
	args := m.Called(ctx, authUserID, period, isPaid, p)
	return args.Get(0).([]models.AccountBillingWithDetails), args.Int(1), args.Error(2)
}

func (m *MockBillingRepository) BulkInsertForPeriod(ctx context.Context, period int, inserts []models.PeriodBillingInsert) error {
	args := m.Called(ctx, period, inserts)
	return args.Error(0)
}

func (m *MockBillingRepository) Update(ctx context.Context, id, authUserID string, req *models.UpdateBillingRequest) (*models.AccountBilling, error) {
	args := m.Called(ctx, id, authUserID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AccountBilling), args.Error(1)
}

func (m *MockBillingRepository) MarkPaid(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockBillingRepository) SoftDelete(ctx context.Context, id, authUserID string) error {
	args := m.Called(ctx, id, authUserID)
	return args.Error(0)
}

func (m *MockBillingRepository) SoftDeleteByAccount(ctx context.Context, accountID string) error {
	args := m.Called(ctx, accountID)
	return args.Error(0)
}

func (m *MockBillingRepository) GetAll(ctx context.Context, authUserID string, filters models.BillingFilters, p models.PaginationParams) ([]models.AccountBilling, int, error) {
	args := m.Called(ctx, authUserID, filters, p)
	var result []models.AccountBilling
	if args.Get(0) != nil {
		result = args.Get(0).([]models.AccountBilling)
	}
	return result, args.Int(1), args.Error(2)
}

// MockAccountRepository is a mock for testing BillingService
type MockAccountRepository struct {
	mock.Mock
}

func (m *MockAccountRepository) Create(ctx context.Context, companyID, authUserID string, req *models.CreateAccountRequest) (*models.Account, error) {
	args := m.Called(ctx, companyID, authUserID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Account), args.Error(1)
}

func (m *MockAccountRepository) GetByID(ctx context.Context, id, authUserID string) (*models.Account, error) {
	args := m.Called(ctx, id, authUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Account), args.Error(1)
}

func (m *MockAccountRepository) GetAllFiltered(ctx context.Context, authUserID string, companyID *string, sort, order string, p models.PaginationParams) ([]models.Account, int, error) {
	args := m.Called(ctx, authUserID, companyID, sort, order, p)
	return args.Get(0).([]models.Account), args.Int(1), args.Error(2)
}

func (m *MockAccountRepository) GetAllActiveByUser(ctx context.Context, authUserID string) ([]models.Account, error) {
	args := m.Called(ctx, authUserID)
	return args.Get(0).([]models.Account), args.Error(1)
}

func (m *MockAccountRepository) GetActiveIDsByCompany(ctx context.Context, companyID string) ([]string, error) {
	args := m.Called(ctx, companyID)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockAccountRepository) Update(ctx context.Context, id, authUserID string, req *models.UpdateAccountRequest) (*models.Account, error) {
	args := m.Called(ctx, id, authUserID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Account), args.Error(1)
}

func (m *MockAccountRepository) SoftDelete(ctx context.Context, id, authUserID string) error {
	args := m.Called(ctx, id, authUserID)
	return args.Error(0)
}

func (m *MockAccountRepository) SoftDeleteByCompany(ctx context.Context, companyID string) error {
	args := m.Called(ctx, companyID)
	return args.Error(0)
}

// Interface compliance
var _ BillingService = (*billingService)(nil)
var _ repository.BillingRepository = (*MockBillingRepository)(nil)
var _ repository.AccountRepository = (*MockAccountRepository)(nil)

// BillingService Unit Tests

func TestBillingService_GetAll(t *testing.T) {
	mockBillings := new(MockBillingRepository)
	mockAccounts := new(MockAccountRepository)
	svc := NewBillingService(mockBillings, mockAccounts)

	authUserID := "user_123"
	filters := models.BillingFilters{}
	pagination := models.PaginationParams{Limit: 10, Page: 1}

	expectedBillings := []models.AccountBilling{
		{ID: "billing-1", AccountID: "account-1", Period: 202604, AmountBilled: 15000.00},
		{ID: "billing-2", AccountID: "account-2", Period: 202603, AmountBilled: 12000.00},
	}

	mockBillings.On("GetAll", mock.Anything, authUserID, filters, pagination).Return(expectedBillings, 2, nil)

	result, total, err := svc.GetAll(context.Background(), authUserID, filters, pagination)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(result))
	assert.Equal(t, 2, total)
	assert.Equal(t, "billing-1", result[0].ID)
	mockBillings.AssertExpectations(t)
}

func TestBillingService_GetAll_WithFilters(t *testing.T) {
	mockBillings := new(MockBillingRepository)
	mockAccounts := new(MockAccountRepository)
	svc := NewBillingService(mockBillings, mockAccounts)

	authUserID := "user_123"
	accountID := "account-123"
	isPaid := true
	filters := models.BillingFilters{
		AccountID: &accountID,
		IsPaid:    &isPaid,
	}
	pagination := models.PaginationParams{Limit: 10, Page: 1}

	expectedBillings := []models.AccountBilling{
		{ID: "billing-1", AccountID: accountID, Period: 202604, IsPaid: true},
	}

	mockBillings.On("GetAll", mock.Anything, authUserID, mock.MatchedBy(func(f models.BillingFilters) bool {
		return f.AccountID != nil && *f.AccountID == accountID && f.IsPaid != nil && *f.IsPaid == isPaid
	}), pagination).Return(expectedBillings, 1, nil)

	result, total, err := svc.GetAll(context.Background(), authUserID, filters, pagination)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(result))
	assert.Equal(t, 1, total)
	assert.True(t, result[0].IsPaid)
	mockBillings.AssertExpectations(t)
}

func TestBillingService_GetAll_EmptyResult(t *testing.T) {
	mockBillings := new(MockBillingRepository)
	mockAccounts := new(MockAccountRepository)
	svc := NewBillingService(mockBillings, mockAccounts)

	authUserID := "user_123"
	filters := models.BillingFilters{}
	pagination := models.PaginationParams{Limit: 10, Page: 1}

	mockBillings.On("GetAll", mock.Anything, authUserID, filters, pagination).Return([]models.AccountBilling(nil), 0, nil)

	result, total, err := svc.GetAll(context.Background(), authUserID, filters, pagination)

	assert.NoError(t, err)
	assert.Nil(t, result)
	assert.Equal(t, 0, total)
	mockBillings.AssertExpectations(t)
}

func TestBillingService_GetAll_Error(t *testing.T) {
	mockBillings := new(MockBillingRepository)
	mockAccounts := new(MockAccountRepository)
	svc := NewBillingService(mockBillings, mockAccounts)

	authUserID := "user_123"
	filters := models.BillingFilters{}
	pagination := models.PaginationParams{Limit: 10, Page: 1}

	mockBillings.On("GetAll", mock.Anything, authUserID, filters, pagination).Return(nil, 0, errors.New("database error"))

	result, total, err := svc.GetAll(context.Background(), authUserID, filters, pagination)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, 0, total)
	assert.Contains(t, err.Error(), "database error")
	mockBillings.AssertExpectations(t)
}

func TestBillingService_Delete_Success(t *testing.T) {
	mockBillings := new(MockBillingRepository)
	mockAccounts := new(MockAccountRepository)
	svc := NewBillingService(mockBillings, mockAccounts)

	billingID := "billing-123"
	authUserID := "user_123"

	billing := &models.AccountBilling{ID: billingID, AccountID: "account-1"}
	mockBillings.On("GetByID", mock.Anything, billingID, authUserID).Return(billing, nil)
	mockBillings.On("SoftDelete", mock.Anything, billingID, authUserID).Return(nil)

	err := svc.Delete(context.Background(), billingID, authUserID)

	assert.NoError(t, err)
	mockBillings.AssertExpectations(t)
}

func TestBillingService_Delete_NotFound(t *testing.T) {
	mockBillings := new(MockBillingRepository)
	mockAccounts := new(MockAccountRepository)
	svc := NewBillingService(mockBillings, mockAccounts)

	billingID := "non-existent"
	authUserID := "user_123"

	mockBillings.On("GetByID", mock.Anything, billingID, authUserID).Return(nil, nil)

	err := svc.Delete(context.Background(), billingID, authUserID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	mockBillings.AssertExpectations(t)
}

func TestBillingService_Delete_NotFoundDueToAuth(t *testing.T) {
	// When GetByID returns nil (not found) due to auth check failure,
	// Delete should return "not found" (this is correct - don't leak existence)
	mockBillings := new(MockBillingRepository)
	mockAccounts := new(MockAccountRepository)
	svc := NewBillingService(mockBillings, mockAccounts)

	billingID := "billing-123"
	authUserID := "other_user"

	// Even if billing exists but belongs to different user, GetByID returns nil
	mockBillings.On("GetByID", mock.Anything, billingID, authUserID).Return(nil, nil)

	err := svc.Delete(context.Background(), billingID, authUserID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	mockBillings.AssertExpectations(t)
}

func TestBillingService_Delete_RepoError(t *testing.T) {
	mockBillings := new(MockBillingRepository)
	mockAccounts := new(MockAccountRepository)
	svc := NewBillingService(mockBillings, mockAccounts)

	billingID := "billing-123"
	authUserID := "user_123"

	billing := &models.AccountBilling{ID: billingID, AccountID: "account-1"}
	mockBillings.On("GetByID", mock.Anything, billingID, authUserID).Return(billing, nil)
	mockBillings.On("SoftDelete", mock.Anything, billingID, authUserID).Return(errors.New("delete failed"))

	err := svc.Delete(context.Background(), billingID, authUserID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "delete failed")
	mockBillings.AssertExpectations(t)
}

func TestBillingService_Delete_GetByIDError(t *testing.T) {
	mockBillings := new(MockBillingRepository)
	mockAccounts := new(MockAccountRepository)
	svc := NewBillingService(mockBillings, mockAccounts)

	billingID := "billing-123"
	authUserID := "user_123"

	mockBillings.On("GetByID", mock.Anything, billingID, authUserID).Return(nil, errors.New("db connection error"))

	err := svc.Delete(context.Background(), billingID, authUserID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db connection error")
	mockBillings.AssertExpectations(t)
}

func TestBillingService_Delete_SoftDeleteNotFound(t *testing.T) {
	// GetByID succeeds (billing exists and belongs to user)
	// But SoftDelete returns pgx.ErrNoRows (billing already deleted or other edge case)
	mockBillings := new(MockBillingRepository)
	mockAccounts := new(MockAccountRepository)
	svc := NewBillingService(mockBillings, mockAccounts)

	billingID := "billing-123"
	authUserID := "user_123"

	billing := &models.AccountBilling{ID: billingID, AccountID: "account-1"}
	mockBillings.On("GetByID", mock.Anything, billingID, authUserID).Return(billing, nil)
	mockBillings.On("SoftDelete", mock.Anything, billingID, authUserID).Return(pgx.ErrNoRows)

	err := svc.Delete(context.Background(), billingID, authUserID)

	assert.Error(t, err)
	mockBillings.AssertExpectations(t)
}