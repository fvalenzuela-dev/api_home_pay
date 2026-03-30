package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/homepay/api/internal/middleware"
	"github.com/homepay/api/internal/models"
	"github.com/homepay/api/internal/repository"
	"github.com/jackc/pgx/v5"
)

type CategoryHandler struct {
	repo repository.CategoryRepository
}

func NewCategoryHandler(repo repository.CategoryRepository) *CategoryHandler {
	return &CategoryHandler{repo: repo}
}

// List godoc
// @Summary     Listar categorías
// @Description Retorna las categorías activas del usuario autenticado
// @Tags        categories
// @Security    BearerAuth
// @Produce     json
// @Success     200  {array}   models.Category
// @Failure     401  {object}  map[string]string
// @Failure     500  {object}  map[string]string
// @Router      /categories [get]
func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	authUserID := middleware.GetAuthUserID(r)
	cats, err := h.repo.GetAll(r.Context(), authUserID)
	if err != nil {
		writeInternalError(w, r, err)
		return
	}
	if cats == nil {
		cats = []models.Category{}
	}
	writeJSON(w, http.StatusOK, cats)
}

// GetOne godoc
// @Summary     Obtener categoría
// @Description Retorna una categoría por ID
// @Tags        categories
// @Security    BearerAuth
// @Produce     json
// @Param       id   path      int  true  "Category ID"
// @Success     200  {object}  models.Category
// @Failure     400  {object}  map[string]string
// @Failure     401  {object}  map[string]string
// @Failure     404  {object}  map[string]string
// @Failure     500  {object}  map[string]string
// @Router      /categories/{id} [get]
func (h *CategoryHandler) GetOne(w http.ResponseWriter, r *http.Request) {
	authUserID := middleware.GetAuthUserID(r)
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	cat, err := h.repo.GetByID(r.Context(), id, authUserID)
	if err != nil {
		writeInternalError(w, r, err)
		return
	}
	if cat == nil {
		writeError(w, http.StatusNotFound, "no encontrado")
		return
	}
	writeJSON(w, http.StatusOK, cat)
}

// Create godoc
// @Summary     Crear categoría
// @Description Crea una nueva categoría para el usuario autenticado
// @Tags        categories
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       body  body      models.CreateCategoryRequest  true  "Datos de la categoría"
// @Success     201   {object}  models.Category
// @Failure     400   {object}  map[string]string
// @Failure     401   {object}  map[string]string
// @Failure     500   {object}  map[string]string
// @Router      /categories [post]
func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	authUserID := middleware.GetAuthUserID(r)
	var req models.CreateCategoryRequest
	if err := decode(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	cat, err := h.repo.Create(r.Context(), authUserID, &req)
	if err != nil {
		writeInternalError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, cat)
}

// Update godoc
// @Summary     Editar categoría
// @Description Actualiza el nombre de una categoría
// @Tags        categories
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id    path      int                           true  "Category ID"
// @Param       body  body      models.UpdateCategoryRequest  true  "Campos a actualizar"
// @Success     200   {object}  models.Category
// @Failure     400   {object}  map[string]string
// @Failure     401   {object}  map[string]string
// @Failure     404   {object}  map[string]string
// @Failure     500   {object}  map[string]string
// @Router      /categories/{id} [put]
func (h *CategoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	authUserID := middleware.GetAuthUserID(r)
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var req models.UpdateCategoryRequest
	if err := decode(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	cat, err := h.repo.Update(r.Context(), id, authUserID, &req)
	if err != nil {
		writeInternalError(w, r, err)
		return
	}
	if cat == nil {
		writeError(w, http.StatusNotFound, "no encontrado")
		return
	}
	writeJSON(w, http.StatusOK, cat)
}

// Delete godoc
// @Summary     Eliminar categoría
// @Description Soft delete de la categoría
// @Tags        categories
// @Security    BearerAuth
// @Produce     json
// @Param       id   path  int  true  "Category ID"
// @Success     204
// @Failure     400  {object}  map[string]string
// @Failure     401  {object}  map[string]string
// @Failure     404  {object}  map[string]string
// @Failure     500  {object}  map[string]string
// @Router      /categories/{id} [delete]
func (h *CategoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	authUserID := middleware.GetAuthUserID(r)
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.repo.Delete(r.Context(), id, authUserID); err != nil {
		if err == pgx.ErrNoRows {
			writeError(w, http.StatusNotFound, "no encontrado")
			return
		}
		writeInternalError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
