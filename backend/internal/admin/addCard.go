package admin

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
	jsonwrite "unicard-go/backend/internal/pkg/handler"
)

// Card struct represents a card and its attributes.
type Card struct {
	CardUID       string
	CardNumber    string
	CardType      string
	InitialAmount float64
	ExpiryDate    string
	CreatedAt     string
}

// AddCardsView renders the addCards.html template after checking the admin session.
func (h *Handler) AddCardsView(w http.ResponseWriter, r *http.Request) {
	fmt.Println("AddCardsView running...")
	

	
	h.Tpl.ExecuteTemplate(w, "addCards.html", nil)
}

// AddCardHandler handles card creation and returns JSON response.
func (h *Handler) AddCardHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("AddCardHandler running...")

	var req struct {
		CardUID       string `json:"cardUID"`
		InitialAmount string `json:"initialAmount"`
	}

	// Try reading JSON body first
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Fallback to post form values
		if err := r.ParseForm(); err == nil {
			req.CardUID = r.PostFormValue("cardUID")
			req.InitialAmount = r.PostFormValue("initialAmount")
		}
	}

	cardUID := strings.TrimSpace(req.CardUID)
	initialAmount := strings.TrimSpace(req.InitialAmount)

	// Validate required fields
	if cardUID == "" || initialAmount == "" {
		jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{
			Success: false,
			Message: "Please fill in all required fields.",
		})
		return
	}

	// Auto-generate card number
	fmt.Println("Calling generateCardNumber()...")
	cardNumber := h.generateCardNumber()
	fmt.Printf("Generated card number: %s\n", cardNumber)

	// Set default card type
	cardType := "Regular"

	// Auto-calculate expiry date as 10 years from now
	expiryDate := time.Now().AddDate(10, 0, 0).Format("2006-01-02")

	// Convert Initial Amount (String -> Float64)
	amount, err := strconv.ParseFloat(initialAmount, 64)
	if err != nil {
		fmt.Printf("Error parsing amount: %v\n", err)
		jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{
			Success: false,
			Message: "Invalid amount format. Must be a number.",
		})
		return
	}

	// Check for existing card UID
	cardUidExist, err := h.cardUIDExist(cardUID)
	if err != nil {
		jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{
			Success: false,
			Message: "Error checking card UID.",
		})
		return
	}
	if cardUidExist {
		jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{
			Success: false,
			Message: "Card UID already exists.",
		})
		return
	}

	// Create a new Card struct
	card := Card{
		CardUID:       cardUID,
		CardNumber:    cardNumber,
		CardType:      cardType,
		InitialAmount: amount,
		ExpiryDate:    expiryDate,
	}

	// Check for existing card number
	cardNumExists, err := h.cardNumberExist(card)
	if err != nil {
		jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{
			Success: false,
			Message: "Error checking card number.",
		})
		return
	}
	if cardNumExists {
		jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{
			Success: false,
			Message: "Card number already exists.",
		})
		return
	}

	createdAt := time.Now().Format("2006-01-02 15:04:05")

	// Insert card into database
	query := "INSERT INTO cards (card_uid, card_number, card_type, initial_amount, expiry_date, created_at) VALUES (?, ?, ?, ?, ?, ?)"
	_, err = h.DB.Exec(
		query,
		card.CardUID,
		card.CardNumber,
		card.CardType,
		card.InitialAmount,
		card.ExpiryDate,
		createdAt,
	)
	if err != nil {
		fmt.Println("Error inserting card into database:", err)
		jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{
			Success: false,
			Message: "Error while adding card.",
		})
		return
	}

	// Successfully added the card
	jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{
		Success: true,
		Message: "Card added successfully!",
	})
}

//--- HELPER FUNCTIONS ---

func (h *Handler) cardUIDExist(card string) (bool, error) {
	var uid string
	query := "SELECT card_uid FROM cards WHERE card_uid = ?"
	err := h.DB.QueryRow(query, card).Scan(&uid)
	if err == sql.ErrNoRows {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func (h *Handler) cardNumberExist(card Card) (bool, error) {
	var cardNum string
	query := "SELECT card_number FROM cards WHERE card_number = ?"
	err := h.DB.QueryRow(query, card.CardNumber).Scan(&cardNum)
	if err == sql.ErrNoRows {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func (h *Handler) generateCardNumber() string {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	now := time.Now()
	year := now.Format("06")
	month := now.Format("01")
	day := now.Format("02")
	datePrefix := year + day + month

	randomNum := rng.Intn(1000000000)
	return fmt.Sprintf("%s%010d", datePrefix, randomNum)
}
