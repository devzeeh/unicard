package admin

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"
	jsonwrite "unicard-go/backend/internal/pkg/handler"

	"github.com/go-playground/validator/v10"
)

// CardRequest struct mapped directly to your frontend JSON payload
type CardRequest struct {
	CardUID       string  `json:"card_uid" db:"card_uid" validate:"required"`
	InitialAmount float64 `json:"initial_amount" db:"balance" validate:"required,min=0"` // Native float, no string parsing needed!
}

// AddCardsView renders the addCards.html template after checking the admin session.
func (h *Handler) AddCardsView(w http.ResponseWriter, r *http.Request) {
	fmt.Println("AddCardsView running...")
	h.Tpl.ExecuteTemplate(w, "addCards.html", nil)
}

// AddCardHandler handles card creation and returns JSON response.
func (h *Handler) AddCardHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("AddCardHandler running...")

	var req CardRequest

	// Decode JSON
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("Error decoding JSON:", err)
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "Invalid request format",
		})
		return
	}

	cardUID := strings.TrimSpace(req.CardUID)

	// Validate required fields
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		errorMessage := "Invalid input provided."
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			errorMap := map[string]string{
				"CardUID":       "Card UID is required.",
				"InitialAmount": "Initial amount is required and cannot be negative.",
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

	// Auto-generate data
	cardNumber := h.generateCardNumber()
	cardType := "regular"                                          // Lowercase to match your database ENUM
	expiryDate := time.Now().AddDate(2, 0, 0).Format("2006-01-02") // 2 years from now unlinked

	// Check for existing UID
	uidExists, err := h.cardUIDExist(cardUID)
	if err != nil {
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Error verifying Card UID.",
		})
		return
	}
	if uidExists {
		jsonwrite.WriteJSON(w, http.StatusConflict, jsonwrite.APIResponse{
			Success: false,
			Message: "This Card UID is already registered in the system.",
		})
		return
	}

	// Insert card into database (Fixed column name: 'balance')
	// We omit created_at because MySQL handles it automatically via DEFAULT CURRENT_TIMESTAMP
	query := `
        INSERT INTO cards (card_uid, card_number, card_type, balance, expiry_date, status) 
        VALUES (?, ?, ?, ?, ?, 'active')
    `

	_, err = h.DB.Exec(
		query,
		cardUID,
		cardNumber,
		cardType,
		req.InitialAmount, // Maps to 'balance'
		expiryDate,
	)

	if err != nil {
		fmt.Println("Error inserting card into database:", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Failed to save card to database.",
		})
		return
	}

	// Success Response
	jsonwrite.WriteJSON(w, http.StatusCreated, jsonwrite.APIResponse{
		Success: true,
		Message: "Card added successfully!",
	})
}

// --- HELPER FUNCTIONS ---

func (h *Handler) cardUIDExist(uid string) (bool, error) {
	var existingUID string
	query := "SELECT card_uid FROM cards WHERE card_uid = ?"
	err := h.DB.QueryRow(query, uid).Scan(&existingUID)
	if err == sql.ErrNoRows {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

// Generate Card Number of format HHMMSSYYDD + 6 random digits
func (h *Handler) generateCardNumber() string {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	now := time.Now()

	// Format: HHMMSSYYDD + 6 random digits
	datePrefix := now.Format("150405060201")
	randomNum := rng.Intn(100000)

	return fmt.Sprintf("%s%06d", datePrefix, randomNum)
}
