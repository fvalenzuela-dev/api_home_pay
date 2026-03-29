package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/homepay/api/internal/middleware"
	"github.com/homepay/api/internal/models"
	"github.com/homepay/api/internal/service"
)

type CompanyHandler struct {
	svc service.CompanyService
}

func NewCompanyHandler(svc service.CompanyService) *CompanyHandler {
	return &CompanyHandler{svc: svc}
}

// List godoc
// @Summary     Listar empresas
// @Description Retorna todas las empresas activas del usuario autenticado
// @Tags        companies
// @Security    BearerAuth
// @Produce     json
// @Success     200  {object}  map[string][]models.Company
// @Failure     401  {object}  map[string]string
// @Failure     500  {object}  map[string]string
// @Router      /companies [get]
func (h *CompanyHandler) List(w http.ResponseWriter, r *http.Request) {
	authUserID := middleware.GetAuthUserID(r)
	companies, err := h.svc.GetAll(r.Context(), authUserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "error interno")
		return
	}
	if companies == nil {
		companies = []models.Company{}
	}
	writeJSON(w, http.StatusOK, companies)
}

// Create godoc
// @Summary     Crear empresa
// @Description Crea una nueva empresa para el usuario autenticado
// @Tags        companies
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       body  body      models.CreateCompanyRequest  true  "Datos de la empresa"
// @Success     201   {object}  map[string]models.Company
// @Failure     400   {object}  map[string]string
// @Failure     401   {object}  map[string]string
// @Router      /companies [post]
func (h *CompanyHandler) Create(w http.ResponseWriter, r *http.Request) {
	authUserID := middleware.GetAuthUserID(r)
	var req models.CreateCompanyRequest
	if err := decode(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	company, err := h.svc.Create(r.Context(), authUserID, &req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, company)
}

// Update godoc
// @Summary     Editar empresa
// @Description Actualiza nombre o categoría de una empresa
// @Tags        companies
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id    path      string                       true  "Company ID"
// @Param       body  body      models.UpdateCompanyRequest  true  "Campos a actualizar"
// @Success     200   {object}  map[string]models.Company
// @Failure     400   {object}  map[string]string
// @Failure     401   {object}  map[string]string
// @Failure     404   {object}  map[string]string
// @Router      /companies/{id} [put]
func (h *CompanyHandler) Update(w http.ResponseWriter, r *http.Request) {
	authUserID := middleware.GetAuthUserID(r)
	id := chi.URLParam(r, "id")
	var req models.UpdateCompanyRequest
	if err := decode(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	company, err := h.svc.Update(r.Context(), id, authUserID, &req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "error interno")
		return
	}
	if company == nil {
		writeError(w, http.StatusNotFound, "no encontrado")
		return
	}
	writeJSON(w, http.StatusOK, company)
}

// Delete godoc
// @Summary     Eliminar empresa
// @Description Soft delete de la empresa. Propaga a sus cuentas y facturas activas.
// @Tags        companies
// @Security    BearerAuth
// @Produce     json
// @Param       id   path  string  true  "Company ID"
// @Success     204
// @Failure     401  {object}  map[string]string
// @Failure     404  {object}  map[string]string
// @Failure     500  {object}  map[string]string
// @Router      /companies/{id} [delete]
func (h *CompanyHandler) Delete(w http.ResponseWriter, r *http.Request) {
	authUserID := middleware.GetAuthUserID(r)
	id := chi.URLParam(r, "id")
	if err := h.svc.Delete(r.Context(), id, authUserID); err != nil {
		if err.Error() == "not found" {
			writeError(w, http.StatusNotFound, "no encontrado")
			return
		}
		writeError(w, http.StatusInternalServerError, "error interno")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
