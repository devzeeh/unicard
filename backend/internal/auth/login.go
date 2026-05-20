package authentication

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	jsonwrite "unicard-go/backend/internal/pkg/handler"

	"github.com/go-playground/validator/v10" // For struct validation
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	ID         string `json:"id,omitempty" db:"ID"`                                        // Optional: can be used for direct ID login
	UserID     string `json:"userId,omitempty" db:"user_id"`                               // Optional: can be used for direct user ID login
	Identifier string `json:"identifier" validate:"required" db:"full_name, email, phone"` // Allow login via email or username
	Password   string `json:"password" db:"password_hash" validate:"required"`             // Expecting the password hash in the database
}

// This struct is used to get the data from the database
type Login struct {
	ID string `db:"ID"`
	UserID string `db:"user_id"`
	Username string `db:"username"`
	Password string `db:"password_hash"`
}

// Validator instance for struct validation
// Initialize the validator for all handlers 
var Validate = validator.New()

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
	log.Printf("Login attempt for: %s", loginReq.Identifier)

	// validate the login request
	err := Validate.Struct(loginReq)
	if err != nil {
		log.Printf("Validation failed: %v", err)

		// Set a default generic message just in case
		errorMessage := "Invalid input provided."

		// Parse the specific validation errors
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			firstErr := validationErrs[0] // Just look at the first error to keep it simple

			// Update the message based on exactly what failed
			if firstErr.Field() == "Identifier" {
				errorMessage = "Please enter a valid email or username."
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

	// Query to check if credential matches in admin_users table first
	var adminID int
	var adminUsername string
	var adminHash string
	stmtAdmin := "SELECT id, username, password_hash FROM admin_users WHERE email = ? OR username = ?"
	errAdmin := h.DB.QueryRow(stmtAdmin, loginReq.Identifier, loginReq.Identifier).Scan(&adminID, &adminUsername, &adminHash)
	if errAdmin == nil {
		// Admin user found! Verify password
		log.Println("Admin hash found, verifying password...")
		if err = bcrypt.CompareHashAndPassword([]byte(adminHash), []byte(loginReq.Password)); err != nil {
			log.Printf("Password mismatch for admin user: %s", loginReq.Identifier)
			jsonwrite.WriteJSON(w, http.StatusUnauthorized, jsonwrite.APIResponse{
				Success: false,
				Message: "Password is incorrect",
			})
			return
		}
		// SUCCESS Admin Login
		log.Printf("Admin login success for user: %s", loginReq.Identifier)
		http.SetCookie(w, &http.Cookie{
			Name:     "session_admin_username",
			Value:    adminUsername,
			Path:     "/",
			HttpOnly: true,
		})
		jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.LoginResponse{
			Success:     true,
			Message:     "Admin login successful",
			ID:          fmt.Sprintf("%d", adminID),
			UserID:      adminUsername,
			RedirectURL: "/admin/dashboard",
		})
		return
	}

	// If not an admin, query the users table
	var (
		hash   string // Store the password hash from the database
		ID     string // Store the ID
		userID string // Store the user ID for successful login response
	)

	stmt := "SELECT ID, user_id, password_hash FROM users WHERE email = ? OR username = ? OR phone = ?"
	err = h.DB.QueryRow(stmt, loginReq.Identifier, loginReq.Identifier, loginReq.Identifier).Scan(&ID, &userID, &hash)

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
		log.Printf("Password mismatch for user: %s", loginReq.Identifier)
		jsonwrite.WriteJSON(w, http.StatusUnauthorized, jsonwrite.APIResponse{
			Success: false,
			Message: "Password is incorrect",
		})
		return
	}

	// SUCCESS User Login
	log.Printf("Login success for user: %s", loginReq.Identifier)
	http.SetCookie(w, &http.Cookie{
		Name:     "session_user_id",
		Value:    userID,
		Path:     "/",
		HttpOnly: true,
	})
	jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.LoginResponse{
		Success:     true,
		Message:     "Login successful",
		ID:          ID,
		UserID:      userID,
		RedirectURL: "/dashboard",
	})
}
