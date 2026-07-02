package user

import (
	"html/template"
	"unicard-go/backend/internal/pkg/database"
)

type Handler struct {
	Store database.Store
	Tpl   *template.Template
}


func NewHandler(store database.Store, tpl *template.Template) *Handler {
	return &Handler{Store: store, Tpl: tpl}
}
