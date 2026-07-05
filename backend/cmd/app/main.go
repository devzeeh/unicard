package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"unicard-go/backend/internal/admin"
	authentication "unicard-go/backend/internal/auth"
	"unicard-go/backend/internal/merchant"
	"unicard-go/backend/internal/middleware"
	"unicard-go/backend/internal/pkg/database"
	"unicard-go/backend/internal/user"

	"github.com/joho/godotenv"
)

var (
	tpl *template.Template
)

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

	// Setup Templates
	tpl, err = template.ParseGlob("./frontend/templates/*/*.html")
	if err != nil {
		log.Fatalf("Failed to load templates: %v. Check your folder path.", err)
	}

	// Setup Database using the new database package
	db, err := database.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	store := database.NewStore(db)

	// Initialize the Handler from the auth package
	authHandler := authentication.NewHandler(store, tpl)
	adminHanlder := admin.NewHandler(store, tpl)
	userHandler := user.NewHandler(store, tpl)
	merchantHandler := merchant.NewHandler(store, tpl)

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
	mux.HandleFunc("POST /v1/loginauth", authHandler.LoginAuthHandler)  // Login authentication endpoint
	mux.HandleFunc("POST /v1/refresh", authHandler.RefreshTokenHandler) // Refresh token endpoint
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
	// Middleware definitions
	requireCustomer := middleware.RequireAuth("customer")
	requireMerchant := middleware.RequireAuth("merchant_admin", "merchant_staff")
	requireAdmin := middleware.RequireAuth("super_admin")

	// Customer Routes
	mux.Handle("GET /u/{username}", requireCustomer(http.HandlerFunc(userHandler.ProfileView)))
	mux.Handle("PATCH /u/{username}/profile/edit", requireCustomer(http.HandlerFunc(userHandler.ProfileEdit)))
	mux.Handle("POST /v1/user/{username}/profile/verify-password", requireCustomer(http.HandlerFunc(userHandler.ProfileVerifyPassword)))
	mux.Handle("PUT /u/{username}/profile/password", requireCustomer(http.HandlerFunc(userHandler.ProfileChangePassword)))
	mux.Handle("GET /u/{username}/dashboard", requireCustomer(http.HandlerFunc(userHandler.DashboardView)))
	mux.Handle("GET /u/{username}/card", requireCustomer(http.HandlerFunc(userHandler.CardView)))
	mux.Handle("POST /v1/user/{username}/card/status", requireCustomer(http.HandlerFunc(userHandler.UpdateCardStatus)))
	mux.Handle("POST /v1/user/{username}/card/replace", requireCustomer(http.HandlerFunc(userHandler.RequestReplacement)))
	mux.Handle("GET /u/{username}/settings", requireCustomer(http.HandlerFunc(userHandler.SettingsView)))
	mux.Handle("GET /u/{username}/topup", requireCustomer(http.HandlerFunc(userHandler.TopUpView)))
	// Your frontend calls this to get the Xendit URL
	mux.Handle("POST /api/topup/create-session/{username}", requireCustomer(http.HandlerFunc(userHandler.CreateXenditInvoice)))

	// Payment gateway endpoints
	// Xendit's servers call this behind the scenes when the payment is done
	mux.HandleFunc("POST /api/webhooks/xendit/invoice", userHandler.XenditWebhook)
	mux.HandleFunc("POST /api/webhooks/xendit/disbursement", merchantHandler.XenditDisbursementWebhook)
	mux.Handle("POST /v1/user/{username}/topup/checkout", requireCustomer(http.HandlerFunc(userHandler.CreateXenditInvoice)))
	mux.Handle("GET /u/{username}/transaction", requireCustomer(http.HandlerFunc(userHandler.TransactionView)))
	mux.Handle("GET /u/{username}/transactions", requireCustomer(http.HandlerFunc(userHandler.TransactionView)))

	mux.Handle("GET /v1/user/{username}", requireCustomer(http.HandlerFunc(userHandler.DashboardHandler)))
	mux.Handle("GET /v1/user/{username}/transactions", requireCustomer(http.HandlerFunc(userHandler.TransactionsJSONHandler)))
	mux.HandleFunc("GET /logout", authHandler.LogoutHandler)
	mux.HandleFunc("POST /logout", authHandler.LogoutHandler) // Allow POST as well just in cases

	// merchant endpoints
	mux.Handle("GET /merchant/{username}/dashboard", requireMerchant(http.HandlerFunc(merchantHandler.MerchantDashboardView)))
	mux.Handle("GET /v1/merchant/{username}/dashboard", requireMerchant(http.HandlerFunc(merchantHandler.MerchantDashboardDataHandler)))
	mux.Handle("GET /merchant/{username}/transactions", requireMerchant(http.HandlerFunc(merchantHandler.MerchantTransactionsView)))
	mux.Handle("GET /v1/merchant/{username}/transactions", requireMerchant(http.HandlerFunc(merchantHandler.TransactionHandler)))

	mux.Handle("GET /v1/merchant/{username}/incomes", requireMerchant(http.HandlerFunc(merchantHandler.IncomeHandler)))
	mux.Handle("GET /merchant/{username}/account", requireMerchant(http.HandlerFunc(merchantHandler.MerchantAccountView)))
	mux.Handle("GET /v1/merchant/{username}/account", requireMerchant(http.HandlerFunc(merchantHandler.MerchantAccountDataHandler)))
	mux.Handle("POST /v1/merchant/{username}/update-bank", requireMerchant(http.HandlerFunc(merchantHandler.UpdateBankDetails)))
	mux.Handle("POST /v1/merchant/{username}/upload-document", requireMerchant(http.HandlerFunc(merchantHandler.UploadDocument)))
	mux.Handle("POST /v1/merchant/{username}/withdraw", requireMerchant(http.HandlerFunc(merchantHandler.WithdrawHandler)))
	mux.Handle("POST /v1/merchant/{username}/terminals/request", requireMerchant(http.HandlerFunc(merchantHandler.RequestTerminalHandler)))

	// super admin endpoints
	mux.Handle("GET /admin/{username}", requireAdmin(http.HandlerFunc(adminHanlder.AdminDashboardView)))
	mux.Handle("GET /v1/admin/{username}/dashboard-data", requireAdmin(http.HandlerFunc(adminHanlder.AdminDashboardDataHandler)))
	mux.Handle("GET /admin/{username}/merchants", requireAdmin(http.HandlerFunc(adminHanlder.MerchantManagementView)))
	mux.Handle("GET /v1/admin/{username}/merchants-data", requireAdmin(http.HandlerFunc(adminHanlder.MerchantManagementDataHandler)))
	mux.Handle("GET /admin/{username}/terminals", requireAdmin(http.HandlerFunc(adminHanlder.TerminalRegistryView)))
	mux.Handle("GET /v1/admin/{username}/terminals-data", requireAdmin(http.HandlerFunc(adminHanlder.TerminalRegistryDataHandler)))
	mux.Handle("GET /v1/admin/{username}/terminals/unassigned", requireAdmin(http.HandlerFunc(adminHanlder.GetUnassignedTerminalsHandler)))
	mux.Handle("POST /v1/admin/{username}/terminals/add", requireAdmin(http.HandlerFunc(adminHanlder.AddTerminalHandler)))
	mux.Handle("GET /admin/{username}/terminal-requests", requireAdmin(http.HandlerFunc(adminHanlder.TerminalRequestsView)))
	mux.Handle("GET /v1/admin/{username}/terminal-requests-data", requireAdmin(http.HandlerFunc(adminHanlder.TerminalRequestsDataHandler)))
	mux.Handle("POST /v1/admin/{username}/terminal-requests/{id}/approve", requireAdmin(http.HandlerFunc(adminHanlder.ApproveTerminalRequestHandler)))
	mux.Handle("POST /v1/admin/{username}/terminal-requests/{id}/reject", requireAdmin(http.HandlerFunc(adminHanlder.RejectTerminalRequestHandler)))
	mux.Handle("GET /admin/{username}/settings", requireAdmin(http.HandlerFunc(adminHanlder.SystemSettingsView)))
	mux.Handle("GET /admin/{username}/transactions", requireAdmin(http.HandlerFunc(adminHanlder.TransactionsView)))
	mux.Handle("GET /v1/admin/{username}/transactions", requireAdmin(http.HandlerFunc(adminHanlder.AllTransactionsJSONHandler)))
	mux.Handle("POST /v1/admin/{username}/merchants/add", requireAdmin(http.HandlerFunc(adminHanlder.AddMerchantHandler)))
	mux.Handle("GET /admin/{username}/merchants/{id}", requireAdmin(http.HandlerFunc(adminHanlder.MerchantInfoView)))
	mux.Handle("GET /v1/admin/{username}/merchants/{id}/data", requireAdmin(http.HandlerFunc(adminHanlder.MerchantInfoDataHandler)))
	mux.Handle("POST /v1/admin/{username}/merchants/{id}/approve", requireAdmin(http.HandlerFunc(adminHanlder.ApproveMerchantHandler)))
	mux.Handle("POST /v1/admin/{username}/merchants/{id}/reject", requireAdmin(http.HandlerFunc(adminHanlder.RejectMerchantHandler)))
	mux.Handle("POST /v1/admin/{username}/merchants/{id}/suspend", requireAdmin(http.HandlerFunc(adminHanlder.SuspendMerchantHandler)))
	mux.Handle("DELETE /v1/admin/{username}/merchants/{id}/delete", requireAdmin(http.HandlerFunc(adminHanlder.DeleteMerchantHandler)))
	mux.Handle("GET /admin/{username}/card-inventory", requireAdmin(http.HandlerFunc(adminHanlder.CardInventoryView)))
	mux.Handle("GET /v1/admin/{username}/card-inventory-data", requireAdmin(http.HandlerFunc(adminHanlder.CardInventoryDataHandler)))
	mux.Handle("POST /v1/admin/{username}/cards/{id}/block", requireAdmin(http.HandlerFunc(adminHanlder.BlockCardHandler)))
	mux.Handle("GET /admin/{username}/addcard", requireAdmin(http.HandlerFunc(adminHanlder.AddCardsView)))
	mux.Handle("GET /admin/{username}/deactivatecard", requireAdmin(http.HandlerFunc(adminHanlder.DeactivateView)))
	mux.Handle("POST /v1/admin/{username}/addcardauth", requireAdmin(http.HandlerFunc(adminHanlder.AddCardHandler)))
	mux.Handle("POST /v1/admin/{username}/deactivatecardauth", requireAdmin(http.HandlerFunc(adminHanlder.DeactivateCardHanlder)))
	mux.Handle("POST /v1/admin/{username}/deletecardauth", requireAdmin(http.HandlerFunc(adminHanlder.DeleteCardHandler)))
	mux.Handle("GET /admin/{username}/delete-cards", requireAdmin(http.HandlerFunc(adminHanlder.DeleteCardView)))

	// terminal endpoints for Fare and Retails.
	mux.HandleFunc("GET /terminal-sim", adminHanlder.TerminalSimView) // kept public since it's a sim
	mux.HandleFunc("POST /v1/terminal-sim/transact", adminHanlder.TerminalSimTransactionHandler)

	// Wrap mux with custom handler for root redirect
	customHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		mux.ServeHTTP(w, r)
	})

	// Wrap with CORS middleware
	handler := corsMiddleware(customHandler)

	// Start Server
	fmt.Println("Server started on: http://" + serverAddress + ":" + port)
	if err := http.ListenAndServe(serverAddress+":"+port, handler); err != nil {
		log.Fatal(err)
	}
}
func corsMiddleware(next http.Handler) http.Handler {
	allowedOrigins := map[string]bool{
		os.Getenv("CORS_ALLOWED_ORIGINS"): true, // Load from .env
		//"http://localhost:5173":           true, // Vue dev
		//"http://localhost:3001":           true, // Go dev
		//"https://unicard.app":   		   true, // production
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if allowedOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
