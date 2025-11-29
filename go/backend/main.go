package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"

	"unicard-go/go/backend/login"

	_ "github.com/go-sql-driver/mysql"
)

var tpl *template.Template
var db *sql.DB

func main() {
	// 1. Setup Templates
	tpl, _ = template.ParseGlob("templates/*.html")

	// 2. Setup Database
	var err error
	db, err = sql.Open("mysql", "root:devengr@tcp(localhost:3306)/testdb")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err.Error())
	}

	// 3. Initialize the Login Handler
	// We inject the 'db' and 'tpl' from main into the login package
	loginH := &login.Handler{
		DB:  db,
		Tpl: tpl,
	}

	// 4. Routes
	http.HandleFunc("/login", loginHandler)

	// Use the method from the struct
	http.HandleFunc("/loginauth", loginH.LoginAuthHandler)

	fmt.Println("Server started on: http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	tpl.ExecuteTemplate(w, "login.html", nil)
}
