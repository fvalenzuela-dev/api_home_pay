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

func (h *InstallmentHandler) List(w http.ResponseWriter, r *http.Request) {
	authUserID := middleware.GetAuthUserID(r)
	plans, err := h.svc.GetAll(r.Context(), authUserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "error interno")
		return
	}
	if plans == nil {
		plans = []models.InstallmentPlanWithPayments{}
	}
	writeJSON(w, http.StatusOK, plans)
}

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
		writeError(w, http.StatusInternalServerError, "error interno")
		return
	}
	writeJSON(w, http.StatusOK, payment)
}

func (h *InstallmentHandler) Delete(w http.ResponseWriter, r *http.Request) {
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
