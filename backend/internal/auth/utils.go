package authentication

import (
	"crypto/rand"
	"log"
	"math/big"
	"time"
)

// This function returns the current timestamp formatted as "YYYY-MM-DD HH:MM:SS"
// in the "Asia/Manila" timezone. If there's an error loading the timezone,
// it returns an empty string and the error.
func CurrentTimestamp() (string, error) {
	// Load Asia/Manila location
	loc, err := time.LoadLocation("Asia/Manila")
	if err != nil {
		log.Printf("CurrentTimestamp error loading timezone: %v", err)
		return "", err
	}
	time.Local = loc

	// Format the current time
	// Pro-Tip: Use .In(loc) instead of changing the global time.Local!
	// Also, use 15:04:05 to get 24-hour time (03:04:05 gives 12-hour time without AM/PM)
	return time.Now().In(loc).Format("2006-01-02 03:04:05"), nil
}

// It generates random numbers and checks the database for uniqueness.
// If a generated ID already exists, it retries until a unique one is found.
// Returns the unique user ID as int64 or an error if any occurs.
func (h *Handler) GenerateUserID() (int64, error) {
	// Generate random 12 digits number
	// Range: 100,000,000,000 to 999,999,999,999
	min := int64(100000000000)
	max := int64(999999999999)

	for {
		// Calculate the range size (max - min + 1)
		diff := new(big.Int).Sub(big.NewInt(max), big.NewInt(min))
		diff.Add(diff, big.NewInt(1))

		number, err := rand.Int(rand.Reader, diff)
		if err != nil {
			return 0, err
		}

		// Add min to get the final ID within the range
		userID := number.Int64() + min

		// Ask the Repository if it exists!
		exists, err := h.isUserIDExist(userID)
		if err != nil {
			return 0, err // Real DB error
		}
		if !exists {
			return userID, nil // Unique ID found!
		}

		log.Println("Collision detected! Retrying...")
	}
}
