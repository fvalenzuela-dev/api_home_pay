package handlers

import (
	"net/http"
	"strconv"

	"github.com/fernandovalenzuela/api-home-pay/internal/middleware"
	"github.com/fernandovalenzuela/api-home-pay/internal/models"
	"github.com/fernandovalenzuela/api-home-pay/internal/repository"
	"github.com/fernandovalenzuela/api-home-pay/internal/services"
	"github.com/fernandovalenzuela/api-home-pay/internal/utils"
	"github.com/gin-gonic/gin"
)

type ExpenseHandler struct {
	service services.ExpenseService
}

func NewExpenseHandler(service services.ExpenseService) *ExpenseHandler {
	return &ExpenseHandler{service: service}
}

// @Summary Create expense
// @Description Create a new expense
// @Tags expenses
// @Accept json
// @Produce json
// @Param expense body models.CreateExpenseRequest true "Expense data"
// @Success 201 {object} models.Expense
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /api/expenses [post]
func (h *ExpenseHandler) Create(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.ErrorResponseClient(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req models.CreateExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponseClient(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	expense := models.Expense{
		CategoryID:         req.CategoryID,
		PeriodID:           req.PeriodID,
		AccountID:          req.AccountID,
		Description:        req.Description,
		DueDate:            req.DueDate,
		CurrentAmount:      req.CurrentAmount,
		AmountPaid:         req.AmountPaid,
		CurrentInstallment: req.CurrentInstallment,
		TotalInstallments:  req.TotalInstallments,
		InstallmentGroupID: req.InstallmentGroupID,
		IsRecurring:        req.IsRecurring,
		Notes:              req.Notes,
	}

	if err := h.service.Create(userID, &expense); err != nil {
		utils.ErrorResponseClient(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, expense)
}

// @Summary Get expense by ID
// @Description Get a specific expense by ID
// @Tags expenses
// @Accept json
// @Produce json
// @Param id path int true "Expense ID"
// @Success 200 {object} models.Expense
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /api/expenses/{id} [get]
func (h *ExpenseHandler) GetByID(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.ErrorResponseClient(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.ErrorResponseClient(c, http.StatusBadRequest, "Invalid expense ID")
		return
	}

	expense, err := h.service.GetByID(userID, id)
	if err != nil {
		utils.ErrorResponseClient(c, http.StatusNotFound, err.Error())
		return
	}

	utils.SuccessResponse(c, expense)
}

// @Summary List all expenses
// @Description Get all expenses for the authenticated user with optional filters
// @Tags expenses
// @Accept json
// @Produce json
// @Param period_id query int false "Filter by period ID"
// @Param category_id query int false "Filter by category ID"
// @Param account_id query int false "Filter by service account ID"
// @Param payment_status query string false "Filter by payment status (paid/pending)"
// @Success 200 {array} models.Expense
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /api/expenses [get]
func (h *ExpenseHandler) GetAll(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.ErrorResponseClient(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	filters := repository.ExpenseFilters{}

	if periodIDStr := c.Query("period_id"); periodIDStr != "" {
		id, err := strconv.Atoi(periodIDStr)
		if err == nil && id > 0 {
			filters.PeriodID = &id
		}
	}

	if categoryIDStr := c.Query("category_id"); categoryIDStr != "" {
		id, err := strconv.Atoi(categoryIDStr)
		if err == nil && id > 0 {
			filters.CategoryID = &id
		}
	}

	if accountIDStr := c.Query("account_id"); accountIDStr != "" {
		id, err := strconv.Atoi(accountIDStr)
		if err == nil && id > 0 {
			filters.AccountID = &id
		}
	}

	if paymentStatus := c.Query("payment_status"); paymentStatus != "" {
		filters.PaymentStatus = &paymentStatus
	}

	expenses, err := h.service.GetAll(userID, filters)
	if err != nil {
		utils.ErrorResponseClient(c, http.StatusInternalServerError, "Failed to retrieve expenses")
		return
	}

	utils.SuccessResponse(c, expenses)
}

// @Summary Update expense
// @Description Update an existing expense
// @Tags expenses
// @Accept json
// @Produce json
// @Param id path int true "Expense ID"
// @Param expense body models.UpdateExpenseRequest true "Expense data"
// @Success 200 {object} models.Expense
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /api/expenses/{id} [put]
func (h *ExpenseHandler) Update(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.ErrorResponseClient(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.ErrorResponseClient(c, http.StatusBadRequest, "Invalid expense ID")
		return
	}

	var req models.UpdateExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponseClient(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	expense := models.Expense{
		ID:                 id,
		CategoryID:         req.CategoryID,
		PeriodID:           req.PeriodID,
		AccountID:          req.AccountID,
		Description:        req.Description,
		DueDate:            req.DueDate,
		CurrentAmount:      req.CurrentAmount,
		AmountPaid:         req.AmountPaid,
		CurrentInstallment: req.CurrentInstallment,
		TotalInstallments:  req.TotalInstallments,
		InstallmentGroupID: req.InstallmentGroupID,
		IsRecurring:        req.IsRecurring,
		Notes:              req.Notes,
	}

	if err := h.service.Update(userID, &expense); err != nil {
		utils.ErrorResponseClient(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, expense)
}

// @Summary Delete expense
// @Description Delete an expense by ID
// @Tags expenses
// @Accept json
// @Produce json
// @Param id path int true "Expense ID"
// @Success 200 {object} utils.ErrorResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /api/expenses/{id} [delete]
func (h *ExpenseHandler) Delete(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.ErrorResponseClient(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.ErrorResponseClient(c, http.StatusBadRequest, "Invalid expense ID")
		return
	}

	if err := h.service.Delete(userID, id); err != nil {
		utils.ErrorResponseClient(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "Expense deleted successfully"})
}

// @Summary Mark expense as paid
// @Description Mark an expense as paid
// @Tags expenses
// @Accept json
// @Produce json
// @Param id path int true "Expense ID"
// @Success 200 {object} utils.ErrorResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /api/expenses/{id}/pay [patch]
func (h *ExpenseHandler) MarkAsPaid(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.ErrorResponseClient(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.ErrorResponseClient(c, http.StatusBadRequest, "Invalid expense ID")
		return
	}

	if err := h.service.MarkAsPaid(userID, id); err != nil {
		utils.ErrorResponseClient(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "Expense marked as paid"})
}

// @Summary Get pending expenses
// @Description Get pending expenses with optional filters for days ahead and overdue
// @Tags expenses
// @Accept json
// @Produce json
// @Param days_ahead query int false "Number of days to look ahead (default: 7)"
// @Param overdue_only query bool false "Only return overdue expenses"
// @Success 200 {array} models.Expense
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /api/expenses/pending [get]
func (h *ExpenseHandler) GetPending(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.ErrorResponseClient(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse days_ahead parameter (default: 7)
	daysAhead := 7
	if daysStr := c.Query("days_ahead"); daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d >= 0 {
			daysAhead = d
		}
	}

	// Parse overdue_only parameter
	overdueOnly := false
	if overdueStr := c.Query("overdue_only"); overdueStr != "" {
		if overdueStr == "true" || overdueStr == "1" {
			overdueOnly = true
		}
	}

	expenses, err := h.service.GetPendingExpenses(userID, daysAhead, overdueOnly)
	if err != nil {
		utils.ErrorResponseClient(c, http.StatusInternalServerError, "Failed to retrieve pending expenses")
		return
	}

	utils.SuccessResponse(c, expenses)
}
