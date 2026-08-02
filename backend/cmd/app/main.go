package main

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

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

func main() {
	if err := loadEnv(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	tpl, err := loadTemplates("./frontend/templates/*/*.html")
	if err != nil {
		log.Fatalf("Failed to load templates: %v", err)
	}

	db, err := database.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	store := database.NewStore(db)

	// NOTE: if mqtt.Service exposes a cleanup method (Close/Stop/Disconnect),
	// add it back here once confirmed. Left out for now since the method
	// name wasn't verified against your actual mqtt package.
	_, err = mqtt.NewMQTTService(store)
	if err != nil {
		log.Printf("Warning: Failed to initialize MQTT service: %v", err)
	}

	r2Storage, err := storage.NewR2Storage()
	if err != nil {
		log.Printf("Warning: Failed to initialize R2 storage (uploads may fail): %v", err)
	}

	deps := buildHandlers(store, r2Storage, tpl)

	mux := buildRouter(deps, r2Storage, tpl)
	handler := corsMiddleware(mux)

	runServer(handler)
}

// loadEnv loads .env from the working directory, falling back to the
// current directory if the primary path isn't found.
func loadEnv() error {
	if err := godotenv.Load("./.env"); err != nil {
		return godotenv.Load()
	}
	return nil
}

func loadTemplates(pattern string) (*template.Template, error) {
	return template.ParseGlob(pattern)
}

// handlers groups every package's wired-up Handler so they can be passed
// around and registered as a unit.
type handlers struct {
	auth     *auth.Handler
	admin    *admin.Handler
	merchant *merchant.Handler
	user     *user.Handler
}

func buildHandlers(store database.Store, r2Storage storage.Service, tpl *template.Template) handlers {
	authRepo := auth.NewRepository(store)
	authSvc := auth.NewService(authRepo, r2Storage)
	authHandler := auth.NewHandler(authSvc, tpl)

	adminRepo := admin.NewRepository(store)
	adminSvc := admin.NewService(adminRepo)
	adminHandler := admin.NewHandler(adminSvc, tpl)

	merchantRepo := merchant.NewRepository(store)
	payoutGateway := merchant.NewXenditPayoutGateway(os.Getenv("XENDIT_SECRET_KEY"))
	merchantSvc := merchant.NewService(merchantRepo, r2Storage, payoutGateway)
	merchantHandler := merchant.NewHandler(merchantSvc, tpl)

	userRepo := user.NewRepository(store)
	invoiceGateway := user.NewXenditInvoiceGateway(os.Getenv("XENDIT_SECRET_KEY"))
	mailer := user.NewSMTPMailerFromEnv(
		os.Getenv("SMTP_HOST"), os.Getenv("SMTP_SENDER"), os.Getenv("SMTP_EMAIL"), os.Getenv("SMTP_PASSWORD"), 587,
	)
	userSvc := user.NewService(userRepo, mailer, invoiceGateway)
	userHandler := user.NewHandler(userSvc, tpl)

	return handlers{
		auth:     authHandler,
		admin:    adminHandler,
		merchant: merchantHandler,
		user:     userHandler,
	}
}

func buildRouter(h handlers, r2Storage storage.Service, tpl *template.Template) http.Handler {
	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("./frontend/assets"))
	mux.Handle("/assets/", http.StripPrefix("/assets/", fileServer))

	registerStorageRoute(mux, r2Storage)

	requireCustomer := middleware.RequireAuth("customer")
	requireMerchant := middleware.RequireAuth("merchant_admin", "merchant_staff")
	requireAdmin := middleware.RequireAuth("super_admin")

	auth.RegisterRoutes(mux, h.auth)
	admin.RegisterRoutes(mux, h.admin, requireAdmin)
	merchant.RegisterRoutes(mux, h.merchant, requireMerchant)
	user.RegisterRoutes(mux, h.user, requireCustomer)

	// Kept public since it's a simulator, not a real terminal.
	mux.HandleFunc("GET /terminal-sim", h.admin.TerminalSimView)
	mux.HandleFunc("POST /v1/terminal-sim/transact", h.admin.TerminalSimTransactionHandler)

	mux.HandleFunc("/", notFoundHandler(tpl))

	// Root redirect wrapper
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		mux.ServeHTTP(w, r)
	})
}

func registerStorageRoute(mux *http.ServeMux, r2Storage storage.Service) {
	if r2Storage == nil {
		// Fallback to local storage if R2 failed to initialize.
		storageServer := http.FileServer(http.Dir("./storage"))
		mux.Handle("/storage/", http.StripPrefix("/storage/", storageServer))
		return
	}

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
		w.Header().Set("Cache-Control", "public, max-age=86400")

		if _, err := io.Copy(w, body); err != nil {
			log.Printf("Error streaming file from R2 (%s): %v", key, err)
		}
	})
}

func notFoundHandler(tpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		if tpl == nil {
			http.Error(w, "404 page not found", http.StatusNotFound)
			return
		}
		if err := tpl.ExecuteTemplate(w, "404.html", nil); err != nil {
			http.Error(w, "404 page not found", http.StatusNotFound)
		}
	}
}

// corsAllowedOrigins parses a comma-separated list of allowed origins from
// the environment. Returns an empty set (nothing allowed) if unset, rather
// than accidentally matching an empty Origin header.
func corsAllowedOrigins() map[string]bool {
	raw := strings.TrimSpace(os.Getenv("CORS_ALLOWED_ORIGINS"))
	allowed := map[string]bool{}
	if raw == "" {
		return allowed
	}
	for _, origin := range strings.Split(raw, ",") {
		origin = strings.TrimSpace(origin)
		if origin != "" {
			allowed[origin] = true
		}
	}
	return allowed
}

func corsMiddleware(next http.Handler) http.Handler {
	allowedOrigins := corsAllowedOrigins()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && allowedOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// runServer starts the HTTP server with sane timeouts and shuts it down
// gracefully on SIGINT/SIGTERM, giving in-flight requests time to finish.
func runServer(handler http.Handler) {
	addr := os.Getenv("SERVER_PORT") + ":" + os.Getenv("PORT")

	srv := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		fmt.Println("Server started on: http://" + addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Graceful shutdown failed: %v", err)
	} else {
		log.Println("Server shut down cleanly")
	}
}