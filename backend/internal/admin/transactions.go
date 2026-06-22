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
				t.transaction_id, 
				COALESCE(t.terminal_id, ''), 
				DATE(t.created_at) as date, 
				TIME(t.created_at) as time, 
				COALESCE(t.transaction_type, ''), 
				t.amount, 
				COALESCE(t.status, ''),
				COALESCE(t.description, ''),
				COALESCE(m.business_name, ''),
				COALESCE(m.merchant_id, ''),
				COALESCE(t.points_earned, 0),
				c.card_number,
				COALESCE(u.name, 'Unknown Customer') as customer_name
			FROM transactions t
			LEFT JOIN cards c ON t.card_number = c.card_number 
			LEFT JOIN users u ON c.user_id = u.user_id
			LEFT JOIN merchants m ON t.merchant_id = m.merchant_id
			ORDER BY t.created_at DESC
		`
	rows, err := h.DB.Query(txnQuery)

	type TxnResponse struct {
		TransactionID string          `json:"transaction_id"`
		TerminalID    string          `json:"terminal_id"`
		Date          string          `json:"date"`
		Time          string          `json:"time"`
		Description   string          `json:"description"`
		Type          string          `json:"type"`
		Amount        float64         `json:"amount"`
		Status        string          `json:"status"`
		MerchantName  string          `json:"merchant_name"`
		MerchantID    string          `json:"merchant_id"`
		ServiceFee    float64         `json:"service_fee"`
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

			err := rows.Scan(&t.TransactionID, &t.TerminalID, &t.Date, &t.Time, &t.Type, &t.Amount, &t.Status, &description, &businessName, &merchantId, &pointsEarned, &cardNumber, &customerName)
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

			t.ServiceFee = 0.00
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
