package handlers

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/fernandovalenzuela/api-home-pay/internal/middleware"
	"github.com/fernandovalenzuela/api-home-pay/internal/repository"
	"github.com/fernandovalenzuela/api-home-pay/internal/utils"
	"github.com/gin-gonic/gin"
)

type SummaryHandler struct {
	expenseRepo repository.ExpenseRepository
	incomeRepo  repository.IncomeRepository
	periodRepo  repository.PeriodRepository
}

type FinancialSummaryResponse struct {
	Period          *PeriodInfo `json:"period"`
	TotalIncomes    float64     `json:"total_incomes"`
	TotalExpenses   float64     `json:"total_expenses"`
	Balance         float64     `json:"balance"`
	PaidExpenses    float64     `json:"paid_expenses"`
	PendingExpenses float64     `json:"pending_expenses"`
	ExpenseCount    int         `json:"expense_count"`
	IncomeCount     int         `json:"income_count"`
}

type PeriodInfo struct {
	ID    int `json:"id"`
	Month int `json:"month"`
	Year  int `json:"year"`
}

func NewSummaryHandler(expenseRepo repository.ExpenseRepository, incomeRepo repository.IncomeRepository, periodRepo repository.PeriodRepository) *SummaryHandler {
	return &SummaryHandler{
		expenseRepo: expenseRepo,
		incomeRepo:  incomeRepo,
		periodRepo:  periodRepo,
	}
}

// @Summary Get financial summary
// @Description Get financial summary for a specific period including incomes, expenses, balance
// @Tags summary
// @Accept json
// @Produce json
// @Param period_id path int true "Period ID"
// @Success 200 {object} handlers.FinancialSummaryResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /api/summary/{period_id} [get]
func (h *SummaryHandler) GetByPeriod(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.ErrorResponseClient(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	periodID, err := strconv.Atoi(c.Param("period_id"))
	if err != nil {
		utils.ErrorResponseClient(c, http.StatusBadRequest, "Invalid period ID")
		return
	}

	// Verify period belongs to user
	period, err := h.periodRepo.GetByID(userID, periodID)
	if err != nil {
		slog.Error("failed to verify period", "error", err, "user_id", userID)
		utils.ErrorResponseClient(c, http.StatusInternalServerError, "Failed to verify period")
		return
	}
	if period == nil {
		slog.Warn("business error", "path", c.Request.URL.Path, "error", "period not found", "user_id", userID)
		utils.ErrorResponseClient(c, http.StatusNotFound, "Period not found")
		return
	}

	// Get expense summary
	expenseSummary, err := h.expenseRepo.GetSummaryByPeriod(userID, periodID)
	if err != nil {
		slog.Error("failed to get expense summary", "error", err, "user_id", userID)
		utils.ErrorResponseClient(c, http.StatusInternalServerError, "Failed to get expense summary")
		return
	}

	// Get income summary
	totalIncomes, incomeCount, err := h.incomeRepo.GetTotalByPeriod(userID, periodID)
	if err != nil {
		slog.Error("failed to get income summary", "error", err, "user_id", userID)
		utils.ErrorResponseClient(c, http.StatusInternalServerError, "Failed to get income summary")
		return
	}

	// Calculate balance
	balance := totalIncomes - expenseSummary.TotalAmount

	response := FinancialSummaryResponse{
		Period: &PeriodInfo{
			ID:    period.ID,
			Month: period.MonthNumber,
			Year:  period.YearNumber,
		},
		TotalIncomes:    totalIncomes,
		TotalExpenses:   expenseSummary.TotalAmount,
		Balance:         balance,
		PaidExpenses:    expenseSummary.PaidAmount,
		PendingExpenses: expenseSummary.PendingAmount,
		ExpenseCount:    expenseSummary.ExpenseCount,
		IncomeCount:     incomeCount,
	}

	utils.SuccessResponse(c, response)
}
