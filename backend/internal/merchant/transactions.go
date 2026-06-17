package merchant

import (
	"log"
	"net/http"
	jsonwrite "unicard-go/backend/internal/pkg/handler"
)

// MerchantTransactionsView renders the merchant_transactions.html template
func (h *Handler) MerchantTransactionsView(w http.ResponseWriter, r *http.Request) {
	log.Println("MerchantTransactionsView running...")
	data := MerchantPageData{
		Page:     "transactions",
		Username: r.PathValue("username"),
	}
	err := h.Tpl.ExecuteTemplate(w, "merchant_transactions.html", data)
	if err != nil {
		log.Println("Error executing template:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (h *Handler) TransactionHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("TransactionHandler running...")

	ctx := r.Context()

	username := r.PathValue("username")
	if username == "" {
		log.Println("Username is required")
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "username is required",
		})
		return
	}

	// Resolve merchant_id from username
	var merchantID string
	err := h.DB.QueryRowContext(ctx, `
		SELECT m.merchant_id 
		FROM merchants m
		JOIN users u ON m.user_id = u.user_id
		WHERE u.username = ?
		LIMIT 1`,
		username).Scan(&merchantID)
	if err != nil {
		log.Println("Error fetching merchant ID:", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Error fetching merchant data",
		})
		return
	}

	// getting all transactions
	query := `SELECT 
			transaction_id, card_number,
			merchant_id, terminal_id,
			transaction_type, amount,
			points_earned, service_fee,
			net_merchant_payout, processed_by,
			status, description, created_at
		FROM transactions 
		WHERE merchant_id = ? 
		ORDER BY created_at DESC LIMIT 15`
	rows, err := h.DB.QueryContext(ctx, query, merchantID)
	if err != nil {
		log.Println("Error fetching transactions:", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Error fetching transactions",
		})
		return
	}
	defer rows.Close()

	var transactions []MerchantTransaction
	for rows.Next() {
		var t MerchantTransaction
		if err := rows.Scan(
			&t.TransactionID, &t.CardNumber,
			&t.MerchantID, &t.TerminalID,
			&t.TransactionType, &t.Amount,
			&t.Points, &t.ServiceFee,
			&t.NetMerchantPayout, &t.ProcessedBy,
			&t.Status, &t.Description, &t.Date); err != nil {
			log.Println("Error scanning transaction row:", err)
			continue
		}
		transactions = append(transactions, t)
	}

	jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{
		Success: true,
		Message: "Transactions fetched successfully",
		Data:    transactions,
	})
}
