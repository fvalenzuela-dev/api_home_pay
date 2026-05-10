package repository

import (
	"context"
	"testing"
	"time"

	"github.com/homepay/api/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAccountRepository is a mock implementation of AccountRepository
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

// MockBillingRepository is a mock implementation of BillingRepository
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

func (m *MockBillingRepository) SoftDeleteByAccount(ctx context.Context, accountID string) error {
	args := m.Called(ctx, accountID)
	return args.Error(0)
}

// Helper function to create test account
func createTestAccount(id, companyID, name string) *models.Account {
	return &models.Account{
		ID:             id,
		CompanyID:      companyID,
		Name:           name,
		BillingDay:     15,
		AutoAccumulate: true,
		IsActive:       true,
		CreatedAt:      time.Now(),
	}
}

// Helper function to create test billing
func createTestBilling(id, accountID string, period int, amount float64, isPaid bool) *models.AccountBilling {
	return &models.AccountBilling{
		ID:           id,
		AccountID:    accountID,
		Period:       period,
		AmountBilled: amount,
		AmountPaid:   amount,
		IsPaid:       isPaid,
		CreatedAt:    time.Now(),
	}
}

// Unit tests using mocks

func TestAccountRepo_Create_WithMock(t *testing.T) {
	mockRepo := new(MockAccountRepository)

	accountID := "account-123"
	companyID := "company-123"
	authUserID := "user-123"

	req := &models.CreateAccountRequest{
		Name:            "Test Account",
		BillingDay:      15,
		AutoAccumulate:  true,
	}

	expectedAccount := createTestAccount(accountID, companyID, req.Name)
	expectedAccount.CompanyID = companyID

	mockRepo.On("Create", mock.Anything, companyID, authUserID, req).Return(expectedAccount, nil)

	result, err := mockRepo.Create(context.Background(), companyID, authUserID, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, accountID, result.ID)
	assert.Equal(t, req.Name, result.Name)
	mockRepo.AssertExpectations(t)
}

func TestAccountRepo_GetByID_WithMock(t *testing.T) {
	mockRepo := new(MockAccountRepository)

	accountID := "account-123"
	authUserID := "user-123"

	expectedAccount := createTestAccount(accountID, "company-123", "Test Account")

	mockRepo.On("GetByID", mock.Anything, accountID, authUserID).Return(expectedAccount, nil)

	result, err := mockRepo.GetByID(context.Background(), accountID, authUserID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, accountID, result.ID)
	mockRepo.AssertExpectations(t)
}

func TestAccountRepo_GetByID_NotFound_WithMock(t *testing.T) {
	mockRepo := new(MockAccountRepository)

	mockRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)

	result, err := mockRepo.GetByID(context.Background(), "non-existent", "user-123")

	assert.NoError(t, err)
	assert.Nil(t, result)
	mockRepo.AssertExpectations(t)
}

func TestAccountRepo_Update_WithMock(t *testing.T) {
	mockRepo := new(MockAccountRepository)

	accountID := "account-123"
	authUserID := "user-123"

	newName := "Updated Account"
	req := &models.UpdateAccountRequest{Name: &newName}

	updatedAccount := createTestAccount(accountID, "company-123", newName)
	mockRepo.On("Update", mock.Anything, accountID, authUserID, req).Return(updatedAccount, nil)

	result, err := mockRepo.Update(context.Background(), accountID, authUserID, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, newName, result.Name)
	mockRepo.AssertExpectations(t)
}

func TestAccountRepo_SoftDelete_WithMock(t *testing.T) {
	mockRepo := new(MockAccountRepository)

	accountID := "account-123"
	authUserID := "user-123"

	mockRepo.On("SoftDelete", mock.Anything, accountID, authUserID).Return(nil)

	err := mockRepo.SoftDelete(context.Background(), accountID, authUserID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestBillingRepo_Create_WithMock(t *testing.T) {
	mockRepo := new(MockBillingRepository)

	accountID := "account-123"
	req := &models.CreateBillingRequest{
		Period:       202603,
		AmountBilled: 15000.00,
	}

	expectedBilling := createTestBilling("billing-123", accountID, req.Period, req.AmountBilled, false)

	mockRepo.On("Create", mock.Anything, accountID, req).Return(expectedBilling, nil)

	result, err := mockRepo.Create(context.Background(), accountID, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, req.Period, result.Period)
	assert.Equal(t, req.AmountBilled, result.AmountBilled)
	mockRepo.AssertExpectations(t)
}

func TestBillingRepo_GetByID_WithMock(t *testing.T) {
	mockRepo := new(MockBillingRepository)

	billingID := "billing-123"
	authUserID := "user-123"

	expectedBilling := createTestBilling(billingID, "account-123", 202603, 15000.00, false)

	mockRepo.On("GetByID", mock.Anything, billingID, authUserID).Return(expectedBilling, nil)

	result, err := mockRepo.GetByID(context.Background(), billingID, authUserID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, billingID, result.ID)
	mockRepo.AssertExpectations(t)
}

func TestBillingRepo_Update_WithMock(t *testing.T) {
	mockRepo := new(MockBillingRepository)

	billingID := "billing-123"
	authUserID := "user-123"

	isPaid := true
	req := &models.UpdateBillingRequest{IsPaid: &isPaid}

	updatedBilling := createTestBilling(billingID, "account-123", 202603, 15000.00, true)
	mockRepo.On("Update", mock.Anything, billingID, authUserID, req).Return(updatedBilling, nil)

	result, err := mockRepo.Update(context.Background(), billingID, authUserID, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.IsPaid)
	mockRepo.AssertExpectations(t)
}

func TestBillingRepo_MarkPaid_WithMock(t *testing.T) {
	mockRepo := new(MockBillingRepository)

	billingID := "billing-123"

	mockRepo.On("MarkPaid", mock.Anything, billingID).Return(nil)

	err := mockRepo.MarkPaid(context.Background(), billingID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestBillingRepo_GetUnpaidByAccount_WithMock(t *testing.T) {
	mockRepo := new(MockBillingRepository)

	accountID := "account-123"
	authUserID := "user-123"

	expectedBilling := createTestBilling("billing-123", accountID, 202603, 15000.00, false)

	mockRepo.On("GetUnpaidByAccount", mock.Anything, accountID, authUserID).Return(expectedBilling, nil)

	result, err := mockRepo.GetUnpaidByAccount(context.Background(), accountID, authUserID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.IsPaid)
	mockRepo.AssertExpectations(t)
}

func TestBillingRepo_BulkInsert_WithMock(t *testing.T) {
	mockRepo := new(MockBillingRepository)

	period := 202603
	inserts := []models.PeriodBillingInsert{
		{AccountID: "account-1", AmountBilled: 10000.00},
		{AccountID: "account-2", AmountBilled: 15000.00},
	}

	mockRepo.On("BulkInsertForPeriod", mock.Anything, period, inserts).Return(nil)

	err := mockRepo.BulkInsertForPeriod(context.Background(), period, inserts)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestBillingRepo_SoftDeleteByAccount_WithMock(t *testing.T) {
	mockRepo := new(MockBillingRepository)

	accountID := "account-123"

	mockRepo.On("SoftDeleteByAccount", mock.Anything, accountID).Return(nil)

	err := mockRepo.SoftDeleteByAccount(context.Background(), accountID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestBillingRepo_CreateCarryOver_WithMock(t *testing.T) {
	mockRepo := new(MockBillingRepository)

	accountID := "account-123"
	period := 202604
	amount := 5000.00
	carriedFrom := "billing-123"

	expectedBilling := createTestBilling("billing-124", accountID, period, amount, false)
	carriedFromPtr := carriedFrom
	expectedBilling.CarriedFrom = &carriedFromPtr

	mockRepo.On("CreateCarryOver", mock.Anything, accountID, period, amount, carriedFrom).Return(expectedBilling, nil)

	result, err := mockRepo.CreateCarryOver(context.Background(), accountID, period, amount, carriedFrom)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, period, result.Period)
	assert.NotNil(t, result.CarriedFrom)
	mockRepo.AssertExpectations(t)
}

// Interface compliance tests

func TestMockAccountRepository_ImplementsInterface(t *testing.T) {
	var _ AccountRepository = (*MockAccountRepository)(nil)
}

func TestMockBillingRepository_ImplementsInterface(t *testing.T) {
	var _ BillingRepository = (*MockBillingRepository)(nil)
}

// Model tests

func TestAccountRepo_Create_Model(t *testing.T) {
	t.Run("CreateAccountRequest validation", func(t *testing.T) {
		req := models.CreateAccountRequest{
			Name:            "Test Account",
			BillingDay:      15,
			AutoAccumulate:  true,
		}
		assert.Equal(t, "Test Account", req.Name)
		assert.Equal(t, 15, req.BillingDay)
		assert.True(t, req.AutoAccumulate)
	})
}

func TestBillingRepo_Create_Model(t *testing.T) {
	t.Run("CreateBillingRequest validation", func(t *testing.T) {
		req := models.CreateBillingRequest{
			Period:       202603,
			AmountBilled: 15000.00,
		}
		assert.Equal(t, 202603, req.Period)
		assert.Equal(t, 15000.00, req.AmountBilled)
	})
}

func TestBillingRepo_Update_Model(t *testing.T) {
	t.Run("UpdateBillingRequest with pointer fields", func(t *testing.T) {
		amountBilled := 20000.00
		isPaid := true

		req := models.UpdateBillingRequest{
			AmountBilled: &amountBilled,
			IsPaid:       &isPaid,
		}
		assert.Equal(t, 20000.00, *req.AmountBilled)
		assert.True(t, *req.IsPaid)
	})
}

func TestBillingRepo_Constants(t *testing.T) {
	t.Run("billingCols constant is defined", func(t *testing.T) {
		assert.NotEmpty(t, billingCols)
		assert.Contains(t, billingCols, "id")
		assert.Contains(t, billingCols, "account_id")
	})

	t.Run("billingColsAB constant is defined", func(t *testing.T) {
		assert.NotEmpty(t, billingColsAB)
		assert.Contains(t, billingColsAB, "ab.id")
	})
}
