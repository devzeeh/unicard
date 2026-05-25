package admin

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	jsonwrite "unicard-go/backend/internal/pkg/handler"

	"github.com/go-playground/validator/v10"
)

// card data from the deactivate card page
type CardData struct {
	CardNumber string `json:"cardNumber" db:"card_number" validate:"required"`
	CardHolder string `json:"cardHolder" db:"user_id" validate:"required"`
	CardType   string `json:"cardType" db:"card_type" validate:"required"`
}

func (h *Handler) DeactivateView(w http.ResponseWriter, r *http.Request) {
	fmt.Println("DeactivateView running...")

	h.Tpl.ExecuteTemplate(w, "deactivateCard.html", nil)
}

// DeactivateCardHanlder handles deactivating a card and returns a JSON response.
func (h *Handler) DeactivateCardHanlder(w http.ResponseWriter, r *http.Request) {
	fmt.Println("DeactivateCardHanlder running...")

	var req CardData

	// Try reading JSON body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("Error decoding JSON:", err)
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "Invalid request format",
		})
		return
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		errorMessage := "Invalid input provided."
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			errorMap := map[string]string{
				"CardNumber": "Card number is required.",
				"CardHolder": "Card holder is required.",
				"CardType":   "Card type is required.",
			}
			if msg, ok := errorMap[validationErrs[0].Field()]; ok {
				errorMessage = msg
			}
		}

		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: errorMessage,
		})
		return
	}

	cardNumber := strings.TrimSpace(req.CardNumber)
	cardHolder := strings.TrimSpace(req.CardHolder)
	cardType := strings.TrimSpace(req.CardType)

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
		AND user_id = ? 
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
