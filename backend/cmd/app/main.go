package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"unicard-go/backend/internal/admin"
	authentication "unicard-go/backend/internal/auth"
	"unicard-go/backend/internal/user"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

var tpl *template.Template
var db *sql.DB

func main() {
	// Load .env file
	err := godotenv.Load("./.env")
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
	tpl, err = template.ParseGlob("./frontend/templates/*/*.html")
	if err != nil {
		log.Fatalf("Failed to load templates: %v. Check your folder path.", err)
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
	adminHanlder := admin.NewHandler(db, tpl)
	userHandler := user.NewHandler(db, tpl)

	// Setup Router
	mux := http.NewServeMux()

	// Serve static files (CSS, JS, images)
	fileServer := http.FileServer(http.Dir("./frontend/assets"))
	mux.Handle("/assets/", http.StripPrefix("/assets/", fileServer))

	// general endpoints
	mux.HandleFunc("GET /login", authHandler.LoginView)
	mux.HandleFunc("GET /signup", authHandler.SignupView)
	mux.HandleFunc("POST /v1/loginauth", authHandler.LoginAuthHandler) // Login authentication endpoint
	mux.HandleFunc("POST /v1/signupauth", authHandler.SignupHandler)
	mux.HandleFunc("POST /v1/signup/check-details", authHandler.CheckDetailsHandler)
	mux.HandleFunc("POST /v1/signup/check-card", authHandler.CheckCardHandler)
	mux.HandleFunc("GET /forgot-password", authHandler.ForgotPasswordView)
	mux.HandleFunc("POST /v1/forgot-password/send-otp", authHandler.ForgotPasswordSendOTP)
	mux.HandleFunc("POST /v1/forgot-password/verify-otp", authHandler.ForgotPasswordVerifyOTP)
	mux.HandleFunc("POST /v1/reset-password", authHandler.ResetPassword)
	mux.HandleFunc("GET /dashboard", userHandler.DashboardView)
	mux.HandleFunc("GET /v1/user/dashboard", userHandler.DashboardHandler)
	//mux.HandleFunc("GET /transaction", userHandler.TransactionView)
	//mux.HandleFunc("GET /topup", userHandler.TopupView)
	//mux.HandleFunc("GET /profile", userHandler.ProfileView)
	//mux.HandleFunc("GET /settings", userHandler.SettingsView)
	//mux.HandleFunc("GET /card", userHandler.CardView)
	//mux.HandleFunc("GET /v1/user/transactions", userHandler.TransactionsJSONHandler)
	//mux.HandleFunc("GET /logout",)

	// super admin endpoints
	mux.HandleFunc("GET /admin/platform-overview", adminHanlder.PlatformOverviewView)
	mux.HandleFunc("GET /admin/merchants", adminHanlder.MerchantManagementView)
	mux.HandleFunc("GET /admin/terminals", adminHanlder.TerminalRegistryView)
	mux.HandleFunc("GET /admin/settings", adminHanlder.SystemSettingsView)
	mux.HandleFunc("POST /v1/admin/merchants/add", adminHanlder.AddMerchantHandler)
	mux.HandleFunc("GET /admin/card-inventory", adminHanlder.CardInventoryView)
	mux.HandleFunc("GET /v1/admin/card-inventory-data", adminHanlder.CardInventoryDataHandler)
	mux.HandleFunc("GET /admin/addcard", adminHanlder.AddCardsView)
	mux.HandleFunc("GET /admin/deactivatecard", adminHanlder.DeactivateView)
	mux.HandleFunc("POST /v1/admin/addcardauth", adminHanlder.AddCardHandler)
	mux.HandleFunc("POST /v1/admin/deactivatecardauth", adminHanlder.DeactivateCardHanlder)
	mux.HandleFunc("POST /v1/admin/deletecardauth", adminHanlder.DeleteCardHandler)
	
	
	// Wrap mux with custom handler for root redirect
	customHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		mux.ServeHTTP(w, r)
	})

	// Start Server
	fmt.Println("Server started on: http://" + serverAddress + port)
	if err := http.ListenAndServe(serverAddress+port, customHandler); err != nil {
		log.Fatal(err)
	}
}
