package user

import (
	"database/sql"
	"html/template"
)

type Handler struct {
	DB  *sql.DB
	Tpl *template.Template
}


func NewHandler(db *sql.DB, tpl *template.Template) *Handler {
	return &Handler{DB: db, Tpl: tpl}
}
