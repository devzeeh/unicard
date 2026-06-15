package user

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	jsonwrite "unicard-go/backend/internal/pkg/handler"
)

// TransactionView renders the transaction.html template
func (h *Handler) TransactionView(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Transaction view is running...")

	username := r.PathValue("username")
	data := struct {
		Username string
	}{
		Username: username,
	}

	h.Tpl.ExecuteTemplate(w, "transaction.html", data)
}

// TransactionsJSONHandler returns the user's transactions as JSON
func (h *Handler) TransactionsJSONHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("TransactionsJSONHandler is running...")

	username := r.PathValue("username")
	if username == "" {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "user is required",
		})
		return
	}

	txnQuery := `
	SELECT t.transaction_id, t.created_at, m.business_name, t.transaction_type, t.amount, t.terminal_id, t.description, t.status, c.card_number, m.merchant_id
		FROM transactions t 
		JOIN cards c ON t.card_number = c.card_number 
		JOIN users u ON c.user_id = u.user_id
		LEFT JOIN merchants m ON t.merchant_id = m.merchant_id 
		WHERE u.username = ? 
		ORDER BY t.created_at DESC
	`
	rows, err := h.DB.Query(txnQuery, username)

	type TxnResponse struct {
		TransactionID string  `json:"transaction_id"`
		TerminalID    string  `json:"terminal_id"`
		Date          string  `json:"date"`
		Time          string  `json:"time"`
		Description   string  `json:"description"`
		Type          string  `json:"type"`
		Amount        float64 `json:"amount"`
		Status        string  `json:"status"`
		MerchantName  string  `json:"merchant_name"`
		MerchantID    string  `json:"merchant_id"`
		ServiceFee    float64 `json:"service_fee"`
		PointsEarned  int     `json:"points_earned"`
		Sender        string  `json:"sender"`
		Receiver      string  `json:"receiver"`
	}

	var transactions []TxnResponse
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var t TxnResponse
			var createdAt string
			var businessName sql.NullString
			var terminalId sql.NullString
			var dbDescription sql.NullString
			var dbStatus sql.NullString
			var cardNumber string
			var merchantId sql.NullString
			if err := rows.Scan(&t.TransactionID, &createdAt, &businessName, &t.Type, &t.Amount, &terminalId, &dbDescription, &dbStatus, &cardNumber, &merchantId); err == nil {

				if dbStatus.Valid && dbStatus.String != "" {
					t.Status = dbStatus.String
				} else {
					t.Status = "Completed"
				}

				t.Date = formatDate(createdAt) // Uses formatDate from dashboard.go
				t.Time = formatTime(createdAt) // Uses formatTime from dashboard.go

				if terminalId.Valid {
					t.TerminalID = terminalId.String
				} else {
					t.TerminalID = "N/A"
				}

				if dbDescription.Valid && dbDescription.String != "" {
					t.Description = dbDescription.String
				} else if businessName.Valid {
					t.Description = businessName.String
				} else if t.Type == "topup" {
					t.Description = "Stripe Top-Up"
				} else {
					t.Description = "Transaction"
				}

				if businessName.Valid {
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
				if strings.ToLower(t.Type) == "payment" && t.Amount >= 100 {
					t.PointsEarned = int(t.Amount / 100)
				} else {
					t.PointsEarned = 0
				}

				// Calculate Sender, Receiver, ServiceFee, PointsEarned
				isPayment := strings.ToLower(t.Type) == "payment"
				merchantStr := t.Description
				if t.TerminalID != "" && t.TerminalID != "N/A" {
					merchantStr += " (Terminal: " + t.TerminalID + ")"
				}
				cardStr := "My UniCard"
				if len(cardNumber) >= 4 {
					cardStr += " (**** " + cardNumber[len(cardNumber)-4:] + ")"
				}

				if isPayment {
					t.Sender = cardStr
					t.Receiver = merchantStr
					if t.Amount >= 100 {
						t.PointsEarned = int(t.Amount / 100)
					}
				} else {
					t.Sender = merchantStr
					t.Receiver = cardStr
				}
				t.ServiceFee = 0.00

				transactions = append(transactions, t)
			} else {
				fmt.Printf("Scan error: %v\n", err)
			}
		}
	} else {
		fmt.Printf("Error fetching transactions: %v\n", err)
	}

	response := struct {
		Success      bool          `json:"success"`
		Transactions []TxnResponse `json:"transactions"`
	}{
		Success:      true,
		Transactions: transactions,
	}

	jsonwrite.WriteJSON(w, http.StatusOK, response)
}
