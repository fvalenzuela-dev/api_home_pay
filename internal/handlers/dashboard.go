package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/homepay/api/internal/middleware"
	"github.com/homepay/api/internal/service"
)

type DashboardHandler struct {
	svc service.DashboardService
}

func NewDashboardHandler(svc service.DashboardService) *DashboardHandler {
	return &DashboardHandler{svc: svc}
}

// Get godoc
// @Summary     Resumen financiero mensual
// @Description Retorna totales de facturas, gastos agrupados por categoría y cuotas del mes. Si no se pasan parámetros, usa el mes actual.
// @Tags        dashboard
// @Security    BearerAuth
// @Produce     json
// @Param       month  query     int  false  "Mes (1-12)"
// @Param       year   query     int  false  "Año (ej: 2026)"
// @Success     200    {object}  map[string]service.DashboardSummary
// @Failure     401    {object}  map[string]string
// @Failure     500    {object}  map[string]string
// @Router      /dashboard [get]
func (h *DashboardHandler) Get(w http.ResponseWriter, r *http.Request) {
	authUserID := middleware.GetAuthUserID(r)

	now := time.Now()
	month := now.Month()
	year := now.Year()

	if m := r.URL.Query().Get("month"); m != "" {
		if v, err := strconv.Atoi(m); err == nil && v >= 1 && v <= 12 {
			month = time.Month(v)
		}
	}
	if y := r.URL.Query().Get("year"); y != "" {
		if v, err := strconv.Atoi(y); err == nil && v >= 2000 {
			year = v
		}
	}

	summary, err := h.svc.GetSummary(r.Context(), authUserID, int(month), year)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "error interno")
		return
	}
	writeJSON(w, http.StatusOK, summary)
}
