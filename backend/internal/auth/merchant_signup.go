package authentication

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strings"
	"time"
	"unicard-go/backend/internal/pkg/account"
	jsonwrite "unicard-go/backend/internal/pkg/handler"

	"github.com/go-playground/validator/v10"
)

type MerchantSignupRequest struct {
	BusinessName     string `json:"businessName" validate:"required"`
	BusinessType     string `json:"businessType" validate:"required"`
	BusinessAddress  string `json:"businessAddress" validate:"required"`
	OwnerName        string `json:"ownerName" validate:"required"`
	BusinessPhone    string `json:"businessPhone" validate:"required"`
	BusinessEmail    string `json:"businessEmail" validate:"required,email"`
	Password         string `json:"password" validate:"required,min=6"`
	BusinessDocument string `json:"businessDocument"`
	BirDocument      string `json:"birDocument"`
	OtherDocument    string `json:"otherDocument"`
}

// Helper to save base64 to R2 Storage
func (h *Handler) saveBase64ToR2(ctx context.Context, b64data string) string {
	if b64data == "" || h.Storage == nil {
		return ""
	}
	url, err := h.Storage.UploadBase64(ctx, b64data)
	if err != nil {
		log.Printf("Error uploading base64 to R2: %v", err)
		return ""
	}
	return url
}

func (h *Handler) MerchantSignupView(w http.ResponseWriter, r *http.Request) {
	log.Printf("Merchant Signup view is running...")
	h.Tpl.ExecuteTemplate(w, "merchant_signup.html", nil)
}

func (h *Handler) MerchantSignupHandler(w http.ResponseWriter, r *http.Request) {
	var req MerchantSignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding merchant signup JSON: %v", err)
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "Failed to parse JSON request",
		})
		return
	}

	// Clean inputs
	req.BusinessName = strings.Title(strings.ToLower(strings.TrimSpace(req.BusinessName)))
	req.BusinessAddress = strings.Title(strings.ToLower(strings.TrimSpace(req.BusinessAddress)))
	req.OwnerName = strings.Title(strings.ToLower(strings.TrimSpace(req.OwnerName)))
	req.BusinessEmail = strings.ToLower(strings.TrimSpace(req.BusinessEmail))
	req.BusinessPhone = strings.TrimSpace(req.BusinessPhone)
	req.BusinessType = strings.TrimSpace(req.BusinessType)
	req.Password = strings.TrimSpace(req.Password)

	// Validate inputs
	err := Validate.Struct(req)
	if err != nil {
		log.Printf("Validation failed: %v", err)
		errorMessage := "Invalid input provided."
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			errorMap := map[string]string{
				"BusinessName":    "Business name is required.",
				"BusinessType":    "Business type is required.",
				"BusinessAddress": "Business address is required.",
				"OwnerName":       "Owner name is required.",
				"BusinessPhone":   "Business phone is required.",
				"BusinessEmail":   "Please provide a valid business email address.",
				"Password":        "Password must be at least 6 characters long.",
			}
			if msg, ok := errorMap[validationErrs[0].Field()]; ok {
				errorMessage = msg
			}
		}

		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: errorMessage,
		})
		return
	}

	// Check if email already exists
	exists, err := account.IsEmailExist(h.Store.DB(), req.BusinessEmail)
	if err != nil {
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Database error",
		})
		return
	}
	if exists {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "Email already registered",
		})
		return
	}

	// Hash password
	hashedPassword, err := account.HashPassword(req.Password)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "System error processing password.",
		})
		return
	}

	// Begin transaction
	ctx := r.Context()
	tx, err := h.Store.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "System error starting transaction.",
		})
		return
	}
	defer tx.Rollback()

	// Generate IDs (Format: YYMMminsecxxxxx)
	timestamp := time.Now().Format("01020605") // MMDDYYss

	nUser, _ := rand.Int(rand.Reader, big.NewInt(10000))
	userID := fmt.Sprintf("UNI-%s%04d", timestamp, nUser.Int64())

	nMerchant, _ := rand.Int(rand.Reader, big.NewInt(10000))
	merchantID := fmt.Sprintf("MCH-%s%04d", timestamp, nMerchant.Int64())

	// Insert User
	username := req.BusinessEmail // using email as username
	userStmt := `INSERT INTO users (user_id, username, name, email, phone_number, password_hash, role, status) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err = tx.ExecContext(ctx, userStmt, userID, username, req.OwnerName, req.BusinessEmail, req.BusinessPhone, string(hashedPassword), "merchant_admin", "active")
	if err != nil {
		log.Printf("Error creating user: %v", err)
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "Failed to create user account. Email or phone might already exist.",
		})
		return
	}

	// Generate registration number (UCBZ-MMDDss-xxxxxxxxxx)
	nReg, _ := rand.Int(rand.Reader, big.NewInt(10000000000))
	regNum := fmt.Sprintf("UCBZ-%s-%010d", time.Now().Format("010205"), nReg.Int64())

	// Save Documents
	bizDocPath := h.saveBase64ToR2(ctx, req.BusinessDocument)
	birPath := h.saveBase64ToR2(ctx, req.BirDocument)
	otherPath := h.saveBase64ToR2(ctx, req.OtherDocument)

	// Insert Merchant with placeholder 'PENDING' for settlement fields
	fixedCommissionRate := 2.00
	merchStmt := `INSERT INTO merchants (
		merchant_id, business_name, business_type, business_registration_number, business_address, 
		user_id, owner_name, business_email, business_phone, commission_rate, 
		status, business_document, bir_document, valid_id
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = tx.ExecContext(ctx, merchStmt,
		merchantID, req.BusinessName, req.BusinessType, regNum, req.BusinessAddress,
		userID, req.OwnerName, req.BusinessEmail, req.BusinessPhone, fixedCommissionRate,
		"pending approval",
		bizDocPath, birPath, otherPath,
	)

	if err != nil {
		log.Printf("Error creating merchant: %v", err)
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "Failed to create merchant profile.",
		})
		return
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Error committing tx: %v", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Failed to finalize account creation",
		})
		return
	}

	jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{
		Success: true,
		Message: "Merchant application submitted successfully",
	})
}
