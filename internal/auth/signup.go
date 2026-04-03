package authentication

import (
	"encoding/json" // Added for JSON support
	"fmt"
	"log"
	"net/http"
	"strings"
	jsonwrite "unicard-go/internal/pkg/handler"
)

// View Handler (GET)
// You can now simplify this because JS handles the errors!
func (h *Handler) SignupView(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Signup view is running...")
	// Just serve the template. No need for the huge switch statement anymore.
	h.Tpl.ExecuteTemplate(w, "signup.html", nil)
}

func (h *Handler) SignupHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Signup API is running...")

	// Decode incoming JSON
	var req SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding signup JSON: %v", err)
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{Success: false, Message: "Failed to parse JSON request"})
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
			jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{Success: false, Message: "Please fill in all fields."})
			return
		}
	}

	// Validation: Password Length (fail fast before any DB calls)
	if len(req.Password) < 8 {
		log.Printf("Validation failed: Password must be at least 8 characters long")
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{Success: false, Message: "Password must be at least 8 characters long."})
		return
	}

	// We pass the cleaned 'req' to signup_service.go. It will return an error if something goes wrong.
	err := h.CreateAccount(req)
	if err != nil {
		// If the service fails (e.g., "Email already exists"), we send that specific message back to the user
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{Success: false, Message: err.Error()})
		return
	}

	// If everything is successful, we send a success message back to the user
	jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{Success: true, Message: "Account created successfully!"})
}
