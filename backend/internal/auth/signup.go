package authentication

import (
	//"database/sql"

	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"unicard-go/backend/internal/pkg/account"
	jsonwrite "unicard-go/backend/internal/pkg/handler"

	"github.com/go-playground/validator/v10"
)

const (
	CardStatusInactive = "Inactive"
	UserTypeRegular    = "Regular"
)

// Create a struct to catch the incoming JSON from the frontend
type SignupRequest struct {
	ID            string `json:"id,omitempty"`
	FirstName     string `json:"first_name" validate:"required"`
	LastName      string `json:"last_name" validate:"required"`
	Name          string `json:"name" db:"name"`
	CardNumber    string `json:"card_number" validate:"required,numeric,len=16"`
	Password      string `json:"password" validate:"required,min=8"`
	Email         string `json:"email" validate:"required,email"`
	ContactNumber string `json:"contact_number" validate:"required,numeric,len=11"`
}

// User struct to hold signup data (Keep your existing one)
// Create a struct to get the data from the database
type User struct {
	UserID     string  `db:"user_id"`
	Username   string  `db:"username"`
	Name       string  `db:"name"`
	Email      string  `db:"email"`
	Phone      string  `db:"phone_number"`
	CardNumber string  `db:"card_number"`
	Password   string  `db:"password_hash"`
	UserType   string  `db:"role"`
	Balance    float64 `db:"balance"`
	CreatedAt  string  `db:"created_at"`
	Role       string  `db:"role"`
	Status     string  `db:"status"`
	RegionID   string  `db:"region_id"`
}

type CheckDetailsRequest struct {
	Email         string `json:"email"`
	ContactNumber string `json:"contact_number"`
}

/*type CheckCardRequest struct {
	CardNumber string `json:"card_number"`
}*/

// CheckDetailsHandler checks if email and phone are available
func (h *Handler) CheckDetailsHandler(w http.ResponseWriter, r *http.Request) {
	var req CheckDetailsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "Invalid request",
		})
		return
	}

	// Check Email
	exists, err := account.IsEmailExist(h.DB, req.Email)
	if err != nil {
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Database error",
		})
		return
	}
	if exists {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "Email already registered",
			Field:   "email",
		})
		return
	}

	// Check Phone
	exists, err = h.isPhoneExist(req.ContactNumber)
	if err != nil {
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Database error",
		})
		return
	}
	if exists {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "Invalid phone number",
			Field:   "phone",
		})
		return
	}

	jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{
		Success: true,
		Message: "Details are valid",
	})
}

// CheckCardHandler checks if card is valid for registration
func (h *Handler) CheckCardHandler(w http.ResponseWriter, r *http.Request) {
	var req SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "Invalid request",
		})
		return
	}

	var status string
	err := h.DB.QueryRow("SELECT status FROM cards WHERE card_number = ?", req.CardNumber).Scan(&status)
	if err != nil {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "Card not found",
		})
		return
	}

	if status != "inactive" {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "Card is invalid",
			Field:   "card",
		})
		return
	}

	jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{
		Success: true,
		Message: "Card is valid",
		Field:   "card",
	})
}

// View Handler (GET)
// You can now simplify this because JS handles the errors!
func (h *Handler) SignupView(w http.ResponseWriter, r *http.Request) {
	log.Printf("Signup view is running...")
	// Just serve the template. No need for the huge switch statement anymore.
	h.Tpl.ExecuteTemplate(w, "signup.html", nil)
}

func (h *Handler) SignupHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Signup Handler is running...")

	// get context from request
	ctx := r.Context()

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
	err := Validate.Struct(req)
	if err != nil {
		log.Printf("Validation failed: %v", err)

		errorMessage := "Invalid input provided."
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			errorMap := map[string]string{
				"FirstName":     "First name is required.",
				"LastName":      "Last name is required.",
				"Email":         "Please provide a valid email address.",
				"ContactNumber": "Contact number must be exactly 11 digits.",
				"CardNumber":    "Card number must be exactly 16 digits.",
				"Password":      "Password must be at least 8 characters long.",
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
	log.Printf("Password hashed successfully: %v", hashedPassword)

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
	log.Printf("Timestamp: %v", createdAt)

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
		UserType:   UserTypeRegular,
		Username:   req.FirstName,
		Name:       req.FirstName + " " + req.LastName,
		CardNumber: req.CardNumber,
		Password:   hashedPassword,
		Email:      req.Email,
		Phone:      req.ContactNumber,
		CreatedAt:  createdAt,
		Balance:    balance,
		Role:       "Customer",
		Status:     "active",
	}

	// Begin transaction: insert user + activate card atomically
	tx, err := h.DB.BeginTx(ctx, nil)
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
    (user_id, username, name, email, phone_number, password_hash, role, status, created_at) 
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err = tx.ExecContext(ctx, insertQuery,
		user.UserID, user.Username, user.Name, user.Email, user.Phone,
		user.Password, user.Role, user.Status, user.CreatedAt,
	)
	if err != nil {
		log.Printf("Error inserting user: %v", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "System error creating account. Please try again.",
		})
		return
	}
	log.Printf("User record inserted: %v", user.Name)

	// Activate Card and Link User Details
	// Set expiry date to 5 years from now in expiry_date column
	updateCardQuery := `
        UPDATE cards 
        SET status = 'active', 
            user_id = ?, 
			card_type = 'regular',
			linked_at = CURRENT_TIMESTAMP,
            updated_at = CURRENT_TIMESTAMP,
			expiry_date = DATE_ADD(CURRENT_DATE, INTERVAL 5 YEAR) 
        WHERE card_number = ?`

	_, err = tx.ExecContext(ctx, updateCardQuery, user.UserID, user.CardNumber)

	if err != nil {
		log.Printf("Error activating card for card_number %s: %v", user.CardNumber, err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "System error activating card.",
		})
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
