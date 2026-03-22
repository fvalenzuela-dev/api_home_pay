package services

import (
	"errors"
	"testing"

	"github.com/fernandovalenzuela/api-home-pay/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockCategoryRepository is a mock implementation of CategoryRepository
type MockCategoryRepository struct {
	mock.Mock
}

func (m *MockCategoryRepository) Create(userID string, category *models.Category) error {
	args := m.Called(userID, category)
	return args.Error(0)
}

func (m *MockCategoryRepository) GetByID(userID string, id int) (*models.Category, error) {
	args := m.Called(userID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Category), args.Error(1)
}

func (m *MockCategoryRepository) GetAll(userID string) ([]models.Category, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Category), args.Error(1)
}

func (m *MockCategoryRepository) Update(userID string, category *models.Category) error {
	args := m.Called(userID, category)
	return args.Error(0)
}

func (m *MockCategoryRepository) Delete(userID string, id int) error {
	args := m.Called(userID, id)
	return args.Error(0)
}

func (m *MockCategoryRepository) ExistsByName(userID string, name string) (bool, error) {
	args := m.Called(userID, name)
	return args.Bool(0), args.Error(1)
}

func (m *MockCategoryRepository) HasExpenses(id int) (bool, error) {
	args := m.Called(id)
	return args.Bool(0), args.Error(1)
}

func TestCategoryService_Create_Success(t *testing.T) {
	mockRepo := new(MockCategoryRepository)
	service := NewCategoryService(mockRepo)

	category := &models.Category{Name: "Groceries"}

	mockRepo.On("ExistsByName", "user123", "Groceries").Return(false, nil)
	mockRepo.On("Create", "user123", category).Return(nil)

	err := service.Create("user123", category)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestCategoryService_Create_EmptyName(t *testing.T) {
	mockRepo := new(MockCategoryRepository)
	service := NewCategoryService(mockRepo)

	category := &models.Category{Name: ""}

	err := service.Create("user123", category)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name cannot be empty")
}

func TestCategoryService_Create_WhitespaceName(t *testing.T) {
	mockRepo := new(MockCategoryRepository)
	service := NewCategoryService(mockRepo)

	category := &models.Category{Name: "   "}

	err := service.Create("user123", category)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name cannot be empty")
}

func TestCategoryService_Create_DuplicateName(t *testing.T) {
	mockRepo := new(MockCategoryRepository)
	service := NewCategoryService(mockRepo)

	category := &models.Category{Name: "Groceries"}

	mockRepo.On("ExistsByName", "user123", "Groceries").Return(true, nil)

	err := service.Create("user123", category)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
	mockRepo.AssertExpectations(t)
}

func TestCategoryService_Create_RepositoryError(t *testing.T) {
	mockRepo := new(MockCategoryRepository)
	service := NewCategoryService(mockRepo)

	category := &models.Category{Name: "Groceries"}

	mockRepo.On("ExistsByName", "user123", "Groceries").Return(false, errors.New("db error"))

	err := service.Create("user123", category)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validation error")
	mockRepo.AssertExpectations(t)
}

func TestCategoryService_GetByID_Success(t *testing.T) {
	mockRepo := new(MockCategoryRepository)
	service := NewCategoryService(mockRepo)

	expectedCategory := &models.Category{ID: 1, Name: "Groceries"}

	mockRepo.On("GetByID", "user123", 1).Return(expectedCategory, nil)

	category, err := service.GetByID("user123", 1)

	assert.NoError(t, err)
	assert.NotNil(t, category)
	assert.Equal(t, 1, category.ID)
	assert.Equal(t, "Groceries", category.Name)
	mockRepo.AssertExpectations(t)
}

func TestCategoryService_GetByID_InvalidID(t *testing.T) {
	mockRepo := new(MockCategoryRepository)
	service := NewCategoryService(mockRepo)

	category, err := service.GetByID("user123", 0)

	assert.Error(t, err)
	assert.Nil(t, category)
	assert.Contains(t, err.Error(), "invalid category ID")
}

func TestCategoryService_GetByID_NotFound(t *testing.T) {
	mockRepo := new(MockCategoryRepository)
	service := NewCategoryService(mockRepo)

	mockRepo.On("GetByID", "user123", 999).Return(nil, nil)

	category, err := service.GetByID("user123", 999)

	assert.Error(t, err)
	assert.Nil(t, category)
	assert.Contains(t, err.Error(), "not found")
	mockRepo.AssertExpectations(t)
}

func TestCategoryService_GetAll_Success(t *testing.T) {
	mockRepo := new(MockCategoryRepository)
	service := NewCategoryService(mockRepo)

	expectedCategories := []models.Category{
		{ID: 1, Name: "Groceries"},
		{ID: 2, Name: "Utilities"},
	}

	mockRepo.On("GetAll", "user123").Return(expectedCategories, nil)

	categories, err := service.GetAll("user123")

	assert.NoError(t, err)
	assert.Len(t, categories, 2)
	mockRepo.AssertExpectations(t)
}

func TestCategoryService_GetAll_RepositoryError(t *testing.T) {
	mockRepo := new(MockCategoryRepository)
	service := NewCategoryService(mockRepo)

	mockRepo.On("GetAll", "user123").Return(nil, errors.New("db error"))

	categories, err := service.GetAll("user123")

	assert.Error(t, err)
	assert.Nil(t, categories)
	mockRepo.AssertExpectations(t)
}

func TestCategoryService_Update_Success(t *testing.T) {
	mockRepo := new(MockCategoryRepository)
	service := NewCategoryService(mockRepo)

	existingCategory := &models.Category{ID: 1, Name: "Old Name", CreatedAt: "2024-01-01T00:00:00Z"}
	updatedCategory := &models.Category{ID: 1, Name: "New Name"}

	mockRepo.On("GetByID", "user123", 1).Return(existingCategory, nil)
	mockRepo.On("ExistsByName", "user123", "New Name").Return(false, nil)
	mockRepo.On("Update", "user123", updatedCategory).Return(nil)

	err := service.Update("user123", updatedCategory)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestCategoryService_Update_InvalidID(t *testing.T) {
	mockRepo := new(MockCategoryRepository)
	service := NewCategoryService(mockRepo)

	category := &models.Category{ID: 0, Name: "Test"}

	err := service.Update("user123", category)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid category ID")
}

func TestCategoryService_Update_EmptyName(t *testing.T) {
	mockRepo := new(MockCategoryRepository)
	service := NewCategoryService(mockRepo)

	category := &models.Category{ID: 1, Name: ""}

	err := service.Update("user123", category)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name cannot be empty")
}

func TestCategoryService_Update_NotFound(t *testing.T) {
	mockRepo := new(MockCategoryRepository)
	service := NewCategoryService(mockRepo)

	category := &models.Category{ID: 999, Name: "Test"}

	mockRepo.On("GetByID", "user123", 999).Return(nil, nil)

	err := service.Update("user123", category)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	mockRepo.AssertExpectations(t)
}

func TestCategoryService_Update_DuplicateName(t *testing.T) {
	mockRepo := new(MockCategoryRepository)
	service := NewCategoryService(mockRepo)

	existingCategory := &models.Category{ID: 1, Name: "Old Name"}
	updatedCategory := &models.Category{ID: 1, Name: "Existing Name"}

	mockRepo.On("GetByID", "user123", 1).Return(existingCategory, nil)
	mockRepo.On("ExistsByName", "user123", "Existing Name").Return(true, nil)

	err := service.Update("user123", updatedCategory)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
	mockRepo.AssertExpectations(t)
}

func TestCategoryService_Update_SameName(t *testing.T) {
	mockRepo := new(MockCategoryRepository)
	service := NewCategoryService(mockRepo)

	existingCategory := &models.Category{ID: 1, Name: "Same Name"}
	updatedCategory := &models.Category{ID: 1, Name: "Same Name"}

	mockRepo.On("GetByID", "user123", 1).Return(existingCategory, nil)
	mockRepo.On("Update", "user123", updatedCategory).Return(nil)

	err := service.Update("user123", updatedCategory)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestCategoryService_Delete_Success(t *testing.T) {
	mockRepo := new(MockCategoryRepository)
	service := NewCategoryService(mockRepo)

	mockRepo.On("HasExpenses", 1).Return(false, nil)
	mockRepo.On("Delete", "user123", 1).Return(nil)

	err := service.Delete("user123", 1)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestCategoryService_Delete_InvalidID(t *testing.T) {
	mockRepo := new(MockCategoryRepository)
	service := NewCategoryService(mockRepo)

	err := service.Delete("user123", 0)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid category ID")
}

func TestCategoryService_Delete_HasExpenses(t *testing.T) {
	mockRepo := new(MockCategoryRepository)
	service := NewCategoryService(mockRepo)

	mockRepo.On("HasExpenses", 1).Return(true, nil)

	err := service.Delete("user123", 1)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot delete category with associated expenses")
	mockRepo.AssertExpectations(t)
}

func TestCategoryService_Delete_RepositoryError(t *testing.T) {
	mockRepo := new(MockCategoryRepository)
	service := NewCategoryService(mockRepo)

	mockRepo.On("HasExpenses", 1).Return(false, nil)
	mockRepo.On("Delete", "user123", 1).Return(errors.New("db error"))

	err := service.Delete("user123", 1)

	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}
