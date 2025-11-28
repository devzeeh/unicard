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

// LoginAuthHandler is now a method of *Handler. 
// It can access h.DB and h.Tpl
func (h *Handler) LoginAuthHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Println("loginauth running")

    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    r.ParseForm()
    username := r.FormValue("username")
    password := r.FormValue("password")
    
    // Retrieve user from database using h.DB
    var hash string
    stmt := "SELECT Hash FROM Persons WHERE username = ?"

    // Use h.DB here (which is passed from main)
    err := h.DB.QueryRow(stmt, username).Scan(&hash)
	fmt.Println("Hashed password:", hash)

    if err != nil {
        fmt.Println("db error:", err)
        // Use h.Tpl here
        h.Tpl.ExecuteTemplate(w, "login.html", "User not found")
        return
    }

    // Check password
    err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))

    if err == nil {
        fmt.Println("login success")
		http.Redirect(w,r, "/login", http.StatusMovedPermanently)
        h.Tpl.ExecuteTemplate(w, "login.html", "Login Success!") 
        return 
    }

    fmt.Println("Password mismatch error:", err)
    h.Tpl.ExecuteTemplate(w, "login.html", "Wrong password")
}