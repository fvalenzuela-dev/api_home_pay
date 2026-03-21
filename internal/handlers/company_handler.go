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

type CompanyHandler struct {
	service services.CompanyService
}

func NewCompanyHandler(service services.CompanyService) *CompanyHandler {
	return &CompanyHandler{service: service}
}

// @Summary Create company
// @Description Create a new company (service provider)
// @Tags companies
// @Accept json
// @Produce json
// @Param company body models.CreateCompanyRequest true "Company data"
// @Success 201 {object} models.Company
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /api/companies [post]
func (h *CompanyHandler) Create(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.ErrorResponseClient(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req models.CreateCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponseClient(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	company := models.Company{
		Name:       req.Name,
		WebsiteURL: req.WebsiteURL,
	}

	if err := h.service.Create(userID, &company); err != nil {
		utils.ErrorResponseClient(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, company)
}

// @Summary Get company by ID
// @Description Get a specific company by ID
// @Tags companies
// @Accept json
// @Produce json
// @Param id path int true "Company ID"
// @Success 200 {object} models.Company
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /api/companies/{id} [get]
func (h *CompanyHandler) GetByID(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.ErrorResponseClient(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.ErrorResponseClient(c, http.StatusBadRequest, "Invalid company ID")
		return
	}

	company, err := h.service.GetByID(userID, id)
	if err != nil {
		utils.ErrorResponseClient(c, http.StatusNotFound, err.Error())
		return
	}

	utils.SuccessResponse(c, company)
}

// @Summary List all companies
// @Description Get all companies for the authenticated user
// @Tags companies
// @Accept json
// @Produce json
// @Success 200 {array} models.Company
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /api/companies [get]
func (h *CompanyHandler) GetAll(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.ErrorResponseClient(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	companies, err := h.service.GetAll(userID)
	if err != nil {
		utils.ErrorResponseClient(c, http.StatusInternalServerError, "Failed to retrieve companies")
		return
	}

	utils.SuccessResponse(c, companies)
}

// @Summary Update company
// @Description Update an existing company
// @Tags companies
// @Accept json
// @Produce json
// @Param id path int true "Company ID"
// @Param company body models.UpdateCompanyRequest true "Company data"
// @Success 200 {object} models.Company
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /api/companies/{id} [put]
func (h *CompanyHandler) Update(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.ErrorResponseClient(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.ErrorResponseClient(c, http.StatusBadRequest, "Invalid company ID")
		return
	}

	var req models.UpdateCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponseClient(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	company := models.Company{
		ID:         id,
		Name:       req.Name,
		WebsiteURL: req.WebsiteURL,
	}

	if err := h.service.Update(userID, &company); err != nil {
		utils.ErrorResponseClient(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, company)
}

// @Summary Delete company
// @Description Delete a company by ID
// @Tags companies
// @Accept json
// @Produce json
// @Param id path int true "Company ID"
// @Success 200 {object} utils.ErrorResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /api/companies/{id} [delete]
func (h *CompanyHandler) Delete(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.ErrorResponseClient(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.ErrorResponseClient(c, http.StatusBadRequest, "Invalid company ID")
		return
	}

	if err := h.service.Delete(userID, id); err != nil {
		utils.ErrorResponseClient(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "Company deleted successfully"})
}
