package admin

import (
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strings"
	"time"
	jsonwrite "unicard-go/backend/internal/pkg/handler"
	structure "unicard-go/backend/internal/pkg/structs"

	"github.com/go-playground/validator/v10"
)

const cardType = "regular" // Lowercase to match your database ENUM

// AddCardsView renders the addCards.html template after checking the admin session.
func (h *Handler) AddCardsView(w http.ResponseWriter, r *http.Request) {
	fmt.Println("AddCardsView running...")
	data := AdminPageData{
		Page:     "addcard",
		Username: r.PathValue("username"),
	}
	h.Tpl.ExecuteTemplate(w, "addCards.html", data)
}

// AddCardHandler handles card creation and returns JSON response.
func (h *Handler) AddCardHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("AddCardHandler running...")

	// Define request struct
	var req structure.CardData

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
				"CardUID": "Card UID is required.",
				"Balance": "Initial amount is required and cannot be negative.",
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
	// Set card to be unlinked for 2 years
	expiryDate := time.Now().AddDate(2, 0, 0).Format("2006-01-02") 

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
        VALUES (?, ?, ?, ?, ?, 'inactive')
    `

	_, err = h.Store.Exec(
		query,
		cardUID,
		cardNumber,
		cardType,
		req.Balance, // Maps to 'balance'
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
	err := h.Store.QueryRow(query, uid).Scan(&existingUID)
	if err == sql.ErrNoRows {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

// Generate Card Number of format YYSS + 10 random digits + DD
func (h *Handler) generateCardNumber() string {
	loc, err := time.LoadLocation("Asia/Manila")
	if err != nil {
		loc = time.Local
	}
	now := time.Now().In(loc)

	yy := now.Format("06")
	ss := now.Format("05")
	dd := now.Format("02")

	max10 := big.NewInt(10000000000) // 10^10 for 10 digits
	random10, errRand := rand.Int(rand.Reader, max10)

	randomDigits := "0000000000" // Default fallback format
	if errRand == nil {
		randomDigits = fmt.Sprintf("%010d", random10.Int64())
	}

	return fmt.Sprintf("%s%s%s%s", yy, ss, randomDigits, dd)
}
