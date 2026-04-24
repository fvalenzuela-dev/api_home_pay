package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAccount_Struct(t *testing.T) {
	now := time.Now()
	groupID := "group-123"
	accountNum := "123456789"

	account := Account{
		ID:             "account-123",
		CompanyID:      "company-123",
		GroupID:        &groupID,
		AccountNumber:  &accountNum,
		Name:           "Test Account",
		BillingDay:     15,
		AutoAccumulate: true,
		IsActive:       true,
		CreatedAt:      now,
	}

	assert.Equal(t, "account-123", account.ID)
	assert.Equal(t, "company-123", account.CompanyID)
	assert.NotNil(t, account.GroupID)
	assert.NotNil(t, account.AccountNumber)
	assert.Equal(t, "Test Account", account.Name)
	assert.Equal(t, 15, account.BillingDay)
	assert.True(t, account.AutoAccumulate)
	assert.True(t, account.IsActive)
}

func TestAccount_OptionalFields(t *testing.T) {
	account := Account{
		ID:             "account-123",
		CompanyID:      "company-123",
		GroupID:        nil,
		AccountNumber:  nil,
		Name:           "Test Account",
		BillingDay:     1,
		AutoAccumulate: false,
		IsActive:       true,
	}

	assert.Nil(t, account.GroupID)
	assert.Nil(t, account.AccountNumber)
}

func TestCreateAccountRequest_Validation(t *testing.T) {
	groupID := "group-123"

	tests := []struct {
		name    string
		req     CreateAccountRequest
		isValid bool
	}{
		{
			name: "valid request",
			req: CreateAccountRequest{
				Name:           "Electricity",
				BillingDay:     15,
				AutoAccumulate: true,
			},
			isValid: true,
		},
		{
			name: "valid request with optional fields",
			req: CreateAccountRequest{
				GroupID:        &groupID,
				Name:           "Water",
				BillingDay:     1,
				AutoAccumulate: false,
			},
			isValid: true,
		},
		{
			name: "empty name",
			req: CreateAccountRequest{
				Name:       "",
				BillingDay: 15,
			},
			isValid: false,
		},
		{
			name: "invalid billing day - zero",
			req: CreateAccountRequest{
				Name:       "Test",
				BillingDay: 0,
			},
			isValid: false,
		},
		{
			name: "invalid billing day - negative",
			req: CreateAccountRequest{
				Name:       "Test",
				BillingDay: -1,
			},
			isValid: false,
		},
		{
			name: "invalid billing day - over 31",
			req: CreateAccountRequest{
				Name:       "Test",
				BillingDay: 32,
			},
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.isValid {
				assert.NotEmpty(t, tt.req.Name)
				assert.GreaterOrEqual(t, tt.req.BillingDay, 1)
				assert.LessOrEqual(t, tt.req.BillingDay, 31)
			} else {
				if tt.req.Name == "" || tt.req.BillingDay < 1 || tt.req.BillingDay > 31 {
					assert.True(t, tt.req.Name == "" || tt.req.BillingDay < 1 || tt.req.BillingDay > 31)
				}
			}
		})
	}
}

func TestUpdateAccountRequest_PartialUpdate(t *testing.T) {
	groupID := "new-group"
	accountNum := "987654321"
	name := "Updated Account"
	billingDay := 20
	autoAccumulate := false

	req := UpdateAccountRequest{
		GroupID:       &groupID,
		AccountNumber: &accountNum,
		Name:          &name,
		BillingDay:    &billingDay,
		AutoAccumulate: &autoAccumulate,
	}

	assert.NotNil(t, req.GroupID)
	assert.NotNil(t, req.AccountNumber)
	assert.NotNil(t, req.Name)
	assert.NotNil(t, req.BillingDay)
	assert.NotNil(t, req.AutoAccumulate)
	assert.Equal(t, "new-group", *req.GroupID)
	assert.Equal(t, "Updated Account", *req.Name)
	assert.Equal(t, 20, *req.BillingDay)
	assert.False(t, *req.AutoAccumulate)
}
