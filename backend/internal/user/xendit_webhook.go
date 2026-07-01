package user

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/shopspring/decimal"
)

// XenditWebhookPayload represents the expected payload from Xendit Invoice webhook
type XenditWebhookPayload struct {
	ID                     string          `json:"id"`
	ExternalID             string          `json:"external_id"`
	UserID                 string          `json:"user_id"`
	IsHigh                 bool            `json:"is_high"`
	PaymentMethod          string          `json:"payment_method"`
	Status                 string          `json:"status"`
	MerchantName           string          `json:"merchant_name"`
	Amount                 decimal.Decimal `json:"amount"`
	PaidAmount             decimal.Decimal `json:"paid_amount"`
	BankCode               string          `json:"bank_code"`
	PaidAt                 string          `json:"paid_at"`
	PayerEmail             string          `json:"payer_email"`
	Description            string          `json:"description"`
	AdjustedReceivedAmount decimal.Decimal `json:"adjusted_received_amount"`
	FeesPaidAmount         decimal.Decimal `json:"fees_paid_amount"`
	Updated                string          `json:"updated"`
	Created                string          `json:"created"`
	Currency               string          `json:"currency"`
	PaymentChannel         string          `json:"payment_channel"`
	PaymentDestination     string          `json:"payment_destination"`
}

// XenditWebhook handles incoming webhook notifications from Xendit for invoice payments.
// It validates the callback token, processes the payment payload, and updates the user's balance.
// more info here: https://developers.xendit.co/docs/invoices#handling-invoice-completion-via-webhooks
func (h *Handler) XenditWebhook(w http.ResponseWriter, r *http.Request) {
	// Verify Xendit Callback Token
	xenditToken := os.Getenv("XENDIT_WEBHOOK_KEY")
	callbackToken := r.Header.Get("x-callback-token")

	// If no token is configured, skip validation for development, but it's recommended to have it
	if xenditToken != "" && callbackToken != xenditToken {
		log.Println("Invalid x-callback-token")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// read body from xendit invoice webhook
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("Failed to read body")
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// parse body
	var payload XenditWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		log.Println("Failed to parse webhook JSON:", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// switch case for payment status
	switch payload.Status {
	case "PAID", "SETTLED":
		externalID := payload.ExternalID // This maps to our topup_id

		// Start Database Transaction
		tx, err := h.DB.Begin()
		if err != nil {
			log.Println("Failed to start transaction:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		var topUp XenditWebhookPayload
		var cardNumber string
		//var amount float64
		var currentStatus string

		// Fetch the top-up record
		err = tx.QueryRow(`SELECT card_number, amount, status FROM top_ups WHERE topup_id = ?`, externalID).Scan(&cardNumber, &topUp.Amount, &currentStatus)
		if err != nil {
			log.Println("Failed to find top-up record or invalid external_id:", err)
			w.WriteHeader(http.StatusOK) // Ignore if not found
			return
		}

		// Prevent double processing
		if currentStatus == "completed" {
			log.Println("Top-up already completed, skipping.")
			w.WriteHeader(http.StatusOK)
			return
		}

		// Update the User's Balance
		if _, err := tx.Exec(`UPDATE cards SET balance = balance + ? WHERE card_number = ?`, topUp.Amount, cardNumber); err != nil {
			log.Println("Failed to update card balance:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Mark the top-up ledger as completed
		if _, err := tx.Exec(`UPDATE top_ups SET status = 'completed' WHERE topup_id = ?`, externalID); err != nil {
			log.Println("Failed to update top_ups status:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Mark the transaction ledger as completed. We find the related pending transaction by card_number, amount, and status.
		// We use LIMIT 1 to ensure we only update one pending transaction if there are duplicates.
		if _, err := tx.Exec(`UPDATE transactions SET status = 'completed', description = 'Successful topup via Xendit' WHERE card_number = ? AND transaction_type = 'topup' AND status = 'pending' AND amount = ? ORDER BY created_at DESC LIMIT 1`, cardNumber, topUp.Amount); err != nil {
			log.Println("Failed to update transactions status:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err := tx.Commit(); err != nil {
			log.Println("Failed to commit transaction:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		log.Printf("Successfully loaded ₱%s onto card %s via Xendit", topUp.Amount, cardNumber)

	// log if payment failed, expired, or canceled
	case "EXPIRED", "FAILED", "PENDING", "CANCELLED":
		log.Printf("Payment Status for external ID: %s, status: %s", payload.ExternalID, payload.Status)
		// Update the database records to failed so users see it as failed
		_, _ = h.DB.Exec(`UPDATE top_ups SET status = ? WHERE topup_id = ?`, payload.Status, payload.ExternalID)

		var description string
		switch payload.Status {
		case "EXPIRED":
			description = "topup via Xendit"
		case "FAILED":
			description = "topup failed via Xendit"
		case "PENDING":
			description = "topup pending via Xendit"
		case "CANCELLED":
			description = "topup cancelled via Xendit"
		}

		var cardNumber string
		var amount float64
		if err := h.DB.QueryRow(`SELECT card_number, amount FROM top_ups WHERE topup_id = ?`, payload.ExternalID).Scan(&cardNumber, &amount); err == nil {
			_, _ = h.DB.Exec(`UPDATE transactions SET status = ?,
			description = ?,
			WHERE card_number = ? AND transaction_type = 'topup' 
			AND status = 'pending' AND amount = ? 
			ORDER BY created_at DESC LIMIT 1`,
				payload.Status, description, cardNumber, amount)
		}
	}

	// Always 200 OK to Xendit
	w.WriteHeader(http.StatusOK)
}
