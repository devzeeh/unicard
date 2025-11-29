package login

import (
    "database/sql"
    "fmt"
    "html/template"
    "net/http"

    "golang.org/x/crypto/bcrypt"
)

// Handler holds the dependencies (DB and Templates) for this package
type Handler struct {
    DB  *sql.DB
    Tpl *template.Template
}

// LoginAuthHandler processes the login form submission
func (h *Handler) LoginAuthHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Println("loginauth running...")

    // Ensure this is a POST request
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    // Get Form Values
    r.ParseForm()
    username := r.FormValue("username")
    password := r.FormValue("password")

    // Retrieve user hash from database
    var hash string
    // Assuming you have a table 'Persons' with columns 'username' and 'Hash'
    stmt := "SELECT Hash FROM Persons WHERE username = ?"

    err := h.DB.QueryRow(stmt, username).Scan(&hash)

    if err != nil {
        fmt.Println("User not found or DB error:", err)
        // Render template with error message
        h.Tpl.ExecuteTemplate(w, "login.html", "User not found")
        return
    }

    fmt.Println("Hash found, verifying...")

    // 4. Check password
    err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))

    if err == nil {
        // SUCCESS CASE
        fmt.Println("Login success")
        
        // We do NOT redirect here because we want to show the "Success" message.
        // If you redirect, the message "Login Success!" will be lost.
        h.Tpl.ExecuteTemplate(w, "login.html", "Login Success!")
        return
    }

    // FAILURE CASE
    fmt.Println("Password mismatch")
    h.Tpl.ExecuteTemplate(w, "login.html", "Wrong password")
}