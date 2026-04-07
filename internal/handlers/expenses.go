package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/homepay/api/internal/middleware"
	"github.com/homepay/api/internal/models"
	"github.com/homepay/api/internal/service"
)

type ExpenseHandler struct {
	svc service.ExpenseService
}

func NewExpenseHandler(svc service.ExpenseService) *ExpenseHandler {
	return &ExpenseHandler{svc: svc}
}

// List godoc
// @Summary     Listar gastos
// @Description Retorna gastos del usuario (paginado). Soporta filtros opcionales por mes/año y empresa.
// @Tags        expenses
// @Security    BearerAuth
// @Produce     json
// @Param       month      query     int     false  "Mes (1-12)"
// @Param       year       query     int     false  "Año (ej: 2026)"
// @Param       company_id query     string  false  "Filtrar por empresa"
// @Param       page       query     int     false  "Página (default: 1)"
// @Param       limit      query     int     false  "Resultados por página (default: 20, max: 100)"
// @Success     200  {object}  map[string]interface{}
// @Failure     401  {object}  map[string]string
// @Failure     500  {object}  map[string]string
// @Router      /expenses [get]
func (h *ExpenseHandler) List(w http.ResponseWriter, r *http.Request) {
	authUserID := middleware.GetAuthUserID(r)

	var filters models.ExpenseFilters
	if m := r.URL.Query().Get("month"); m != "" {
		if yr := r.URL.Query().Get("year"); yr != "" {
			month, err1 := strconv.Atoi(m)
			year, err2 := strconv.Atoi(yr)
			if err1 == nil && err2 == nil {
				filters.Month = &month
				filters.Year = &year
			}
		}
	}
	if cid := r.URL.Query().Get("company_id"); cid != "" {
		filters.CompanyID = &cid
	}

	p := parsePagination(r)
	expenses, total, err := h.svc.GetAll(r.Context(), authUserID, filters, p)
	if err != nil {
		writeInternalError(w, r, err)
		return
	}
	if expenses == nil {
		expenses = []models.Expense{}
	}
	writePaginatedJSON(w, expenses, models.NewPaginationMeta(p.Page, p.Limit, total))
}

// GetOne godoc
// @Summary     Obtener gasto
// @Description Retorna un gasto por ID
// @Tags        expenses
// @Security    BearerAuth
// @Produce     json
// @Param       id   path      string  true  "Expense ID"
// @Success     200  {object}  models.Expense
// @Failure     401  {object}  map[string]string
// @Failure     404  {object}  map[string]string
// @Failure     500  {object}  map[string]string
// @Router      /expenses/{id} [get]
func (h *ExpenseHandler) GetOne(w http.ResponseWriter, r *http.Request) {
	authUserID := middleware.GetAuthUserID(r)
	id := chi.URLParam(r, "id")
	expense, err := h.svc.GetByID(r.Context(), id, authUserID)
	if err != nil {
		writeInternalError(w, r, err)
		return
	}
	if expense == nil {
		writeError(w, http.StatusNotFound, "no encontrado")
		return
	}
	writeJSON(w, http.StatusOK, expense)
}

// Create godoc
// @Summary     Registrar gasto
// @Description Registra un nuevo gasto variable para el usuario autenticado
// @Tags        expenses
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       body  body      models.CreateExpenseRequest  true  "Datos del gasto"
// @Success     201   {object}  map[string]models.Expense
// @Failure     400   {object}  map[string]string
// @Failure     401   {object}  map[string]string
// @Router      /expenses [post]
func (h *ExpenseHandler) Create(w http.ResponseWriter, r *http.Request) {
	authUserID := middleware.GetAuthUserID(r)
	var req models.CreateExpenseRequest
	if err := decode(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	expense, err := h.svc.Create(r.Context(), authUserID, &req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, expense)
}

// Update godoc
// @Summary     Editar gasto
// @Description Actualiza descripción, monto, categoría o fecha de un gasto
// @Tags        expenses
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id    path      string                      true  "Expense ID"
// @Param       body  body      models.UpdateExpenseRequest  true  "Campos a actualizar"
// @Success     200   {object}  map[string]models.Expense
// @Failure     400   {object}  map[string]string
// @Failure     401   {object}  map[string]string
// @Failure     404   {object}  map[string]string
// @Router      /expenses/{id} [put]
func (h *ExpenseHandler) Update(w http.ResponseWriter, r *http.Request) {
	authUserID := middleware.GetAuthUserID(r)
	id := chi.URLParam(r, "id")
	var req models.UpdateExpenseRequest
	if err := decode(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	expense, err := h.svc.Update(r.Context(), id, authUserID, &req)
	if err != nil {
		if err.Error() == "not found" {
			writeError(w, http.StatusNotFound, "no encontrado")
			return
		}
		writeInternalError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, expense)
}

// Delete godoc
// @Summary     Eliminar gasto
// @Description Soft delete de un gasto
// @Tags        expenses
// @Security    BearerAuth
// @Produce     json
// @Param       id   path  string  true  "Expense ID"
// @Success     204
// @Failure     401  {object}  map[string]string
// @Failure     404  {object}  map[string]string
// @Failure     500  {object}  map[string]string
// @Router      /expenses/{id} [delete]
func (h *ExpenseHandler) Delete(w http.ResponseWriter, r *http.Request) {
	authUserID := middleware.GetAuthUserID(r)
	id := chi.URLParam(r, "id")
	if err := h.svc.Delete(r.Context(), id, authUserID); err != nil {
		if err.Error() == "not found" {
			writeError(w, http.StatusNotFound, "no encontrado")
			return
		}
		writeInternalError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
