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

type CategoryHandler struct {
	service services.CategoryService
}

func NewCategoryHandler(service services.CategoryService) *CategoryHandler {
	return &CategoryHandler{service: service}
}

// @Summary Create category
// @Description Create a new expense category
// @Tags categories
// @Accept json
// @Produce json
// @Param category body models.CreateCategoryRequest true "Category data"
// @Success 201 {object} models.Category
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /api/categories [post]
func (h *CategoryHandler) Create(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.ErrorResponseClient(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req models.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponseClient(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	category := models.Category{
		Name: req.Name,
	}

	if err := h.service.Create(userID, &category); err != nil {
		utils.ErrorResponseClient(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, category)
}

// @Summary Get category by ID
// @Description Get a specific category by ID
// @Tags categories
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Success 200 {object} models.Category
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /api/categories/{id} [get]
func (h *CategoryHandler) GetByID(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.ErrorResponseClient(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.ErrorResponseClient(c, http.StatusBadRequest, "Invalid category ID")
		return
	}

	category, err := h.service.GetByID(userID, id)
	if err != nil {
		utils.ErrorResponseClient(c, http.StatusNotFound, err.Error())
		return
	}

	utils.SuccessResponse(c, category)
}

// @Summary List all categories
// @Description Get all categories for the authenticated user
// @Tags categories
// @Accept json
// @Produce json
// @Success 200 {array} models.Category
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /api/categories [get]
func (h *CategoryHandler) GetAll(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.ErrorResponseClient(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	categories, err := h.service.GetAll(userID)
	if err != nil {
		utils.ErrorResponseClient(c, http.StatusInternalServerError, "Failed to retrieve categories")
		return
	}

	utils.SuccessResponse(c, categories)
}

// @Summary Update category
// @Description Update an existing category
// @Tags categories
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Param category body models.UpdateCategoryRequest true "Category data"
// @Success 200 {object} models.Category
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /api/categories/{id} [put]
func (h *CategoryHandler) Update(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.ErrorResponseClient(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.ErrorResponseClient(c, http.StatusBadRequest, "Invalid category ID")
		return
	}

	var req models.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponseClient(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	category := models.Category{
		ID:   id,
		Name: req.Name,
	}

	if err := h.service.Update(userID, &category); err != nil {
		utils.ErrorResponseClient(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, category)
}

// @Summary Delete category
// @Description Delete a category by ID
// @Tags categories
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Success 200 {object} utils.ErrorResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /api/categories/{id} [delete]
func (h *CategoryHandler) Delete(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.ErrorResponseClient(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.ErrorResponseClient(c, http.StatusBadRequest, "Invalid category ID")
		return
	}

	if err := h.service.Delete(userID, id); err != nil {
		utils.ErrorResponseClient(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "Category deleted successfully"})
}
