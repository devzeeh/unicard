package user

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// XenditWebhookPayload represents the expected payload from Xendit Invoice webhook
type XenditWebhookPayload struct {
	ID                     string  `json:"id"`
	ExternalID             string  `json:"external_id"`
	UserID                 string  `json:"user_id"`
	IsHigh                 bool    `json:"is_high"`
	PaymentMethod          string  `json:"payment_method"`
	Status                 string  `json:"status"`
	MerchantName           string  `json:"merchant_name"`
	Amount                 float64 `json:"amount"`
	PaidAmount             float64 `json:"paid_amount"`
	BankCode               string  `json:"bank_code"`
	PaidAt                 string  `json:"paid_at"`
	PayerEmail             string  `json:"payer_email"`
	Description            string  `json:"description"`
	AdjustedReceivedAmount float64 `json:"adjusted_received_amount"`
	FeesPaidAmount         float64 `json:"fees_paid_amount"`
	Updated                string  `json:"updated"`
	Created                string  `json:"created"`
	Currency               string  `json:"currency"`
	PaymentChannel         string  `json:"payment_channel"`
	PaymentDestination     string  `json:"payment_destination"`
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
		externalID := payload.ExternalID
		// check if external id is for top up
		if strings.HasPrefix(externalID, "TOPUP") && len(externalID) > 21 {
			// TOPUP is 5 chars, CardNumber is strictly 16 chars
			// indexes 5 to 21
			cardNumber := externalID[5:21]

			remainder := externalID[21:]
			// find first dot
			firstDot := strings.Index(remainder, ".")
			if firstDot == -1 || len(remainder) < firstDot+3 {
				log.Println("Invalid external ID format: missing base amount")
				http.Error(w, "Invalid external ID data", http.StatusBadRequest)
				return
			}
			baseAmountStr := remainder[:firstDot+3] // e.g. 50.00

			remainder2 := remainder[firstDot+3:]
			// find second dot
			secondDot := strings.Index(remainder2, ".")
			if secondDot == -1 || len(remainder2) < secondDot+3 {
				log.Println("Invalid external ID format: missing fee amount")
				http.Error(w, "Invalid external ID data", http.StatusBadRequest)
				return
			}
			convenienceFeeStr := remainder2[:secondDot+3] // e.g. 15.00

			// convert string to float
			baseAmount, err1 := strconv.ParseFloat(baseAmountStr, 64)
			convenienceFee, err2 := strconv.ParseFloat(convenienceFeeStr, 64)

			if err1 != nil || err2 != nil {
				log.Println("Failed to parse amounts from external ID")
				http.Error(w, "Invalid external ID data", http.StatusBadRequest)
				return
			}

			// process successful top up
			err := h.processSuccessfulTopUp(cardNumber, baseAmount, convenienceFee, 0.0, "xendit", payload.ExternalID)
			if err != nil {
				log.Println("Database transaction failed:", err)
				w.WriteHeader(http.StatusOK) // don't return 500 to Xendit
				return
			}

			log.Printf("Successfully loaded ₱%.2f onto card %s via Xendit", baseAmount, cardNumber)
		} else {
			log.Println("Unrecognized External ID format:", payload.ExternalID)
		}
	// log if payment failed or expired
	case "EXPIRED", "FAILED":
		log.Printf("Payment failed for external ID: %s, status: %s", payload.ExternalID, payload.Status)
	}

	// Always 200 OK to Xendit
	w.WriteHeader(http.StatusOK)
}

// Helper function to handle the database transaction
func (h *Handler) processSuccessfulTopUp(cardNumber string, baseAmount float64, convenienceFee float64, paymentGatewayCost float64, paymentMethod string, externalID string) error {
	topupID := fmt.Sprintf("TOPUP-%d", time.Now().UnixNano())
	transactionID := fmt.Sprintf("TX-%d", time.Now().UnixNano())

	// begin transaction to make sure that all database operations are performed in a single unit of work
	// if any operation fails, the entire transaction will be rolled back and the database will be left unchanged
	tx, err := h.DB.Begin()
	if err != nil {
		return err
	}
	// if transaction fails, rollback not close the transaction
	// Rollback() will do nothing if the transaction is already closed (committed or rolled back)
	// so that's why we use defer tx.Rollback()
	// This pattern ensures that the database connection is properly managed
	// and that any errors during the transaction are handled gracefully.
	defer tx.Rollback()

	// Check if already processed
	// this prevents double top up if xendit sends multiple webhooks or network delay causes multiple requests
	// using external_id as unique identifier for each top up
	var exists bool
	if err := tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM top_ups WHERE external_id = ?)`, externalID).Scan(&exists); err != nil {
		return err
	}
	if exists {
		return nil // already processed, skip silently
	}

	// Update User Balance
	// Using atomic update to prevent race conditions
	if _, err := tx.Exec(`UPDATE cards SET balance = balance + ? WHERE card_number = ?`, baseAmount, cardNumber); err != nil {
		return err
	}

	// Insert into Loading Ledger (top_ups table)
	// top_ups table is our loading ledger
	queryTopUp := `INSERT INTO top_ups (topup_id, card_number, amount, convenience_fee, gateway_cost, payment_method, handled_by, external_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	if _, err := tx.Exec(queryTopUp, topupID, cardNumber, baseAmount, convenienceFee, paymentGatewayCost, paymentMethod, "payment gateway", externalID); err != nil {
		return err
	}

	// Insert into Spending Ledger (transactions table)
	// transactions table is our spending ledger
	queryTx := `INSERT INTO transactions (transaction_id, card_number, merchant_id, terminal_id, transaction_type, amount, service_fee, processed_by, description) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	if _, err := tx.Exec(queryTx, transactionID, cardNumber, "xendit", "xendit", "topup", baseAmount, convenienceFee, "xendit", "Successful topup via Xendit"); err != nil {
		return err
	}

	// Commit the transaction if everything is fine
	return tx.Commit()
}
