package handlers

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/fernandovalenzuela/api-home-pay/internal/middleware"
	"github.com/fernandovalenzuela/api-home-pay/internal/models"
	"github.com/fernandovalenzuela/api-home-pay/internal/services"
	"github.com/fernandovalenzuela/api-home-pay/internal/utils"
	"github.com/gin-gonic/gin"
)

type IncomeHandler struct {
	service services.IncomeService
}

func NewIncomeHandler(service services.IncomeService) *IncomeHandler {
	return &IncomeHandler{service: service}
}

// @Summary Create income
// @Description Create a new income
// @Tags incomes
// @Accept json
// @Produce json
// @Param income body models.CreateIncomeRequest true "Income data"
// @Success 201 {object} models.Income
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /api/incomes [post]
func (h *IncomeHandler) Create(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.ErrorResponseClient(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req models.CreateIncomeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponseClient(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	income := models.Income{
		PeriodID:    req.PeriodID,
		Description: req.Description,
		Amount:      req.Amount,
		IsRecurring: req.IsRecurring,
		ReceivedAt:  req.ReceivedAt,
	}

	if err := h.service.Create(userID, &income); err != nil {
		slog.Warn("business error", "path", c.Request.URL.Path, "error", err.Error(), "user_id", userID)
		utils.ErrorResponseClient(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, income)
}

// @Summary Get income by ID
// @Description Get a specific income by ID
// @Tags incomes
// @Accept json
// @Produce json
// @Param id path int true "Income ID"
// @Success 200 {object} models.Income
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /api/incomes/{id} [get]
func (h *IncomeHandler) GetByID(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.ErrorResponseClient(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.ErrorResponseClient(c, http.StatusBadRequest, "Invalid income ID")
		return
	}

	income, err := h.service.GetByID(userID, id)
	if err != nil {
		slog.Warn("business error", "path", c.Request.URL.Path, "error", err.Error(), "user_id", userID)
		utils.ErrorResponseClient(c, http.StatusNotFound, err.Error())
		return
	}

	utils.SuccessResponse(c, income)
}

// @Summary List all incomes
// @Description Get all incomes for the authenticated user, optionally filtered by period
// @Tags incomes
// @Accept json
// @Produce json
// @Param period_id query int false "Filter by period ID"
// @Success 200 {array} models.Income
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /api/incomes [get]
func (h *IncomeHandler) GetAll(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.ErrorResponseClient(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var periodID *int
	if periodIDStr := c.Query("period_id"); periodIDStr != "" {
		id, err := strconv.Atoi(periodIDStr)
		if err == nil && id > 0 {
			periodID = &id
		}
	}

	incomes, err := h.service.GetAll(userID, periodID)
	if err != nil {
		slog.Error("failed to retrieve incomes", "error", err, "user_id", userID)
		utils.ErrorResponseClient(c, http.StatusInternalServerError, "Failed to retrieve incomes")
		return
	}

	utils.SuccessResponse(c, incomes)
}

// @Summary Update income
// @Description Update an existing income
// @Tags incomes
// @Accept json
// @Produce json
// @Param id path int true "Income ID"
// @Param income body models.UpdateIncomeRequest true "Income data"
// @Success 200 {object} models.Income
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /api/incomes/{id} [put]
func (h *IncomeHandler) Update(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.ErrorResponseClient(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.ErrorResponseClient(c, http.StatusBadRequest, "Invalid income ID")
		return
	}

	var req models.UpdateIncomeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponseClient(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	income := models.Income{
		ID:          id,
		PeriodID:    req.PeriodID,
		Description: req.Description,
		Amount:      req.Amount,
		IsRecurring: req.IsRecurring,
		ReceivedAt:  req.ReceivedAt,
	}

	if err := h.service.Update(userID, &income); err != nil {
		slog.Warn("business error", "path", c.Request.URL.Path, "error", err.Error(), "user_id", userID)
		utils.ErrorResponseClient(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, income)
}

// @Summary Delete income
// @Description Delete an income by ID
// @Tags incomes
// @Accept json
// @Produce json
// @Param id path int true "Income ID"
// @Success 200 {object} utils.ErrorResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /api/incomes/{id} [delete]
func (h *IncomeHandler) Delete(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.ErrorResponseClient(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.ErrorResponseClient(c, http.StatusBadRequest, "Invalid income ID")
		return
	}

	if err := h.service.Delete(userID, id); err != nil {
		slog.Warn("business error", "path", c.Request.URL.Path, "error", err.Error(), "user_id", userID)
		utils.ErrorResponseClient(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "Income deleted successfully"})
}
