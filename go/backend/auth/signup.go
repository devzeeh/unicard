package auth

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

// Struct to handle error message
type ErrorMessage struct {
	Error string
}

// User struct to hold signup data

// View Handler (GET)
// This function checks the URL for errors (e.g., ?error=invalid)
// and displays the red text if needed.
func (h *Handler) SignupView(w http.ResponseWriter, r *http.Request) {
	h.Tpl.ExecuteTemplate(w, "signup.html", nil)
}

// HashPassword (Helper function, doesn't need Handler)
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// Signup handler
func (h *Handler) SignupHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Signup is running...")

	// parse the formdata
	r.ParseForm()

	// Get form values

	username := r.FormValue("username")
	cardNumber := r.FormValue("cardNumber")
	password := r.FormValue("password")

	// Check if fields is empty
	if username == "" || cardNumber == "" || password == "" {
		h.Tpl.ExecuteTemplate(w, "signup.html", ErrorMessage{Error: "Please check all fields."})
		return
	}

	// Check if username exist
	var existingUsername string // to hold existing username
	query := "SELECT Username FROM persons WHERE Username = ? "
	err := h.DB.QueryRow(query, username).Scan(&existingUsername)

	if err == sql.ErrNoRows {
		log.Printf("Username %s is available.", username)
	} else if err != nil {
		log.Printf("Username check error: %v", err)
		h.Tpl.ExecuteTemplate(w, "signup.html", ErrorMessage{Error: "System Error checking username."})
		return
	} else {
		log.Printf("Username %s already exists.", username)
		h.Tpl.ExecuteTemplate(w, "signup.html", ErrorMessage{Error: "Username already registered."})
		return
	}

	// Check cardNumber if exist
	var existingCardNumber string
	query = "SELECT CardNumber FROM cardDetails WHERE CardNumber = ?"
	err = h.DB.QueryRow(query, cardNumber).Scan(&existingCardNumber)
	if err == sql.ErrNoRows {
		log.Printf("CardNumber %s is available.", cardNumber)
	} else if err != nil {
		log.Printf("CardNumber check error: %v", err)
		h.Tpl.ExecuteTemplate(w, "signup.html", ErrorMessage{Error: "System Error checking card."})
		return
	} else {
		log.Printf("CardNumber %s already exists.", cardNumber)
		h.Tpl.ExecuteTemplate(w, "signup.html", ErrorMessage{Error: "Card Number already registered."})
		return
	}

	// Generate Hash for password
	hashedPassword, err := HashPassword(password)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		h.Tpl.ExecuteTemplate(w, "signup.html", "Error hashing password.")
		return
	}
	password = hashedPassword // Store the hashed password
	log.Printf("hass password is: %v", password)

	// Insert to DB
	query = "INSERT INTO persons (Username, Hash, CardNumber) values (?, ?, ?)"

	stmt, err := h.DB.Prepare(query)
	if err != nil {
		log.Printf("Prepare statement error: %v", err)
	}
	defer stmt.Close()

	// Execute
	_, err = stmt.Exec(username, password, cardNumber)
	if err != nil {
		log.Printf("Insert execution error: %v", err)
		return // you may want to handle this error properly
	}

	// Succesfully account creation
	log.Printf("account successfully created!")
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
