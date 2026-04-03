package authentication

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"log"
	"math/big"
	"time"
)

// isUserIDExist checks if a given user ID already exists in the database.
// It queries the users table and returns true if the ID is found, false otherwise.
func (h *Handler) isUserIDExist(userID int64) (bool, error) {
    var tmpId int64
    query := "SELECT user_id FROM users WHERE user_id = ?"
    err := h.DB.QueryRow(query, userID).Scan(&tmpId)
    
    if err == sql.ErrNoRows {
        return false, nil // Doesn't exist!
    } else if err != nil {
        return false, err // Real DB error
    }
    return true, nil // It exists
}

// GenerateUniqueUsername creates a random unique username
// Format: user + 12 random lowercase characters/numbers
// Example: user9d8a7c2b3e4f
func (h *Handler) GenerateUniqueUsername() (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	usernamePrefix := "user"
	length := 7

	for {
		// Get the current date in DDYY format (Go uses "010206" as reference)
		userDate := time.Now().Format("06") // e.g., "012624" for Jan 26, 2024
		// time .Now().Format("1504") // e.g., "1530" for 3:30 PM
		timePart := time.Now().Format("0405") // e.g., "1530" for 3:30 PM

		// Combine date and time to form part of the username
		//usernamePrefix = fmt.Sprintf("user%s%s", userDate, timePart)

		// Generate the random suffix
		randomPart := ""
		for i := 0; i < length; i++ {
			num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
			if err != nil {
				return "", err
			}
			randomPart += string(charset[num.Int64()])
		}

		// Combine prefix + date + time + random part
		username := fmt.Sprintf("%s%s%s%s", usernamePrefix, userDate, randomPart, timePart)

		// Check DB for uniqueness
		var existing string
		query := "SELECT username FROM users WHERE username = ?"
		err := h.DB.QueryRow(query, username).Scan(&existing)

		if err == sql.ErrNoRows {
			//fmt.Println("Generated unique username:", username)
			return username, nil // Found a unique one!
		} else if err != nil {
			return "", err // Real DB Error
		}

		// Collision detected, loop runs again...
		log.Println("Username collision! Retrying...")
	}
}

// It generates the unique cardID for every card users
// Checks the database for uniqueness.
// Returns the unique card ID as string or an error if any occurs.
// Example format: CARD-XXXXXXX
func (h *Handler) GenerateCardID() (string, error) {
	// Define charset: letters and numbers
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	cardIDPrefix := "CARD-"
	randomLength := 7

	for {
		// Get the current date in MMDDYY format (Go uses "010206" as reference)
		datePart := time.Now().Format("010206")

		// Generate the 7 random characters
		randomPart := ""
		for i := 0; i < randomLength; i++ {
			num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
			if err != nil {
				return "", err
			}
			randomPart += string(charset[num.Int64()])
		}

		// Combine them: CARD- + Date + Random
		// Example output: CARD-012626Ab7z9X1
		cardID := fmt.Sprintf("%s%s%s", cardIDPrefix, datePart, randomPart)

		// Check database for uniqueness
		var tmpCardID string
		query := "SELECT card_id FROM users WHERE card_id = ?"
		err := h.DB.QueryRow(query, cardID).Scan(&tmpCardID)

		if err == sql.ErrNoRows {
			return cardID, nil // Unique ID found
		} else if err != nil {
			return "", err // DB error
		}
		log.Printf("Collision detected! Retrying... conflicting ID: %s", cardID)
	}
}

// It check the initial balance based on card number prefix
// Returns the initial balance as float64 or an error if any occurs.
// Gets the initial balance from the "card" table in the database.
// Example: Card Number "1234567890" has initial balance of 100.0
func (h *Handler) GetInitialBalance(cardNumber string) (float64, error) {
	var initialBalance float64 // to hold the initial balance

	query := "SELECT initial_amount FROM cards WHERE card_number = ?"
	err := h.DB.QueryRow(query, cardNumber).Scan(&initialBalance)
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
	query := "SELECT phone FROM users WHERE phone = ?"
	err := h.DB.QueryRow(query, phone).Scan(&existingPhone)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		log.Printf("Phone number check error: %v", err)
		return false, err
	}
	return true, nil
}
