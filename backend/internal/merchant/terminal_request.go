package merchant

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
	jsonwrite "unicard-go/backend/internal/pkg/handler"
	structs "unicard-go/backend/internal/pkg/structs"
)

// RequestTerminalHandler allows a merchant to request a terminal assignment.
// Payload: { "terminal_sn": "optional-serial" }
func (h *Handler) RequestTerminalHandler(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	if username == "" {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{Success: false, Message: "Username required"})
		return
	}

	var payload struct {
		TerminalSN string `json:"terminal_sn"`
		Notes      string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{Success: false, Message: "Invalid request payload"})
		return
	}

	// resolve merchant_id
	var merchantID string
	err := h.Store.QueryRow("SELECT merchant_id FROM merchants WHERE user_id = (SELECT user_id FROM users WHERE username = ?)", username).Scan(&merchantID)
	if err != nil {
		log.Println("Error finding merchant for terminal request:", err)
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{Success: false, Message: "Merchant not found"})
		return
	}

	// generate request id
	requestID := fmt.Sprintf("TRQ-%d", time.Now().UnixNano()/1000000)

	_, err = h.Store.Exec("INSERT INTO terminal_requests (request_id, merchant_id, terminal_sn, status, requested_at, notes) VALUES (?, ?, ?, 'pending', CURRENT_TIMESTAMP, ?)", requestID, merchantID, payload.TerminalSN, payload.Notes)
	if err != nil {
		log.Println("Failed to create terminal request:", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{Success: false, Message: "Failed to create terminal request"})
		return
	}

	jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{Success: true, Message: "Terminal request submitted", Data: structs.Terminal{TerminalSN: payload.TerminalSN}})
}
