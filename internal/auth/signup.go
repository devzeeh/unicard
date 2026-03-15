package authentication

import (
	"crypto/rand"
	"database/sql"
	"encoding/json" // Added for JSON support
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"
	"unicard-go/internal/pkg/account"
)

// 1. Create a struct to catch the incoming JSON from the frontend
type SignupRequest struct {
	FirstName     string `json:"firstName"`
	LastName      string `json:"lastName"`
	CardNumber    string `json:"cardNumber"`
	Password      string `json:"password"`
	Email         string `json:"email"`
	ContactNumber string `json:"contactNumber"`
}

// 2. Create a standard API response struct
type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// User struct to hold signup data (Keep your existing one)
type User struct {
	UserID     string
	Username   string
	Fullname   string
	Email      string
	Phone      string
	CardNumber string
	Password   string
	CardID     string
	Usertype   string
	Balance    float64
	CreatedAt  string
}

// View Handler (GET)
// You can now simplify this because JS handles the errors!
func (h *Handler) SignupView(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Signup view is running...")
	// Just serve the template. No need for the huge switch statement anymore.
	h.Tpl.ExecuteTemplate(w, "signup.html", nil)
}

// Auth Handler (POST) - Converted to JSON API
func (h *Handler) SignupHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Signup API is running...")
	w.Header().Set("Content-Type", "application/json")

	// Enforce POST method
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(APIResponse{Success: false, Message: "Method not allowed"})
		return
	}

	// Decode incoming JSON
	var req SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(APIResponse{Success: false, Message: "Failed to parse JSON request"})
		return
	}

	// Clean the inputs
	req.FirstName = strings.TrimSpace(req.FirstName)
	req.LastName = strings.TrimSpace(req.LastName)
	req.CardNumber = strings.TrimSpace(req.CardNumber)
	req.Password = strings.TrimSpace(req.Password)
	req.Email = strings.TrimSpace(req.Email)
	req.ContactNumber = strings.TrimSpace(req.ContactNumber)

	// Validation: Empty Fields
	if req.FirstName == "" || req.LastName == "" || req.CardNumber == "" || req.Password == "" || req.Email == "" || req.ContactNumber == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(APIResponse{Success: false, Message: "Please fill in all fields."})
		return
	}

	// Prepare User struct
	user := User{
		Usertype:   "Regular",
		Username:   req.FirstName, // Using First Name as temporary Username
		Fullname:   req.FirstName + " " + req.LastName,
		CardNumber: req.CardNumber,
		Password:   req.Password,
		Email:      req.Email,
		Phone:      req.ContactNumber,
	}

	// Generate Username
	generatedUsername, err := h.GenerateUniqueUsername()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(APIResponse{Success: false, Message: "System error generating username. Please try again."})
		return
	}
	user.Username = generatedUsername

	// Check Card Status
	var cardStatus string
	query := "SELECT status FROM cards WHERE card_number = ?"
	err = h.DB.QueryRow(query, user.CardNumber).Scan(&cardStatus)

	if err == sql.ErrNoRows {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(APIResponse{Success: false, Message: "Card number not found. Please check your card."})
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(APIResponse{Success: false, Message: "System error checking card status."})
		return
	}

	if cardStatus != "Inactive" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(APIResponse{Success: false, Message: fmt.Sprintf("Card is currently '%s'. Please contact support.", cardStatus)})
		return
	}

	// Password Length Check
	if len(user.Password) < 8 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(APIResponse{Success: false, Message: "Password must be at least 8 characters long."})
		return
	}

	// Hash Password
	hashedPassword, err := account.HashPassword(user.Password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(APIResponse{Success: false, Message: "System error processing password."})
		return
	}
	user.Password = hashedPassword

	// Check Email
	if exists, _ := account.IsEmailExist(h.DB, user.Email); exists {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(APIResponse{Success: false, Message: "Email already registered. Please use a different email."})
		return
	}

	// Check Phone
	if exists, _ := h.isPhoneExist(user.Phone); exists {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(APIResponse{Success: false, Message: "Phone number already registered. Please use a different number."})
		return
	}

	// Generate IDs
	generateUserId, err := h.GenerateUserID()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(APIResponse{Success: false, Message: "System error generating UserID."})
		return
	}
	user.UserID = fmt.Sprintf("%d", generateUserId)

	generateCardID, err := h.GenerateCardID()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(APIResponse{Success: false, Message: "System error generating CardID."})
		return
	}
	user.CardID = generateCardID

	// Get Timestamp & Balance
	user.CreatedAt, _ = CurrentTimestamp()
	user.Balance, err = h.GetInitialBalance(user.CardNumber)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(APIResponse{Success: false, Message: "Invalid Card Number."})
		return
	}

	// Insert into DB
	insertQuery := "INSERT INTO users (user_id, username, full_name, email, phone, password_hash, card_id, card_number, user_type, balance, created_at) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
	_, err = h.DB.Exec(insertQuery, user.UserID, user.Username, user.Fullname, user.Email, user.Phone, user.Password, user.CardID, user.CardNumber, user.Usertype, user.Balance, user.CreatedAt)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(APIResponse{Success: false, Message: "System error creating account. Please try again."})
		return
	}

	// Activate Card
	updateQuery := "UPDATE cards SET status = 'Active' WHERE card_number = ?"
	_, err = h.DB.Exec(updateQuery, user.CardNumber)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(APIResponse{Success: false, Message: "System error activating card."})
		return
	}

	// SUCCESS! Send a JSON success message instead of redirecting.
	fmt.Println("Account successfully created!")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(APIResponse{Success: true, Message: "Account created successfully!"})
}

// ... [KEEP ALL YOUR HELPER FUNCTIONS DOWN HERE EXACTLY AS THEY WERE] ...
// ---Helper Functions---

// GenerateUniqueUsername creates a random unique username
// Format: user + 12 random lowercase characters/numbers
// Example: user9d8a7c2b3e4f
func (h *Handler) GenerateUniqueUsername() (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	usernamePrefix := "user"
	length := 7

	for {
		// Get the current date in DDYY format (Go uses "010206" as reference)
		userDate := time.Now().Format("06") // e.g., "012624" for Jan 26, 2024
		// time .Now().Format("1504") // e.g., "1530" for 3:30 PM
		timePart := time.Now().Format("0405") // e.g., "1530" for 3:30 PM

		// Combine date and time to form part of the username
		//usernamePrefix = fmt.Sprintf("user%s%s", userDate, timePart)

		// Generate the random suffix
		randomPart := ""
		for i := 0; i < length; i++ {
			num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
			if err != nil {
				return "", err
			}
			randomPart += string(charset[num.Int64()])
		}

		// Combine prefix + date + time + random part
		username := fmt.Sprintf("%s%s%s%s", usernamePrefix, userDate, randomPart, timePart)

		// Check DB for uniqueness
		var existing string
		query := "SELECT username FROM users WHERE username = ?"
		err := h.DB.QueryRow(query, username).Scan(&existing)

		if err == sql.ErrNoRows {
			//fmt.Println("Generated unique username:", username)
			return username, nil // Found a unique one!
		} else if err != nil {
			return "", err // Real DB Error
		}

		// Collision detected, loop runs again...
		fmt.Println("Username collision! Retrying...")
	}
}

// This function checks if a given phone number already exists in the database.
// It executes a SQL query to search for the phone number in the users table.
// If the phone number is found, it returns true. If not found, it returns false.
// If an error occurs during the query, it returns the error.
func (h *Handler) isPhoneExist(phone string) (bool, error) {
	// Hold the existing phone number
	var existingPhone string

	// Check query
	query := "SELECT phone FROM users WHERE phone = ?"
	err := h.DB.QueryRow(query, phone).Scan(&existingPhone)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		fmt.Println("Phone number check error:", err)
		return false, err
	}
	return true, nil
}

// It generates random numbers and checks the database for uniqueness.
// If a generated ID already exists, it retries until a unique one is found.
// Returns the unique user ID as int64 or an error if any occurs.
func (h *Handler) GenerateUserID() (int64, error) {
	// Generate random 12 digits number
	// Range: 100,000,000,000 to 999,999,999,999
	min := int64(100000000000)
	max := int64(999999999999)

	for {
		// Calculate the range size (max - min + 1)
		diff := new(big.Int).Sub(big.NewInt(max), big.NewInt(min))
		diff.Add(diff, big.NewInt(1))

		number, err := rand.Int(rand.Reader, diff)
		if err != nil {
			return 0, err
		}

		// Add min to get the final ID within the range
		userID := number.Int64() + min

		// Check DB directly
		var tmpId int64
		query := "SELECT user_id FROM users WHERE user_id = ?"

		err = h.DB.QueryRow(query, userID).Scan(&tmpId)
		if err == sql.ErrNoRows {
			//	fmt.Println("Unique UserID generated:", userID)
			return userID, nil // Unique ID found
		} else if err != nil {
			return 0, err // Real DB error
		}
		// If it exists, loop runs again
		fmt.Println("Collision detected! Retrying...")
	}
}

// It generates the unique cardID for every card users
// Checks the database for uniqueness.
// Returns the unique card ID as string or an error if any occurs.
// Example format: CARD-XXXXXXX
func (h *Handler) GenerateCardID() (string, error) {
	// Define charset: letters and numbers
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	cardIDPrefix := "CARD-"
	randomLength := 7

	for {
		// Get the current date in MMDDYY format (Go uses "010206" as reference)
		datePart := time.Now().Format("010206")

		// Generate the 7 random characters
		randomPart := ""
		for i := 0; i < randomLength; i++ {
			num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
			if err != nil {
				return "", err
			}
			randomPart += string(charset[num.Int64()])
		}

		// Combine them: CARD- + Date + Random
		// Example output: CARD-012626Ab7z9X1
		cardID := fmt.Sprintf("%s%s%s", cardIDPrefix, datePart, randomPart)

		// Check database for uniqueness
		var tmpCardID string
		query := "SELECT card_id FROM users WHERE card_id = ?"
		err := h.DB.QueryRow(query, cardID).Scan(&tmpCardID)

		if err == sql.ErrNoRows {
			return cardID, nil // Unique ID found
		} else if err != nil {
			return "", err // DB error
		}
		fmt.Println("Collision detected! Retrying...", cardID)
	}
}

// It check the initial balance based on card number prefix
// Returns the initial balance as float64 or an error if any occurs.
// Gets the initial balance from the "card" table in the database.
// Example: Card Number "1234567890" has initial balance of 100.0
func (h *Handler) GetInitialBalance(cardNumber string) (float64, error) {
	var initialBalance float64 // to hold the initial balance

	query := "SELECT initial_amount FROM cards WHERE card_number = ?"
	err := h.DB.QueryRow(query, cardNumber).Scan(&initialBalance)
	if err != nil {
		return 0, err
	}
	return initialBalance, nil
}

// This function returns the current timestamp formatted as "YYYY-MM-DD HH:MM:SS"
// in the "Asia/Manila" timezone. If there's an error loading the timezone,
// it returns an empty string and the error.
func CurrentTimestamp() (string, error) {
	// Load Asia/Manila location
	loc, err := time.LoadLocation("Asia/Manila")
	if err != nil {
		return "", err
	}
	time.Local = loc

	// Format the current time
	return time.Now().Format("2006-01-02 03:04:05"), nil
}
