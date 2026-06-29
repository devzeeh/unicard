package merchant

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
)

// XenditPayoutWebhookPayload represents the expected payload from Xendit Payout webhook
type XenditPayoutWebhookPayload struct {
	Event string `json:"event"`
	Data  struct {
		ReferenceID string  `json:"reference_id"`
		Status      string  `json:"status"` // SUCCEEDED, FAILED
		ChannelCode string  `json:"channel_code"`
		Amount      float64 `json:"amount"`
		FailureCode string  `json:"failure_code,omitempty"`
	} `json:"data"`
}

// XenditDisbursementWebhook handles incoming webhook notifications from Xendit for disbursements.
func (h *Handler) XenditDisbursementWebhook(w http.ResponseWriter, r *http.Request) {
	// Verify Xendit Callback Token
	xenditToken := os.Getenv("XENDIT_WEBHOOK_KEY")
	callbackToken := r.Header.Get("x-callback-token")

	if xenditToken != "" && callbackToken != xenditToken {
		log.Println("Invalid x-callback-token for disbursement")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("Failed to read body")
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	var payload XenditPayoutWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		log.Println("Failed to parse webhook JSON:", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	externalID := payload.Data.ReferenceID

	// Payout events: payout.succeeded, payout.failed
	switch payload.Event {
	case "payout.succeeded":
		_, err := h.DB.Exec(`UPDATE transactions SET status = 'completed' WHERE transaction_id = ? AND transaction_type = 'withdrawal' AND status = 'pending'`, externalID)
		if err != nil {
			log.Println("Failed to update withdrawal transaction to completed:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Printf("Successfully disbursed ₱%.2f for transaction %s", payload.Data.Amount, externalID)

	case "payout.failed":
		_, err := h.DB.Exec(`UPDATE transactions SET status = 'failed' WHERE transaction_id = ? AND transaction_type = 'withdrawal' AND status = 'pending'`, externalID)
		if err != nil {
			log.Println("Failed to update withdrawal transaction to failed:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Printf("Disbursement failed for transaction %s. Reason: %s", externalID, payload.Data.FailureCode)
	}

	// Always 200 OK to Xendit
	w.WriteHeader(http.StatusOK)
}
