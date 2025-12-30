package test

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"regexp"
	"strings"
	"testing"
	"unicard-go/backend/auth"

	"github.com/DATA-DOG/go-sqlmock"
	_ "github.com/go-sql-driver/mysql" // Import MySQL driver for real DB tests
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

// TestLoginView tests the GET handler (rendering HTML)
func TestLoginView(t *testing.T) {
	// 1. SETUP
	mockTpl, err := template.New("login.html").Parse("{{.Error}}")
	assert.NoError(t, err)

	h := &auth.Handler{Tpl: mockTpl}

	tests := []struct {
		name           string
		queryParam     string
		expectedOutput string
	}{
		{"No Error", "", ""},
		{"Invalid Password", "?error=invalid", "Wrong password"},
		{"User Not Found", "?error=notfound", "User not found"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/login"+tt.queryParam, nil)
			rr := httptest.NewRecorder()

			h.LoginView(rr, req)

			assert.Equal(t, tt.expectedOutput, strings.TrimSpace(rr.Body.String()))
		})
	}
}

// TestLoginAuthHandler tests the POST handler (Database Logic)
func TestLoginAuthHandler(t *testing.T) {
	// SETUP MOCK DB
	// sqlmock.New() returns a *sql.DB and a mock controller
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	// Initialize handler with the Mock DB
	h := &auth.Handler{DB: db}

	// Helper to create a proper regex for the query
	// We use QuoteMeta to safely escape characters like '?'
	queryRegex := regexp.QuoteMeta("SELECT Hash FROM persons WHERE username = ?")

	// SCENARIO 1: SUCCESS (Valid User & Password)
	t.Run("Login Success", func(t *testing.T) {
		password := "secret123"
		hashedPwd, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

		// MOCK EXPECTATION:
		rows := sqlmock.NewRows([]string{"Hash"}).AddRow(string(hashedPwd))
		mock.ExpectQuery(queryRegex).WithArgs("validUser").WillReturnRows(rows)

		// REQUEST
		form := url.Values{}
		form.Add("username", "validUser")
		form.Add("password", password)
		req := httptest.NewRequest("POST", "/login-auth", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		// EXECUTE
		h.LoginAuthHandler(rr, req)

		// ASSERT
		assert.Equal(t, http.StatusSeeOther, rr.Code)
		assert.Equal(t, "/dashboard", rr.Header().Get("Location"))
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	// SCENARIO 2: USER NOT FOUND
	t.Run("User Not Found", func(t *testing.T) {
		mock.ExpectQuery(queryRegex).WithArgs("ghostUser").WillReturnError(sqlmock.ErrCancelled)

		form := url.Values{}
		form.Add("username", "ghostUser")
		form.Add("password", "any")
		req := httptest.NewRequest("POST", "/login-auth", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		h.LoginAuthHandler(rr, req)

		assert.Equal(t, http.StatusSeeOther, rr.Code)
		assert.Equal(t, "/login?error=notfound", rr.Header().Get("Location"))
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	// SCENARIO 3: WRONG PASSWORD
	t.Run("Wrong Password", func(t *testing.T) {
		realPassword := "secret123"
		hashedPwd, _ := bcrypt.GenerateFromPassword([]byte(realPassword), bcrypt.DefaultCost)

		rows := sqlmock.NewRows([]string{"Hash"}).AddRow(string(hashedPwd))
		mock.ExpectQuery(queryRegex).WithArgs("validUser").WillReturnRows(rows)

		form := url.Values{}
		form.Add("username", "validUser")
		form.Add("password", "WRONG_PASSWORD")
		req := httptest.NewRequest("POST", "/login-auth", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		h.LoginAuthHandler(rr, req)

		assert.Equal(t, http.StatusSeeOther, rr.Code)
		assert.Equal(t, "/login?error=invalid", rr.Header().Get("Location"))
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// HELPER: Real Environment Setup
// setupRealEnv attempts to connect to the REAL database and load REAL templates.
// It returns a Handler ready for integration testing.
func setupRealEnv() (*auth.Handler, error) {
	// Load .env file
	// Try loading from parent directory first (common when running tests in ./test)
	err := godotenv.Load("../.env")
	if err != nil {
		// Fallback: try loading from current directory
		if err := godotenv.Load(); err != nil {
			return nil, fmt.Errorf("error loading .env file: %v", err)
		}
	}

	// Read .env VALUES
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")

	if dbUser == "" || dbHost == "" {
		return nil, fmt.Errorf("environment variables (DB_USER/DB_HOST) are empty")
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUser, dbPass, dbHost, dbPort, dbName)

	// Setup Templates
	// Uses := to declare new variable 'tpl'
	tpl, err := template.ParseGlob("../templates/*.html")
	if err != nil {
		log.Printf("Warning: Could not load templates from ../templates: %v", err)
		// Try current directory as fallback
		tpl, err = template.ParseGlob("templates/*.html")
		if err != nil {
			return nil, fmt.Errorf("failed to load templates: %v", err)
		}
	}

	// Setup Database
	// Uses := to declare new variable 'db'
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("sql.Open failed: %v", err)
	}

	// Always verify connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("database connection failed: %v", err)
	}

	return &auth.Handler{
		DB:  db,
		Tpl: tpl,
	}, nil
}

// TestIntegration_RealDB uses the setupRealEnv helper.
// It will SKIP if it cannot connect to the real database.
func TestIntegration_RealDB(t *testing.T) {
	// Try to setup real env
	h, err := setupRealEnv()
	if err != nil {
		t.Skipf("Skipping integration test: %v", err)
	}
	defer h.DB.Close()

	// You can now write a test that hits the REAL database.
	// Be careful: this modifies real data!
	t.Log("Successfully connected to Real DB, running integration check...")

	// Example check: Ensure template is loaded
	if h.Tpl == nil {
		t.Error("Template should not be nil")
	}
}
