package authentication

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	jsonwrite "unicard-go/internal/pkg/handler"

	"github.com/go-playground/validator/v10" // For struct validation
	"golang.org/x/crypto/bcrypt"
)

// LoginRequest represents the JSON request for login
type LoginRequest struct {
	ID          string `json:"id,omitempty" db:"ID"`                                      // Optional: can be used for direct ID login
	UserID      string `json:"userId,omitempty" db:"user_id"`                             // Optional: can be used for direct user ID login
	PhoneNumber string `json:"phoneNumber" db:"phone" validate:"required,numeric,len=11"` // Allow login via phone number
	Password    string `json:"password" db:"password_hash" validate:"required"`           // Expecting the password hash in the database
}

// Validator instance for struct validation
var validate = validator.New()

// View Handler (GET)
// Serves the login page template
func (h *Handler) LoginView(w http.ResponseWriter, r *http.Request) {
	log.Println("Login view is running...")
	h.Tpl.ExecuteTemplate(w, "login.html", nil)
}

// Accepts JSON login credentials: username, email, or full_name
// Returns JSON response with success status and message
func (h *Handler) LoginAuthHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("LoginAuth is running...")

	// Parse JSON request body
	var loginReq LoginRequest // Define a struct to hold the login request data
	if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
		log.Printf("Error decoding login JSON: %v", err)
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "Invalid request format",
		})
		return
	}
	log.Printf("Login attempt for: %s", loginReq.PhoneNumber)

	err := validate.Struct(loginReq)
	if err != nil {
		log.Printf("Validation failed: %v", err)

		// Set a default generic message just in case
		errorMessage := "Invalid input provided."

		// Parse the specific validation errors
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			firstErr := validationErrs[0] // Just look at the first error to keep it simple

			// 3. Update the message based on exactly what failed
			if firstErr.Field() == "PhoneNumber" {
				errorMessage = "Please enter a valid 11-digit phone number."
			} else if firstErr.Field() == "Password" {
				errorMessage = "Please enter your password."
			}
		}
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: errorMessage,
		})
		return
	}

	// Query to check if credential matches
	var (
		hash   string // Store the password hash from the database
		ID     string // Store the ID
		userID string // Store the user ID for successful login response
	)

	stmt := "SELECT ID, user_id, password_hash FROM users WHERE phone = ?"
	err = h.DB.QueryRow(stmt, loginReq.PhoneNumber).Scan(&ID, &userID, &hash)

	// User not found
	if err != nil {
		log.Printf("User not found or DB error: %v", err)
		jsonwrite.WriteJSON(w, http.StatusUnauthorized, jsonwrite.APIResponse{
			Success: false,
			Message: "Incorrect phone number or password",
		})
		return
	}

	// Verify password
	log.Println("Hash found, verifying password...")
	if err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(loginReq.Password)); err != nil {
		log.Printf("Password mismatch for user: %s", loginReq.PhoneNumber)
		jsonwrite.WriteJSON(w, http.StatusUnauthorized, jsonwrite.APIResponse{
			Success: false,
			Message: "Password is incorrect",
		})
		return
	}

	// SUCCESS
	log.Printf("Login success for user: %s", loginReq.PhoneNumber)
	jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.LoginResponse{
		Success: true,
		Message: "Login successful",
		ID:      ID,
		UserID:  userID,
	})
}
