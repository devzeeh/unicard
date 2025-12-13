package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"unicard-go/go/backend/auth"

	_ "github.com/go-sql-driver/mysql"
)

var tpl *template.Template
var db *sql.DB

func main() {
	// Setup Templates
	// Ensure you have a folder named 'templates' with .html files inside
	var err error
	tpl, _ = template.ParseGlob("../templates/*.html")

	// Setup Database
	// Change 'root:devengr' to your actual db username/password
	db, err = sql.Open("mysql", "root:devengr@tcp(localhost:3306)/testdb")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic("Database connection failed: " + err.Error())
	}

	// Initialize the Login Handler (Dependency Injection)
	// We inject the running 'db' and 'tpl' into the login package
	loginH := &auth.Handler{
		DB:  db,
		Tpl: tpl,
	}

	// Routes
	// This handles the GET request (showing the form)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/dashboard", dashboardHandler)

	// This handles the POST request (processing the form) from the external package
	http.HandleFunc("/loginauth", loginH.LoginAuthHandler)

    // Start Server
	fmt.Println("Server started on: http://localhost:8080/login")
	http.ListenAndServe(":8080", nil)
}

// Simple handler just to show the login page initially
func loginHandler(w http.ResponseWriter, r *http.Request) {
	tpl.ExecuteTemplate(w, "login.html", nil)
}

// Dashboard handler to show after successful login
func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Dashboard is running")
	tpl.ExecuteTemplate(w, "dashboard.html", nil)
}
