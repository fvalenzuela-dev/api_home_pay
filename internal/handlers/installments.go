package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/homepay/api/internal/middleware"
	"github.com/homepay/api/internal/models"
	"github.com/homepay/api/internal/service"
)

type InstallmentHandler struct {
	svc service.InstallmentService
}

func NewInstallmentHandler(svc service.InstallmentService) *InstallmentHandler {
	return &InstallmentHandler{svc: svc}
}

// GetOne godoc
// @Summary     Obtener plan de cuotas
// @Description Retorna un plan de cuotas por ID con sus pagos individuales
// @Tags        installments
// @Security    BearerAuth
// @Produce     json
// @Param       id   path      string  true  "Plan ID"
// @Success     200  {object}  models.InstallmentPlanWithPayments
// @Failure     401  {object}  map[string]string
// @Failure     404  {object}  map[string]string
// @Failure     500  {object}  map[string]string
// @Router      /installments/{id} [get]
func (h *InstallmentHandler) GetOne(w http.ResponseWriter, r *http.Request) {
	authUserID := middleware.GetAuthUserID(r)
	id := chi.URLParam(r, "id")
	plan, err := h.svc.GetByID(r.Context(), id, authUserID)
	if err != nil {
		writeInternalError(w, r, err)
		return
	}
	if plan == nil {
		writeError(w, http.StatusNotFound, "no encontrado")
		return
	}
	writeJSON(w, http.StatusOK, plan)
}

// List godoc
// @Summary     Listar planes de cuotas
// @Description Retorna todos los planes activos del usuario con sus pagos individuales (paginado)
// @Tags        installments
// @Security    BearerAuth
// @Produce     json
// @Param       page   query     int  false  "Página (default: 1)"
// @Param       limit  query     int  false  "Resultados por página (default: 20, max: 100)"
// @Success     200  {object}  map[string]interface{}
// @Failure     401  {object}  map[string]string
// @Failure     500  {object}  map[string]string
// @Router      /installments [get]
func (h *InstallmentHandler) List(w http.ResponseWriter, r *http.Request) {
	authUserID := middleware.GetAuthUserID(r)
	p := parsePagination(r)
	plans, total, err := h.svc.GetAll(r.Context(), authUserID, p)
	if err != nil {
		writeInternalError(w, r, err)
		return
	}
	if plans == nil {
		plans = []models.InstallmentPlanWithPayments{}
	}
	writePaginatedJSON(w, plans, models.NewPaginationMeta(p.Page, p.Limit, total))
}

// Create godoc
// @Summary     Crear plan de cuotas
// @Description Crea un plan y genera automáticamente todos los installment_payments con sus due_dates
// @Tags        installments
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       body  body      models.CreateInstallmentRequest  true  "Datos del plan"
// @Success     201   {object}  map[string]models.InstallmentPlanWithPayments
// @Failure     400   {object}  map[string]string
// @Failure     401   {object}  map[string]string
// @Router      /installments [post]
func (h *InstallmentHandler) Create(w http.ResponseWriter, r *http.Request) {
	authUserID := middleware.GetAuthUserID(r)
	var req models.CreateInstallmentRequest
	if err := decode(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	plan, err := h.svc.Create(r.Context(), authUserID, &req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, plan)
}

// PayInstallment godoc
// @Summary     Pagar cuota
// @Description Marca una cuota como pagada. Incrementa installments_paid en el plan. Si se completan todas, is_completed=true.
// @Tags        installments
// @Security    BearerAuth
// @Produce     json
// @Param       id         path      string  true  "Plan ID"
// @Param       paymentID  path      string  true  "Payment ID"
// @Success     200        {object}  map[string]models.InstallmentPayment
// @Failure     401        {object}  map[string]string
// @Failure     404        {object}  map[string]string
// @Failure     500        {object}  map[string]string
// @Router      /installments/{id}/payments/{paymentID} [put]
func (h *InstallmentHandler) PayInstallment(w http.ResponseWriter, r *http.Request) {
	authUserID := middleware.GetAuthUserID(r)
	planID := chi.URLParam(r, "id")
	paymentID := chi.URLParam(r, "paymentID")
	payment, err := h.svc.PayInstallment(r.Context(), planID, paymentID, authUserID)
	if err != nil {
		if err.Error() == "not found" || err.Error() == "not found or already paid" {
			writeError(w, http.StatusNotFound, "no encontrado")
			return
		}
		writeInternalError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, payment)
}

// Delete godoc
// @Summary     Eliminar plan de cuotas
// @Description Soft delete del plan y sus pagos
// @Tags        installments
// @Security    BearerAuth
// @Produce     json
// @Param       id   path  string  true  "Plan ID"
// @Success     204
// @Failure     401  {object}  map[string]string
// @Failure     404  {object}  map[string]string
// @Failure     500  {object}  map[string]string
// @Router      /installments/{id} [delete]
func (h *InstallmentHandler) Delete(w http.ResponseWriter, r *http.Request) {
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
