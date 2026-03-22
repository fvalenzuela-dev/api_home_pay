package services

import (
	"testing"

	"github.com/fernandovalenzuela/api-home-pay/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockCompanyRepository is a mock implementation of CompanyRepository
type MockCompanyRepository struct {
	mock.Mock
}

func (m *MockCompanyRepository) Create(userID string, company *models.Company) error {
	args := m.Called(userID, company)
	return args.Error(0)
}

func (m *MockCompanyRepository) GetByID(userID string, id int) (*models.Company, error) {
	args := m.Called(userID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Company), args.Error(1)
}

func (m *MockCompanyRepository) GetAll(userID string) ([]models.Company, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Company), args.Error(1)
}

func (m *MockCompanyRepository) Update(userID string, company *models.Company) error {
	args := m.Called(userID, company)
	return args.Error(0)
}

func (m *MockCompanyRepository) Delete(userID string, id int) error {
	args := m.Called(userID, id)
	return args.Error(0)
}

func (m *MockCompanyRepository) ExistsByName(userID string, name string) (bool, error) {
	args := m.Called(userID, name)
	return args.Bool(0), args.Error(1)
}

func (m *MockCompanyRepository) HasServiceAccounts(id int) (bool, error) {
	args := m.Called(id)
	return args.Bool(0), args.Error(1)
}

func TestCompanyService_Create_Success(t *testing.T) {
	mockRepo := new(MockCompanyRepository)
	service := NewCompanyService(mockRepo)

	company := &models.Company{Name: "Acme Corp", WebsiteURL: "https://acme.com"}

	mockRepo.On("ExistsByName", "user123", "Acme Corp").Return(false, nil)
	mockRepo.On("Create", "user123", company).Return(nil)

	err := service.Create("user123", company)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestCompanyService_Create_EmptyName(t *testing.T) {
	mockRepo := new(MockCompanyRepository)
	service := NewCompanyService(mockRepo)

	company := &models.Company{Name: ""}

	err := service.Create("user123", company)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name cannot be empty")
}

func TestCompanyService_Create_WhitespaceName(t *testing.T) {
	mockRepo := new(MockCompanyRepository)
	service := NewCompanyService(mockRepo)

	company := &models.Company{Name: "   "}

	err := service.Create("user123", company)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name cannot be empty")
}

func TestCompanyService_Create_DuplicateName(t *testing.T) {
	mockRepo := new(MockCompanyRepository)
	service := NewCompanyService(mockRepo)

	company := &models.Company{Name: "Acme Corp"}

	mockRepo.On("ExistsByName", "user123", "Acme Corp").Return(true, nil)

	err := service.Create("user123", company)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
	mockRepo.AssertExpectations(t)
}

func TestCompanyService_Create_InvalidWebsite(t *testing.T) {
	mockRepo := new(MockCompanyRepository)
	service := NewCompanyService(mockRepo)

	company := &models.Company{Name: "Acme Corp", WebsiteURL: "not-a-url"}

	mockRepo.On("ExistsByName", "user123", "Acme Corp").Return(false, nil)

	err := service.Create("user123", company)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid website URL")
}

func TestCompanyService_Create_InvalidWebsiteNoProtocol(t *testing.T) {
	mockRepo := new(MockCompanyRepository)
	service := NewCompanyService(mockRepo)

	company := &models.Company{Name: "Acme Corp", WebsiteURL: "acme.com"}

	mockRepo.On("ExistsByName", "user123", "Acme Corp").Return(false, nil)

	err := service.Create("user123", company)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid website URL")
}

func TestCompanyService_GetByID_Success(t *testing.T) {
	mockRepo := new(MockCompanyRepository)
	service := NewCompanyService(mockRepo)

	expectedCompany := &models.Company{ID: 1, Name: "Acme Corp", WebsiteURL: "https://acme.com"}

	mockRepo.On("GetByID", "user123", 1).Return(expectedCompany, nil)

	company, err := service.GetByID("user123", 1)

	assert.NoError(t, err)
	assert.NotNil(t, company)
	assert.Equal(t, "Acme Corp", company.Name)
	mockRepo.AssertExpectations(t)
}

func TestCompanyService_GetByID_InvalidID(t *testing.T) {
	mockRepo := new(MockCompanyRepository)
	service := NewCompanyService(mockRepo)

	company, err := service.GetByID("user123", 0)

	assert.Error(t, err)
	assert.Nil(t, company)
	assert.Contains(t, err.Error(), "invalid company ID")
}

func TestCompanyService_GetByID_NotFound(t *testing.T) {
	mockRepo := new(MockCompanyRepository)
	service := NewCompanyService(mockRepo)

	mockRepo.On("GetByID", "user123", 999).Return(nil, nil)

	company, err := service.GetByID("user123", 999)

	assert.Error(t, err)
	assert.Nil(t, company)
	assert.Contains(t, err.Error(), "not found")
	mockRepo.AssertExpectations(t)
}

func TestCompanyService_GetAll_Success(t *testing.T) {
	mockRepo := new(MockCompanyRepository)
	service := NewCompanyService(mockRepo)

	expectedCompanies := []models.Company{
		{ID: 1, Name: "Acme Corp"},
		{ID: 2, Name: "Tech Inc"},
	}

	mockRepo.On("GetAll", "user123").Return(expectedCompanies, nil)

	companies, err := service.GetAll("user123")

	assert.NoError(t, err)
	assert.Len(t, companies, 2)
	mockRepo.AssertExpectations(t)
}

func TestCompanyService_Update_Success(t *testing.T) {
	mockRepo := new(MockCompanyRepository)
	service := NewCompanyService(mockRepo)

	existingCompany := &models.Company{ID: 1, Name: "Old Corp", WebsiteURL: "https://old.com"}
	updatedCompany := &models.Company{ID: 1, Name: "New Corp", WebsiteURL: "https://new.com"}

	mockRepo.On("GetByID", "user123", 1).Return(existingCompany, nil)
	mockRepo.On("ExistsByName", "user123", "New Corp").Return(false, nil)
	mockRepo.On("Update", "user123", updatedCompany).Return(nil)

	err := service.Update("user123", updatedCompany)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestCompanyService_Update_InvalidID(t *testing.T) {
	mockRepo := new(MockCompanyRepository)
	service := NewCompanyService(mockRepo)

	company := &models.Company{ID: 0, Name: "Test Corp"}

	err := service.Update("user123", company)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid company ID")
}

func TestCompanyService_Update_EmptyName(t *testing.T) {
	mockRepo := new(MockCompanyRepository)
	service := NewCompanyService(mockRepo)

	company := &models.Company{ID: 1, Name: ""}

	err := service.Update("user123", company)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name cannot be empty")
}

func TestCompanyService_Update_InvalidWebsite(t *testing.T) {
	mockRepo := new(MockCompanyRepository)
	service := NewCompanyService(mockRepo)

	company := &models.Company{ID: 1, Name: "Test Corp", WebsiteURL: "invalid"}

	err := service.Update("user123", company)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid website URL")
}

func TestCompanyService_Update_NotFound(t *testing.T) {
	mockRepo := new(MockCompanyRepository)
	service := NewCompanyService(mockRepo)

	company := &models.Company{ID: 999, Name: "Test Corp"}

	mockRepo.On("GetByID", "user123", 999).Return(nil, nil)

	err := service.Update("user123", company)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	mockRepo.AssertExpectations(t)
}

func TestCompanyService_Update_DuplicateName(t *testing.T) {
	mockRepo := new(MockCompanyRepository)
	service := NewCompanyService(mockRepo)

	existingCompany := &models.Company{ID: 1, Name: "Old Corp"}
	updatedCompany := &models.Company{ID: 1, Name: "Existing Corp"}

	mockRepo.On("GetByID", "user123", 1).Return(existingCompany, nil)
	mockRepo.On("ExistsByName", "user123", "Existing Corp").Return(true, nil)

	err := service.Update("user123", updatedCompany)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
	mockRepo.AssertExpectations(t)
}

func TestCompanyService_Update_SameName(t *testing.T) {
	mockRepo := new(MockCompanyRepository)
	service := NewCompanyService(mockRepo)

	existingCompany := &models.Company{ID: 1, Name: "Same Corp"}
	updatedCompany := &models.Company{ID: 1, Name: "Same Corp", WebsiteURL: "https://updated.com"}

	mockRepo.On("GetByID", "user123", 1).Return(existingCompany, nil)
	mockRepo.On("Update", "user123", updatedCompany).Return(nil)

	err := service.Update("user123", updatedCompany)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestCompanyService_Delete_Success(t *testing.T) {
	mockRepo := new(MockCompanyRepository)
	service := NewCompanyService(mockRepo)

	mockRepo.On("HasServiceAccounts", 1).Return(false, nil)
	mockRepo.On("Delete", "user123", 1).Return(nil)

	err := service.Delete("user123", 1)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestCompanyService_Delete_InvalidID(t *testing.T) {
	mockRepo := new(MockCompanyRepository)
	service := NewCompanyService(mockRepo)

	err := service.Delete("user123", 0)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid company ID")
}

func TestCompanyService_Delete_HasServiceAccounts(t *testing.T) {
	mockRepo := new(MockCompanyRepository)
	service := NewCompanyService(mockRepo)

	mockRepo.On("HasServiceAccounts", 1).Return(true, nil)

	err := service.Delete("user123", 1)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot delete company with associated service accounts")
	mockRepo.AssertExpectations(t)
}
