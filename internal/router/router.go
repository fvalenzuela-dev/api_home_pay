package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/homepay/api/internal/handlers"
	"github.com/homepay/api/internal/middleware"
)

func New(
	webhook *handlers.WebhookHandler,
	companies *handlers.CompanyHandler,
	accounts *handlers.AccountHandler,
	billings *handlers.BillingHandler,
	expenses *handlers.ExpenseHandler,
	installments *handlers.InstallmentHandler,
	dashboard *handlers.DashboardHandler,
) http.Handler {
	r := chi.NewRouter()
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)

	// Rutas públicas
	r.Post("/webhooks/clerk", webhook.Handle)

	// Rutas protegidas
	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth)

		r.Route("/companies", func(r chi.Router) {
			r.Get("/", companies.List)
			r.Post("/", companies.Create)
			r.Put("/{id}", companies.Update)
			r.Delete("/{id}", companies.Delete)

			r.Route("/{companyID}/accounts", func(r chi.Router) {
				r.Get("/", accounts.List)
				r.Post("/", accounts.Create)
				r.Put("/{id}", accounts.Update)
				r.Delete("/{id}", accounts.Delete)
			})
		})

		r.Route("/accounts/{accountID}/billings", func(r chi.Router) {
			r.Get("/", billings.List)
			r.Post("/", billings.Create)
			r.Put("/{id}", billings.Update)
		})

		r.Route("/expenses", func(r chi.Router) {
			r.Get("/", expenses.List)
			r.Post("/", expenses.Create)
			r.Put("/{id}", expenses.Update)
			r.Delete("/{id}", expenses.Delete)
		})

		r.Route("/installments", func(r chi.Router) {
			r.Get("/", installments.List)
			r.Post("/", installments.Create)
			r.Put("/{id}/payments/{paymentID}", installments.PayInstallment)
			r.Delete("/{id}", installments.Delete)
		})

		r.Get("/dashboard", dashboard.Get)
	})

	return r
}
