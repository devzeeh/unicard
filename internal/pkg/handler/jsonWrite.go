package jsonwrite

import (
	"encoding/json"
	"net/http"
)

// Create a standard API response struct
type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// Login specific response — returns user data after login
type LoginResponse struct {
    Success bool   `json:"success"`
    Message string `json:"message"`
    UserID  string `json:"user_id"`
}

// Auth Handler (POST) - Converted to JSON API
func WriteJSON(w http.ResponseWriter, status int, resp any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(resp)
}
