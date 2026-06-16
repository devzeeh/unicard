package admin

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/shopspring/decimal"
	jsonwrite "unicard-go/backend/internal/pkg/handler"
)

// TerminalSimView renders the terminal simulation page
func (h *Handler) TerminalSimView(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Terminal Simulation view is running...")

	type Merchant struct {
		ID   string
		Name string
	}
	var merchants []Merchant

	rows, err := h.DB.Query("SELECT merchant_id, business_name FROM merchants")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var m Merchant
			if err := rows.Scan(&m.ID, &m.Name); err == nil {
				merchants = append(merchants, m)
			}
		}
	}

	data := struct {
		Merchants []Merchant
	}{
		Merchants: merchants,
	}

	h.Tpl.ExecuteTemplate(w, "terminal_sim.html", data)
}

// TerminalSimTransactionHandler handles simulated terminal transactions
func (h *Handler) TerminalSimTransactionHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("TerminalSimTransactionHandler is running...")

	type SimRequest struct {
		CardNumber string  `json:"card_number"`
		Type       string  `json:"type"`
		Amount     float64 `json:"amount"`
		MerchantID string  `json:"merchant_id"`
	}

	var req SimRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "Invalid request payload",
		})
		return
	}

	if req.CardNumber == "" || req.Amount <= 0 || req.Type == "" {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "Missing required fields",
		})
		return
	}

	// 1. Check if card exists, is active, and linked to a user
	var balance float64
	var status string
	var userID sql.NullString

	err := h.DB.QueryRow(`
		SELECT balance, status, user_id 
		FROM cards 
		WHERE card_number = ?
	`, req.CardNumber).Scan(&balance, &status, &userID)

	if err != nil {
		if err == sql.ErrNoRows {
			jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
				Success: false,
				Message: "Card not found",
			})
			return
		}
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Database error",
		})
		return
	}

	if status != "active" {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "Card is not active",
		})
		return
	}

	if !userID.Valid || userID.String == "" {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "Card is not linked to any user",
		})
		return
	}

	// 2. Check balance
	if req.Type != "Refund" && balance < req.Amount {
		jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{
			Success: false,
			Message: fmt.Sprintf("Insufficient balance. Current balance: %.2f", balance),
		})
		return
	}

	// 2.5 Get merchant commission rate
	var commissionRate float64
	err = h.DB.QueryRow("SELECT commission_rate FROM merchants WHERE merchant_id = ?", req.MerchantID).Scan(&commissionRate)
	if err != nil {
		commissionRate = 2.00 // default fallback
	}

	serviceFee := req.Amount * (commissionRate / 100.0)

	amountDec := decimal.NewFromFloat(req.Amount)
	loyaltyPoints := amountDec.Mul(decimal.NewFromFloat(0.002))

	// 3. Process Transaction (Start TX)
	tx, err := h.DB.Begin()
	if err != nil {
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Failed to start transaction",
		})
		return
	}

	// Deduct or add balance and adjust loyalty points based on type
	if req.Type == "Refund" {
		_, err = tx.Exec(`UPDATE cards SET balance = balance + ?, loyalty_points = loyalty_points - ? WHERE card_number = ?`, req.Amount, loyaltyPoints, req.CardNumber)
	} else {
		_, err = tx.Exec(`UPDATE cards SET balance = balance - ?, loyalty_points = loyalty_points + ? WHERE card_number = ?`, req.Amount, loyaltyPoints, req.CardNumber)
	}
	if err != nil {
		tx.Rollback()
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Failed to deduct balance",
		})
		return
	}

	// Insert transaction record
	transactionID := fmt.Sprintf("TXN-SIM-%d", time.Now().UnixNano())

	// Get a dummy terminal ID
	var terminalID string
	err = h.DB.QueryRow("SELECT terminal_id FROM terminals LIMIT 1").Scan(&terminalID)
	if err != nil {
		terminalID = "TRM-SIM-001" // Fallback if no terminals exist
	}

	// Get a dummy processed_by user ID
	var processedBy string
	err = h.DB.QueryRow("SELECT user_id FROM users LIMIT 1").Scan(&processedBy)
	if err != nil {
		processedBy = "USR-SIM-001" // Fallback if no users exist
	}

	// Determine transaction type
	dbTransactionType := "payment"
	if req.Type == "Refund" {
		dbTransactionType = "refund"
	}

	_, err = tx.Exec(`
		INSERT INTO transactions (transaction_id, card_number, merchant_id, terminal_id, transaction_type, amount, service_fee, processed_by, points_earned) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, transactionID, req.CardNumber, req.MerchantID, terminalID, dbTransactionType, req.Amount, serviceFee, processedBy, loyaltyPoints)

	if err != nil {
		// If `merchant_id` is not nullable and causes error or id doesn't auto-increment
		// let's try with dummy data if needed, but we rollback first
		tx.Rollback()
		fmt.Printf("Error inserting transaction: %v\n", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Failed to record transaction",
		})
		return
	}

	if err = tx.Commit(); err != nil {
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Failed to commit transaction",
		})
		return
	}

	jsonwrite.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"success":     true,
		"message":     "Transaction successful",
		"service_fee": serviceFee,
	})
}
