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
