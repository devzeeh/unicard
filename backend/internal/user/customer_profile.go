package user

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"unicard-go/backend/internal/pkg/account"
	jsonwrite "unicard-go/backend/internal/pkg/handler"

	"golang.org/x/crypto/bcrypt"
)

// ProfileUpdateRequest represents the expected payload for updating user profile information
type ProfileUpdateRequest struct {
	Username        string `json:"username,omitempty" db:"username"`
	FullName        string `json:"full_name,omitempty" db:"name"`
	Email           string `json:"email,omitempty" db:"email"`
	Phone           string `json:"phone_number,omitempty" db:"phone_number"`
	CurrentPassword string `json:"current_password,omitempty"`
	NewPassword     string `json:"new_password,omitempty"`
	ConfirmPassword string `json:"confirm_password,omitempty"`
}

// ProfileView handles the display of the user's profile page
// It retrieves the username from the URL and renders the profile template
func (h *Handler) ProfileView(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Profile view is running...")

	// Extract the username from the URL path
	username := r.PathValue("username")
	data := struct {
		Username string
	}{
		Username: username,
	}

	// Render the profile template with the username data
	h.Tpl.ExecuteTemplate(w, "profile.html", data)
}

// ProfileEdit renders the profile edit page (assumes router already filtered by method)
func (h *Handler) ProfileEdit(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	ctx := r.Context()

	var req ProfileUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "Invalid request payload",
		})
		return
	}

	// PATCH — at least one field required
	if req.FullName == "" && req.Email == "" && req.Phone == "" {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "At least one field is required",
		})
		return
	}

	// Build query dynamically based on what was sent
	fields := []string{}
	args := []any{}

	if req.FullName != "" {
		fields = append(fields, "name = ?")
		args = append(args, req.FullName)
	}
	if req.Email != "" {
		fields = append(fields, "email = ?")
		args = append(args, req.Email)
	}
	if req.Phone != "" {
		fields = append(fields, "phone_number = ?")
		args = append(args, req.Phone)
	}

	// Append username as last arg for WHERE clause
	args = append(args, username)

	query := "UPDATE users SET " + strings.Join(fields, ", ") + " WHERE username = ?"

	_, err := h.DB.ExecContext(ctx, query, args...)
	if err != nil {
		log.Printf("ProfileEdit DB error: %v | query: %s | args: %v", err, query, args)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Failed to update profile",
		})
		return
	}

	jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{
		Success: true,
		Message: "Profile updated successfully",
	})
}

// current password verification is not implemented yet, so this endpoint just updates the password without checking the old one
// password hashing is also not implemented yet, so the password is stored in plaintext (this will be fixed in the future). (FIXED: password hashing is now implemented using bcrypt in the account package)
// lowercase the password field names in the struct to avoid accidentally exposing them in JSON responses. FIXED in the struct definition above.
func (h *Handler) ProfileChangePassword(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")

	var req ProfileUpdateRequest
	var ctx = r.Context()

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "Invalid request payload",
		})
		return
	}

	if req.CurrentPassword == "" || req.NewPassword == "" || req.ConfirmPassword == "" {
		log.Printf("One or more password fields are empty for user: %s", username)
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "All password fields are required",
		})
		return
	}

	if req.NewPassword != req.ConfirmPassword {
		log.Printf("New password and confirm password do not match for user: %s", username)
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "Password do not match",
		})
		return
	}

	// Fetch current password hash
	log.Printf("Verifying current password for user: %s", username)
	var currentHash string
	err := h.DB.QueryRowContext(ctx, "SELECT password_hash FROM users WHERE username = ?", username).Scan(&currentHash)
	if err != nil {
		log.Printf("Error fetching current password hash: %v", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Failed to verify current password",
		})
		return
	}

	// Verify current password is correct
	if err := bcrypt.CompareHashAndPassword([]byte(currentHash), []byte(req.CurrentPassword)); err != nil {
		log.Printf("Current password verification failed for user: %s", username)
		jsonwrite.WriteJSON(w, http.StatusUnauthorized, jsonwrite.APIResponse{
			Success: false,
			Message: "Current password is incorrect",
		})
		return
	}
	log.Printf("Current password verified for user: %s", username)

	// Prevent setting the same password again
	if err := bcrypt.CompareHashAndPassword([]byte(currentHash), []byte(req.NewPassword)); err == nil {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "New password must be different from current password",
		})
		return
	}

	// Hash the new password using the account package's HashPassword function
	hashedPassword, err := account.HashPassword(req.NewPassword)
	if err != nil {
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Failed to hash password",
		})
		return
	}

	// Here you would hash the password and update it in the database
	// Save new password hash
	query := "UPDATE users SET password_hash = ? WHERE username = ?"
	_, err = h.DB.ExecContext(ctx, query, hashedPassword, username)
	if err != nil {
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Failed to update password",
		})
		return
	}

	// Respond with success message
	jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{
		Success: true,
		Message: "Password updated successfully",
	})
}
