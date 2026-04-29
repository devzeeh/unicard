package authentication

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
	"unicard-go/backend/internal/pkg/account"
	jsonwrite "unicard-go/backend/internal/pkg/handler"
)

// In-memory store for OTPs (In production, use Redis or DB)
var (
	otpStore = make(map[string]string)
	otpMutex sync.RWMutex
)

type CheckDetailsRequest struct {
	Email         string `json:"email"`
	ContactNumber string `json:"contact_number"`
}

type VerifyOTPRequest struct {
	Email string `json:"email"`
	OTP   string `json:"otp"`
}

type CheckCardRequest struct {
	CardNumber string `json:"card_number"`
}

// CheckDetailsHandler checks if email and phone are available and sends OTP
func (h *Handler) CheckDetailsHandler(w http.ResponseWriter, r *http.Request) {
	var req CheckDetailsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{Success: false, Message: "Invalid request"})
		return
	}

	// Check Email
	exists, err := account.IsEmailExist(h.DB, req.Email)
	if err != nil {
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{Success: false, Message: "Database error"})
		return
	}
	if exists {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{Success: false, Message: "Email already registered", Field: "email"})
		return
	}

	// Check Phone
	exists, err = h.isPhoneExist(req.ContactNumber)
	if err != nil {
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{Success: false, Message: "Database error"})
		return
	}
	if exists {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{Success: false, Message: "Phone number already registered", Field: "phone"})
		return
	}

	// Generate OTP
	otp := fmt.Sprintf("%06d", rand.Intn(1000000))
	otpMutex.Lock()
	otpStore[req.Email] = otp
	otpMutex.Unlock()

	// Log OTP for development (since we don't have SMTP configured here)
	log.Printf("SIGNUP OTP for %s: %s", req.Email, otp)

	jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{
		Success: true,
		Message: "OTP sent successfully",
	})
}

// VerifyOTPHandler checks the entered OTP
func (h *Handler) VerifyOTPHandler(w http.ResponseWriter, r *http.Request) {
	var req VerifyOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{Success: false, Message: "Invalid request"})
		return
	}

	otpMutex.RLock()
	storedOTP, ok := otpStore[req.Email]
	otpMutex.RUnlock()

	if !ok || storedOTP != req.OTP {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{Success: false, Message: "Invalid or expired OTP"})
		return
	}

	// Optional: Clear OTP after successful verification
	// otpMutex.Lock()
	// delete(otpStore[req.Email])
	// otpMutex.Unlock()

	jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{
		Success: true,
		Message: "OTP verified successfully",
	})
}

// CheckCardHandler checks if card is valid for registration
func (h *Handler) CheckCardHandler(w http.ResponseWriter, r *http.Request) {
	var req CheckCardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{Success: false, Message: "Invalid request"})
		return
	}

	var status string
	err := h.DB.QueryRow("SELECT status FROM cards WHERE card_number = ?", req.CardNumber).Scan(&status)
	if err != nil {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{Success: false, Message: "Card not found"})
		return
	}

	if status != "Inactive" {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{Success: false, Message: "Card is already active or restricted"})
		return
	}

	jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{
		Success: true,
		Message: "Card is valid",
	})
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
