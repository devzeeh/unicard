package auth

import (
	"database/sql"
	"html/template"
)

// One Handler to rule them all
type Handler struct {
	DB  *sql.DB
	Tpl *template.Template
}