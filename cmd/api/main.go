package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	clerkSDK "github.com/clerk/clerk-sdk-go/v2"
	"github.com/homepay/api/internal/config"
	"github.com/homepay/api/internal/database"
	"github.com/homepay/api/internal/handlers"
	"github.com/homepay/api/internal/repository"
	"github.com/homepay/api/internal/router"
	"github.com/homepay/api/internal/service"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("config error", "error", err)
		os.Exit(1)
	}

	clerkSDK.SetKey(cfg.ClerkSecretKey)

	db, err := database.Connect(context.Background(), cfg.DatabaseURL)
	if err != nil {
		slog.Error("database connection error", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Repositories
	userRepo := repository.NewUserRepository(db)
	companyRepo := repository.NewCompanyRepository(db)
	accountRepo := repository.NewAccountRepository(db)
	billingRepo := repository.NewBillingRepository(db)
	expenseRepo := repository.NewExpenseRepository(db)
	installmentRepo := repository.NewInstallmentRepository(db)

	// Services
	companySvc := service.NewCompanyService(companyRepo, accountRepo, billingRepo)
	accountSvc := service.NewAccountService(accountRepo, billingRepo)
	billingSvc := service.NewBillingService(billingRepo, accountRepo)
	expenseSvc := service.NewExpenseService(expenseRepo)
	installmentSvc := service.NewInstallmentService(installmentRepo)
	dashboardSvc := service.NewDashboardService(billingRepo, expenseRepo, installmentRepo)

	// Handlers
	webhookHandler := handlers.NewWebhookHandler(userRepo, cfg.ClerkWebhookSecret)
	companyHandler := handlers.NewCompanyHandler(companySvc)
	accountHandler := handlers.NewAccountHandler(accountSvc)
	billingHandler := handlers.NewBillingHandler(billingSvc)
	expenseHandler := handlers.NewExpenseHandler(expenseSvc)
	installmentHandler := handlers.NewInstallmentHandler(installmentSvc)
	dashboardHandler := handlers.NewDashboardHandler(dashboardSvc)

	r := router.New(
		webhookHandler,
		companyHandler,
		accountHandler,
		billingHandler,
		expenseHandler,
		installmentHandler,
		dashboardHandler,
	)

	addr := ":" + cfg.Port
	slog.Info("server starting", "addr", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}
