package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/homepay/api/internal/middleware"
	"github.com/homepay/api/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock DashboardService
type MockDashboardService struct {
	mock.Mock
}

func (m *MockDashboardService) GetSummary(ctx context.Context, authUserID string, month, year int) (*service.DashboardSummary, error) {
	args := m.Called(ctx, authUserID, month, year)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.DashboardSummary), args.Error(1)
}

// Tests for DashboardHandler
func TestDashboardHandler_Get(t *testing.T) {
	mockSvc := new(MockDashboardService)
	handler := NewDashboardHandler(mockSvc)

	t.Run("success - get dashboard summary", func(t *testing.T) {
		summary := &service.DashboardSummary{
			Month:        3,
			Year:         2026,
			TotalBilled:  150000,
			TotalPaid:    100000,
			TotalPending: 50000,
		}
		mockSvc.On("GetSummary", mock.Anything, "user_123", 3, 2026).Return(summary, nil)

		req := httptest.NewRequest("GET", "/dashboard?month=3&year=2026", nil)
		req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserIDKey, "user_123"))
		w := httptest.NewRecorder()

		handler.Get(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockSvc.AssertExpectations(t)
	})
}
