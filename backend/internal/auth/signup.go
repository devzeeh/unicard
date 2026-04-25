package authentication

import (
	"database/sql"
	"encoding/json" // Added for JSON support
	"fmt"
	"log"
	"net/http"
	"strings"
	"unicard-go/backend/internal/pkg/account"
	jsonwrite "unicard-go/backend/internal/pkg/handler"
)

// Create a struct to catch the incoming JSON from the frontend
type SignupRequest struct {
	FirstName     string `json:"firstName"`
	LastName      string `json:"lastName"`
	CardNumber    string `json:"cardNumber"`
	Password      string `json:"password"`
	Email         string `json:"email"`
	ContactNumber string `json:"contactNumber"`
}

// Create a standard API response struct
type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// User struct to hold signup data (Keep your existing one)
type User struct {
	UserID     string  `db:"user_id"`
	Username   string  `db:"username"`
	Fullname   string  `db:"full_name"`
	Email      string  `db:"email"`
	Phone      string  `db:"phone"`
	CardNumber string  `db:"card_number"`
	Password   string  `db:"password_hash"`
	CardID     string  `db:"card_id"`
	Usertype   string  `db:"user_type"`
	Balance    float64 `db:"balance"`
	CreatedAt  string  `db:"created_at"`
}

// View Handler (GET)
// You can now simplify this because JS handles the errors!
func (h *Handler) SignupView(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Signup view is running...")
	// Just serve the template. No need for the huge switch statement anymore.
	h.Tpl.ExecuteTemplate(w, "signup.html", nil)
}

func (h *Handler) SignupHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Signup Handler is running...")

	// Decode incoming JSON
	var req SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding signup JSON: %v", err)
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "Failed to parse JSON request",
		})
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
	fields := []string{req.FirstName, req.LastName, req.CardNumber, req.Password, req.Email, req.ContactNumber}
	for _, f := range fields {
		if f == "" {
			log.Printf("Validation failed: Empty fields")
			jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
				Success: false, 
				Message: "Please fill in all fields.",
			})
			return
		}
	}

	// Validation: Password Length (fail fast before any DB calls)
	if len(req.Password) < 8 {
		log.Printf("Validation failed: Password must be at least 8 characters long")
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false, 
			Message: "Password must be at least 8 characters long.",
		})
		return
	}

	// Check Card Status
	var cardStatus string
	err := h.DB.QueryRow("SELECT status FROM cards WHERE card_number = ?", req.CardNumber).Scan(&cardStatus)
	if err == sql.ErrNoRows {
		log.Printf("Card not found: %v", req.CardNumber)
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false, 
			Message: "Card number not found. Please check your card.",
		})
		return
	} else if err != nil {
		log.Printf("Error checking card status: %v", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false, 
			Message: "System error checking card status.",
		})
		return
	}

	if cardStatus != "Inactive" {
		log.Printf("Card is not inactive: %v", cardStatus)
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false, 
			Message: fmt.Sprintf("Card is currently '%s'. Please contact support.", cardStatus),
		})
		return
	}

	// Check Email
	exists, err := account.IsEmailExist(h.DB, req.Email)
	if err != nil {
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false, 
			Message: "System error checking email.",
		})
		return
	}
	if exists {
		jsonwrite.WriteJSON(w, http.StatusConflict, jsonwrite.APIResponse{
			Success: false, 
			Message: "Email already registered. Please use a different email.",
		})
		return
	}

	// Check Phone
	exists, err = h.isPhoneExist(req.ContactNumber)
	if err != nil {
		fmt.Println("Phone number check error:", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false, 
			Message: "System error checking phone number.",
		})
		return
	}
	if exists {
		jsonwrite.WriteJSON(w, http.StatusConflict, jsonwrite.APIResponse{
			Success: false, 
			Message: "Phone number already registered. Please use a different number.",
		})
		return
	}

	// Hash Password
	hashedPassword, err := account.HashPassword(req.Password)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false, 
			Message: "System error processing password.",
		})
		return
	}
	log.Printf("Password: %v", req.Password)
	log.Printf("Password hashed successfully: %v", hashedPassword)

	// Generate Username
	generatedUsername, err := h.GenerateUniqueUsername()
	if err != nil {
		fmt.Println("Error generating username:", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false, 
			Message: "System error generating username. Please try again.",
		})
		return
	}
	log.Printf("Generated Username: %v", generatedUsername)

	// Generate IDs
	generateUserId, err := h.GenerateUserID()
	if err != nil {
		log.Printf("Error generating UserID: %v", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false, 
			Message: "System error generating UserID.",
		})
		return
	}

	generateCardID, err := h.GenerateCardID()
	if err != nil {
		log.Printf("Error generating CardID: %v", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false, Message: "System error generating CardID."})
		return
	}

	// Get Timestamp
	createdAt, err := CurrentTimestamp()
	if err != nil {
		log.Printf("Error getting timestamp: %v", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false, 
			Message: "System error getting timestamp.",
		})
		return
	}
	fmt.Println("Timestamp:", createdAt)

	// Get Initial Balance
	balance, err := h.GetInitialBalance(req.CardNumber)
	if err != nil {
		log.Printf("Error getting initial balance: %v", err)
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false, 
			Message: "Invalid Card Number.",
		})
		return
	}

	// Build User struct
	user := User{
		UserID:     fmt.Sprintf("%d", generateUserId),
		CardID:     generateCardID,
		Usertype:   "Regular",
		Username:   generatedUsername,
		Fullname:   req.FirstName + " " + req.LastName,
		CardNumber: req.CardNumber,
		Password:   hashedPassword,
		Email:      req.Email,
		Phone:      req.ContactNumber,
		CreatedAt:  createdAt,
		Balance:    balance,
	}

	// Begin transaction: insert user + activate card atomically
	tx, err := h.DB.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false, 
			Message: "System error starting transaction.",
		})
		return
	}
	defer tx.Rollback() // no-op if tx.Commit() is called

	// Insert User
	insertQuery := `INSERT INTO users 
    (user_id, username, full_name, email, phone, password_hash, card_id, card_number, user_type, balance, created_at) 
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err = tx.Exec(insertQuery,
		user.UserID, user.Username, user.Fullname, user.Email, user.Phone,
		user.Password, user.CardID, user.CardNumber, user.Usertype, user.Balance, user.CreatedAt,
	)
	if err != nil {
		log.Printf("Error inserting user: %v", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false, 
			Message: "System error creating account. Please try again.",
		})
		return
	}
	log.Printf("User record inserted: %v", user.UserID)

	// Activate Card and Link User Details
	// Set expiry date to 10 years from now in expiry_date column
	updateCardQuery := `
        UPDATE cards 
        SET status = 'Active', 
            user_id = ?, 
            card_holder = ?, 
            linked_at = CURRENT_TIMESTAMP,
			expiry_date = DATE_ADD(CURRENT_DATE, INTERVAL 10 YEAR) 
        WHERE card_number = ?`

	_, err = tx.Exec(updateCardQuery, user.UserID, user.Fullname, user.CardNumber)

	if err != nil {
		log.Printf("Error activating card for card_number %s: %v", user.CardNumber, err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false, 
			Message: "System error activating card.",
		} )
		return
	}

	if err = tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false, 
			Message: "System error finalizing account creation.",
		})
		return
	}

	log.Printf("Account successfully created! UserID: %s", user.UserID) // moved here
	jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{
		Success: true, 
		Message: "Account created successfully!",
	})
}
