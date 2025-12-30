package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"unicard-go/backend/auth"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

var tpl *template.Template
var db *sql.DB

func main() {

	// Load .env file
	err := godotenv.Load("../.env")
	if err != nil {
		// Fallback: try loading from current directory
		if err := godotenv.Load(); err != nil {
			log.Fatalf("Error loading .env file: %v", err)
		}
	}

	// read .env VALUES
	port := os.Getenv("PORT")
	serverAddress := os.Getenv("SERVER_ADDR")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUser, dbPass, dbHost, dbPort, dbName)

	// Setup Templates
	tpl, err = template.ParseGlob("../templates/*.html")
	if err != nil {
		log.Fatal("Templates loaded but variable is nil. Check your folder path.")
	}
	
	// Setup Database
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	// Always verify connection
	if err := db.Ping(); err != nil {
		panic("Database connection failed: " + err.Error())
	}

	// Initialize the Handler from the auth package
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
	fmt.Println("Server started on: http://", serverAddress, port)
	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatal(err)
	}
}

// Dashboard handler
func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Dashboard is running")
	tpl.ExecuteTemplate(w, "dashboard.html", nil)
}
