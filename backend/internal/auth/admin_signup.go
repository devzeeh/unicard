package authentication

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"unicard-go/backend/internal/pkg/account"
	jsonwrite "unicard-go/backend/internal/pkg/handler"
)

type AdminSignupRequest struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) AdminSignupView(w http.ResponseWriter, r *http.Request) {
	log.Printf("Admin Signup view is running...")
	h.Tpl.ExecuteTemplate(w, "admin_signup.html", nil)
}

func (h *Handler) AdminSignupHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Admin Signup Handler is running...")

	ctx := r.Context()

	var req AdminSignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "Invalid request format",
		})
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(req.Email)
	req.Password = strings.TrimSpace(req.Password)

	if req.Name == "" || req.Username == "" || req.Email == "" || req.Password == "" {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "All fields are required",
		})
		return
	}

	hashedPassword, err := account.HashPassword(req.Password)
	if err != nil {
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Error processing password",
		})
		return
	}

	generateUserId, err := h.GenerateUserID()
	if err != nil {
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Error generating User ID",
		})
		return
	}
	userIDStr := fmt.Sprintf("%d", generateUserId)

	insertQuery := `INSERT INTO users 
    (user_id, username, name, email, password_hash, role, status) 
    VALUES (?, ?, ?, ?, ?, 'super_admin', 'active')`
	
	_, err = h.DB.ExecContext(ctx, insertQuery, userIDStr, req.Username, req.Name, req.Email, hashedPassword)
	if err != nil {
		log.Printf("Error creating admin user: %v", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Database error creating account. Email or Username might be taken.",
		})
		return
	}

	jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{
		Success: true,
		Message: "Super Admin account created successfully!",
	})
}
