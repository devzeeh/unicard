package authentication

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"time"
	"unicode"

	"unicard-go/backend/internal/pkg/account"
	jsonwrite "unicard-go/backend/internal/pkg/handler"
	smtp "unicard-go/backend/internal/pkg/smtpbody"

	"github.com/go-playground/validator/v10"
	"gopkg.in/gomail.v2"
)

type OTPData struct {
	OTP    string
	Expiry time.Time
}

// Forgot and Reset Password Request
type ForgotPasswordRequest struct {
	Email       string `json:"email" validate:"required,email" db:"email"`
	OTP         string `json:"otp" validate:"required,numeric,len=6"`
	NewPassword string `json:"new_password" validate:"required,min=8" db:"password_hash"`
}

var otpStore = make(map[string]OTPData)
var validate *validator.Validate

func init() {
	validate = validator.New()
}

// Forgot Password View
func (h *Handler) ForgotPasswordView(w http.ResponseWriter, r *http.Request) {
	log.Println("Forgot Password View")
	h.Tpl.ExecuteTemplate(w, "forgot-password.html", nil)
}

// Generate OTP Code
func generateOTP() string {
	max := big.NewInt(1000000)
	n, _ := rand.Int(rand.Reader, max)

	return fmt.Sprintf("%06d", n.Int64())
}

// Send OTP to email
func sendEmailOTP(email, name, otp string) error {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := 587
	smtpEmail := os.Getenv("SMTP_EMAIL")
	smtpSender := os.Getenv("SMTP_SENDER")
	smtpPass := os.Getenv("SMTP_PASSWORD")

	m := gomail.NewMessage()
	m.SetHeader("From", smtpSender+" <"+smtpEmail+">")
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Unicard Password Reset OTP")

	htmlBody := fmt.Sprintf(smtp.OTPCode(), name, otp)

	m.SetBody("text/html", htmlBody)

	d := gomail.NewDialer(
		smtpHost,
		smtpPort,
		smtpEmail,
		smtpPass,
	)

	err := d.DialAndSend(m)
	if err != nil {
		return err
	}

	log.Printf("OTP sent successfully to %s", email)
	return nil
}

// Forgot Password Send OTP
func (h *Handler) ForgotPasswordSendOTP(w http.ResponseWriter, r *http.Request) {
	// get context from request
	ctx := r.Context()

	var req ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("Error decoding request:", err)
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "Invalid input",
		})
		return
	}

	exists, err := account.IsEmailExist(h.DB, req.Email)
	if err != nil {
		log.Println("Error checking email existence:", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "System error",
		})
		return
	}
	if !exists {
		// Even if not exists, return success to prevent email enumeration
		jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{
			Success: true,
			Message: "If the email is found, an OTP has been sent.",
		})
		return
	}

	// Fetch the user's name
	var fullName string
	err = h.DB.QueryRowContext(ctx, "SELECT full_name FROM users WHERE email = ?", req.Email).Scan(&fullName)
	if err != nil {
		fullName = "there" // Fallback if name is not found
	}

	// Generate OTP that valid for 5 minutes
	otp := generateOTP()
	otpStore[req.Email] = OTPData{
		OTP:    otp,
		Expiry: time.Now().Add(5 * time.Minute),
	}

	if err := sendEmailOTP(req.Email, fullName, otp); err != nil {
		log.Println("Error sending email:", err)
		http.Error(w, "Failed to send OTP", http.StatusInternalServerError)
		return
	}

	jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{
		Success: true,
		Message: "OTP sent successfully",
	})
}

// Forgot Password Verify OTP
func (h *Handler) ForgotPasswordVerifyOTP(w http.ResponseWriter, r *http.Request) {
	var req ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("Error decoding request:", err)
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "Invalid input",
		})
		return
	}

	data, ok := otpStore[req.Email]
	if !ok || data.OTP != req.OTP {
		http.Error(w, "Invalid OTP", http.StatusUnauthorized)
		return
	}

	if time.Now().After(data.Expiry) {
		delete(otpStore, req.Email)
		http.Error(w, "OTP expired", http.StatusUnauthorized)
		return
	}

	jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{
		Success: true,
		Message: "OTP verified",
	})
}

// Reset Password Handler
func (h *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("Error decoding request:", err)
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "Invalid input",
		})
		return
	}

	// Verify OTP again
	data, ok := otpStore[req.Email]
	if !ok || data.OTP != req.OTP {
		jsonwrite.WriteJSON(w, http.StatusUnauthorized, jsonwrite.APIResponse{
			Success: false,
			Message: "Invalid or expired OTP",
		})
		return
	}

	// Validate Password
	if err := validatePassword(req.NewPassword); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Hash password
	hashedPassword, err := account.HashPassword(req.NewPassword)
	if err != nil {
		http.Error(w, "System error", http.StatusInternalServerError)
		return
	}

	// Update DB
	if err := h.updatePassword(req.Email, hashedPassword); err != nil {
		http.Error(w, "System error", http.StatusInternalServerError)
		return
	}

	// Clean up OTP
	delete(otpStore, req.Email)

	jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{
		Success: true,
		Message: "Password updated successfully",
	})
}

// Validate password helper function
func validatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, c := range password {
		switch {
		case unicode.IsUpper(c):
			hasUpper = true

		case unicode.IsLower(c):
			hasLower = true

		case unicode.IsNumber(c):
			hasNumber = true

		case unicode.IsPunct(c), unicode.IsSymbol(c):
			hasSpecial = true
		}
	}

	switch {
	case !hasUpper:
		return fmt.Errorf("password must contain at least one uppercase letter")

	case !hasLower:
		return fmt.Errorf("password must contain at least one lowercase letter")

	case !hasNumber:
		return fmt.Errorf("password must contain at least one number")

	case !hasSpecial:
		return fmt.Errorf("password must contain at least one special character")
	}

	return nil
}

// Update Password Handler
func (h *Handler) updatePassword(email, hashedPassword string) error {
	query := "UPDATE users SET password_hash = ? WHERE email = ?"
	_, err := h.DB.Exec(query, hashedPassword, email)
	if err != nil {
		log.Printf("failed to update password: %v", err)
		return err
	}
	return nil
}
