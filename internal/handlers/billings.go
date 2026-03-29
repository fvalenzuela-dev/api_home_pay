package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/homepay/api/internal/middleware"
	"github.com/homepay/api/internal/models"
	"github.com/homepay/api/internal/service"
)

type BillingHandler struct {
	svc service.BillingService
}

func NewBillingHandler(svc service.BillingService) *BillingHandler {
	return &BillingHandler{svc: svc}
}

// List godoc
// @Summary     Listar facturas
// @Description Retorna todas las facturas de una cuenta, ordenadas por año/mes desc
// @Tags        billings
// @Security    BearerAuth
// @Produce     json
// @Param       accountID  path      string  true  "Account ID"
// @Success     200        {object}  map[string][]models.AccountBilling
// @Failure     401        {object}  map[string]string
// @Failure     500        {object}  map[string]string
// @Router      /accounts/{accountID}/billings [get]
func (h *BillingHandler) List(w http.ResponseWriter, r *http.Request) {
	authUserID := middleware.GetAuthUserID(r)
	accountID := chi.URLParam(r, "accountID")
	billings, err := h.svc.GetAllByAccount(r.Context(), accountID, authUserID)
	if err != nil {
		writeInternalError(w, r, err)
		return
	}
	if billings == nil {
		billings = []models.AccountBilling{}
	}
	writeJSON(w, http.StatusOK, billings)
}

// Create godoc
// @Summary     Registrar factura
// @Description Registra la factura del mes para una cuenta. Si auto_accumulate=true y hay una factura impaga, se crea un carry-over automáticamente.
// @Tags        billings
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       accountID  path      string                      true  "Account ID"
// @Param       body       body      models.CreateBillingRequest  true  "Datos de la factura"
// @Success     201        {object}  map[string]models.AccountBilling
// @Failure     400        {object}  map[string]string
// @Failure     401        {object}  map[string]string
// @Failure     404        {object}  map[string]string
// @Router      /accounts/{accountID}/billings [post]
func (h *BillingHandler) Create(w http.ResponseWriter, r *http.Request) {
	authUserID := middleware.GetAuthUserID(r)
	accountID := chi.URLParam(r, "accountID")
	var req models.CreateBillingRequest
	if err := decode(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	billing, err := h.svc.Create(r.Context(), accountID, authUserID, &req)
	if err != nil {
		if err.Error() == "not found" {
			writeError(w, http.StatusNotFound, "no encontrado")
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, billing)
}

// Update godoc
// @Summary     Actualizar factura
// @Description Actualiza monto pagado. Si amount_paid >= amount_billed se marca automáticamente como pagada.
// @Tags        billings
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       accountID  path      string                      true  "Account ID"
// @Param       id         path      string                      true  "Billing ID"
// @Param       body       body      models.UpdateBillingRequest  true  "Campos a actualizar"
// @Success     200        {object}  map[string]models.AccountBilling
// @Failure     400        {object}  map[string]string
// @Failure     401        {object}  map[string]string
// @Failure     404        {object}  map[string]string
// @Router      /accounts/{accountID}/billings/{id} [put]
func (h *BillingHandler) Update(w http.ResponseWriter, r *http.Request) {
	authUserID := middleware.GetAuthUserID(r)
	id := chi.URLParam(r, "id")
	var req models.UpdateBillingRequest
	if err := decode(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	billing, err := h.svc.Update(r.Context(), id, authUserID, &req)
	if err != nil {
		if err.Error() == "not found" {
			writeError(w, http.StatusNotFound, "no encontrado")
			return
		}
		writeInternalError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, billing)
}
