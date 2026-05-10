package handlers

import (
	"net/http"
	"strconv"

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

// GetOne godoc
// @Summary     Obtener factura
// @Description Retorna una factura por ID
// @Tags        billings
// @Security    BearerAuth
// @Produce     json
// @Param       accountID  path      string  true  "Account ID"
// @Param       id         path      string  true  "Billing ID"
// @Success     200        {object}  models.AccountBilling
// @Failure     401        {object}  map[string]string
// @Failure     404        {object}  map[string]string
// @Failure     500        {object}  map[string]string
// @Router      /billings/{id} [get]
func (h *BillingHandler) GetOne(w http.ResponseWriter, r *http.Request) {
	authUserID := middleware.GetAuthUserID(r)
	id := chi.URLParam(r, "id")
	billing, err := h.svc.GetByID(r.Context(), id, authUserID)
	if err != nil {
		writeInternalError(w, r, err)
		return
	}
	if billing == nil {
		writeError(w, http.StatusNotFound, "no encontrado")
		return
	}
	writeJSON(w, http.StatusOK, billing)
}

// Create godoc
// @Summary     Registrar factura
// @Description Registra la factura del mes para una cuenta. Si auto_accumulate=true y hay una factura impaga, se crea un carry-over automáticamente.
// En /accounts/{accountID}/billings el accountID se toma del path.
// En /billings (top-level) se requiere account_id en el body.
// @Tags        billings
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       accountID  path      string                      false "Account ID (nested routes)"
// @Param       body       body      models.CreateBillingRequest  true  "Datos de la factura"
// @Success     201        {object}  map[string]models.AccountBilling
// @Failure     400        {object}  map[string]string
// @Failure     401        {object}  map[string]string
// @Failure     404        {object}  map[string]string
// @Router      /billings [post]
func (h *BillingHandler) Create(w http.ResponseWriter, r *http.Request) {
	authUserID := middleware.GetAuthUserID(r)
	var req models.CreateBillingRequest
	if err := decode(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	// Get accountID: from path param (nested) or from body field (top-level)
	accountID := chi.URLParam(r, "accountID")
	if accountID == "" {
		accountID = req.AccountID
	}
	if accountID == "" {
		writeError(w, http.StatusBadRequest, "account_id es requerido")
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
// @Router      /billings/{id} [put]
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

// OpenPeriod godoc
// @Summary     Abrir periodo
// @Description Genera un billing por cada cuenta activa del usuario para el periodo indicado. Idempotente: si el billing ya existe, lo saltea. Aplica carry-over del periodo anterior si hay deuda pendiente.
// @Tags        periods
// @Security    BearerAuth
// @Produce     json
// @Param       period  path      int  true  "Periodo YYYYMM (ej: 202605)"
// @Success     200     {object}  models.OpenPeriodResponse
// @Failure     400     {object}  map[string]string
// @Failure     401     {object}  map[string]string
// @Failure     500     {object}  map[string]string
// @Router      /periods/{period}/open [post]
func (h *BillingHandler) OpenPeriod(w http.ResponseWriter, r *http.Request) {
	authUserID := middleware.GetAuthUserID(r)
	periodStr := chi.URLParam(r, "period")
	period, err := strconv.Atoi(periodStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "period inválido: debe ser un entero YYYYMM")
		return
	}
	resp, err := h.svc.OpenPeriod(r.Context(), authUserID, period)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// ListByPeriod godoc
// @Summary     Listar billings de un periodo
// @Description Retorna todos los billings del usuario para el periodo indicado. Filtrable por estado de pago.
// @Tags        periods
// @Security    BearerAuth
// @Produce     json
// @Param       period    path      int     true   "Periodo YYYYMM (ej: 202605)"
// @Param       status    query     string  false  "Filtro: all (default), paid, unpaid"
// @Param       page      query     int     false  "Página (default: 1)"
// @Param       page_size query     int     false  "Resultados por página (default: 20, max: 100)"
// @Success     200       {object}  map[string]interface{}
// @Failure     400       {object}  map[string]string
// @Failure     401       {object}  map[string]string
// @Failure     500       {object}  map[string]string
// @Router      /periods/{period}/billings [get]
func (h *BillingHandler) ListByPeriod(w http.ResponseWriter, r *http.Request) {
	authUserID := middleware.GetAuthUserID(r)
	periodStr := chi.URLParam(r, "period")
	period, err := strconv.Atoi(periodStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "period inválido: debe ser un entero YYYYMM")
		return
	}

	var isPaid *bool
	if status := r.URL.Query().Get("status"); status != "" {
		switch status {
		case "paid":
			v := true
			isPaid = &v
		case "unpaid":
			v := false
			isPaid = &v
		case "all":
			// isPaid queda nil = sin filtro
		default:
			writeError(w, http.StatusBadRequest, "status inválido: usar all, paid o unpaid")
			return
		}
	}

	p := parsePagination(r)
	billings, total, err := h.svc.GetAllByPeriod(r.Context(), authUserID, period, isPaid, p)
	if err != nil {
		writeInternalError(w, r, err)
		return
	}
	if billings == nil {
		billings = []models.AccountBillingWithDetails{}
	}
	writePaginatedJSON(w, billings, models.NewPaginationMeta(p.Page, p.Limit, total))
}

// ListAll godoc
// @Summary     Listar todas las facturas
// @Description Retorna todas las facturas del usuario con filtros opcionales. Si no se pasa account_id, busca en todas las cuentas.
// @Tags        billings
// @Security    BearerAuth
// @Produce     json
// @Param       account_id  query     string  false  "Filtrar por account ID"
// @Param       from_period query     int     false  "Periodo inicio YYYYMM (inclusive)"
// @Param       to_period   query     int     false  "Periodo fin YYYYMM (inclusive)"
// @Param       is_paid     query     bool    false  "Filtrar por estado: true=pagadas, false=impagas"
// @Param       page        query     int     false  "Página (default: 1)"
// @Param       page_size   query     int     false  "Resultados por página (default: 20, max: 100)"
// @Success     200         {object}  map[string]interface{}
// @Failure     401         {object}  map[string]string
// @Failure     500         {object}  map[string]string
// @Router      /billings [get]
func (h *BillingHandler) ListAll(w http.ResponseWriter, r *http.Request) {
	authUserID := middleware.GetAuthUserID(r)
	filters := parseBillingFilters(r)
	p := parsePagination(r)

	billings, total, err := h.svc.GetAll(r.Context(), authUserID, filters, p)
	if err != nil {
		writeInternalError(w, r, err)
		return
	}
	if billings == nil {
		billings = []models.AccountBilling{}
	}
	writePaginatedJSON(w, billings, models.NewPaginationMeta(p.Page, p.Limit, total))
}

// Delete godoc
// @Summary     Eliminar factura
// @Description Realiza un soft-delete de una factura por ID.
// @Tags        billings
// @Security    BearerAuth
// @Param       id  path  string  true  "Billing ID"
// @Success     204
// @Failure     401  {object}  map[string]string
// @Failure     404  {object}  map[string]string
// @Failure     500  {object}  map[string]string
// @Router      /billings/{id} [delete]
func (h *BillingHandler) Delete(w http.ResponseWriter, r *http.Request) {
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

// parseBillingFilters extrae los query params de filtrado desde la request.
func parseBillingFilters(r *http.Request) models.BillingFilters {
	filters := models.BillingFilters{}

	if accountID := r.URL.Query().Get("account_id"); accountID != "" {
		filters.AccountID = &accountID
	}
	if fromStr := r.URL.Query().Get("from_period"); fromStr != "" {
		if from, err := strconv.Atoi(fromStr); err == nil {
			filters.FromPeriod = &from
		}
	}
	if toStr := r.URL.Query().Get("to_period"); toStr != "" {
		if to, err := strconv.Atoi(toStr); err == nil {
			filters.ToPeriod = &to
		}
	}
	if isPaidStr := r.URL.Query().Get("is_paid"); isPaidStr != "" {
		if isPaid, err := strconv.ParseBool(isPaidStr); err == nil {
			filters.IsPaid = &isPaid
		}
	}

	return filters
}
