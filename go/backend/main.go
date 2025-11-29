package main

import (
    "database/sql"
    "fmt"
    "html/template"
    "net/http"

    // IMPORT YOUR LOCAL PACKAGE
    // This path must match your folder structure and go.mod name
    "unicard-go/go/backend/login"

    _ "github.com/go-sql-driver/mysql"
)

var tpl *template.Template
var db *sql.DB

func main() {
    // 1. Setup Templates
    // Ensure you have a folder named 'templates' with .html files inside
    var err error
    tpl, _ = template.ParseGlob("../templates/*.html")

    // 2. Setup Database
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

    // 3. Initialize the Login Handler (Dependency Injection)
    // We inject the running 'db' and 'tpl' into the login package
    loginH := &login.Handler{
        DB:  db,
        Tpl: tpl,
    }

    // 4. Routes
    // This handles the GET request (showing the form)
    http.HandleFunc("/login", loginHandler)

    // This handles the POST request (processing the form) from the external package
    http.HandleFunc("/loginauth", loginH.LoginAuthHandler)

    fmt.Println("Server started on: http://localhost:8080/login")
    http.ListenAndServe(":8080", nil)
}

// Simple handler just to show the login page initially
func loginHandler(w http.ResponseWriter, r *http.Request) {
    tpl.ExecuteTemplate(w, "login.html", nil)
}