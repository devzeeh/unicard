package merchant

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	jsonwrite "unicard-go/backend/internal/pkg/handler"

	"github.com/shopspring/decimal"
)

type WithdrawRequest struct {
	Amount float64 `json:"amount"`
}

// WithdrawHandler handles the merchant's request to withdraw their available balance.
func (h *Handler) WithdrawHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("WithdrawHandler running...")

	// 1. Get username from URL
	username := r.PathValue("username")
	if username == "" {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "Username is required",
		})
		return
	}

	// 2. Parse request payload
	var req WithdrawRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "Invalid request payload",
		})
		return
	}

	if req.Amount <= 0 {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "Withdrawal amount must be greater than zero",
		})
		return
	}

	// 3. Fetch Merchant Info (ID, Settlement Details)
	var (
		merchantID        string
		settlementBank    *string
		settlementAccount *string
	)

	err := h.DB.QueryRow(`
		SELECT m.merchant_id, m.settlement_bank_name, m.settlement_account_number
		FROM merchants m
		JOIN users u ON m.user_id = u.user_id
		WHERE u.username = ? LIMIT 1
	`, username).Scan(&merchantID, &settlementBank, &settlementAccount)

	if err != nil {
		if err == sql.ErrNoRows {
			jsonwrite.WriteJSON(w, http.StatusNotFound, jsonwrite.APIResponse{
				Success: false,
				Message: "Merchant not found",
			})
			return
		}
		log.Println("Error fetching merchant for withdrawal:", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Internal server error fetching merchant details",
		})
		return
	}

	// Validate settlement details
	if settlementBank == nil || settlementAccount == nil {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "Please set up your settlement bank account details in your profile before withdrawing.",
		})
		return
	}

	// 4. Calculate Available Balance
	stats, err := h.GetMerchantIncomeStats(r.Context(), merchantID)
	if err != nil {
		log.Println("Error fetching income stats:", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Error fetching available balance",
		})
		return
	}

	withdrawAmount := decimal.NewFromFloat(req.Amount)
	if withdrawAmount.GreaterThan(stats.AvailableBalance) {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: fmt.Sprintf("Insufficient available balance. You can only withdraw up to %.2f", stats.AvailableBalance.InexactFloat64()),
		})
		return
	}

	// 5. Generate Transaction ID
	// Format: TXN-WD-timestamp
	txnID := fmt.Sprintf("TXN-WD-%d", time.Now().UnixNano())

	accountNum := *settlementAccount
	last4 := accountNum
	if len(accountNum) > 4 {
		last4 = accountNum[len(accountNum)-4:]
	}
	description := fmt.Sprintf("Withdrawal to %s ending in %s", *settlementBank, last4)
	
	insertTxnQuery := `
		INSERT INTO transactions (
			transaction_id, merchant_id, transaction_type, amount, status, description, card_number
		) VALUES (?, ?, 'withdrawal', ?, 'completed', ?, NULL)
	`
	_, err = h.DB.Exec(insertTxnQuery, txnID, merchantID, req.Amount, description)
	if err != nil {
		log.Println("Error inserting withdrawal transaction:", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Failed to process withdrawal",
		})
		return
	}

	jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{
		Success: true,
		Message: "Withdrawal processed successfully",
		Data: map[string]interface{}{
			"transaction_id": txnID,
			"amount":         req.Amount,
		},
	})
}
