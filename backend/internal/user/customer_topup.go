package user

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
	jsonwrite "unicard-go/backend/internal/pkg/handler"

	"github.com/stripe/stripe-go/v85"
	"github.com/stripe/stripe-go/v85/checkout/session"
)

type TopUpRequest struct {
	CardNumber string  `json:"card_number"`
	Amount     float64 `json:"amount"`
}

type TopUpRecord struct {
	TopupID        string  `json:"topup_id" db:"topup_id"`
	CardNumber     string  `json:"card_number" db:"card_number"`
	Amount         float64 `json:"amount" db:"amount"`
	ConvenienceFee float64 `json:"convenience_fee" db:"convenience_fee"`
	GatewayCost    float64 `json:"gateway_cost" db:"gateway_cost"`
	PaymentMethod  string  `json:"payment_method" db:"payment_method"`
}

func (h *Handler) TopUpView(w http.ResponseWriter, r *http.Request) {
	fmt.Println("TopUp view is running...")

	username := r.PathValue("username")
	data := struct {
		Username string
	}{
		Username: username,
	}

	h.Tpl.ExecuteTemplate(w, "customer_topup.html", data)
}

func (h *Handler) CreateStripeCheckoutSession(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")

	var req TopUpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "Invalid request payload",
		})
		return
	}
	if req.Amount < 50 {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "Minimum topup amount is 50 PHP",
		})
		return
	}

	// Fetch card number and email securely from DB instead of trusting the frontend
	var cardNumber, email string
	err := h.DB.QueryRow(`
		SELECT c.card_number, u.email 
		FROM cards c 
		JOIN users u ON c.user_id = u.user_id 
		WHERE u.username = ? LIMIT 1
	`, username).Scan(&cardNumber, &email)

	if err != nil {
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Failed to find card for user",
		})
		return
	}

	// Stripe takes amount in cents/lowest denomination. For PHP, it's centavos.
	topupCentavos := int64(req.Amount * 100) // ex: 100 -> 10000
	feeCentavos := int64(15 * 100)           // 15 pesos for fee
	domain := "http://" + os.Getenv("SERVER_PORT") + os.Getenv("PORT")
	// Fallback if domain is malformed
	if domain == "http://" {
		domain = "http://localhost:3000"
	}
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",
		}),
		Metadata: map[string]string{
			"card_number":     cardNumber, // Use securely fetched card number from DB!
			"base_amount":     fmt.Sprintf("%.2f", req.Amount),
			"convenience_fee": "15.00",
		},
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String("php"),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name:        stripe.String("Unicard Top-Up"),
						Description: stripe.String("Top up your unicard with stripe payment"),
					},
					UnitAmount: stripe.Int64(topupCentavos),
				},
				Quantity: stripe.Int64(1),
			},
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String("php"),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name:        stripe.String("Convienence Fee"),
						Description: stripe.String("15 peso convienence fee"),
					},
					UnitAmount: stripe.Int64(feeCentavos),
				},
				Quantity: stripe.Int64(1),
			},
		},
		Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL: stripe.String(domain + "/u/" + username + "/dashboard"),
		CancelURL:  stripe.String(domain + "/u/" + username + "/topup"),
	}

	// creating the checkout session
	s, err := session.New(params)

	// Handle error if session creation fails
	if err != nil {
		log.Println("Failed to create checkout session:", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Failed to create checkout session",
		})
		return
	}

	// Handle success case
	jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{
		Success: true,
		Message: "Checkout session created successfully",
		Data:    map[string]string{"url": s.URL},
	})

	log.Println("Checkout session created successfully:", s.URL)
}

// save topup tp database
func (h *Handler) SaveTopUpToDatabase(w http.ResponseWriter, r *http.Request) {
	log.Println("SaveTopUpToDatabase running...")

	var req TopUpRecord
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "Invalid request payload",
		})
		return
	}

	// Generate Unique IDs for both ledgers
	topupID := fmt.Sprintf("TOPUP-%d", time.Now().UnixNano())
	transactionID := fmt.Sprintf("TX-%d", time.Now().UnixNano())

	// Start the Database Transaction
	tx, err := h.DB.Begin()
	if err != nil {
		log.Println("Failed to start transaction:", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Failed to start transaction",
		})
		return
	}

	// SAFETY NET: This automatically rolls back if the function exits before tx.Commit()
	defer tx.Rollback()

	// Update User Balance (cards table)
	if _, err := tx.Exec(`UPDATE cards SET balance = balance + ? WHERE card_number = ?`, req.Amount, req.CardNumber); err != nil {
		log.Println("Failed to update user balance:", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Failed to update user balance",
		})
		return // The defer statement will automatically handle the rollback
	}

	// Insert into Loading Ledger (top_ups table)
	queryTopUp := `INSERT INTO top_ups (topup_id, card_number, amount, convenience_fee, gateway_cost, payment_method) VALUES (?, ?, ?, ?, ?, ?)`
	if _, err := tx.Exec(queryTopUp, topupID, req.CardNumber, req.Amount, req.ConvenienceFee, req.GatewayCost, req.PaymentMethod); err != nil {
		log.Println("Failed to record top-up:", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Failed to record top-up",
		})
		return
	}

	// Insert into Spending Ledger (transactions table)
	// *Note: Adjust 'category' if your enum doesn't include 'top_up'
	queryTx := `INSERT INTO transactions (transaction_id, card_number, merchant_id, terminal_id, transaction_type, amount, service_fee, processed_by) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	// "Stripe" as the merchant_id since this is an internal load, not a retail/Fare purchase
	if _, err := tx.Exec(queryTx, transactionID, req.CardNumber, sql.NullString{}, sql.NullString{}, "topup", req.Amount, req.ConvenienceFee, sql.NullString{}); err != nil {
		log.Println("Failed to record global transaction:", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Failed to record global transaction",
		})
		return
	}

	// THE MOST IMPORTANT PART: Commit the transaction!
	if err := tx.Commit(); err != nil {
		log.Println("Failed to finalize database changes:", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Failed to finalize database changes",
		})
		return
	}

	// Success Response
	jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{
		Success: true,
		Message: "Top-up fully processed and saved",
	})
	log.Println("Transaction saved successfully")
}
