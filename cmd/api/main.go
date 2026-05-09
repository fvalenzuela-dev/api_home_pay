// @title           HomePay API
// @version         ${VERSION}
// @description     Backend REST para gestión de finanzas personales HomePay. Todos los endpoints protegidos requieren un JWT de Clerk en el header Authorization.
// @BasePath        /
// @securityDefinitions.apikey BearerAuth
// @in              header
// @name            Authorization
// @description     JWT de Clerk. Formato: "Bearer <token>"
package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"

	clerkSDK "github.com/clerk/clerk-sdk-go/v2"
	_ "github.com/homepay/api/docs"
	"github.com/homepay/api/internal/config"
	"github.com/homepay/api/internal/database"
	"github.com/homepay/api/internal/handlers"
	"github.com/homepay/api/internal/repository"
	"github.com/homepay/api/internal/router"
	"github.com/homepay/api/internal/service"
)

// ServerConfig holds server configuration
type ServerConfig struct {
	Addr string
}

var version = "dev"

var app *App

// App holds all application dependencies
type App struct {
	Config *config.Config
	DB     interface {
		Close()
		Ping(ctx context.Context) error
	}
	Router http.Handler
}

// InitializeApp creates and wires up all application dependencies
// This function is exposed for testing and integration purposes
func InitializeApp(cfg *config.Config) (*App, error) {
	clerkSDK.SetKey(cfg.ClerkSecretKey)

	db, err := database.Connect(context.Background(), cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	// Repositories
	userRepo := repository.NewUserRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	companyRepo := repository.NewCompanyRepository(db)
	accountGroupRepo := repository.NewAccountGroupRepository(db)
	accountRepo := repository.NewAccountRepository(db)
	billingRepo := repository.NewBillingRepository(db)
	expenseRepo := repository.NewExpenseRepository(db)
	installmentRepo := repository.NewInstallmentRepository(db)

	// Services
	companySvc := service.NewCompanyService(companyRepo, accountRepo, billingRepo)
	accountGroupSvc := service.NewAccountGroupService(accountGroupRepo)
	accountSvc := service.NewAccountService(accountRepo, companyRepo, billingRepo)
	billingSvc := service.NewBillingService(billingRepo, accountRepo)
	expenseSvc := service.NewExpenseService(expenseRepo)
	installmentSvc := service.NewInstallmentService(installmentRepo)
	dashboardSvc := service.NewDashboardService(billingRepo, expenseRepo, installmentRepo)

	// Handlers
	webhookHandler := handlers.NewWebhookHandler(userRepo, cfg.ClerkWebhookSecret)
	categoryHandler := handlers.NewCategoryHandler(categoryRepo)
	companyHandler := handlers.NewCompanyHandler(companySvc)
	accountGroupHandler := handlers.NewAccountGroupHandler(accountGroupSvc)
	accountHandler := handlers.NewAccountHandler(accountSvc)
	billingHandler := handlers.NewBillingHandler(billingSvc)
	expenseHandler := handlers.NewExpenseHandler(expenseSvc)
	installmentHandler := handlers.NewInstallmentHandler(installmentSvc)
	dashboardHandler := handlers.NewDashboardHandler(dashboardSvc)

	r := router.New(
		webhookHandler,
		categoryHandler,
		companyHandler,
		accountGroupHandler,
		accountHandler,
		billingHandler,
		expenseHandler,
		installmentHandler,
		dashboardHandler,
	)

	return &App{
		Config: cfg,
		DB:     db,
		Router: r,
	}, nil
}

// healthReady godoc
// @Summary     Health check - readiness probe
// @Description Returns 200 if the service and database are ready. Used by GCP Cloud Run readiness probe.
// @Tags        health
// @Produce     json
// @Success     200  {object}  map[string]string
// @Failure     503  {object}  map[string]string
// @Router      /health/ready [get]
func healthReady(w http.ResponseWriter, r *http.Request) {
	if err := app.DB.Ping(r.Context()); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"error": "database unavailable"})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"data": `{"status":"ready"}`})
}

func main() {
	setupLogger()

	cfg, err := loadConfig()
	if err != nil {
		slog.Error("config error", "error", err)
		os.Exit(1)
	}

	app, err := initializeApp(cfg)
	if err != nil {
		slog.Error("app initialization error", "error", err)
		os.Exit(1)
	}
	defer app.DB.Close()

	mux := setupMux(app.Router)

	serverCfg := getServerConfig(cfg)
	slog.Info("server starting", "addr", serverCfg.Addr)

	startServer(serverCfg, mux)
}

func loadConfig() (*config.Config, error) {
	return config.Load()
}

func initializeApp(cfg *config.Config) (*App, error) {
	return InitializeApp(cfg)
}

func startServer(serverCfg ServerConfig, mux *http.ServeMux) {
	// TLS is handled by Cloud Run load balancer
	// Cloud Run terminates TLS and forwards HTTP to the container
	// codacy-ignore-next-line G114
	if err := http.ListenAndServe(serverCfg.Addr, mux); err != nil {
		slog.Error("server error", "errors", err)
		os.Exit(1)
	}
}

func setupLogger() *slog.Logger {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	return logger
}

func setupMux(router http.Handler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health/ready", healthReady)
	mux.Handle("/", router)
	return mux
}

// getServerConfig determines server configuration
func getServerConfig(cfg *config.Config) ServerConfig {
	return ServerConfig{
		Addr: ":" + cfg.Port,
	}
}
