package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	jsonwrite "unicard-go/backend/internal/pkg/handler"
)

// DeleteCardHandler handles deleting a card by card_number and returns JSON.
func (h *Handler) DeleteCardHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("DeleteCardHandler running...")


	if r.Method != http.MethodPost {
		jsonwrite.WriteJSON(w, http.StatusMethodNotAllowed, jsonwrite.APIResponse{
			Success: false,
			Message: "Method not allowed",
		})
		return
	}

	var req struct {
		CardNumber string `json:"cardNumber"`
	}

	// Try reading JSON body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Fallback to post form values
		if err := r.ParseForm(); err == nil {
			req.CardNumber = r.PostFormValue("cardNumber")
		}
	}

	cardNumber := strings.TrimSpace(req.CardNumber)

	if cardNumber == "" {
		jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{
			Success: false,
			Message: "Card number is required for deletion.",
		})
		return
	}

	// Delete from the database
	result, err := h.DB.Exec("DELETE FROM cards WHERE card_number = ?", cardNumber)
	if err != nil {
		fmt.Println("Error deleting card from DB:", err)
		jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{
			Success: false,
			Message: "Failed to delete card due to database error.",
		})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		fmt.Println("Error reading rows affected:", err)
		jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{
			Success: false,
			Message: "Failed to confirm card deletion.",
		})
		return
	}

	if rowsAffected == 0 {
		fmt.Printf("Card %s not found for deletion\n", cardNumber)
		jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{
			Success: false,
			Message: "Card not found.",
		})
		return
	}

	fmt.Println("Card deleted successfully:", cardNumber)
	jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{
		Success: true,
		Message: "Card deleted successfully!",
	})
}
