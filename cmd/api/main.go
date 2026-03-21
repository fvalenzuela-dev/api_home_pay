package main

import (
	"fmt"
	"log"
	"os"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/fernandovalenzuela/api-home-pay/docs"
	"github.com/fernandovalenzuela/api-home-pay/internal/config"
	"github.com/fernandovalenzuela/api-home-pay/internal/handlers"
	"github.com/fernandovalenzuela/api-home-pay/internal/middleware"
	"github.com/fernandovalenzuela/api-home-pay/internal/repository"
	"github.com/fernandovalenzuela/api-home-pay/internal/services"
	"github.com/fernandovalenzuela/api-home-pay/internal/utils"
)

// @title api-home-pay API
// @version 1.0
// @description API REST para gestión de gastos e ingresos del hogar
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@example.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and the JWT token.

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	clerk.SetKey(cfg.ClerkSecretKey)

	db, err := repository.NewDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	gin.SetMode(cfg.GinMode)

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.ResponseMiddleware())

	router.GET("/health", func(c *gin.Context) {
		utils.SuccessResponse(c, gin.H{
			"status":  "healthy",
			"service": "api-home-pay",
		})
	})

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, func(c *ginSwagger.Config) {
		c.URL = "/swagger/doc.json"
	}))

	api := router.Group("/api")
	api.Use(middleware.ClerkAuthMiddleware())
	{
		api.GET("/me", func(c *gin.Context) {
			userID, _ := middleware.GetUserID(c)
			utils.SuccessResponse(c, gin.H{
				"user_id": userID,
			})
		})

		// Initialize repositories
		categoryRepo := repository.NewCategoryRepository(db.Conn)
		periodRepo := repository.NewPeriodRepository(db.Conn)
		companyRepo := repository.NewCompanyRepository(db.Conn)
		serviceAccountRepo := repository.NewServiceAccountRepository(db.Conn)
		expenseRepo := repository.NewExpenseRepository(db.Conn)
		incomeRepo := repository.NewIncomeRepository(db.Conn)

		// Initialize services
		categoryService := services.NewCategoryService(categoryRepo)
		periodService := services.NewPeriodService(periodRepo)
		companyService := services.NewCompanyService(companyRepo)
		serviceAccountService := services.NewServiceAccountService(serviceAccountRepo)
		expenseService := services.NewExpenseService(expenseRepo)
		incomeService := services.NewIncomeService(incomeRepo)

		// Initialize handlers
		categoryHandler := handlers.NewCategoryHandler(categoryService)
		periodHandler := handlers.NewPeriodHandler(periodService)
		companyHandler := handlers.NewCompanyHandler(companyService)
		serviceAccountHandler := handlers.NewServiceAccountHandler(serviceAccountService)
		expenseHandler := handlers.NewExpenseHandler(expenseService)
		incomeHandler := handlers.NewIncomeHandler(incomeService)
		summaryHandler := handlers.NewSummaryHandler(expenseRepo, incomeRepo, periodRepo)

		// Category routes
		categories := api.Group("/categories")
		{
			categories.POST("", categoryHandler.Create)
			categories.GET("", categoryHandler.GetAll)
			categories.GET("/:id", categoryHandler.GetByID)
			categories.PUT("/:id", categoryHandler.Update)
			categories.DELETE("/:id", categoryHandler.Delete)
		}

		// Period routes
		periods := api.Group("/periods")
		{
			periods.POST("", periodHandler.Create)
			periods.GET("", periodHandler.GetAll)
			periods.GET("/:id", periodHandler.GetByID)
			periods.PUT("/:id", periodHandler.Update)
			periods.DELETE("/:id", periodHandler.Delete)
		}

		// Company routes
		companies := api.Group("/companies")
		{
			companies.POST("", companyHandler.Create)
			companies.GET("", companyHandler.GetAll)
			companies.GET("/:id", companyHandler.GetByID)
			companies.PUT("/:id", companyHandler.Update)
			companies.DELETE("/:id", companyHandler.Delete)
		}

		// Service Account routes
		serviceAccounts := api.Group("/service-accounts")
		{
			serviceAccounts.POST("", serviceAccountHandler.Create)
			serviceAccounts.GET("", serviceAccountHandler.GetAll)
			serviceAccounts.GET("/:id", serviceAccountHandler.GetByID)
			serviceAccounts.PUT("/:id", serviceAccountHandler.Update)
			serviceAccounts.DELETE("/:id", serviceAccountHandler.Delete)
		}

		// Expense routes
		expenses := api.Group("/expenses")
		{
			expenses.POST("", expenseHandler.Create)
			expenses.GET("", expenseHandler.GetAll)
			expenses.GET("/pending", expenseHandler.GetPending)
			expenses.GET("/:id", expenseHandler.GetByID)
			expenses.PUT("/:id", expenseHandler.Update)
			expenses.DELETE("/:id", expenseHandler.Delete)
			expenses.PATCH("/:id/pay", expenseHandler.MarkAsPaid)
		}

		// Summary route
		api.GET("/summary/:period_id", summaryHandler.GetByPeriod)

		// Income routes
		incomes := api.Group("/incomes")
		{
			incomes.POST("", incomeHandler.Create)
			incomes.GET("", incomeHandler.GetAll)
			incomes.GET("/:id", incomeHandler.GetByID)
			incomes.PUT("/:id", incomeHandler.Update)
			incomes.DELETE("/:id", incomeHandler.Delete)
		}
	}

	port := cfg.Port
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server starting on port %s\n", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
		os.Exit(1)
	}
}
