package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"unicard-go/go/backend/auth"

	_ "github.com/go-sql-driver/mysql"
)

var tpl *template.Template
var db *sql.DB

func main() {
	// Setup Templates
	var err error
	tpl, _ = template.ParseGlob("../templates/*.html")

	// Setup Database
	db, err = sql.Open("mysql", "root:devengr@tcp(localhost:3306)/testdb")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		panic("Database connection failed: " + err.Error())
	}

	// Initialize the Handler
	loginH := &auth.Handler{
		DB:  db,
		Tpl: tpl,
	}

	// Setup Router
	mux := http.NewServeMux()

	// --- ROUTES ---
	// GET Request: Shows the page AND handles the ?error=invalid logic
	// We use the function from the auth package now
	mux.HandleFunc("GET /login", loginH.LoginView)
	mux.HandleFunc("GET /signup", loginH.SignupHandler)
	// POST Request: Processes the form
	// This matches <form action="/loginauth"> in your HTML
	mux.HandleFunc("POST /loginauth", loginH.LoginAuthHandler)
	mux.HandleFunc("POST /signupauth", loginH.SignupHandler)

	// Dashboard
	mux.HandleFunc("/dashboard", dashboardHandler)

	// Start Server
	fmt.Println("Server started on: http://localhost:8001/login")
	if err := http.ListenAndServe(":8001", mux); err != nil {
		log.Fatal(err)
	}
}

// Dashboard handler
func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Dashboard is running")
	tpl.ExecuteTemplate(w, "dashboard.html", nil)
}
