package admin

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	jsonwrite "unicard-go/backend/internal/pkg/handler"

	"github.com/shopspring/decimal"
)

type Merchant struct {
	ID   string
	Name string
}

type SimRequest struct {
	CardNumber string          `json:"card_number"`
	Type       string          `json:"type"`
	Amount     decimal.Decimal `json:"amount"`
	MerchantID string          `json:"merchant_id"`
	Balance    decimal.Decimal `json:"balance"`
	Status     string          `json:"status"`
	UserID     *string         `json:"user_id"`
}

// TerminalSimView renders the terminal simulation page
func (h *Handler) TerminalSimView(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Terminal Simulation view is running...")

	var merchants []Merchant

	rows, err := h.Store.Query("SELECT merchant_id, business_name FROM merchants")
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

	var req SimRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "Invalid request payload",
		})
		return
	}

	if req.CardNumber == "" || req.Amount.LessThanOrEqual(decimal.Zero) || req.Type == "" {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "Missing required fields",
		})
		return
	}

	// Check if card exists, is active, and linked to a user

	err := h.Store.QueryRow(`
		SELECT balance, status, user_id 
		FROM cards 
		WHERE card_number = ?
	`, req.CardNumber).Scan(&req.Balance, &req.Status, &req.UserID)

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

	if req.Status != "active" {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "Card is not active",
		})
		return
	}

	if req.UserID == nil || *req.UserID == "" {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "Card is not linked to any user",
		})
		return
	}

	// Check balance
	if req.Type != "Refund" && req.Balance.LessThan(req.Amount) {
		log.Printf("Insufficient Balance : %s, req Amount: %s", req.Balance, req.Amount)
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "Insufficient balance",
		})
		return
	}

	// Get merchant commission rate
	var commissionRate decimal.Decimal
	err = h.Store.QueryRow("SELECT commission_rate FROM merchants WHERE merchant_id = ?", req.MerchantID).Scan(&commissionRate)
	if err != nil {
		commissionRate = decimal.NewFromFloat(2) // default fallback
	}

	// Calculate service fee
	serviceFee := req.Amount.Mul(commissionRate.Div(decimal.NewFromFloat(100)))

	// Calculate loyalty points
	loyaltyPoints := req.Amount.Mul(decimal.NewFromFloat(0.002))

	// Process Transaction (Start TX)
	tx, err := h.Store.Begin()
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
	err = h.Store.QueryRow("SELECT terminal_id FROM terminals LIMIT 1").Scan(&terminalID)
	if err != nil {
		terminalID = "TRM-SIM-001" // Fallback if no terminals exist
	}

	// Get a dummy processed_by user ID
	var processedBy string
	err = h.Store.QueryRow("SELECT user_id FROM users LIMIT 1").Scan(&processedBy)
	if err != nil {
		processedBy = "USR-SIM-001" // Fallback if no users exist
	}

	// Determine transaction type
	dbTransactionType := "payment"
	if req.Type == "Refund" {
		dbTransactionType = "refund"
	}

	var userID string
	_ = tx.QueryRow("SELECT user_id FROM cards WHERE card_number = ?", req.CardNumber).Scan(&userID)

	_, err = tx.Exec(`
		INSERT INTO transactions (transaction_id, card_number, user_id, merchant_id, terminal_id, transaction_type, amount, service_fee, processed_by, points_earned) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, transactionID, req.CardNumber, userID, req.MerchantID, terminalID, dbTransactionType, req.Amount, serviceFee, processedBy, loyaltyPoints)

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

	jsonwrite.WriteJSON(w, http.StatusOK, map[string]any{
		"success":     true,
		"message":     "Transaction successful",
		"service_fee": serviceFee,
	})
}
