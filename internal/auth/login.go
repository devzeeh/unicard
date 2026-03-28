package authentication

import (
	"encoding/json"
	"log"
	"net/http"
	jsonwrite "unicard-go/internal/pkg/handler"

	"golang.org/x/crypto/bcrypt"
)

// LoginRequest represents the JSON request for login
type LoginRequest struct {
	Username string `json:"username"` // username, email, or full_name
	Password string `json:"password"`
}

// View Handler (GET)
// Serves the login page template
func (h *Handler) LoginView(w http.ResponseWriter, r *http.Request) {
	log.Println("Login view is running...")
	h.Tpl.ExecuteTemplate(w, "login.html", nil)
}

// Auth Handler (POST)
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
	log.Printf("Login attempt for: %s", loginReq.Username)

	// Validate input
	fields := []string{loginReq.Username, loginReq.Password}
	for _, f := range fields {
		if f == "" {
			log.Println("Validation failed: empty fields")
			jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
				Success: false,
				Message: "Please fill in all fields.",
			})
			return
		}
	}

	// Query to check if credential matches username, email, or full_name
	var hash string
	var userID string
	stmt := "SELECT user_id, password_hash FROM users WHERE username = ? OR email = ? OR full_name = ?"
	err := h.DB.QueryRow(stmt, loginReq.Username, loginReq.Username, loginReq.Username).Scan(&userID, &hash)

	// User not found
	if err != nil {
		log.Printf("User not found or DB error: %v", err)
		jsonwrite.WriteJSON(w, http.StatusUnauthorized, jsonwrite.APIResponse{
			Success: false,
			Message: "Invalid credentials",
		})
		return
	}

	// Verify password
	log.Println("Hash found, verifying password...")
	if err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(loginReq.Password)); err != nil {
		log.Printf("Password mismatch for user: %s", loginReq.Username)
		jsonwrite.WriteJSON(w, http.StatusUnauthorized, jsonwrite.APIResponse{
			Success: false,
			Message: "Invalid credentials",
		})
		return
	}

	// SUCCESS
	log.Printf("Login success for user: %s", loginReq.Username)
	jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.LoginResponse{
		Success: true,
		Message: "Login successful",
		UserID:  userID,
	})
}
