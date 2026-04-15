package admin

import (
	"database/sql"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
	message "unicard-go/internal/pkg"
)

// This struct represents a card and its attributes.
// We can use it to easily pass card data around in our functions.
// It also helps to keep our code organized and makes it easier to manage card-related data.
type Card struct {
	CardUID       string
	CardNumber    string
	CardType      string
	InitialAmount float64
	ExpiryDate    string
	CreatedAt     string
}

// This function renders the addCards.html template when the admin visits the /admin/addcard page.
// It doesn't do any processing yet, it just shows the form to the admin.
// We can also pass an empty AddCardsData struct to the template, which allows us to easily display error or success messages later on when we process the form submission.
func (h *Handler) AddCardsView(w http.ResponseWriter, r *http.Request) {
	fmt.Println("AddCardsView running...")
	// Render the addCards.html template
	h.Tpl.ExecuteTemplate(w, "addCards.html", message.MessageData{})
}

// This function handles the form submission from the addCards.html page.
// It processes the form data, validates it, generates a card number, checks for duplicates, and inserts the new card into the database.
// Also have error handling at each step, and we pass error or success messages back to the template to inform the admin of the result.
func (h *Handler) AddCardHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("addcardshandler running...")

	if err := r.ParseForm(); err != nil {
		h.Tpl.ExecuteTemplate(w, "addCards.html", message.MessageData{Error: "Failed to parse form"})
		return
	}

	// Process the form data here
	cardUID := strings.TrimSpace(r.PostFormValue("cardUID"))
	initialAmount := strings.TrimSpace(r.PostFormValue("initialAmount"))

	// Validate required fields
	if cardUID == "" || initialAmount == "" {
		h.Tpl.ExecuteTemplate(w, "addCards.html", message.MessageData{Error: "Please fill in all required fields."})
		return
	}

	// Auto-generate card number
	fmt.Println("Calling generateCardNumber()...")
	cardNumber := h.generateCardNumber()
	fmt.Printf("Generated card number: %s\n", cardNumber)

	// Set default card type
	cardType := "Regular"
	fmt.Printf("Set card type: %s\n", cardType)

	// Auto-calculate expiry date as 10 years from now
	expiryDate := time.Now().AddDate(10, 0, 0).Format("2006-01-02")
	fmt.Printf("Set expiry date: %s\n", expiryDate)

	// Convert Initial Amount (String -> Float64)
	fmt.Println("Parsing initial amount...")
	amount, err := strconv.ParseFloat(initialAmount, 64)
	if err != nil {
		fmt.Printf("Error parsing amount: %v\n", err)
		h.Tpl.ExecuteTemplate(w, "addCards.html", message.MessageData{Error: "Invalid amount format. Must be a number."})
		return
	}
	fmt.Printf("Parsed amount: %.2f\n", amount)

	// Check for existing card UID
	fmt.Printf("Calling cardUIDExist() for UID: %s\n", cardUID)
	cardUidExist, err := h.cardUIDExist(cardUID)
	if err != nil {
		fmt.Println("Error checking card UID existence:", err)
		h.Tpl.ExecuteTemplate(w, "addCards.html", message.MessageData{Error: "Error checking card UID."})
		return
	}
	if cardUidExist {
		fmt.Println("Card UID already exists")
		h.Tpl.ExecuteTemplate(w, "addCards.html", message.MessageData{Error: "Card UID already exists."})
		return
	}
	fmt.Println("Card UID is unique")

	// Create a new Card struct
	card := Card{
		CardUID:       cardUID,
		CardNumber:    cardNumber,
		CardType:      cardType,
		InitialAmount: amount,
		ExpiryDate:    expiryDate,
	}

	// Check for existing card number
	fmt.Printf("Calling cardNumberExist() for card number: %s\n", card.CardNumber)
	cardNumExists, err := h.cardNumberExist(card)
	if err != nil {
		fmt.Println("Error checking card number existence:", err)
		h.Tpl.ExecuteTemplate(w, "addCards.html", message.MessageData{Error: "Error checking card number."})
		return
	}
	if cardNumExists {
		fmt.Println("Card number already exists")
		h.Tpl.ExecuteTemplate(w, "addCards.html", message.MessageData{Error: "Card number already exists."})
		return
	}
	fmt.Println("Card number is unique")

	// Create current timestamp
	// Output example: 2024-06-15 14:30:00
	fmt.Println("Generating timestamp...")
	createdAt := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf("Timestamp: %s\n", createdAt)

	// Insert card into database
	fmt.Println("Executing database insert...")
	query := "INSERT INTO cards (card_uid, card_number, card_type, initial_amount, expiry_date, created_at) VALUES (?, ?, ?, ?, ?, ?)"
	_, err = h.DB.Exec(
		query,
		card.CardUID,
		card.CardNumber,
		card.CardType,
		card.InitialAmount,
		card.ExpiryDate,
		createdAt)
	if err != nil {
		fmt.Println("Error inserting card into database:", err)
		h.Tpl.ExecuteTemplate(w, "addCards.html", message.MessageData{Error: "Error while adding card."})
		return
	}

	// Successfully added the card
	fmt.Printf("Card added successfully: %s\n", card.CardNumber)
	h.Tpl.ExecuteTemplate(w, "addCards.html", message.MessageData{Success: "Card added successfully!"})
}

//--- HELPER FUNCTIONS ---

// Check card UID existence
// We can use this before adding a new card to avoid duplicates
func (h *Handler) cardUIDExist(card string) (bool, error) {
	var uid string
	query := "SELECT card_uid FROM cards WHERE card_uid = ?"
	err := h.DB.QueryRow(query, card).Scan(&uid)
	if err == sql.ErrNoRows {
		return false, nil // UID is unique
	} else if err != nil {
		fmt.Println("Error checking card UID existence:", err)
		return false, err // Exit on DB error
	}
	return true, nil // UID exists
}

// Check if card number already exists.
// We can use this before adding a new card to avoid duplicates
func (h *Handler) cardNumberExist(card Card) (bool, error) {
	var cardNum string
	query := "SELECT card_number FROM cards WHERE card_number = ?"
	err := h.DB.QueryRow(query, card.CardNumber).Scan(&cardNum)
	if err == sql.ErrNoRows {
		return false, nil // Card number is unique
	} else if err != nil {
		fmt.Println("Error checking card existence:", err)
		return false, err
	}
	return true, nil // Card number exists
}

// Generate a unique card number with format YYDDMM + 10 random digits
func (h *Handler) generateCardNumber() string {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	// Get current date components
	now := time.Now()                // Fixed prefix for all cards
	year := now.Format("06")         // YY format
	month := now.Format("01")        // MM format
	day := now.Format("02")          // DD format
	datePrefix := year + day + month // YYDDMM format

	// Generate remaining 10 random digits to make 16 total
	randomNum := rng.Intn(10000000000)
	return fmt.Sprintf("%s%010d", datePrefix, randomNum)
}
