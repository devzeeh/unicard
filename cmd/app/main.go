package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	adminauth "unicard-go/internal/admin"
	authentication "unicard-go/internal/auth"

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
	serverAddress := os.Getenv("SERVER_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUser, dbPass, dbHost, dbPort, dbName)

	// Setup Templates
	tpl, err = template.ParseGlob("./templates/*.html")
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
	authHandler := authentication.NewHandler(db, tpl)
	adminHanlder := adminauth.NewHandler(db, tpl)

	// Setup Router
	mux := http.NewServeMux()

	// Serve static files (CSS, JS, images)
	fileServer := http.FileServer(http.Dir("./assets"))
	mux.Handle("/assets/", http.StripPrefix("/assets/", fileServer))

	// --- ROUTES ---
	// POST Request: JSON API endpoints
	mux.HandleFunc("POST /api/v1/loginauth", authHandler.LoginAuthHandler) // Login authentication endpoint
	mux.HandleFunc("POST /api/v1/signupauth", authHandler.SignupHandler)

	// GET Request: SSR endpoints (for signup, admin)
	mux.HandleFunc("GET /login", authHandler.LoginView) // 
	mux.HandleFunc("GET /signup", authHandler.SignupView)
	mux.HandleFunc("GET /admin/addcard", adminHanlder.AddCardsView)
	mux.HandleFunc("GET /admin/deactivatecard", adminHanlder.DeactivateView)

	mux.HandleFunc("POST /api/v1/admin/addcardauth", adminHanlder.AddCardHandler)
	mux.HandleFunc("POST /api/v1/admin/deactivatecardauth", adminHanlder.DeactivateCardHanlder)

	// Dashboard
	mux.HandleFunc("/dashboard", dashboardHandler)

	// Start Server
	fmt.Println("Server started on: http://" + serverAddress + port)
	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatal(err)
	}
}

// Dashboard handler
func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Dashboard is running")
	tpl.ExecuteTemplate(w, "dashboard.html", nil)
}

// View handler - just serve the template
