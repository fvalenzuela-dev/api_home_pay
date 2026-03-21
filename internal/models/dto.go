package models

type PeriodSummary struct {
	PeriodID           int               `json:"period_id"`
	Period             *Period           `json:"period,omitempty"`
	TotalIncomes       float64           `json:"total_incomes"`
	TotalExpenses      float64           `json:"total_expenses"`
	Balance            float64           `json:"balance"`
	PaidAmount         float64           `json:"paid_amount"`
	PendingAmount      float64           `json:"pending_amount"`
	ExpensesByCategory []CategorySummary `json:"expenses_by_category,omitempty"`
}

type CategorySummary struct {
	CategoryID   int     `json:"category_id"`
	CategoryName string  `json:"category_name"`
	TotalAmount  float64 `json:"total_amount"`
	Count        int     `json:"count"`
}

type PayExpenseRequest struct {
	Amount float64 `json:"amount" binding:"required,gt=0"`
}

type PendingExpensesFilter struct {
	DaysAhead int `json:"days_ahead,omitempty"`
}

// Request DTOs - only user-provided fields
type CreateCategoryRequest struct {
	Name string `json:"name" binding:"required"`
}

type UpdateCategoryRequest struct {
	Name string `json:"name" binding:"required"`
}

type CreateCompanyRequest struct {
	Name       string `json:"name" binding:"required"`
	WebsiteURL string `json:"website_url,omitempty"`
}

type UpdateCompanyRequest struct {
	Name       string `json:"name" binding:"required"`
	WebsiteURL string `json:"website_url,omitempty"`
}

type CreatePeriodRequest struct {
	MonthNumber int `json:"month_number,omitempty"`
	YearNumber  int `json:"year_number" binding:"required"`
}

type UpdatePeriodRequest struct {
	MonthNumber int `json:"month_number,omitempty"`
	YearNumber  int `json:"year_number" binding:"required"`
}

type CreateServiceAccountRequest struct {
	CompanyID         int    `json:"company_id" binding:"required"`
	AccountIdentifier string `json:"account_identifier" binding:"required"`
	Alias             string `json:"alias,omitempty"`
}

type UpdateServiceAccountRequest struct {
	CompanyID         int    `json:"company_id" binding:"required"`
	AccountIdentifier string `json:"account_identifier" binding:"required"`
	Alias             string `json:"alias,omitempty"`
}

type CreateExpenseRequest struct {
	CategoryID         int     `json:"category_id" binding:"required"`
	PeriodID           int     `json:"period_id" binding:"required"`
	AccountID          *int    `json:"account_id,omitempty"`
	Description        string  `json:"description" binding:"required"`
	DueDate            *string `json:"due_date,omitempty"`
	CurrentAmount      float64 `json:"current_amount" binding:"required"`
	AmountPaid         float64 `json:"amount_paid,omitempty"`
	CurrentInstallment int     `json:"current_installment,omitempty"`
	TotalInstallments  int     `json:"total_installments,omitempty"`
	InstallmentGroupID *string `json:"installment_group_id,omitempty"`
	IsRecurring        bool    `json:"is_recurring"`
	Notes              string  `json:"notes,omitempty"`
}

type UpdateExpenseRequest struct {
	CategoryID         int     `json:"category_id" binding:"required"`
	PeriodID           int     `json:"period_id" binding:"required"`
	AccountID          *int    `json:"account_id,omitempty"`
	Description        string  `json:"description" binding:"required"`
	DueDate            *string `json:"due_date,omitempty"`
	CurrentAmount      float64 `json:"current_amount" binding:"required"`
	AmountPaid         float64 `json:"amount_paid,omitempty"`
	CurrentInstallment int     `json:"current_installment,omitempty"`
	TotalInstallments  int     `json:"total_installments,omitempty"`
	InstallmentGroupID *string `json:"installment_group_id,omitempty"`
	IsRecurring        bool    `json:"is_recurring"`
	Notes              string  `json:"notes,omitempty"`
}

type CreateIncomeRequest struct {
	PeriodID    int     `json:"period_id" binding:"required"`
	Description string  `json:"description" binding:"required"`
	Amount      float64 `json:"amount" binding:"required"`
	IsRecurring bool    `json:"is_recurring"`
	ReceivedAt  string  `json:"received_at,omitempty"`
}

type UpdateIncomeRequest struct {
	PeriodID    int     `json:"period_id" binding:"required"`
	Description string  `json:"description" binding:"required"`
	Amount      float64 `json:"amount" binding:"required"`
	IsRecurring bool    `json:"is_recurring"`
	ReceivedAt  string  `json:"received_at,omitempty"`
}
