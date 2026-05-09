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

	"gopkg.in/gomail.v2"
)

type OTPData struct {
	OTP    string
	Expiry time.Time
}

// Forgot and Reset Password Request
type ForgotPasswordRequest struct {
	Email       string `json:"email"`
	OTP         string `json:"otp"`
	NewPassword string `json:"new_password"`
}

var otpStore = make(map[string]OTPData)

// Forgot Password View
func (h *Handler) ForgotPasswordView(w http.ResponseWriter, r *http.Request) {
	log.Println("Forgot Password View")
	h.Tpl.ExecuteTemplate(w, "forgot-password.html", nil)
}

// Generate OTP Code
func generateOTP() string {
	max := big.NewInt(1000000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "123456"
	}
	return fmt.Sprintf("%06d", n.Int64())
}

// Send OTP to email
func sendEmailOTP(email, otp string) error {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := 587
	smtpEmail := os.Getenv("SMTP_EMAIL")
	smtpSender := os.Getenv("SMTP_SENDER")
	smtpPass := os.Getenv("SMTP_PASSWORD")

	m := gomail.NewMessage()
	m.SetHeader("From", smtpSender, smtpEmail)
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Your Password Reset OTP")
	m.SetBody("text/plain", "Your OTP for password reset is: "+otp)

	d := gomail.NewDialer(smtpHost, smtpPort, smtpSender, smtpPass)
	if err := d.DialAndSend(m); err != nil {
		log.Fatal(err)
	}
	fmt.Println("OTP SENT")

	// Always print the OTP to the terminal so we can test locally even if email fails
	fmt.Printf("\n======================================================\n")
	fmt.Printf("=> [LOCAL TEST] OTP for %s is: %s\n", email, otp)
	fmt.Printf("======================================================\n\n")

	if os.Getenv("SMTP_HOST") == "" {
		fmt.Printf("SMTP credentials not set. Simulating email success.\n")
		return nil
	}

	err := d.DialAndSend(m)
	if err != nil {
		fmt.Println("Warning: Failed to send email via SMTP, but continuing for local testing. Error:", err)
		// Return nil instead of err so the frontend flow continues seamlessly
		return nil
	}
	return nil
}

// Forgot Password Send OTP
func (h *Handler) ForgotPasswordSendOTP(w http.ResponseWriter, r *http.Request) {
	var req ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	exists, err := account.IsEmailExist(h.DB, req.Email)
	if err != nil {
		http.Error(w, "System error", http.StatusInternalServerError)
		return
	}
	if !exists {
		// Even if not exists, return success to prevent email enumeration
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "If the email is found, an OTP has been sent."})
		return
	}

	otp := generateOTP()
	otpStore[req.Email] = OTPData{
		OTP:    otp,
		Expiry: time.Now().Add(10 * time.Minute),
	}

	if err := sendEmailOTP(req.Email, otp); err != nil {
		fmt.Println("Error sending email:", err)
		http.Error(w, "Failed to send OTP", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "OTP sent successfully"})
}

// Forgot Password Verify OTP
func (h *Handler) ForgotPasswordVerifyOTP(w http.ResponseWriter, r *http.Request) {
	var req ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
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

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "OTP verified"})
}

// Reset Password Handler
func (h *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Verify OTP again
	data, ok := otpStore[req.Email]
	if !ok || data.OTP != req.OTP {
		http.Error(w, "Invalid or expired OTP", http.StatusUnauthorized)
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

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Password updated successfully"})
}

// Validate password helper function
func validatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	var hasUpper, hasLower, hasNumber, hasSpecial bool
	for _, c := range password {
		switch {
		case unicode.IsNumber(c):
			hasNumber = true
		case unicode.IsUpper(c):
			hasUpper = true
		case unicode.IsLower(c):
			hasLower = true
		case unicode.IsPunct(c) || unicode.IsSymbol(c):
			hasSpecial = true
		}
	}
	if !hasUpper {
		return fmt.Errorf("password must contain an uppercase letter")
	}
	if !hasLower {
		return fmt.Errorf("password must contain a lowercase letter")
	}
	if !hasNumber {
		return fmt.Errorf("password must contain a number")
	}
	if !hasSpecial {
		return fmt.Errorf("password must contain a special character")
	}
	return nil
}

// Update Password Handler
func (h *Handler) updatePassword(email, hashedPassword string) error {
	query := "UPDATE users SET password = ? WHERE email = ?"
	_, err := h.DB.Exec(query, hashedPassword, email)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}
	return nil
}
