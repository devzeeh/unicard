package authentication

import (
	"database/sql"
	"log"
)

// isUserIDExist checks if a given user ID already exists in the database.
// It queries the users table and returns true if the ID is found, false otherwise.
func (h *Handler) isUserIDExist(userID int64) (bool, error) {
	var tmpId int64
	query := "SELECT user_id FROM users WHERE user_id = ?"
	err := h.Store.QueryRow(query, userID).Scan(&tmpId)

	if err == sql.ErrNoRows {
		return false, nil // Doesn't exist!
	} else if err != nil {
		return false, err // Real DB error
	}
	return true, nil // It exists
}

// It check the initial balance based on card number prefix
// Returns the initial balance as float64 or an error if any occurs.
// Gets the initial balance from the "card" table in the database.
// Example: Card Number "1234567890" has initial balance of 100.0
func (h *Handler) GetInitialBalance(cardNumber string) (float64, error) {
	var initialBalance float64 // to hold the initial balance

	query := "SELECT balance FROM cards WHERE card_number = ?"
	err := h.Store.QueryRow(query, cardNumber).Scan(&initialBalance)
	if err != nil {
		log.Printf("GetInitialBalance error for card %s: %v", cardNumber, err)
		return 0, err
	}
	return initialBalance, nil
}

// This function checks if a given phone number already exists in the database.
// It executes a SQL query to search for the phone number in the users table.
// If the phone number is found, it returns true. If not found, it returns false.
// If an error occurs during the query, it returns the error.
func (h *Handler) isPhoneExist(phone string) (bool, error) {
	// Hold the existing phone number
	var existingPhone string

	// Check query
	query := "SELECT phone_number FROM users WHERE phone_number = ?"
	err := h.Store.QueryRow(query, phone).Scan(&existingPhone)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		log.Printf("Phone number check error: %v", err)
		return false, err
	}
	return true, nil
}
