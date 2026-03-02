package authentication

import (
	"fmt"
	"net/http"
	structMessage "unicard-go/internal/pkg"

	"golang.org/x/crypto/bcrypt"
)

// View Handler (GET)
// This function checks the URL for errors (e.g., ?error=invalid)
// and displays the red text if needed.
func (h *Handler) LoginView(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Login view is running...")
	// Get the error code from the URL
	errCode := r.URL.Query().Get("user")

	var msg string

	// Determine the message based on the error code
	switch errCode {
	case "invalid":
		msg = "Wrong password"
	case "notfound":
		msg = "User not found"
	}

	// Render the template with the message
	h.Tpl.ExecuteTemplate(w, "login.html", structMessage.MessageData{Error: msg})
}

// Auth Handler (POST)
// Accepts login credentials: username, email, or full_name
func (h *Handler) LoginAuthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("loginauth running...")

	r.ParseForm()
	credential := r.PostFormValue("username") // This can be username, email, or full_name
	password := r.PostFormValue("password")

	var hash string
	// Query to check if credential matches username, email, or full_name
	stmt := "SELECT password_hash FROM users WHERE username = ? OR email = ? OR full_name = ?"

	err := h.DB.QueryRow(stmt, credential, credential, credential).Scan(&hash)

	// User not found
	if err != nil {
		fmt.Println("User not found or DB error:", err)
		// Redirect with ?user=notfound
		http.Redirect(w, r, "/login?user=notfound", http.StatusSeeOther)
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

	// Redirect with ?user=invalid
	http.Redirect(w, r, "/login?user=invalid", http.StatusSeeOther)
}
