package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	jsonwrite "unicard-go/backend/internal/pkg/handler"
)

// DeactivateView renders the deactivateCard.html template after verifying session.
func (h *Handler) DeactivateView(w http.ResponseWriter, r *http.Request) {
	fmt.Println("DeactivateView running...")
	

	
	h.Tpl.ExecuteTemplate(w, "deactivateCard.html", nil)
}

// DeactivateCardHanlder handles deactivating a card and returns a JSON response.
func (h *Handler) DeactivateCardHanlder(w http.ResponseWriter, r *http.Request) {
	fmt.Println("DeactivateCardHanlder running...")

	var req struct {
		CardNumber string `json:"cardNumber"`
		Name       string `json:"name"`
		CardType   string `json:"cardType"`
	}

	// Try reading JSON body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Fallback to post form values
		if err := r.ParseForm(); err == nil {
			req.CardNumber = r.PostFormValue("cardNumber")
			req.Name = r.PostFormValue("name")
			req.CardType = r.PostFormValue("cardType")
		}
	}

	cardNumber := strings.TrimSpace(req.CardNumber)
	cardHolder := strings.TrimSpace(req.Name)
	cardType := strings.TrimSpace(req.CardType)

	if cardNumber == "" || cardHolder == "" || cardType == "" {
		jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{
			Success: false,
			Message: "Please fill in all required fields.",
		})
		return
	}

	// Check if the card exists and is active, then deactivate it
	ok, err := h.deactivateCardIfActive(cardNumber, cardHolder, cardType)
	if err != nil {
		fmt.Println("Error while deactivating card:", err)
		jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{
			Success: false,
			Message: "Failed to deactivate card.",
		})
		return
	}

	if !ok {
		fmt.Printf("Card not found or already inactive: %s with Card Type: %s\n", cardNumber, cardType)
		jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{
			Success: false,
			Message: "Card not found, name/type mismatch, or card is already inactive.",
		})
		return
	}

	fmt.Println("Card deactivated successfully:", cardNumber, cardType)
	jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{
		Success: true,
		Message: "Card deactivated successfully!",
	})
}

// --- Helper functions ---

func (h *Handler) deactivateCardIfActive(cardNumber, cardHolder, cardType string) (bool, error) {
	result, err := h.DB.Exec(`
		UPDATE cards
		SET status = 'Blocked'
		WHERE card_number = ?
		AND card_holder = ? 
		AND card_type = ?
		AND status = 'Active'
	`, cardNumber, cardHolder, cardType)
	if err != nil {
		return false, err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return rows > 0, nil
}
