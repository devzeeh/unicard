package authentication

import (
	"fmt"
	"net/http"
)

// Transaction struct represents a user's transaction for the dashboard view
type Transaction struct {
	Date        string  `db:"date" json:"date"`
	Description string  `db:"description" json:"description"`
	Type        string  `db:"transaction_type" json:"type"`
	Amount      float64 `db:"transaction_amount" json:"amount"`
}

// DashboardUser info struct for the user dashboard view
type DashboardUser struct {
	ID                 int           `db:"id" json:"id,omitempty"`
	UserID             string        `db:"user_id" json:"user_id,omitempty"`
	Username           string        `db:"username" json:"username"`
	Name               string        `db:"name" json:"name"`
	Transaction        string        `db:"transaction" json:"transaction"`
	Balance            float64       `db:"balance" json:"balance"`
	LoyaltyPoints      int           `db:"loyalty_points" json:"loyalty_points"`
	AccountType        string        `db:"account_type" json:"account_type"`
	RecentTransactions []Transaction `json:"recent_transactions"` // Add recent transactions to the dashboard response
}

func (h *Handler) DashboardView(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Dashboard view is running...")
	h.Tpl.ExecuteTemplate(w, "dashboard.html", nil)
}

func (h *Handler) DashboardHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Dashboard handler is running...")
	
// For demonstration, we'll create a dummy user with some transactions
	dashboardUser := DashboardUser{
		ID:            1,
		UserID:        "user123",
		Username:      "johndoe",
		Name:          "John Doe",
		Transaction:   "Transaction",
		Balance:       150.75,
		LoyaltyPoints: 200,
		AccountType:   "Premium",
		RecentTransactions: []Transaction{
			{Date: "2024-06-01", Description: "Grocery Store", Type: "Debit", Amount: 50.25},
			{Date: "2024-06-03", Description: "Salary", Type: "Credit", Amount: 2000.00},
			{Date: "2024-06-05", Description: "Online Shopping", Type: "Debit", Amount: 75.50},
		},
	}
	h.Tpl.ExecuteTemplate(w, "dashboard.html", dashboardUser)
}