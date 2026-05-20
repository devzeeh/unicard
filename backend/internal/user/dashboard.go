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

	// Check if session cookie is present
	cookie, err := r.Cookie("session_user_id")
	if err != nil || cookie.Value == "" {
		fmt.Println("No user session found in view, redirecting to login")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	h.Tpl.ExecuteTemplate(w, "dashboard.html", nil)
}

func (h *Handler) DashboardHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Dashboard JSON handler is running...")

	// Get session cookie
	cookie, err := r.Cookie("session_user_id")
	if err != nil || cookie.Value == "" {
		fmt.Println("No user session found, returning unauthorized JSON")
		jsonwrite.WriteJSON(w, http.StatusUnauthorized, jsonwrite.APIResponse{
			Success: false,
			Message: "Unauthorized",
		})
		return
	}
	userID := cookie.Value

	// Fetch user and card details
	var (
		id            int
		username      string
		fullName      string
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
    u.full_name,
    u.user_type,
    u.balance,
    u.loyalty_points,
    u.card_number,
    u.status
	FROM users u
	LEFT JOIN cards c 
    ON u.card_number = c.card_number
	WHERE u.user_id = ?
	`
	err = h.DB.QueryRow(stmt, userID).Scan(&id, &username, &fullName, &userType, &balance, &loyaltyPoints, &cardNumber, &cardStatus)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Printf("User %s not found in DB\n", userID)
		} else {
			fmt.Printf("Error fetching user %s from DB: %v\n", userID, err)
		}
		// Clear invalid session cookie and return unauthorized response
		http.SetCookie(w, &http.Cookie{
			Name:   "session_user_id",
			Value:  "",
			Path:   "/",
			MaxAge: -1,
		})
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
	rows, err := h.DB.Query("SELECT created_at, description, transaction_type, amount FROM transactions WHERE user_id = ? ORDER BY created_at DESC LIMIT 5", userID)
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
