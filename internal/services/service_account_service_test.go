package services

import (
	"testing"

	"github.com/fernandovalenzuela/api-home-pay/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockServiceAccountRepository is a mock implementation of ServiceAccountRepository
type MockServiceAccountRepository struct {
	mock.Mock
}

func (m *MockServiceAccountRepository) Create(userID string, account *models.ServiceAccount) error {
	args := m.Called(userID, account)
	return args.Error(0)
}

func (m *MockServiceAccountRepository) GetByID(userID string, id int) (*models.ServiceAccount, error) {
	args := m.Called(userID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ServiceAccount), args.Error(1)
}

func (m *MockServiceAccountRepository) GetAll(userID string, companyID *int) ([]models.ServiceAccount, error) {
	args := m.Called(userID, companyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.ServiceAccount), args.Error(1)
}

func (m *MockServiceAccountRepository) Update(userID string, account *models.ServiceAccount) error {
	args := m.Called(userID, account)
	return args.Error(0)
}

func (m *MockServiceAccountRepository) Delete(userID string, id int) error {
	args := m.Called(userID, id)
	return args.Error(0)
}

func (m *MockServiceAccountRepository) ExistsByIdentifier(userID string, companyID int, identifier string) (bool, error) {
	args := m.Called(userID, companyID, identifier)
	return args.Bool(0), args.Error(1)
}

func (m *MockServiceAccountRepository) HasExpenses(id int) (bool, error) {
	args := m.Called(id)
	return args.Bool(0), args.Error(1)
}

func (m *MockServiceAccountRepository) CompanyExistsAndBelongsToUser(userID string, companyID int) (bool, error) {
	args := m.Called(userID, companyID)
	return args.Bool(0), args.Error(1)
}

func TestServiceAccountService_Create_Success(t *testing.T) {
	mockRepo := new(MockServiceAccountRepository)
	service := NewServiceAccountService(mockRepo)

	account := &models.ServiceAccount{
		CompanyID:         1,
		AccountIdentifier: "ACC123456",
		Alias:             "My Account",
	}

	mockRepo.On("CompanyExistsAndBelongsToUser", "user123", 1).Return(true, nil)
	mockRepo.On("ExistsByIdentifier", "user123", 1, "ACC123456").Return(false, nil)
	mockRepo.On("Create", "user123", account).Return(nil)

	err := service.Create("user123", account)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestServiceAccountService_Create_InvalidCompanyID(t *testing.T) {
	mockRepo := new(MockServiceAccountRepository)
	service := NewServiceAccountService(mockRepo)

	account := &models.ServiceAccount{
		CompanyID:         0,
		AccountIdentifier: "ACC123456",
	}

	mockRepo.On("Create", "user123", account).Return(nil)

	err := service.Create("user123", account)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestServiceAccountService_Create_EmptyIdentifier(t *testing.T) {
	mockRepo := new(MockServiceAccountRepository)
	service := NewServiceAccountService(mockRepo)

	account := &models.ServiceAccount{
		CompanyID:         1,
		AccountIdentifier: "",
	}

	err := service.Create("user123", account)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "account identifier cannot be empty")
}

func TestServiceAccountService_Create_WhitespaceIdentifier(t *testing.T) {
	mockRepo := new(MockServiceAccountRepository)
	service := NewServiceAccountService(mockRepo)

	account := &models.ServiceAccount{
		CompanyID:         1,
		AccountIdentifier: "   ",
	}

	err := service.Create("user123", account)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "account identifier cannot be empty")
}

func TestServiceAccountService_Create_CompanyNotFound(t *testing.T) {
	mockRepo := new(MockServiceAccountRepository)
	service := NewServiceAccountService(mockRepo)

	account := &models.ServiceAccount{
		CompanyID:         999,
		AccountIdentifier: "ACC123456",
	}

	mockRepo.On("CompanyExistsAndBelongsToUser", "user123", 999).Return(false, nil)

	err := service.Create("user123", account)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "company not found or access denied")
	mockRepo.AssertExpectations(t)
}

func TestServiceAccountService_Create_DuplicateIdentifier(t *testing.T) {
	mockRepo := new(MockServiceAccountRepository)
	service := NewServiceAccountService(mockRepo)

	account := &models.ServiceAccount{
		CompanyID:         1,
		AccountIdentifier: "ACC123456",
	}

	mockRepo.On("CompanyExistsAndBelongsToUser", "user123", 1).Return(true, nil)
	mockRepo.On("ExistsByIdentifier", "user123", 1, "ACC123456").Return(true, nil)

	err := service.Create("user123", account)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
	mockRepo.AssertExpectations(t)
}

func TestServiceAccountService_GetByID_Success(t *testing.T) {
	mockRepo := new(MockServiceAccountRepository)
	service := NewServiceAccountService(mockRepo)

	expectedAccount := &models.ServiceAccount{
		ID:                1,
		CompanyID:         1,
		AccountIdentifier: "ACC123456",
		Alias:             "My Account",
	}

	mockRepo.On("GetByID", "user123", 1).Return(expectedAccount, nil)

	account, err := service.GetByID("user123", 1)

	assert.NoError(t, err)
	assert.NotNil(t, account)
	assert.Equal(t, "ACC123456", account.AccountIdentifier)
	mockRepo.AssertExpectations(t)
}

func TestServiceAccountService_GetByID_InvalidID(t *testing.T) {
	mockRepo := new(MockServiceAccountRepository)
	service := NewServiceAccountService(mockRepo)

	account, err := service.GetByID("user123", 0)

	assert.Error(t, err)
	assert.Nil(t, account)
	assert.Contains(t, err.Error(), "invalid service account ID")
}

func TestServiceAccountService_GetByID_NotFound(t *testing.T) {
	mockRepo := new(MockServiceAccountRepository)
	service := NewServiceAccountService(mockRepo)

	mockRepo.On("GetByID", "user123", 999).Return(nil, nil)

	account, err := service.GetByID("user123", 999)

	assert.Error(t, err)
	assert.Nil(t, account)
	assert.Contains(t, err.Error(), "not found")
	mockRepo.AssertExpectations(t)
}

func TestServiceAccountService_GetAll_Success(t *testing.T) {
	mockRepo := new(MockServiceAccountRepository)
	service := NewServiceAccountService(mockRepo)

	expectedAccounts := []models.ServiceAccount{
		{ID: 1, AccountIdentifier: "ACC001"},
		{ID: 2, AccountIdentifier: "ACC002"},
	}

	mockRepo.On("GetAll", "user123", (*int)(nil)).Return(expectedAccounts, nil)

	accounts, err := service.GetAll("user123", nil)

	assert.NoError(t, err)
	assert.Len(t, accounts, 2)
	mockRepo.AssertExpectations(t)
}

func TestServiceAccountService_GetAll_WithCompanyFilter(t *testing.T) {
	mockRepo := new(MockServiceAccountRepository)
	service := NewServiceAccountService(mockRepo)

	companyID := 1

	mockRepo.On("CompanyExistsAndBelongsToUser", "user123", 1).Return(true, nil)

	expectedAccounts := []models.ServiceAccount{
		{ID: 1, CompanyID: 1, AccountIdentifier: "ACC001"},
	}

	mockRepo.On("GetAll", "user123", &companyID).Return(expectedAccounts, nil)

	accounts, err := service.GetAll("user123", &companyID)

	assert.NoError(t, err)
	assert.Len(t, accounts, 1)
	mockRepo.AssertExpectations(t)
}

func TestServiceAccountService_Update_Success(t *testing.T) {
	mockRepo := new(MockServiceAccountRepository)
	service := NewServiceAccountService(mockRepo)

	existingAccount := &models.ServiceAccount{
		ID:                1,
		CompanyID:         1,
		AccountIdentifier: "OLD123",
	}
	updatedAccount := &models.ServiceAccount{
		ID:                1,
		CompanyID:         1,
		AccountIdentifier: "NEW123",
		Alias:             "Updated Account",
	}

	mockRepo.On("GetByID", "user123", 1).Return(existingAccount, nil)
	mockRepo.On("CompanyExistsAndBelongsToUser", "user123", 1).Return(true, nil)
	mockRepo.On("ExistsByIdentifier", "user123", 1, "NEW123").Return(false, nil)
	mockRepo.On("Update", "user123", updatedAccount).Return(nil)

	err := service.Update("user123", updatedAccount)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestServiceAccountService_Update_InvalidID(t *testing.T) {
	mockRepo := new(MockServiceAccountRepository)
	service := NewServiceAccountService(mockRepo)

	account := &models.ServiceAccount{
		ID:                0,
		CompanyID:         1,
		AccountIdentifier: "ACC123",
	}

	err := service.Update("user123", account)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid service account ID")
}

func TestServiceAccountService_Update_InvalidCompanyID(t *testing.T) {
	mockRepo := new(MockServiceAccountRepository)
	service := NewServiceAccountService(mockRepo)

	existingAccount := &models.ServiceAccount{
		ID:                1,
		CompanyID:         1,
		AccountIdentifier: "ACC123",
	}
	account := &models.ServiceAccount{
		ID:                1,
		CompanyID:         0,
		AccountIdentifier: "ACC123",
	}

	mockRepo.On("GetByID", "user123", 1).Return(existingAccount, nil)
	mockRepo.On("Update", "user123", account).Return(nil)

	err := service.Update("user123", account)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestServiceAccountService_Update_EmptyIdentifier(t *testing.T) {
	mockRepo := new(MockServiceAccountRepository)
	service := NewServiceAccountService(mockRepo)

	account := &models.ServiceAccount{
		ID:                1,
		CompanyID:         1,
		AccountIdentifier: "",
	}

	err := service.Update("user123", account)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "account identifier cannot be empty")
}

func TestServiceAccountService_Update_NotFound(t *testing.T) {
	mockRepo := new(MockServiceAccountRepository)
	service := NewServiceAccountService(mockRepo)

	account := &models.ServiceAccount{
		ID:                999,
		CompanyID:         1,
		AccountIdentifier: "ACC123",
	}

	mockRepo.On("GetByID", "user123", 999).Return(nil, nil)

	err := service.Update("user123", account)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	mockRepo.AssertExpectations(t)
}

func TestServiceAccountService_Update_CompanyNotFound(t *testing.T) {
	mockRepo := new(MockServiceAccountRepository)
	service := NewServiceAccountService(mockRepo)

	existingAccount := &models.ServiceAccount{
		ID:                1,
		CompanyID:         1,
		AccountIdentifier: "ACC123",
	}
	updatedAccount := &models.ServiceAccount{
		ID:                1,
		CompanyID:         999,
		AccountIdentifier: "ACC123",
	}

	mockRepo.On("GetByID", "user123", 1).Return(existingAccount, nil)
	mockRepo.On("CompanyExistsAndBelongsToUser", "user123", 999).Return(false, nil)

	err := service.Update("user123", updatedAccount)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "company not found or access denied")
	mockRepo.AssertExpectations(t)
}

func TestServiceAccountService_Update_DuplicateIdentifier(t *testing.T) {
	mockRepo := new(MockServiceAccountRepository)
	service := NewServiceAccountService(mockRepo)

	existingAccount := &models.ServiceAccount{
		ID:                1,
		CompanyID:         1,
		AccountIdentifier: "OLD123",
	}
	updatedAccount := &models.ServiceAccount{
		ID:                1,
		CompanyID:         1,
		AccountIdentifier: "NEW123",
	}

	mockRepo.On("GetByID", "user123", 1).Return(existingAccount, nil)
	mockRepo.On("CompanyExistsAndBelongsToUser", "user123", 1).Return(true, nil)
	mockRepo.On("ExistsByIdentifier", "user123", 1, "NEW123").Return(true, nil)

	err := service.Update("user123", updatedAccount)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
	mockRepo.AssertExpectations(t)
}

func TestServiceAccountService_Update_SameIdentifier(t *testing.T) {
	mockRepo := new(MockServiceAccountRepository)
	service := NewServiceAccountService(mockRepo)

	existingAccount := &models.ServiceAccount{
		ID:                1,
		CompanyID:         1,
		AccountIdentifier: "SAME123",
	}
	updatedAccount := &models.ServiceAccount{
		ID:                1,
		CompanyID:         1,
		AccountIdentifier: "SAME123",
		Alias:             "Updated Alias",
	}

	mockRepo.On("GetByID", "user123", 1).Return(existingAccount, nil)
	mockRepo.On("CompanyExistsAndBelongsToUser", "user123", 1).Return(true, nil)
	mockRepo.On("Update", "user123", updatedAccount).Return(nil)

	err := service.Update("user123", updatedAccount)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestServiceAccountService_Delete_Success(t *testing.T) {
	mockRepo := new(MockServiceAccountRepository)
	service := NewServiceAccountService(mockRepo)

	mockRepo.On("HasExpenses", 1).Return(false, nil)
	mockRepo.On("Delete", "user123", 1).Return(nil)

	err := service.Delete("user123", 1)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestServiceAccountService_Delete_InvalidID(t *testing.T) {
	mockRepo := new(MockServiceAccountRepository)
	service := NewServiceAccountService(mockRepo)

	err := service.Delete("user123", 0)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid service account ID")
}

func TestServiceAccountService_Delete_HasExpenses(t *testing.T) {
	mockRepo := new(MockServiceAccountRepository)
	service := NewServiceAccountService(mockRepo)

	mockRepo.On("HasExpenses", 1).Return(true, nil)

	err := service.Delete("user123", 1)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot delete service account with associated expenses")
	mockRepo.AssertExpectations(t)
}
