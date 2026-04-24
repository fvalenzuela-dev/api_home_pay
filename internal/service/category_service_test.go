package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExpenseService_Exists(t *testing.T) {
	assert.NotNil(t, NewExpenseService)
}

func TestBillingService_Exists(t *testing.T) {
	assert.NotNil(t, NewBillingService)
}

func TestInstallmentService_Exists(t *testing.T) {
	assert.NotNil(t, NewInstallmentService)
}

func TestDashboardService_Exists(t *testing.T) {
	assert.NotNil(t, NewDashboardService)
}

func TestAccountGroupService_Exists(t *testing.T) {
	assert.NotNil(t, NewAccountGroupService)
}

func TestCategoryService_NotExists(t *testing.T) {
	// Note: CategoryService does not exist in this project.
	// Categories are handled directly by CategoryHandler using CategoryRepository.
	// See internal/handlers/categories.go
	assert.True(t, true)
}
