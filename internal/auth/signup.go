package authentication

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"
	"unicard-go/internal/pkg/account"
)

// Struct to handle error message display in signup template
type ErrorMessage struct {
	Error   string
	Success string
}

// User struct to hold signup data
type User struct {
	UserID     string
	Username   string
	Fullname   string
	Email      string
	Phone      string
	CardNumber string
	Password   string
	CardID     string
	Usertype   string
	Balance    float64
	CreatedAt  string
}

// View Handler (GET)
// This function checks the URL for errors (e.g., ?error=invalid)
// and displays the red text if needed.
func (h *Handler) SignupView(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Signup view is running...")
	// Get the error code from the URL
	errCode := r.URL.Query().Get("error")

	var msg string
	// Determine the message based on the error code
	switch errCode {
	case "cardnotfound":
		msg = "Card number not found. Please check your card or contact support."
	case "userid":
		msg = "System error generating UserID. Please try again."
	case "cardnumber":
		msg = "System error generating CardNumber. Please try again."
	case "invalidcard":
		msg = "Invalid Card Number. Please check your card."
	case "contactnumber":
		msg = "Phone number already registered. Please use a different number."
	case "email":
		msg = "Email already registered. Please use a different email."
	case "emptyfields":
		msg = "Please fill in all fields."
	case "genusername":
		msg = "System error generating username. Please try again."
	case "cardstatuscheck":
		msg = "System error checking card status. Please try again."
	case "cardstatus":
		msg = "Card is invalid. Please contact support."
	case "passwordshort":
		msg = "Password must be at least 8 characters long."
	case "gencardid":
		msg = "System error generating CardID. Please try again."
	case "dbinsert":
		msg = "System error creating account. Please try again."
	case "parseform":
		msg = "Failed to parse form. Please try again."
	case "hashpassword":
		msg = "System error processing password. Please try again."
	case "cardupdate":
		msg = "System error activating card. Please contact support."
	}
	// Render the signup template with error message
	h.Tpl.ExecuteTemplate(w, "signup.html", ErrorMessage{Error: msg})
}

// This function processes the signup form submission.
// It validates the input, checks for existing usernames and card numbers,
// hashes the password, and inserts the new user into the database.
// On success, it redirects to the login page.
// On failure, it re-renders the signup page with an error message.
// POST /signupauth
func (h *Handler) SignupHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Signup is running...")

	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/signup?error=parseform", http.StatusSeeOther)
		return
	}

	// 1. Get and CLEAN the form values
	// strings.TrimSpace removes leading/trailing spaces.
	// Example: "  John  " becomes "John". "   " becomes "".
	firstName := strings.TrimSpace(r.PostFormValue("firstName"))
	lastName := strings.TrimSpace(r.PostFormValue("lastName"))
	cardNum := strings.TrimSpace(r.PostFormValue("cardNumber"))
	password := strings.TrimSpace(r.PostFormValue("password"))
	email := strings.TrimSpace(r.PostFormValue("email"))
	phone := strings.TrimSpace(r.PostFormValue("contactNumber"))

	// 2. Check if ANY required field is empty
	// We check individual variables before putting them in the struct
	if firstName == "" || lastName == "" || cardNum == "" || password == "" || email == "" || phone == "" {
		fmt.Println("Validation Failed: One or more fields are empty.")
		http.Redirect(w, r, "/signup?error=emptyfields", http.StatusSeeOther)
		return
	}

	// 3. Hydrate User Struct
	// Now we know the data is clean and present
	user := User{
		Usertype:   "Regular",
		Username:   firstName, // Using First Name as Username
		Fullname:   firstName + " " + lastName,
		CardNumber: cardNum,
		Password:   password,
		Email:      email,
		Phone:      phone,
	}

	// Check if username exists using helper function
	generatedUsername, err := h.GenerateUniqueUsername()
	if err != nil {
		fmt.Printf("Error generating username: %v\n", err)
		http.Redirect(w, r, "/signup?error=genusername", http.StatusSeeOther)
		return
	}
	user.Username = generatedUsername
	fmt.Printf("Assigned Username: %s\n", user.Username)

	// Check card number if exist
	// If exist push thru
	// Variable to hold the status from the DB
	var cardStatus string
	// Make sure your column name is correct (card_status vs status) based on your table definition
	query := "SELECT status FROM cards WHERE card_number = ?"
	err = h.DB.QueryRow(query, user.CardNumber).Scan(&cardStatus)

	// Card does not exist
	if err == sql.ErrNoRows {
		fmt.Printf("Validation Failed: Card Number %s does not exist in system.\n", user.CardNumber)
		http.Redirect(w, r, "/signup?error=cardnotfound", http.StatusSeeOther)
		return
	}
	// System Error
	if err != nil {
		fmt.Printf("Validation Error: System error checking card: %v\n", err)
		http.Redirect(w, r, "/signup?error=cardstatuscheck", http.StatusSeeOther)
		return
	}
	// Handle: Card Exists, but is NOT "Inactive" (e.g., Active, Blocked)
	// This blocks 'Active', 'Blocked', 'Lost', 'Expired', etc.
	if cardStatus != "Inactive" {
		fmt.Printf("Validation Failed: Card %s is currently '%s'.\n", user.CardNumber, cardStatus)
		http.Redirect(w, r, "/signup?error=cardstatus", http.StatusSeeOther)
		return
	}
	// Success: Proceed
	fmt.Printf("Card %s is Inactive (Valid). Proceeding...\n", user.CardNumber)

	// Password length check
	if len(user.Password) < 8 {
		fmt.Printf("Validation Failed: Password too short (%d chars).\n", len(user.Password))
		http.Redirect(w, r, "/signup?error=passwordshort", http.StatusSeeOther)
		return
	}

	// Generate Hash for password
	hashedPassword, err := account.HashPassword(user.Password)
	if err != nil {
		fmt.Printf("Error hashing password: %v", err)
		http.Redirect(w, r, "/signup?error=hashpassword", http.StatusSeeOther)
		return
	}
	fmt.Println("raw password is:", user.Password)
	user.Password = hashedPassword // Store the hashed password
	fmt.Printf("hashed password is: %v\n", user.Password)

	// Check if Email Exists (Using our helper method)
	if exists, _ := account.IsEmailExist(h.DB, user.Email); exists {
		fmt.Printf("Validation Failed: Email %s already exists.\n", user.Email)
		http.Redirect(w, r, "/signup?error=email", http.StatusSeeOther)
		//h.Tpl.ExecuteTemplate(w, "signup.html", ErrorMessage{Error: "Email already registered."})
		return
	}
	fmt.Printf("Email %s is available.\n", user.Email)

	// Check if Phone Exists (Using our helper method)
	if exists, _ := h.isPhoneExist(user.Phone); exists {
		fmt.Printf("Validation Failed: Phone %s already exists.\n", user.Phone)
		http.Redirect(w, r, "/signup?error=contactnumber", http.StatusSeeOther)
		return
	}
	fmt.Printf("Phone %s is available.\n", user.Phone)
	user.Phone = phone

	// Generate unique UserID (12 digits)
	generateUserId, err := h.GenerateUserID()
	if err != nil {
		fmt.Printf("Error generating UserID: %v", err)
		http.Redirect(w, r, "/signup?error=userid", http.StatusSeeOther)
		return
	}
	fmt.Printf("Generated UserID: %d\n", generateUserId)
	user.UserID = fmt.Sprintf("%d", generateUserId)

	// Generate unique CardID
	generateCardID, err := h.GenerateCardID()
	if err != nil {
		fmt.Printf("Error generating CardID: %v", err)
		http.Redirect(w, r, "/signup?error=gencardid", http.StatusSeeOther)
		return
	}
	user.CardID = generateCardID
	fmt.Printf("Generated CardID: %s\n", user.CardID)

	// Time of account creation
	user.CreatedAt, _ = CurrentTimestamp()

	// Check initial balance base on card
	// Using helper function from account package
	user.Balance, err = h.GetInitialBalance(user.CardNumber)
	if err != nil {
		fmt.Printf("Error: Failed to retrieve initial balance for card %s: %v\n", user.CardNumber, err)
		http.Redirect(w, r, "/signup?error=invalidcard", http.StatusSeeOther)
		return
	}
	fmt.Printf("Initial balance for card %s is %.2f\n", user.CardNumber, user.Balance)

	// Insert to DB
	query = "INSERT INTO users (user_id, username, full_name, email, phone, password_hash, card_id, card_number, user_type, balance, created_at) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"

	// Execute
	_, err = h.DB.Exec(
		query,
		user.UserID,
		user.Username,
		user.Fullname,
		user.Email,
		user.Phone,
		user.Password,
		user.CardID,
		user.CardNumber,
		user.Usertype,
		user.Balance,
		user.CreatedAt,
	)
	if err != nil {
		fmt.Printf("CRITICAL ERROR: Failed to insert user into database: %v\n", err)
		http.Redirect(w, r, "/signup?error=dbinsert", http.StatusSeeOther)
		return
	}

	// Update card status from "Inactive" to "Active"
	updateQuery := "UPDATE cards SET status = 'Active' WHERE card_number = ?"
	_, err = h.DB.Exec(updateQuery, user.CardNumber)
	if err != nil {
		fmt.Printf("ERROR: Failed to update card status to Active: %v\n", err)
		http.Redirect(w, r, "/signup?error=cardupdate", http.StatusSeeOther)
		return
	}
	fmt.Printf("Card %s status updated to Active\n", user.CardNumber)

	// Succesfully account creation
	fmt.Printf("account successfully created!")
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// ---Helper Functions---

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

		// 1. Generate the random suffix
		randomPart := ""
		for i := 0; i < length; i++ {
			num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
			if err != nil {
				return "", err
			}
			randomPart += string(charset[num.Int64()])
		}

		// 2. Combine prefix + date + time + random part
		username := fmt.Sprintf("%s%s%s%s", usernamePrefix, userDate, randomPart, timePart)

		// 3. Check DB for uniqueness
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
		fmt.Println("Username collision! Retrying...")
	}
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
		fmt.Println("Phone number check error:", err)
		return false, err
	}
	return true, nil
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

		// Check DB directly
		var tmpId int64
		query := "SELECT user_id FROM users WHERE user_id = ?"

		err = h.DB.QueryRow(query, userID).Scan(&tmpId)
		if err == sql.ErrNoRows {
			//	fmt.Println("Unique UserID generated:", userID)
			return userID, nil // Unique ID found
		} else if err != nil {
			return 0, err // Real DB error
		}
		// If it exists, loop runs again
		fmt.Println("Collision detected! Retrying...")
	}
}

// It generates the unique cardID for every card users
// Checks the database for uniqueness.
// Returns the unique card ID as string or an error if any occurs.
// Example format: CARD-XXXXXXXXXX
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
		fmt.Println("Collision detected! Retrying...", cardID)
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
		return 0, err
	}
	return initialBalance, nil
}

// This function returns the current timestamp formatted as "YYYY-MM-DD HH:MM:SS"
// in the "Asia/Manila" timezone. If there's an error loading the timezone,
// it returns an empty string and the error.
func CurrentTimestamp() (string, error) {
	// Load Asia/Manila location
	loc, err := time.LoadLocation("Asia/Manila")
	if err != nil {
		return "", err
	}
	time.Local = loc

	// Format the current time
	return time.Now().Format("2006-01-02 03:04:05"), nil
}
