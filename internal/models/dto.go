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
