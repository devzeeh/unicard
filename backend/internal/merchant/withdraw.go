package merchant

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	jsonwrite "unicard-go/backend/internal/pkg/handler"

	"github.com/shopspring/decimal"
	xendit "github.com/xendit/xendit-go/v7"
	"github.com/xendit/xendit-go/v7/payout"
)

type BankDetails struct {
	merchantID            string
	settlementBank        *string
	settlementAccountName *string
	settlementAccount     *string
}

type WithdrawRequest struct {
	Amount float64 `json:"amount"`
}

// WithdrawHandler handles the merchant's request to withdraw their available balance.
func (h *Handler) WithdrawHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("WithdrawHandler running...")

	// Get username from URL
	username := r.PathValue("username")
	if username == "" {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "Username is required",
		})
		return
	}

	// Parse request payload
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

	// Fetch Merchant Info (ID, Settlement Details)
	var bank BankDetails
	err := h.DB.QueryRow(`
		SELECT m.merchant_id, m.settlement_bank_name, m.settlement_account_name, m.settlement_account_number
		FROM merchants m
		JOIN users u ON m.user_id = u.user_id
		WHERE u.username = ? LIMIT 1
	`, username).Scan(&bank.merchantID, &bank.settlementBank, &bank.settlementAccountName, &bank.settlementAccount)

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
	if bank.settlementBank == nil || bank.settlementAccount == nil || bank.settlementAccountName == nil {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "Please set up your settlement bank account details in your profile before withdrawing.",
		})
		return
	}

	// Calculate Available Balance
	stats, err := h.GetMerchantIncomeStats(r.Context(), bank.merchantID)
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

	// Generate Transaction ID
	// Format: TXN-WD-timestamp
	txnID := fmt.Sprintf("TXN-WD-%d", time.Now().UnixNano())

	// getting the last 4 digit of bank account number
	accountNum := *bank.settlementAccount
	last4 := accountNum
	if len(accountNum) > 4 {
		last4 = accountNum[len(accountNum)-4:]
	}
	description := fmt.Sprintf("Withdrawal to %s ending in %s", *bank.settlementBank, last4)

	// Create Xendit Payout (v7)
	xenditClient := xendit.NewClient(os.Getenv("XENDIT_SECRET_KEY"))
	payoutClient := payout.NewPayoutApi(xenditClient)

	channelProps := payout.NewDigitalPayoutChannelProperties(*bank.settlementAccount)
	channelProps.SetAccountHolderName(*bank.settlementAccountName)

	createPayoutReq := payout.NewCreatePayoutRequest(
		txnID,
		"PH_"+strings.ToUpper(*bank.settlementBank), // Channel code e.g. "PH_MAYA"
		*channelProps,
		float32(req.Amount),
		"PHP",
	)
	createPayoutReq.SetDescription(description)

	_, _, payoutErr := payoutClient.CreatePayout(context.Background()).
		IdempotencyKey(txnID).
		CreatePayoutRequest(*createPayoutReq).
		Execute()

	if payoutErr != nil {
		log.Printf("Failed to create Xendit payout: %v", payoutErr.Error())
		log.Printf("Data sent to Xendit: %+v", createPayoutReq)
		log.Printf("Xendit Detailed Error: %v", payoutErr.Error())

		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Failed to process withdrawal via payment gateway",
		})
		return
	}

	insertTxnQuery := `
		INSERT INTO transactions (
			transaction_id, merchant_id, transaction_type, amount, status, description, card_number
		) VALUES (?, ?, 'withdrawal', ?, 'pending', ?, NULL)
	`
	_, err = h.DB.Exec(insertTxnQuery, txnID, bank.merchantID, req.Amount, description)
	if err != nil {
		log.Println("Error inserting withdrawal transaction:", err)
		// We could potentially try to cancel the disbursement here, or have a manual reconciliation process.
		// For now, we return an error.
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Withdrawal initiated, but failed to record in database. Please contact support.",
		})
		return
	}

	jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{
		Success: true,
		Message: "Withdrawal is being processed",
		Data: map[string]interface{}{
			"transaction_id": txnID,
			"amount":         req.Amount,
			"status":         "pending",
		},
	})
}
