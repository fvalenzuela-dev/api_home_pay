package handlers

import (
	"net/http"
	"strconv"

	"github.com/fernandovalenzuela/api-home-pay/internal/middleware"
	"github.com/fernandovalenzuela/api-home-pay/internal/models"
	"github.com/fernandovalenzuela/api-home-pay/internal/services"
	"github.com/fernandovalenzuela/api-home-pay/internal/utils"
	"github.com/gin-gonic/gin"
)

type PeriodHandler struct {
	service services.PeriodService
}

func NewPeriodHandler(service services.PeriodService) *PeriodHandler {
	return &PeriodHandler{service: service}
}

// @Summary Create period
// @Description Create a new period (month/year)
// @Tags periods
// @Accept json
// @Produce json
// @Param period body models.CreatePeriodRequest true "Period data"
// @Success 201 {object} models.Period
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /api/periods [post]
func (h *PeriodHandler) Create(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.ErrorResponseClient(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req models.CreatePeriodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponseClient(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	period := models.Period{
		MonthNumber: req.MonthNumber,
		YearNumber:  req.YearNumber,
	}

	if err := h.service.Create(userID, &period); err != nil {
		utils.ErrorResponseClient(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, period)
}

// @Summary Get period by ID
// @Description Get a specific period by ID
// @Tags periods
// @Accept json
// @Produce json
// @Param id path int true "Period ID"
// @Success 200 {object} models.Period
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /api/periods/{id} [get]
func (h *PeriodHandler) GetByID(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.ErrorResponseClient(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.ErrorResponseClient(c, http.StatusBadRequest, "Invalid period ID")
		return
	}

	period, err := h.service.GetByID(userID, id)
	if err != nil {
		utils.ErrorResponseClient(c, http.StatusNotFound, err.Error())
		return
	}

	utils.SuccessResponse(c, period)
}

// @Summary List all periods
// @Description Get all periods for the authenticated user
// @Tags periods
// @Accept json
// @Produce json
// @Success 200 {array} models.Period
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /api/periods [get]
func (h *PeriodHandler) GetAll(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.ErrorResponseClient(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	periods, err := h.service.GetAll(userID)
	if err != nil {
		utils.ErrorResponseClient(c, http.StatusInternalServerError, "Failed to retrieve periods")
		return
	}

	utils.SuccessResponse(c, periods)
}

// @Summary Update period
// @Description Update an existing period
// @Tags periods
// @Accept json
// @Produce json
// @Param id path int true "Period ID"
// @Param period body models.UpdatePeriodRequest true "Period data"
// @Success 200 {object} models.Period
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /api/periods/{id} [put]
func (h *PeriodHandler) Update(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.ErrorResponseClient(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.ErrorResponseClient(c, http.StatusBadRequest, "Invalid period ID")
		return
	}

	var req models.UpdatePeriodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponseClient(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	period := models.Period{
		ID:          id,
		MonthNumber: req.MonthNumber,
		YearNumber:  req.YearNumber,
	}

	if err := h.service.Update(userID, &period); err != nil {
		utils.ErrorResponseClient(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, period)
}

// @Summary Delete period
// @Description Delete a period by ID
// @Tags periods
// @Accept json
// @Produce json
// @Param id path int true "Period ID"
// @Success 200 {object} utils.ErrorResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /api/periods/{id} [delete]
func (h *PeriodHandler) Delete(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.ErrorResponseClient(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.ErrorResponseClient(c, http.StatusBadRequest, "Invalid period ID")
		return
	}

	if err := h.service.Delete(userID, id); err != nil {
		utils.ErrorResponseClient(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "Period deleted successfully"})
}
