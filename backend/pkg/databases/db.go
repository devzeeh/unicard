// DB connection setup
package db

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"os"

	"github.com/joho/godotenv"
)

var tpl *template.Template
var db *sql.DB

func DBSetup() {
	// Load .env file
	err := godotenv.Load("../.env")
	if err != nil {
		// Fallback: try loading from current directory
		if err := godotenv.Load(); err != nil {
			log.Fatalf("Error loading .env file: %v", err)
		}
	}

	// read .env VALUES
	//port := os.Getenv("PORT")
	//serverAddress := os.Getenv("SERVER_ADDR")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", dbUser, dbPass, dbHost, dbPort, dbName)

	// Setup Templates
	tpl, err = template.ParseGlob("../templates/*.html")
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
}
