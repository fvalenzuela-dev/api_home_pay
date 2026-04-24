package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/homepay/api/internal/middleware"
	"github.com/homepay/api/internal/models"
	"github.com/homepay/api/internal/service"
)

type AccountGroupHandler struct {
	svc service.AccountGroupService
}

func NewAccountGroupHandler(svc service.AccountGroupService) *AccountGroupHandler {
	return &AccountGroupHandler{svc: svc}
}

// List godoc
// @Summary     Listar grupos de cuentas
// @Description Retorna todos los grupos de cuentas del usuario autenticado (paginado)
// @Tags        account-groups
// @Security    BearerAuth
// @Produce     json
// @Param       page   query     int  false  "Página (default: 1)"
// @Param       limit  query     int  false  "Resultados por página (default: 20, max: 100)"
// @Success     200  {object}  map[string]interface{}
// @Failure     401  {object}  map[string]string
// @Failure     500  {object}  map[string]string
// @Router      /account-groups [get]
func (h *AccountGroupHandler) List(w http.ResponseWriter, r *http.Request) {
	authUserID := middleware.GetAuthUserID(r)
	p := parsePagination(r)
	groups, total, err := h.svc.GetAll(r.Context(), authUserID, p)
	if err != nil {
		writeInternalError(w, r, err)
		return
	}
	if groups == nil {
		groups = []models.AccountGroup{}
	}
	writePaginatedJSON(w, groups, models.NewPaginationMeta(p.Page, p.Limit, total))
}

// GetOne godoc
// @Summary     Obtener grupo de cuentas
// @Description Retorna un grupo de cuentas por ID
// @Tags        account-groups
// @Security    BearerAuth
// @Produce     json
// @Param       id   path      string  true  "Account Group ID"
// @Success     200  {object}  models.AccountGroup
// @Failure     401  {object}  map[string]string
// @Failure     404  {object}  map[string]string
// @Failure     500  {object}  map[string]string
// @Router      /account-groups/{id} [get]
func (h *AccountGroupHandler) GetOne(w http.ResponseWriter, r *http.Request) {
	authUserID := middleware.GetAuthUserID(r)
	id := chi.URLParam(r, "id")
	group, err := h.svc.GetByID(r.Context(), id, authUserID)
	if err != nil {
		writeInternalError(w, r, err)
		return
	}
	if group == nil {
		writeError(w, http.StatusNotFound, "no encontrado")
		return
	}
	writeJSON(w, http.StatusOK, group)
}

// Create godoc
// @Summary     Crear grupo de cuentas
// @Description Crea un nuevo grupo de cuentas para el usuario autenticado
// @Tags        account-groups
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       body  body      models.CreateAccountGroupRequest  true  "Datos del grupo"
// @Success     201   {object}  models.AccountGroup
// @Failure     400   {object}  map[string]string
// @Failure     401   {object}  map[string]string
// @Failure     409   {object}  map[string]string
// @Failure     500   {object}  map[string]string
// @Router      /account-groups [post]
func (h *AccountGroupHandler) Create(w http.ResponseWriter, r *http.Request) {
	authUserID := middleware.GetAuthUserID(r)
	var req models.CreateAccountGroupRequest
	if err := decode(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	group, err := h.svc.Create(r.Context(), authUserID, &req)
	if err != nil {
		if err.Error() == "ya existe un grupo con ese nombre" {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, group)
}

// Update godoc
// @Summary     Editar grupo de cuentas
// @Description Actualiza el nombre de un grupo de cuentas
// @Tags        account-groups
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id    path      string                            true  "Account Group ID"
// @Param       body  body      models.UpdateAccountGroupRequest  true  "Campos a actualizar"
// @Success     200   {object}  models.AccountGroup
// @Failure     400   {object}  map[string]string
// @Failure     401   {object}  map[string]string
// @Failure     404   {object}  map[string]string
// @Failure     409   {object}  map[string]string
// @Failure     500   {object}  map[string]string
// @Router      /account-groups/{id} [put]
func (h *AccountGroupHandler) Update(w http.ResponseWriter, r *http.Request) {
	authUserID := middleware.GetAuthUserID(r)
	id := chi.URLParam(r, "id")
	var req models.UpdateAccountGroupRequest
	if err := decode(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	group, err := h.svc.Update(r.Context(), id, authUserID, &req)
	if err != nil {
		if err.Error() == "ya existe un grupo con ese nombre" {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if group == nil {
		writeError(w, http.StatusNotFound, "no encontrado")
		return
	}
	writeJSON(w, http.StatusOK, group)
}

// Delete godoc
// @Summary     Eliminar grupo de cuentas
// @Description Soft delete de un grupo de cuentas
// @Tags        account-groups
// @Security    BearerAuth
// @Produce     json
// @Param       id   path  string  true  "Account Group ID"
// @Success     204
// @Failure     401  {object}  map[string]string
// @Failure     404  {object}  map[string]string
// @Failure     500  {object}  map[string]string
// @Router      /account-groups/{id} [delete]
func (h *AccountGroupHandler) Delete(w http.ResponseWriter, r *http.Request) {
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
