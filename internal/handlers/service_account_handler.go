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

type ServiceAccountHandler struct {
	service services.ServiceAccountService
}

func NewServiceAccountHandler(service services.ServiceAccountService) *ServiceAccountHandler {
	return &ServiceAccountHandler{service: service}
}

// @Summary Create service account
// @Description Create a new service account (e.g., "Electricity - Home")
// @Tags service-accounts
// @Accept json
// @Produce json
// @Param service_account body models.CreateServiceAccountRequest true "Service account data"
// @Success 201 {object} models.ServiceAccount
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /api/service-accounts [post]
func (h *ServiceAccountHandler) Create(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.ErrorResponseClient(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req models.CreateServiceAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponseClient(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	account := models.ServiceAccount{
		CompanyID:         req.CompanyID,
		AccountIdentifier: req.AccountIdentifier,
		Alias:             req.Alias,
	}

	if err := h.service.Create(userID, &account); err != nil {
		utils.ErrorResponseClient(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, account)
}

// @Summary Get service account by ID
// @Description Get a specific service account by ID
// @Tags service-accounts
// @Accept json
// @Produce json
// @Param id path int true "Service Account ID"
// @Success 200 {object} models.ServiceAccount
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /api/service-accounts/{id} [get]
func (h *ServiceAccountHandler) GetByID(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.ErrorResponseClient(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.ErrorResponseClient(c, http.StatusBadRequest, "Invalid service account ID")
		return
	}

	account, err := h.service.GetByID(userID, id)
	if err != nil {
		utils.ErrorResponseClient(c, http.StatusNotFound, err.Error())
		return
	}

	utils.SuccessResponse(c, account)
}

// @Summary List all service accounts
// @Description Get all service accounts for the authenticated user, optionally filtered by company
// @Tags service-accounts
// @Accept json
// @Produce json
// @Param company_id query int false "Filter by company ID"
// @Success 200 {array} models.ServiceAccount
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /api/service-accounts [get]
func (h *ServiceAccountHandler) GetAll(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.ErrorResponseClient(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var companyID *int
	if companyIDStr := c.Query("company_id"); companyIDStr != "" {
		id, err := strconv.Atoi(companyIDStr)
		if err == nil && id > 0 {
			companyID = &id
		}
	}

	accounts, err := h.service.GetAll(userID, companyID)
	if err != nil {
		utils.ErrorResponseClient(c, http.StatusInternalServerError, "Failed to retrieve service accounts")
		return
	}

	utils.SuccessResponse(c, accounts)
}

// @Summary Update service account
// @Description Update an existing service account
// @Tags service-accounts
// @Accept json
// @Produce json
// @Param id path int true "Service Account ID"
// @Param service_account body models.UpdateServiceAccountRequest true "Service account data"
// @Success 200 {object} models.ServiceAccount
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /api/service-accounts/{id} [put]
func (h *ServiceAccountHandler) Update(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.ErrorResponseClient(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.ErrorResponseClient(c, http.StatusBadRequest, "Invalid service account ID")
		return
	}

	var req models.UpdateServiceAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponseClient(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	account := models.ServiceAccount{
		ID:                id,
		CompanyID:         req.CompanyID,
		AccountIdentifier: req.AccountIdentifier,
		Alias:             req.Alias,
	}

	if err := h.service.Update(userID, &account); err != nil {
		utils.ErrorResponseClient(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, account)
}

// @Summary Delete service account
// @Description Delete a service account by ID
// @Tags service-accounts
// @Accept json
// @Produce json
// @Param id path int true "Service Account ID"
// @Success 200 {object} utils.ErrorResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /api/service-accounts/{id} [delete]
func (h *ServiceAccountHandler) Delete(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.ErrorResponseClient(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.ErrorResponseClient(c, http.StatusBadRequest, "Invalid service account ID")
		return
	}

	if err := h.service.Delete(userID, id); err != nil {
		utils.ErrorResponseClient(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "Service account deleted successfully"})
}
