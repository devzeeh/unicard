package admin

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	jsonwrite "unicard-go/backend/internal/pkg/handler"

	"github.com/shopspring/decimal"
)

// TransactionsView serves the System Transactions page
func (h *Handler) TransactionsView(w http.ResponseWriter, r *http.Request) {
	data := AdminPageData{Page: "transactions", Username: r.PathValue("username")}
	err := h.Tpl.ExecuteTemplate(w, "transactions.html", data)
	if err != nil {
		log.Printf("Template execution error: %v", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Internal Server Error",
		})
	}
}

// AllTransactionsJSONHandler returns all transactions across the system as JSON
func (h *Handler) AllTransactionsJSONHandler(w http.ResponseWriter, r *http.Request) {
	txnQuery := `
			SELECT 
				transaction_id, 
				terminal_id,
				date,
				time,
				transaction_type,
				amount,
				status,
				description,
				business_name,
				merchant_id,
				points_earned,
				card_number,
				customer_name,
				service_fee
			FROM (
				SELECT 
					t.transaction_id, 
					COALESCE(t.terminal_id, '') AS terminal_id,
					DATE(t.created_at) as date,
					TIME(t.created_at) as time,
					COALESCE(t.transaction_type, '') AS transaction_type,
					COALESCE(t.amount, 0.00) AS amount,
					COALESCE(t.status, '') AS status,
					COALESCE(t.description, '') AS description,
					COALESCE(m.business_name, '') AS business_name,
					COALESCE(m.merchant_id, '') AS merchant_id,
					COALESCE(t.points_earned, 0) AS points_earned,
					COALESCE(c.card_number, '') AS card_number,
					COALESCE(u.name, 'Unknown Customer') as customer_name,
					COALESCE(t.service_fee, 0) AS service_fee,
					t.created_at
				FROM transactions t
				LEFT JOIN cards c ON t.card_number = c.card_number 
				LEFT JOIN users u ON t.user_id = u.user_id
				LEFT JOIN merchants m ON t.merchant_id = m.merchant_id
				
				UNION ALL
				
				SELECT 
					CONCAT('LOG-', ual.id) AS transaction_id, 
					'' AS terminal_id,
					DATE(ual.created_at) as date,
					TIME(ual.created_at) as time,
					ual.activity_type AS transaction_type,
					0.00 AS amount,
					ual.status AS status,
					COALESCE(ual.description, '') AS description,
					COALESCE(m.business_name, '') AS business_name,
					COALESCE(m.merchant_id, '') AS merchant_id,
					0 AS points_earned,
					'' AS card_number,
					COALESCE(u.name, 'Unknown Customer') as customer_name,
					0.00 AS service_fee,
					ual.created_at
				FROM user_activity_logs ual
				LEFT JOIN users u ON ual.user_id = u.user_id
				LEFT JOIN merchants m ON ual.user_id = m.user_id
			) AS combined_txns
			ORDER BY created_at DESC
		`
	rows, err := h.Store.Query(txnQuery)

	type TxnResponse struct {
		TransactionID string          `json:"transaction_id"`
		TerminalID    string          `json:"terminal_id"`
		Date          string          `json:"date"`
		Time          string          `json:"time"`
		Description   string          `json:"description"`
		Type          string          `json:"type"`
		Amount        decimal.Decimal `json:"amount"`
		Status        string          `json:"status"`
		MerchantName  string          `json:"merchant_name"`
		MerchantID    string          `json:"merchant_id"`
		ServiceFee    decimal.Decimal `json:"service_fee"`
		PointsEarned  decimal.Decimal `json:"points_earned"`
		Sender        string          `json:"sender"`
		Receiver      string          `json:"receiver"`
		CustomerName  string          `json:"customer_name"`
		Source        string          `json:"source"`
		CardNumber    string          `json:"card_number"`
	}

	var transactions []TxnResponse
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var t TxnResponse
			var description string
			var businessName string
			var merchantId string
			var pointsEarned decimal.Decimal
			var cardNumber *string
			var customerName string
			var serviceFee decimal.Decimal

			err := rows.Scan(&t.TransactionID, &t.TerminalID, &t.Date, &t.Time, &t.Type, &t.Amount, &t.Status, &description, &businessName, &merchantId, &pointsEarned, &cardNumber, &customerName, &serviceFee)
			if err != nil {
				fmt.Printf("Error scanning transaction row: %v\n", err)
				continue
			}
			t.Description = description
			t.CustomerName = customerName

			if businessName != "" {
				t.MerchantName = businessName
			} else {
				t.MerchantName = "Transaction"
			}

			if merchantId != "" {
				t.MerchantID = merchantId
			} else {
				t.MerchantID = "N/A"
			}

			t.ServiceFee = serviceFee
			t.PointsEarned = pointsEarned

			// Calculate Source / Details
			if t.TerminalID != "" && t.TerminalID != "N/A" {
				t.Source = "Terminal: " + t.TerminalID
			} else {
				t.Source = "Online/System"
			}

			if cardNumber != nil {
				t.CardNumber = *cardNumber
			} else {
				t.CardNumber = "N/A"
			}

			// Calculate Sender, Receiver
			isPayment := strings.ToLower(t.Type) == "payment"
			merchantStr := t.MerchantName
			if t.TerminalID != "" && t.TerminalID != "N/A" {
				merchantStr += " (Terminal: " + t.TerminalID + ")"
			}

			cardStr := customerName
			if cardNumber != nil && len(*cardNumber) >= 4 {
				cardStr += " (**** " + (*cardNumber)[len(*cardNumber)-4:] + ")"
			}

			if isPayment {
				t.Sender = cardStr
				t.Receiver = merchantStr
			} else {
				t.Sender = merchantStr
				t.Receiver = cardStr
			}

			transactions = append(transactions, t)
		}
	} else {
		fmt.Printf("Error querying transactions: %v\n", err)
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
