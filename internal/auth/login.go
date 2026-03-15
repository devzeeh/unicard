package authentication

import (
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

// LoginRequest represents the JSON request for login
type LoginRequest struct {
	Username string `json:"username"` // username, email, or full_name
	Password   string `json:"password"`
}

// LoginResponse represents the JSON response for login
type LoginResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	UserID  string `json:"user_id,omitempty"`
}

// View Handler (GET)
// Serves the login page template
func (h *Handler) LoginView(w http.ResponseWriter, r *http.Request) {
	h.Tpl.ExecuteTemplate(w, "login.html", nil)
}

// Auth Handler (POST)
// Accepts JSON login credentials: username, email, or full_name
// Returns JSON response with success status and message
func (h *Handler) LoginAuthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("loginauth running...")

	w.Header().Set("Content-Type", "application/json")

	// Parse JSON request body
	var loginReq LoginRequest // Define a struct to hold the login request data
	err := json.NewDecoder(r.Body).Decode(&loginReq)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(LoginResponse{
			Success: false,
			Message: "Invalid request format",
		})
		return
	}

	// Validate input
	if loginReq.Username == "" || loginReq.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(LoginResponse{
			Success: false,
			Message: "Credential and password are required",
		})
		return
	}

	var hash string
	var userID string
	// Query to check if credential matches username, email, or full_name
	stmt := "SELECT id, password_hash FROM users WHERE username = ? OR email = ? OR full_name = ?"

	err = h.DB.QueryRow(stmt, loginReq.Username, loginReq.Username, loginReq.Username).Scan(&userID, &hash)

	// User not found
	if err != nil {
		fmt.Println("User not found or DB error:", err)
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(LoginResponse{
			Success: false,
			Message: "User not found",
		})
		return
	}

	fmt.Println("Hash found, verifying...")
	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(loginReq.Password))

	// SUCCESS
	if err == nil {
		fmt.Println("Login success")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(LoginResponse{
			Success: true,
			Message: "Login successful",
			UserID:  userID,
		})
		return
	}

	// Password mismatch
	fmt.Println("Password Incorrect:", err)
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(LoginResponse{
		Success: false,
		Message: "Password incorrect",
	})
}
