package user

import (
	"database/sql"
	"fmt"
	"net/http"
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

	// Fetch transactions
	txnQuery := `
		SELECT t.transaction_id, t.created_at, m.business_name, t.transaction_type, t.amount, t.terminal_id
		FROM transactions t 
		JOIN cards c ON t.card_number = c.card_number 
		JOIN users u ON c.user_id = u.user_id
		LEFT JOIN merchants m ON t.merchant_id = m.user_id 
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
	}

	var transactions []TxnResponse
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var t TxnResponse
			var createdAt string
			var businessName sql.NullString
			if err := rows.Scan(&t.TransactionID, &createdAt, &businessName, &t.Type, &t.Amount, &t.TerminalID); err == nil {
				t.Status = "Completed"
				t.Date = formatDate(createdAt) // Uses formatDate from dashboard.go
				t.Time = formatTime(createdAt) // Uses formatTime from dashboard.go
				if businessName.Valid {
					t.Description = businessName.String
				} else {
					t.Description = "Transaction"
				}
				transactions = append(transactions, t)
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
