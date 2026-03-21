package models

type Category struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at,omitempty"`
}

type Company struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	WebsiteURL string `json:"website_url,omitempty"`
	CreatedAt  string `json:"created_at,omitempty"`
}

type Period struct {
	ID          int `json:"id"`
	MonthNumber int `json:"month_number,omitempty"`
	YearNumber  int `json:"year_number"`
}

type ServiceAccount struct {
	ID                int      `json:"id"`
	CompanyID         int      `json:"company_id,omitempty"`
	AccountIdentifier string   `json:"account_identifier"`
	Alias             string   `json:"alias,omitempty"`
	Company           *Company `json:"company,omitempty"`
}

type Expense struct {
	ID                 int             `json:"id"`
	CategoryID         int             `json:"category_id"`
	PeriodID           int             `json:"period_id"`
	AccountID          *int            `json:"account_id,omitempty"`
	Description        string          `json:"description"`
	DueDate            *string         `json:"due_date,omitempty"`
	CurrentAmount      float64         `json:"current_amount"`
	AmountPaid         float64         `json:"amount_paid"`
	CurrentInstallment int             `json:"current_installment,omitempty"`
	TotalInstallments  int             `json:"total_installments,omitempty"`
	InstallmentGroupID *string         `json:"installment_group_id,omitempty"`
	IsRecurring        bool            `json:"is_recurring"`
	Notes              string          `json:"notes,omitempty"`
	CreatedAt          string          `json:"created_at,omitempty"`
	UpdatedAt          string          `json:"updated_at,omitempty"`
	Category           *Category       `json:"category,omitempty"`
	Period             *Period         `json:"period,omitempty"`
	ServiceAccount     *ServiceAccount `json:"service_account,omitempty"`
}

type Income struct {
	ID          int     `json:"id"`
	PeriodID    int     `json:"period_id"`
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
	IsRecurring bool    `json:"is_recurring"`
	ReceivedAt  string  `json:"received_at,omitempty"`
	CreatedAt   string  `json:"created_at,omitempty"`
	Period      *Period `json:"period,omitempty"`
}
