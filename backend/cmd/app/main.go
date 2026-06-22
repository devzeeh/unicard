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
	"unicard-go/backend/internal/merchant"
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
	merchantHandler := merchant.NewHandler(db, tpl)

	// Setup Router
	mux := http.NewServeMux()

	// Serve static files (CSS, JS, images)
	fileServer := http.FileServer(http.Dir("./frontend/assets"))
	mux.Handle("/assets/", http.StripPrefix("/assets/", fileServer))

	// Serve storage directory for uploaded documents and images (locally stored)
	storageServer := http.FileServer(http.Dir("./storage"))
	mux.Handle("/storage/", http.StripPrefix("/storage/", storageServer))

	// general endpoints
	mux.HandleFunc("GET /login", authHandler.LoginView)
	mux.HandleFunc("POST /v1/loginauth", authHandler.LoginAuthHandler) // Login authentication endpoint
	mux.HandleFunc("GET /merchant-signup", authHandler.MerchantSignupView)
	mux.HandleFunc("POST /v1/merchant-signup", authHandler.MerchantSignupHandler)
	mux.HandleFunc("GET /admin-signup", authHandler.AdminSignupView)
	mux.HandleFunc("POST /v1/admin-signup", authHandler.AdminSignupHandler)
	// Customer Signup routes
	mux.HandleFunc("GET /signup", authHandler.SignupView)
	mux.HandleFunc("POST /v1/signup/send-otp", authHandler.SignupSendOTP)
	mux.HandleFunc("POST /v1/signup/verify-otp", authHandler.SignupVerifyOTP)
	mux.HandleFunc("POST /v1/signup/check-card", authHandler.CheckCardHandler)
	mux.HandleFunc("POST /v1/signupauth", authHandler.SignupHandler)
	mux.HandleFunc("GET /forgot-password", authHandler.ForgotPasswordView)
	mux.HandleFunc("POST /v1/forgot-password/send-otp", authHandler.ForgotPasswordSendOTP)
	mux.HandleFunc("POST /v1/forgot-password/verify-otp", authHandler.ForgotPasswordVerifyOTP)
	mux.HandleFunc("POST /v1/reset-password", authHandler.ResetPassword)
	mux.HandleFunc("GET /u/{username}", userHandler.ProfileView)
	mux.HandleFunc("PATCH /u/{username}/profile/edit", userHandler.ProfileEdit)
	mux.HandleFunc("GET /v1/verify-email", userHandler.VerifyEmail)
	mux.HandleFunc("POST /v1/user/{username}/profile/verify-password", userHandler.ProfileVerifyPassword)
	mux.HandleFunc("PUT /u/{username}/profile/password", userHandler.ProfileChangePassword)
	mux.HandleFunc("GET /u/{username}/dashboard", userHandler.DashboardView)
	mux.HandleFunc("GET /u/{username}/card", userHandler.CardView)
	mux.HandleFunc("POST /v1/user/{username}/card/status", userHandler.UpdateCardStatus)
	mux.HandleFunc("GET /u/{username}/settings", userHandler.SettingsView)
	mux.HandleFunc("GET /u/{username}/topup", userHandler.TopUpView)
	// Your frontend calls this to get the Xendit URL
	mux.HandleFunc("POST /api/topup/create-session/{username}", userHandler.CreateXenditInvoice)

	// Payment gateway endpoints
	// Xendit's servers call this behind the scenes when the payment is done
	mux.HandleFunc("POST /api/webhooks/xendit", userHandler.XenditWebhook)
	mux.HandleFunc("POST /v1/user/{username}/topup/checkout", userHandler.CreateXenditInvoice) //
	//mux.HandleFunc("GET /v1/user/{username}/topup/success", userHandler.TopUpSuccessHandler)
	mux.HandleFunc("GET /u/{username}/transaction", userHandler.TransactionView)
	mux.HandleFunc("GET /u/{username}/transactions", userHandler.TransactionView)

	mux.HandleFunc("GET /v1/user/{username}", userHandler.DashboardHandler)
	mux.HandleFunc("GET /v1/user/{username}/transactions", userHandler.TransactionsJSONHandler)
	//mux.HandleFunc("GET /logout",)

	// merchant endpoints
	mux.HandleFunc("GET /merchant/{username}/dashboard", merchantHandler.MerchantDashboardView)
	mux.HandleFunc("GET /v1/merchant/{username}/dashboard", merchantHandler.MerchantDashboardDataHandler)
	mux.HandleFunc("GET /merchant/{username}/transactions", merchantHandler.MerchantTransactionsView)
	mux.HandleFunc("GET /v1/merchant/{username}/transactions", merchantHandler.TransactionHandler)

	mux.HandleFunc("GET /v1/merchant/{username}/incomes", merchantHandler.IncomeHandler)
	mux.HandleFunc("GET /merchant/{username}/account", merchantHandler.MerchantAccountView)
	mux.HandleFunc("GET /v1/merchant/{username}/account", merchantHandler.MerchantAccountDataHandler)
	mux.HandleFunc("POST /v1/merchant/{username}/withdraw", merchantHandler.WithdrawHandler)

	// super admin endpoints
	mux.HandleFunc("GET /admin/{username}", adminHanlder.AdminDashboardView)
	mux.HandleFunc("GET /v1/admin/{username}/dashboard-data", adminHanlder.AdminDashboardDataHandler)
	mux.HandleFunc("GET /admin/{username}/merchants", adminHanlder.MerchantManagementView)
	mux.HandleFunc("GET /v1/admin/{username}/merchants-data", adminHanlder.MerchantManagementDataHandler)
	mux.HandleFunc("GET /admin/{username}/terminals", adminHanlder.TerminalRegistryView)
	mux.HandleFunc("GET /v1/admin/{username}/terminals-data", adminHanlder.TerminalRegistryDataHandler)
	mux.HandleFunc("GET /v1/admin/{username}/terminals/unassigned", adminHanlder.GetUnassignedTerminalsHandler)
	mux.HandleFunc("POST /v1/admin/{username}/terminals/add", adminHanlder.AddTerminalHandler)
	mux.HandleFunc("GET /admin/{username}/settings", adminHanlder.SystemSettingsView)
	mux.HandleFunc("GET /admin/{username}/transactions", adminHanlder.TransactionsView)
	mux.HandleFunc("GET /v1/admin/{username}/transactions", adminHanlder.AllTransactionsJSONHandler)
	mux.HandleFunc("POST /v1/admin/{username}/merchants/add", adminHanlder.AddMerchantHandler)
	mux.HandleFunc("GET /admin/{username}/merchants/{id}", adminHanlder.MerchantInfoView)
	mux.HandleFunc("GET /v1/admin/{username}/merchants/{id}/data", adminHanlder.MerchantInfoDataHandler)
	mux.HandleFunc("POST /v1/admin/{username}/merchants/{id}/approve", adminHanlder.ApproveMerchantHandler)
	mux.HandleFunc("POST /v1/admin/{username}/merchants/{id}/reject", adminHanlder.RejectMerchantHandler)
	mux.HandleFunc("POST /v1/admin/{username}/merchants/{id}/suspend", adminHanlder.SuspendMerchantHandler)
	mux.HandleFunc("DELETE /v1/admin/{username}/merchants/{id}/delete", adminHanlder.DeleteMerchantHandler)
	mux.HandleFunc("GET /admin/{username}/card-inventory", adminHanlder.CardInventoryView)
	mux.HandleFunc("GET /v1/admin/{username}/card-inventory-data", adminHanlder.CardInventoryDataHandler)
	mux.HandleFunc("POST /v1/admin/{username}/cards/{id}/block", adminHanlder.BlockCardHandler)
	mux.HandleFunc("GET /admin/{username}/addcard", adminHanlder.AddCardsView)
	mux.HandleFunc("GET /admin/{username}/deactivatecard", adminHanlder.DeactivateView)
	mux.HandleFunc("POST /v1/admin/{username}/addcardauth", adminHanlder.AddCardHandler)
	mux.HandleFunc("POST /v1/admin/{username}/deactivatecardauth", adminHanlder.DeactivateCardHanlder)
	mux.HandleFunc("POST /v1/admin/{username}/deletecardauth", adminHanlder.DeleteCardHandler)
	mux.HandleFunc("GET /admin/{username}/delete-cards", adminHanlder.DeleteCardView)

	// terminal endpoints for Fare and Retails.
	mux.HandleFunc("GET /terminal-sim", adminHanlder.TerminalSimView)
	mux.HandleFunc("POST /v1/terminal-sim/transact", adminHanlder.TerminalSimTransactionHandler)

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
