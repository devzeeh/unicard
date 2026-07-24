package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"unicard-go/backend/internal/admin"
	"unicard-go/backend/internal/auth"
	"unicard-go/backend/internal/merchant"
	"unicard-go/backend/internal/middleware"
	"unicard-go/backend/internal/mqtt"
	"unicard-go/backend/internal/pkg/database"
	"unicard-go/backend/internal/pkg/storage"
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

	// Initialize MQTT Service
	_, err = mqtt.NewMQTTService(store)
	if err != nil {
		log.Printf("Warning: Failed to initialize MQTT service: %v", err)
	}

	// Initialize R2 Storage
	r2Storage, err := storage.NewR2Storage()
	if err != nil {
		log.Printf("Warning: Failed to initialize R2 storage (uploads may fail): %v", err)
	}

	// Initialize the Handler from the auth package
	authRepo := auth.NewRepository(store)
	authSvc := auth.NewService(authRepo, r2Storage)

	// Initialize Handlers for other modules
	authHandler := auth.NewHandler(authSvc, tpl)
	adminRepo := admin.NewRepository(store)
	adminSvc := admin.NewService(adminRepo)
	adminHanlder := admin.NewHandler(adminSvc, tpl)
	userHandler := user.NewHandler(store, tpl)
	merchantHandler := merchant.NewHandler(store, tpl, r2Storage)

	// Setup Router
	mux := http.NewServeMux()

	// Serve static files (CSS, JS, images)
	fileServer := http.FileServer(http.Dir("./frontend/assets"))
	mux.Handle("/assets/", http.StripPrefix("/assets/", fileServer))

	// Serve storage directory for uploaded documents and images via proxy to R2
	if r2Storage != nil {
		mux.HandleFunc("/storage/", func(w http.ResponseWriter, r *http.Request) {
			key := strings.TrimPrefix(r.URL.Path, "/storage/")
			body, contentType, err := r2Storage.DownloadFile(r.Context(), key)
			if err != nil {
				log.Printf("Error downloading file from R2 (%s): %v", key, err)
				http.Error(w, "File not found", http.StatusNotFound)
				return
			}
			defer body.Close()
			w.Header().Set("Content-Type", contentType)

			// Optional: add cache headers to prevent re-fetching the image constantly
			w.Header().Set("Cache-Control", "public, max-age=86400")

			io.Copy(w, body)
		})
	} else {
		// Fallback to local storage if R2 fails
		storageServer := http.FileServer(http.Dir("./storage"))
		mux.Handle("/storage/", http.StripPrefix("/storage/", storageServer))
	}

	// Register auth routes endpoints
	// Holds: login, logout, admin-signup, merchant-signup, customer-signup, forgot-password
	auth.RegisterRoutes(mux, authHandler)
	
	// Middleware definitions
	requireCustomer := middleware.RequireAuth("customer")
	requireMerchant := middleware.RequireAuth("merchant_admin", "merchant_staff")
	requireAdmin := middleware.RequireAuth("super_admin")
	
	// Routes admin endpoints
	admin.RegisterRoutes(mux, adminHanlder, requireAdmin)
	
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

	

	// terminal endpoints for Fare and Retails.
	mux.HandleFunc("GET /terminal-sim", adminHanlder.TerminalSimView) // kept public since it's a sim
	mux.HandleFunc("POST /v1/terminal-sim/transact", adminHanlder.TerminalSimTransactionHandler)

	// Catch-all route for 404
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		if tpl != nil {
			err := tpl.ExecuteTemplate(w, "404.html", nil)
			if err != nil {
				http.Error(w, "404 page not found", http.StatusNotFound)
			}
		} else {
			http.Error(w, "404 page not found", http.StatusNotFound)
		}
	})

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
