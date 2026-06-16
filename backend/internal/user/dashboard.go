package user

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	jsonwrite "unicard-go/backend/internal/pkg/handler"
)

type Transaction struct {
	TransactionID string          `json:"transaction_id"`
	TerminalID    string          `json:"terminal_id"`
	Date          string          `json:"date" db:"date"`
	Time          string          `json:"time"`
	Description   string          `json:"description" db:"description"`
	Type          string          `json:"type" db:"transaction_type"`
	Amount        float64         `json:"amount" db:"transaction_amount"`
	Status        string          `json:"status" db:"status"`
	MerchantName  string          `json:"merchant_name"`
	MerchantID    string          `json:"merchant_id"`
	ServiceFee    float64         `json:"service_fee"`
	PointsEarned  decimal.Decimal `json:"points_earned"`
}

// DashboardUser info struct for the user dashboard view
type DashboardUser struct {
	ID                 int             `json:"id,omitempty" db:"id"`
	UserID             string          `json:"user_id,omitempty" db:"user_id"`
	Username           string          `json:"username" db:"username"`
	Name               string          `json:"name" db:"name"`
	Email              string          `json:"email" db:"email"`
	PendingEmail       string          `json:"pending_email"`
	Phone              string          `json:"phone" db:"phone"`
	Initials           string          `json:"initials"`
	Balance            float64         `json:"balance" db:"balance"`
	LoyaltyPoints      decimal.Decimal `json:"loyalty_points" db:"loyalty_points"`
	AccountType        string          `db:"account_type" json:"account_type"`
	CardNumber         string          `json:"card_number"`
	CardExpiry         string          `json:"card_expiry"`
	CardStatus         string          `json:"card_status"`
	RecentTransactions []Transaction   `json:"recent_transactions"` // Add recent transactions to the dashboard response
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
		pendingEmail  string
		phone         string
		userType      string
		balance       float64
		loyaltyPoints decimal.Decimal
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
			COALESCE(u.pending_email, ''),
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
	err := h.DB.QueryRow(stmt, userID).Scan(&id, &username, &fullName, &email, &pendingEmail, &phone, &userType, &balance, &loyaltyPoints, &cardNumber, &expiryDate, &cardStatus)
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
    SELECT t.transaction_id, t.description, t.created_at, COALESCE(t.transaction_type, ''), t.amount, COALESCE(t.terminal_id, ''), COALESCE(t.status, ''), m.business_name, m.merchant_id, COALESCE(t.points_earned, 0)
    FROM transactions t 
    JOIN cards c ON t.card_number = c.card_number 
    JOIN users u ON c.user_id = u.user_id
    LEFT JOIN merchants m ON t.merchant_id = m.merchant_id
    WHERE u.username = ? 
    ORDER BY t.created_at DESC LIMIT 5
`
	rows, err := h.DB.Query(txnQuery, userID)
	var transactions []Transaction
	if err != nil {
		fmt.Printf("Error fetching transactions: %v\n", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var t Transaction
			var createdAt string
			var description sql.NullString
			var businessName sql.NullString
			var merchantId sql.NullString
			var pointsEarned decimal.Decimal
			if err := rows.Scan(
				&t.TransactionID,
				&description,
				&createdAt,
				&t.Type,
				&t.Amount,
				&t.TerminalID,
				&t.Status,
				&businessName,
				&merchantId,
				&pointsEarned,
			); err != nil {
				fmt.Printf("Error scanning transaction row: %v\n", err)
				continue
			}
			t.Date = formatDate(createdAt)
			t.Time = formatTime(createdAt)
			if description.Valid {
				t.Description = description.String
			} else {
				t.Description = ""
			}
			
			if businessName.Valid && businessName.String != "" {
				t.MerchantName = businessName.String
			} else {
				t.MerchantName = t.Description
			}
			
			if merchantId.Valid {
				t.MerchantID = merchantId.String
			} else {
				t.MerchantID = "N/A"
			}

			t.ServiceFee = 0.00
			t.PointsEarned = pointsEarned

			transactions = append(transactions, t)
		}
	}

	dashboardUser := DashboardUser{
		ID:                 id,
		UserID:             userID,
		Username:           username,
		Name:               fullName,
		Email:              email,
		PendingEmail:       pendingEmail,
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

func formatTime(dbTime string) string {
	t, err := time.Parse("2006-01-02 15:04:05", dbTime)
	if err == nil {
		return t.Format("03:04 PM")
	}
	t2, err := time.Parse(time.RFC3339, dbTime)
	if err == nil {
		return t2.Format("03:04 PM")
	}
	if len(dbTime) > 10 {
		return dbTime[11:16]
	}
	return ""
}
