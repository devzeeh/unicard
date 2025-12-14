package auth

import (
	"fmt"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

// We use this struct to pass data to the HTML
type LoginData struct {
	Error string
}

// View Handler (GET)
// This function checks the URL for errors (e.g., ?error=invalid)
// and displays the red text if needed.
func (h *Handler) LoginView(w http.ResponseWriter, r *http.Request) {
	// Get the error code from the URL
	errCode := r.URL.Query().Get("error")

	var msg string

	// Determine the message based on the error code
	switch errCode {
	case "invalid":
		msg = "Wrong password"
	case "notfound":
		msg = "User not found"
	}

	// Render the template with the message
	h.Tpl.ExecuteTemplate(w, "login.html", LoginData{Error: msg})
}

// Auth Handler (POST)
func (h *Handler) LoginAuthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("loginauth running...")

	r.ParseForm()
	username := r.FormValue("username")
	password := r.FormValue("password")

	var hash string
	stmt := "SELECT Hash FROM Persons WHERE username = ?"

	err := h.DB.QueryRow(stmt, username).Scan(&hash)

	// User not found
	if err != nil {
		fmt.Println("User not found or DB error:", err)
		// Redirect with ?error=notfound
		http.Redirect(w, r, "/login?error=notfound", http.StatusSeeOther)
		return
	}

	fmt.Println("Hash found, verifying...")
	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))

	// SUCCESS
	if err == nil {
		fmt.Println("Login success")
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	// Password mismatch
	fmt.Println("Password mismatch")

	// Redirect with ?error=invalid
	http.Redirect(w, r, "/login?error=invalid", http.StatusSeeOther)
}
