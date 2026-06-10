package user

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"
	jsonwrite "unicard-go/backend/internal/pkg/handler"
)

// Transaction struct represents a user's transaction for the dashboard view
type Transaction struct {
	Date        string  `json:"date" db:"date"`
	Description string  `json:"description" db:"description"`
	Type        string  `json:"type" db:"transaction_type"`
	Amount      float64 `json:"amount" db:"transaction_amount"`
}

// DashboardUser info struct for the user dashboard view
type DashboardUser struct {
	ID                 int           `json:"id,omitempty" db:"id"`
	UserID             string        `json:"user_id,omitempty" db:"user_id"`
	Username           string        `json:"username" db:"username"`
	Name               string        `json:"name" db:"name"`
	Email              string        `json:"email" db:"email"`
	Phone              string        `json:"phone" db:"phone"`
	Initials           string        `json:"initials"`
	Balance            float64       `json:"balance" db:"balance"`
	LoyaltyPoints      float64       `json:"loyalty_points" db:"loyalty_points"`
	AccountType        string        `db:"account_type" json:"account_type"`
	CardNumber         string        `json:"card_number"`
	CardExpiry         string        `json:"card_expiry"`
	CardStatus         string        `json:"card_status"`
	RecentTransactions []Transaction `json:"recent_transactions"` // Add recent transactions to the dashboard response
}

func (h *Handler) DashboardView(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Dashboard view is running...")

	username := r.PathValue("username")
	data := struct {
		Username string
	}{
		Username: username,
	}

	h.Tpl.ExecuteTemplate(w, "dashboard.html", data)
}

func (h *Handler) DashboardHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Dashboard JSON handler is running...")
	
	// Get user ID from path param (No cookies for now)
	userID := r.PathValue("username")
	if userID == "" {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "user is required",
		})
		return
	}

	// Fetch user and card details
	var (
		id            int
		username      string
		fullName      string
		email         string
		phone         string
		userType      string
		balance       float64
		loyaltyPoints float64
		cardNumber    string
		expiryDate    string
		cardStatus    string
	)
	stmt := `
		SELECT 
			u.id,
			u.username,
			u.name,
			u.email,
			COALESCE(u.phone_number, ''),
			u.role,
			COALESCE(c.balance, 0),
			COALESCE(c.loyalty_points, 0),
			COALESCE(c.card_number, ''),
			COALESCE(c.expiry_date, ''),
			COALESCE(c.status, '')
		FROM users u
		LEFT JOIN cards c 
			ON u.user_id = c.user_id
		WHERE u.username = ?
	`
	err := h.DB.QueryRow(stmt, userID).Scan(&id, &username, &fullName, &email, &phone, &userType, &balance, &loyaltyPoints, &cardNumber, &expiryDate, &cardStatus)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Printf("User %s not found in DB\n", userID)
		} else {
			fmt.Printf("Error fetching user %s from DB: %v\n", userID, err)
		}
		// Clear invalid session cookie (Removed)
		jsonwrite.WriteJSON(w, http.StatusUnauthorized, jsonwrite.APIResponse{
			Success: false,
			Message: "Unauthorized: User not found",
		})
		return
	}

	// Generate Initials
	initials := ""
	parts := strings.Fields(fullName)
	if len(parts) > 0 {
		initials += string([]rune(parts[0])[0])
		if len(parts) > 1 {
			initials += string([]rune(parts[len(parts)-1])[0])
		}
	}
	if initials == "" {
		initials = "U"
	}
	initials = strings.ToUpper(initials)

	// Format Expiry Date
	expiryStr := "MM/YY"
	if len(expiryDate) >= 10 {
		tExpiry, errT := time.Parse("2006-01-02", expiryDate[:10])
		if errT == nil {
			expiryStr = tExpiry.Format("01/06")
		}
	}

	// Fetch recent transactions
	txnQuery := `
		SELECT t.created_at, m.business_name, t.transaction_type, t.amount 
		FROM transactions t 
		JOIN cards c ON t.card_number = c.card_number 
		JOIN merchants m ON t.merchant_id = m.id 
		WHERE c.user_id = ? 
		ORDER BY t.created_at DESC LIMIT 5
	`
	rows, err := h.DB.Query(txnQuery, userID)
	var transactions []Transaction
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var t Transaction
			var createdAt string
			if err := rows.Scan(&createdAt, &t.Description, &t.Type, &t.Amount); err == nil {
				t.Date = formatDate(createdAt)
				transactions = append(transactions, t)
			}
		}
	} else {
		fmt.Printf("Error fetching transactions: %v\n", err)
	}

	dashboardUser := DashboardUser{
		ID:                 id,
		UserID:             userID,
		Username:           username,
		Name:               fullName,
		Email:              email,
		Phone:              phone,
		Initials:           initials,
		Balance:            balance,
		LoyaltyPoints:      loyaltyPoints,
		AccountType:        userType,
		CardNumber:         cardNumber,
		CardExpiry:         expiryStr,
		CardStatus:         cardStatus,
		RecentTransactions: transactions,
	}

	jsonwrite.WriteJSON(w, http.StatusOK, dashboardUser)
}

func formatDate(dbTime string) string {
	t, err := time.Parse("2006-01-02 15:04:05", dbTime)
	if err == nil {
		return t.Format("Jan _2, 2006")
	}
	t2, err := time.Parse(time.RFC3339, dbTime)
	if err == nil {
		return t2.Format("Jan _2, 2006")
	}
	if len(dbTime) >= 10 {
		return dbTime[:10]
	}
	return dbTime
}
