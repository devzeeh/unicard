package admin

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	jsonwrite "unicard-go/backend/internal/pkg/handler"
	structure "unicard-go/backend/internal/pkg/structs"

	"github.com/go-playground/validator/v10"
)

func (h *Handler) DeleteCardView(w http.ResponseWriter, r *http.Request) {
	fmt.Println("DeleteCardView running...")
	data := AdminPageData{
		Page:     "delete-cards",
		Username: r.PathValue("username"),
	}
	h.Tpl.ExecuteTemplate(w, "delete_card.html", data)
}

// DeleteCardHandler handles deleting a card by card_number and returns JSON.
func (h *Handler) DeleteCardHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("DeleteCardHandler running...")

	var req structure.CardData

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "Invalid request format",
		})
		return
	}

	cardNumber := strings.TrimSpace(req.CardNumber)

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		errorMessage := "Invalid input provided."
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			errorMap := map[string]string{
				"CardNumber": "Card number is required.",
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

	// Delete from the database
	result, err := h.DB.Exec("DELETE FROM cards WHERE card_number = ?", cardNumber)
	if err != nil {
		fmt.Println("Error deleting card from DB:", err)
		jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{
			Success: false,
			Message: "Failed to delete card.",
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
