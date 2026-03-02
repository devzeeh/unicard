package authentication

import (
	"database/sql"
	"fmt"
	"net/http"
	structMessage "unicard-go/internal/pkg"
	"unicard-go/internal/pkg/account"
)

// This function renders the forgot password HTML template.
// It is triggered when a user navigates to the forgot password page.
// The function uses the template engine to execute and display the "forgotPassword.html" template.
func (h *Handler) ForgotPasswordView(w http.ResponseWriter, r *http.Request) {
	h.Tpl.ExecuteTemplate(w, "forgotPassword.html", nil)
}

// This function handles the forgot password process.
// It retrieves the email and new password from the form submission.
// The function checks if the email exists in the database.
// If the email exists, it hashes the new password and updates it in the database.
// Finally, it provides feedback to the user about the success or failure of the operation.
func (h *Handler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Forgot Password is running...")

	r.ParseForm()
	email := r.FormValue("email")
	password := r.FormValue("password")
	//otp := r.FormValue("otp")
	fmt.Println("email:", email, "\n Password:", password)

	// Check if email exists
	exists, err := h.checkEmailExist(email)
	if err != nil {
		fmt.Println("Error checking email existence:", err)
		h.Tpl.ExecuteTemplate(w, "forgotPassword.html", structMessage.MessageData{Error: "System error. Please try again later."})
		return
	}
	if !exists {
		h.Tpl.ExecuteTemplate(w, "forgotPassword.html", structMessage.MessageData{Error: "Email not found."})
		return
	}

	// Hash the new password
	hashedPassword, err := account.HashPassword(password)
	if err != nil {
		fmt.Println("Error hashing password:", err)
		h.Tpl.ExecuteTemplate(w, "forgotPassword.html", structMessage.MessageData{Error: "System error. Please try again later."})
		return
	}

	// Update the password in the database
	err = h.updatePassword(email, hashedPassword)
	if err != nil {
		fmt.Println("Error updating password:", err)
		h.Tpl.ExecuteTemplate(w, "forgotPassword.html", structMessage.MessageData{Error: "System error. Please try again later."})
		return
	}
	h.Tpl.ExecuteTemplate(w, "forgotPassword.html", structMessage.MessageData{Success: "Password updated successfully."})
}

// ---Helper Function---

// This function checks if a given email already exists in the database.
// It executes a SQL query to search for the email in the users table.
// If the email is found, it returns true. If not found, it returns false.
// If an error occurs during the query, it returns the error.
func (h *Handler) checkEmailExist(email string) (bool, error) {
	// Hold the existing email
	var existingEmail string

	// Check query
	query := "SELECT email FROM users WHERE email = ?"
	err := h.DB.QueryRow(query, email).Scan(&existingEmail)
	if err == sql.ErrNoRows {
		fmt.Println("Email is available.")
		return false, nil
	}
	if err != nil {
		fmt.Println("Email check error:", err)
		return false, err
	}
	fmt.Println("Email already exists.")
	return true, nil
}

// This function updates the user's password in the database.
// It takes the user's email and the new hashed password as parameters.
// It executes an UPDATE SQL query to set the new password for the given email.
// If the update is successful, it returns nil. If an error occurs, it returns the error.
func (h *Handler) updatePassword(email, hashedPassword string) error {
	query := "UPDATE users SET password = ? WHERE email = ?"
	_, err := h.DB.Exec(query, hashedPassword, email)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}
	return nil
}
