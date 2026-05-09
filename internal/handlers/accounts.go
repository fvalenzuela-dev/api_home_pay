package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/homepay/api/internal/middleware"
	"github.com/homepay/api/internal/models"
	"github.com/homepay/api/internal/service"
)

type AccountHandler struct {
	svc service.AccountService
}

func NewAccountHandler(svc service.AccountService) *AccountHandler {
	return &AccountHandler{svc: svc}
}

// List godoc
// @Summary     Listar cuentas
// @Description Retorna todas las cuentas del usuario (paginado, con filtros opcionales)
// @Tags        accounts
// @Security    BearerAuth
// @Produce     json
// @Param       company_id  query     string  false  "Filtrar por empresa"
// @Param       sort        query     string  false  "Campo de orden (created_at, name, billing_day, company_name)"
// @Param       order       query     string  false  "Dirección (asc, desc)"
// @Param       page        query     int     false  "Página (default: 1)"
// @Param       limit       query     int     false  "Resultados por página (default: 20, max: 100)"
// @Success     200        {object}  map[string]interface{}
// @Failure     401        {object}  map[string]string
// @Failure     500        {object}  map[string]string
// @Router      /accounts [get]
func (h *AccountHandler) List(w http.ResponseWriter, r *http.Request) {
	authUserID := middleware.GetAuthUserID(r)
	companyID := r.URL.Query().Get("company_id")
	var companyIDPtr *string
	if companyID != "" {
		if _, err := uuid.Parse(companyID); err != nil {
			writeError(w, http.StatusBadRequest, "invalid company_id format")
			return
		}
		companyIDPtr = &companyID
	}
	sort := r.URL.Query().Get("sort")
	order := r.URL.Query().Get("order")
	p := parsePagination(r)
	accounts, total, err := h.svc.GetAll(r.Context(), authUserID, companyIDPtr, sort, order, p)
	if err != nil {
		writeInternalError(w, r, err)
		return
	}
	if accounts == nil {
		accounts = []models.Account{}
	}
	writePaginatedJSON(w, accounts, models.NewPaginationMeta(p.Page, p.Limit, total))
}

// GetOne godoc
// @Summary     Obtener cuenta
// @Description Retorna una cuenta por ID
// @Tags        accounts
// @Security    BearerAuth
// @Produce     json
// @Param       id         path      string  true  "Account ID"
// @Success     200        {object}  models.Account
// @Failure     401        {object}  map[string]string
// @Failure     404        {object}  map[string]string
// @Failure     500        {object}  map[string]string
// @Router      /accounts/{id} [get]
func (h *AccountHandler) GetOne(w http.ResponseWriter, r *http.Request) {
	authUserID := middleware.GetAuthUserID(r)
	id := chi.URLParam(r, "id")
	account, err := h.svc.GetByID(r.Context(), id, authUserID)
	if err != nil {
		writeInternalError(w, r, err)
		return
	}
	if account == nil {
		writeError(w, http.StatusNotFound, "no encontrado")
		return
	}
	writeJSON(w, http.StatusOK, account)
}

// Create godoc
// @Summary     Crear cuenta
// @Description Crea una nueva cuenta para una empresa
// @Tags        accounts
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       body       body      models.CreateAccountRequest  true  "Datos de la cuenta (company_id requerido)"
// @Success     201        {object}  map[string]models.Account
// @Failure     400        {object}  map[string]string
// @Failure     401        {object}  map[string]string
// @Router      /accounts [post]
func (h *AccountHandler) Create(w http.ResponseWriter, r *http.Request) {
	authUserID := middleware.GetAuthUserID(r)
	var req models.CreateAccountRequest
	if err := decode(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	account, err := h.svc.Create(r.Context(), authUserID, &req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, account)
}

// Update godoc
// @Summary     Editar cuenta
// @Description Actualiza nombre, billing_day o auto_accumulate de una cuenta
// @Tags        accounts
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id         path      string                      true  "Account ID"
// @Param       body       body      models.UpdateAccountRequest  true  "Campos a actualizar"
// @Success     200        {object}  map[string]models.Account
// @Failure     400        {object}  map[string]string
// @Failure     401        {object}  map[string]string
// @Failure     404        {object}  map[string]string
// @Router      /accounts/{id} [put]
func (h *AccountHandler) Update(w http.ResponseWriter, r *http.Request) {
	authUserID := middleware.GetAuthUserID(r)
	id := chi.URLParam(r, "id")
	var req models.UpdateAccountRequest
	if err := decode(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	account, err := h.svc.Update(r.Context(), id, authUserID, &req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if account == nil {
		writeError(w, http.StatusNotFound, "no encontrado")
		return
	}
	writeJSON(w, http.StatusOK, account)
}

// Delete godoc
// @Summary     Eliminar cuenta
// @Description Soft delete de la cuenta. Propaga a sus facturas activas.
// @Tags        accounts
// @Security    BearerAuth
// @Produce     json
// @Param       id         path  string  true  "Account ID"
// @Success     204
// @Failure     401  {object}  map[string]string
// @Failure     404  {object}  map[string]string
// @Failure     500  {object}  map[string]string
// @Router      /accounts/{id} [delete]
func (h *AccountHandler) Delete(w http.ResponseWriter, r *http.Request) {
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
