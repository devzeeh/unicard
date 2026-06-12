package user

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/stripe/stripe-go/v85"
	"github.com/stripe/stripe-go/v85/webhook"
)

// StripeWebhook is the endpoint Stripe will POST to when a payment succeeds
func (h *Handler) StripeWebhook(w http.ResponseWriter, r *http.Request) {
	const MaxBodyBytes = int64(65536)
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusServiceUnavailable)
		return
	}

	// 1. Verify the request actually came from Stripe
	endpointSecret := os.Getenv("STRIPE_WEBHOOK_SECRET") // Get this from your Stripe Dashboard
	signatureHeader := r.Header.Get("Stripe-Signature")

	event, err := webhook.ConstructEventWithOptions(payload, signatureHeader, endpointSecret, webhook.ConstructEventOptions{
		IgnoreAPIVersionMismatch: true,
	})
	if err != nil {
		log.Println("Webhook signature verification failed:", err)
		http.Error(w, "Bad signature", http.StatusBadRequest)
		return
	}

	// 2. Only process successful checkout sessions
	if event.Type == "checkout.session.completed" {
		var session stripe.CheckoutSession
		err := json.Unmarshal(event.Data.Raw, &session)
		if err != nil {
			log.Println("Error parsing webhook JSON:", err)
			http.Error(w, "Error parsing webhook JSON", http.StatusBadRequest)
			return
		}

		// 3. Extract the hidden data we passed earlier
		cardNumber := session.Metadata["card_number"]
		amountStr := session.Metadata["base_amount"]
		feeStr := session.Metadata["convenience_fee"]

		baseAmount, _ := strconv.ParseFloat(amountStr, 64)
		convenienceFee, _ := strconv.ParseFloat(feeStr, 64)

		// 4. Safely write to the database
		err = h.processSuccessfulTopUp(cardNumber, baseAmount, convenienceFee)
		if err != nil {
			log.Println("Database transaction failed:", err)
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		log.Printf("Successfully loaded ₱%.2f onto card %s", baseAmount, cardNumber)
	}

	// Stripe expects a 200 OK immediately to know we received the event
	w.WriteHeader(http.StatusOK)
}

// Helper function to handle the database transaction (This replaces your old HTTP handler)
func (h *Handler) processSuccessfulTopUp(cardNumber string, baseAmount float64, convenienceFee float64) error {
	topupID := fmt.Sprintf("TOPUP-%d", time.Now().UnixNano())
	transactionID := fmt.Sprintf("TX-%d", time.Now().UnixNano())

	tx, err := h.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Update User Balance
	if _, err := tx.Exec(`UPDATE cards SET balance = balance + ? WHERE card_number = ?`, baseAmount, cardNumber); err != nil {
		return err
	}

	// Insert into Loading Ledger (top_ups table)
	// NOTE: handled_by is nil because this is automated
	queryTopUp := `INSERT INTO top_ups (topup_id, card_number, amount, convenience_fee, payment_method) VALUES (?, ?, ?, ?, ?)`
	if _, err := tx.Exec(queryTopUp, topupID, cardNumber, baseAmount, convenienceFee, "stripe"); err != nil {
		return err
	}

	// Insert into Spending Ledger (transactions table)
	// NOTE: Ensure "SYS_STRIPE", "SYS_TERM" and "SYS_BOT" exist in your merchants, terminals and users tables to prevent Foreign Key crashes!
	queryTx := `INSERT INTO transactions (transaction_id, card_number,  transaction_type, amount, service_fee) VALUES (?, ?, ?, ?, ?)`
	if _, err := tx.Exec(queryTx, transactionID, cardNumber, "topup", baseAmount, convenienceFee); err != nil {
		return err
	}

	return tx.Commit()
}
