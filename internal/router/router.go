package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/homepay/api/internal/handlers"
	"github.com/homepay/api/internal/middleware"
	httpswagger "github.com/swaggo/http-swagger"
)

func New(
	webhook *handlers.WebhookHandler,
	categories *handlers.CategoryHandler,
	companies *handlers.CompanyHandler,
	accountGroups *handlers.AccountGroupHandler,
	accounts *handlers.AccountHandler,
	billings *handlers.BillingHandler,
	expenses *handlers.ExpenseHandler,
	installments *handlers.InstallmentHandler,
	dashboard *handlers.DashboardHandler,
) http.Handler {
	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           300,
	}))
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)

	// Swagger UI — público
	r.Get("/docs/*", httpswagger.Handler(
		httpswagger.URL("/docs/doc.json"),
	))

	// Rutas públicas
	r.Post("/webhooks/clerk", webhook.Handle)

	// Rutas protegidas
	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth)

		r.Route("/categories", func(r chi.Router) {
			r.Get("/", categories.List)
			r.Post("/", categories.Create)
			r.Get("/{id}", categories.GetOne)
			r.Put("/{id}", categories.Update)
			r.Delete("/{id}", categories.Delete)
		})

		r.Route("/account-groups", func(r chi.Router) {
			r.Get("/", accountGroups.List)
			r.Post("/", accountGroups.Create)
			r.Get("/{id}", accountGroups.GetOne)
			r.Put("/{id}", accountGroups.Update)
			r.Delete("/{id}", accountGroups.Delete)
		})

		r.Route("/companies", func(r chi.Router) {
			r.Get("/", companies.List)
			r.Post("/", companies.Create)
			r.Get("/{id}", companies.GetOne)
			r.Put("/{id}", companies.Update)
			r.Delete("/{id}", companies.Delete)
		})

		r.Route("/accounts", func(r chi.Router) {
			r.Get("/", accounts.List)
			r.Post("/", accounts.Create)
			r.Get("/{id}", accounts.GetOne)
			r.Put("/{id}", accounts.Update)
			r.Delete("/{id}", accounts.Delete)
		})

		// Top-level billings routes
		r.Route("/billings", func(r chi.Router) {
			r.Get("/", billings.ListAll)
			r.Post("/", billings.Create)
			r.Get("/{id}", billings.GetOne)
			r.Put("/{id}", billings.Update)
			r.Delete("/{id}", billings.Delete)
		})

		r.Route("/periods/{period}", func(r chi.Router) {
			r.Post("/open", billings.OpenPeriod)
			r.Get("/billings", billings.ListByPeriod)
		})

		r.Route("/expenses", func(r chi.Router) {
			r.Get("/", expenses.List)
			r.Post("/", expenses.Create)
			r.Get("/{id}", expenses.GetOne)
			r.Put("/{id}", expenses.Update)
			r.Delete("/{id}", expenses.Delete)
		})

		r.Route("/installments", func(r chi.Router) {
			r.Get("/", installments.List)
			r.Post("/", installments.Create)
			r.Get("/{id}", installments.GetOne)
			r.Put("/{id}/payments/{paymentID}", installments.PayInstallment)
			r.Delete("/{id}", installments.Delete)
		})

		r.Get("/dashboard", dashboard.Get)
	})

	return r
}
