package admin

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strings"
	"time"
	jsonwrite "unicard-go/backend/internal/pkg/handler"

	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
)

// AddMerchantRequest represents the payload for adding a new merchant
type AddMerchantRequest struct {
	BusinessName      string `json:"businessName" validate:"required" db:"business_name"`
	BusinessType      string `json:"businessType" validate:"required" db:"business_type"`
	RegistrationNum   string `json:"registrationNum" db:"registration_num"`
	BusinessAddress   string `json:"businessAddress" validate:"required" db:"business_address"`
	OwnerName         string `json:"ownerName" validate:"required" db:"owner_name"`
	BusinessEmail     string `json:"businessEmail" validate:"required,email" db:"business_email"`
	BusinessPhone     string `json:"businessPhone" validate:"required" db:"business_phone"`
	CommissionRate    string `json:"commissionRate" db:"commission_rate"`
	SettlementName    string `json:"settlementName" validate:"required" db:"settlement_name"`
	SettlementAccount string `json:"settlementAccount" validate:"required" db:"settlement_account_number"`
	SettlementBank    string `json:"settlementBank" validate:"required" db:"settlement_bank_name"`
	TerminalSN        string `json:"terminalSn" validate:"required" db:"terminal_sn"`
	DeviceName        string `json:"deviceName" db:"device_name"`
}

// AddMerchantHandler creates new merchants and their corresponding owner users in bulk
func (h *Handler) AddMerchantHandler(w http.ResponseWriter, r *http.Request) {
	var reqs []AddMerchantRequest
	if err := json.NewDecoder(r.Body).Decode(&reqs); err != nil {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "Invalid JSON payload format. Expected array of merchants.",
		})
		return
	}

	if len(reqs) == 0 {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "No merchants provided.",
		})
		return
	}

	tx, err := h.DB.Begin()
	if err != nil {
		log.Printf("Error starting tx: %v", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Database error",
		})
		return
	}

	// Prepare statements
	userStmt, err := tx.Prepare(`INSERT INTO users (user_id, username, name, email, phone_number, password_hash, role, status) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		tx.Rollback()
		log.Printf("Error preparing user stmt: %v", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Database error",
		})
		return
	}
	defer userStmt.Close()

	merchStmt, err := tx.Prepare(`INSERT INTO merchants (
		merchant_id, business_name, business_type, business_registration_number, business_address, 
		user_id, owner_name, business_email, business_phone, commission_rate, 
		settlement_account_name, settlement_account_number, settlement_bank_name, status
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		tx.Rollback()
		log.Printf("Error preparing merchant stmt: %v", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Database error",
		})
		return
	}
	defer merchStmt.Close()

	// Update terminal status to active and assign merchant_id
	termStmt, err := tx.Prepare(`UPDATE terminals SET merchant_id = ?, location_details = ?, status = 'active' WHERE terminal_sn = ?`)
	if err != nil {
		tx.Rollback()
		log.Printf("Error preparing terminal stmt: %v", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Database error",
		})
		return
	}
	defer termStmt.Close()

	for i, req := range reqs {
		// Clean and format string fields
		req.BusinessName = strings.Title(strings.ToLower(strings.TrimSpace(req.BusinessName)))
		req.BusinessAddress = strings.Title(strings.ToLower(strings.TrimSpace(req.BusinessAddress)))
		req.OwnerName = strings.Title(strings.ToLower(strings.TrimSpace(req.OwnerName)))
		req.SettlementName = strings.Title(strings.ToLower(strings.TrimSpace(req.SettlementName)))
		
		// Some fields don't need title case but should be trimmed
		req.BusinessEmail = strings.ToLower(strings.TrimSpace(req.BusinessEmail))
		req.BusinessPhone = strings.TrimSpace(req.BusinessPhone)
		req.TerminalSN = strings.TrimSpace(req.TerminalSN)
		req.DeviceName = strings.TrimSpace(req.DeviceName)
		
		reqs[i] = req // update back to slice

		err := Validate.Struct(req)
		if err != nil {
			tx.Rollback()
			log.Printf("Validation error on merchant %d: %v", i+1, err)
			var validationErrs validator.ValidationErrors
			msg := fmt.Sprintf("Validation failed on merchant #%d", i+1)
			if errors.As(err, &validationErrs) {
				firstErr := validationErrs[0]
				fieldMessages := map[string]string{
					"BusinessName":      "Business name is required",
					"BusinessType":      "Business type is required",
					"BusinessAddress":   "Business address is required",
					"OwnerName":         "Owner name is required",
					"BusinessEmail":     "A valid business email is required",
					"BusinessPhone":     "Business phone number is required",
					"SettlementName":    "Settlement name is required",
					"SettlementAccount": "Settlement account number is required",
					"SettlementBank":    "Settlement bank name is required",
					"TerminalSN":        "Terminal serial number is required",
				}
				if customMsg, ok := fieldMessages[firstErr.Field()]; ok {
					msg = fmt.Sprintf("Merchant #%d: %s", i+1, customMsg)
				}
			}
			jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
				Success: false,
				Message: msg,
			})
			return
		}

		// Generate IDs (Format: YYMMminsecxxxxx where xxxxx is 5 random numbers)
		timestamp := time.Now().Format("01020605") // MMDDYYss

		nUser, _ := rand.Int(rand.Reader, big.NewInt(10000)) // max 9999
		userID := fmt.Sprintf("UNI-%s%04d", timestamp, nUser.Int64())

		nMerchant, _ := rand.Int(rand.Reader, big.NewInt(10000))
		merchantID := fmt.Sprintf("MCH-%s%04d", timestamp, nMerchant.Int64())

		// Create user for the merchant owner
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("TempPass123!"), bcrypt.DefaultCost)
		if err != nil {
			tx.Rollback()
			log.Printf("Error hashing password: %v", err)
			jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
				Success: false,
				Message: "Failed to secure user credentials",
			})
			return
		}

		// Using business email as username for simplicity
		username := req.BusinessEmail
		_, err = userStmt.Exec(userID, username, req.OwnerName, req.BusinessEmail, req.BusinessPhone, string(hashedPassword), "merchant_admin", "active")
		if err != nil {
			tx.Rollback()
			log.Printf("Error creating user %d: %v", i+1, err)
			jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to create user account for Merchant #%d (email or phone might already exist)", i+1),
			})
			return
		}

		// Generate registration number (UCBZ-MMDDss-xxxxxxxxxx)
		nReg, _ := rand.Int(rand.Reader, big.NewInt(10000000000))
		regNum := fmt.Sprintf("UCBZ-%s-%010d", time.Now().Format("010205"), nReg.Int64())

		// Set commission rate
		fixedCommissionRate := 2.00

		_, err = merchStmt.Exec(
			merchantID, req.BusinessName, req.BusinessType, regNum, req.BusinessAddress,
			userID, req.OwnerName, req.BusinessEmail, req.BusinessPhone, fixedCommissionRate,
			req.SettlementName, req.SettlementAccount, req.SettlementBank, "active",
		)

		if err != nil {
			tx.Rollback()
			log.Printf("Error creating merchant %d: %v", i+1, err)
			jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to create profile for Merchant #%d (registration num or email might exist)", i+1),
			})
			return
		}

		// Update the existing terminal
		_, err = termStmt.Exec(userID, req.BusinessAddress, req.TerminalSN)
		if err != nil {
			tx.Rollback()
			log.Printf("Error creating terminal %d: %v", i+1, err)
			jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to register terminal for Merchant #%d (serial number might exist)", i+1),
			})
			return
		}
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Error committing tx: %v", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Failed to finalize batch creation",
		})
		return
	}

	jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{
		Success: true,
		Message: fmt.Sprintf("Successfully onboarded %d merchant(s)", len(reqs)),
	})
}
