package auth

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"log"
	"math/big"
	"net/http"
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
	Username        string
	Fullname        string
	UserID          string
	CardNumber      string
	Email           string
	Phone           string
	Password        string
	Usertype        string
	Status          string
	CreatedAt       string
	ConfirmPassword string
}

// View Handler (GET)
// This function checks the URL for errors (e.g., ?error=invalid)
// and displays the red text if needed.
func (h *Handler) SignupView(w http.ResponseWriter, r *http.Request) {
	h.Tpl.ExecuteTemplate(w, "signup.html", nil)
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
		// Handle the error, e.g., send a 400 Bad Request response.
		h.Tpl.ExecuteTemplate(w, "signup.html", ErrorMessage{Error: "Failed to parse form."})
		return
	}

	// Get form values using User struct
	user := User{
		Usertype:   "Regular",
		Username:   r.PostFormValue("first_name"),
		Fullname:   r.PostFormValue("first_name") + " " + r.PostFormValue("last_name"),
		CardNumber: r.PostFormValue("card_number"),
		Password:   r.PostFormValue("password"),
		Email:      r.PostFormValue("email"),
		Phone:      r.PostFormValue("contact_number"),
	}

	// Check if fields is empty
	if user.Username == "" || user.CardNumber == "" || user.Password == "" {
		h.Tpl.ExecuteTemplate(w, "signup.html", ErrorMessage{Error: "Please check all fields."})
		return
	}

	// Check if username exist
	var existingUsername string // to hold existing username
	query := "SELECT Username FROM persons WHERE username = ? "
	err := h.DB.QueryRow(query, user.Username).Scan(&existingUsername)

	if err == sql.ErrNoRows {
		log.Printf("Username %s is available.", user.Username)
	} else if err != nil {
		log.Printf("Username check error: %v", err)
		h.Tpl.ExecuteTemplate(w, "signup.html", ErrorMessage{Error: "System Error checking username."})
		return
	} else {
		log.Printf("Username %s already exists.", user.Username)
		h.Tpl.ExecuteTemplate(w, "signup.html", ErrorMessage{Error: "Username already registered."})
		return
	}

	// Check card number if exist
	var existingCardNumber string
	query = "SELECT CardNumber FROM cardDetails WHERE CardNumber = ?"
	err = h.DB.QueryRow(query, user.CardNumber).Scan(&existingCardNumber)
	if err == sql.ErrNoRows {
		log.Printf("CardNumber %s is available.", user.CardNumber)
	} else if err != nil {
		log.Printf("CardNumber check error: %v", err)
		h.Tpl.ExecuteTemplate(w, "signup.html", ErrorMessage{Error: "System Error checking card."})
		return
	} else {
		log.Printf("CardNumber %s already exists.", user.CardNumber)
		h.Tpl.ExecuteTemplate(w, "signup.html", ErrorMessage{Error: "Card Number already registered."})
		return
	}

	// Password length check
	if len(user.Password) < 8 {
		log.Printf("Password too short `%d`", len(user.Password))
		h.Tpl.ExecuteTemplate(w, "signup.html", ErrorMessage{Error: "Password must be at least 8 characters long."})
		return
	}

	// Generate Hash for password
	hashedPassword, err := account.HashPassword(user.Password)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		//h.Tpl.ExecuteTemplate(w, "signup.html", "Error hashing password.")
		return
	}
	user.Password = hashedPassword // Store the hashed password
	log.Printf("hashed password is: %v", user.Password)

	// Check if Email Exists (Using our helper method)
	hasEmail, err := account.IsEmailExist(h.DB, user.Email)
	if err != nil {
		log.Printf("Error checking email existence: %v", err)
		h.Tpl.ExecuteTemplate(w, "signup.html", "System Error checking email.")
		return
	}
	if hasEmail {
		log.Printf("Email %s already registered.", user.Email)
		h.Tpl.ExecuteTemplate(w, "signup.html", "Error: Email already registered.")
		return
	}

	// Check if Phone Exists (Using our helper method)
	hasPhone, err := h.isPhoneExist(user.Phone)
	if err != nil {
		log.Printf("Error checking phone existence: %v", err)
		h.Tpl.ExecuteTemplate(w, "signup.html", "System Error checking phone.")
		return
	}
	if hasPhone {
		log.Printf("Phone %s already registered.", user.Phone)
		h.Tpl.ExecuteTemplate(w, "signup.html", "Error: Phone number already registered.")
		return
	}

	// Generate unique UserID (12 digits)
	generateUserId, err := h.GenerateUserID()
	if err != nil {
		log.Printf("Error generating UserID: %v", err)
		h.Tpl.ExecuteTemplate(w, "signup.html", ErrorMessage{Error: "Error Generating UserID."})
	}
	user.UserID = fmt.Sprintf("%d", generateUserId)

	// Time of account creation
	user.CreatedAt, _ = CurrentTimestamp()

	// Insert to DB
	query = "INSERT INTO persons (role, username, fullname, hash, card_number, email, phone, user_id, created_at) values (?, ?, ?, ?, ?, ?, ?, ?, ?)"

	// Execute
	_, err = h.DB.Exec(
		query,
		user.Usertype,
		user.Username,
		user.Fullname,
		user.Password,
		user.CardNumber,
		user.Email,
		user.Phone,
		user.UserID,
		user.CreatedAt,
	)
	if err != nil {
		log.Printf("Insert execution error: %v", err)
		return // you may want to handle this error properly
	}

	// Succesfully account creation
	log.Printf("account successfully created!")
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// ---Helper Functions---

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
		fmt.Println("Phone number is available.")
		return false, nil
	}
	if err != nil {
		fmt.Println("Phone number check error:", err)
		return false, err
	}
	fmt.Println("Phone number already exists.")
	return true, nil
}

// It generates random numbers and checks the database for uniqueness.
// If a generated ID already exists, it retries until a unique one is found.
// Returns the unique user ID as int64 or an error if any occurs.
func (h *Handler) GenerateUserID() (int64, error) {
	// Generate random 12 digits number
	min := int64(10000000000)
	max := int64(99999999999)

	for {
		number, err := rand.Int(rand.Reader, big.NewInt(max-min+1))
		if err != nil {
			return 0, err
		}
		userID := number.Int64() + min

		// Check DB directly (Faster: 1 Round Trip)
		// to hold queried ID
		var tmpId int64
		err = h.DB.QueryRow("SELECT user_id FROM  users WHERE user_id=?", userID).Scan(&tmpId)
		if err == sql.ErrNoRows {
			return userID, nil // Unique ID found
		} else if err != nil {
			return 0, err
		}
		// If it exists, loop runs again
		fmt.Println("Collision detected! Retrying...")
	}
}

// This function returns the current timestamp formatted as "YYYY-MM-DD HH:MM:SS"
// in the "Asia/Manila" timezone. If there's an error loading the timezone,
// it returns an empty string and the error.
func CurrentTimestamp() (string, error) {
	loc, err := time.LoadLocation("Asia/Manila")
	if err != nil {
		return "", err
	}
	time.Local = loc
	return time.Now().Format("2006-01-02 03:04:05"), nil
}
