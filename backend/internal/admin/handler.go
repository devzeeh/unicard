package admin

import (
	"html/template"
	"unicard-go/backend/internal/pkg/database"
)

// The struct is shared across the files in this package
type Handler struct {
	Store database.Store     // Database store
	Tpl   *template.Template // HTML templates
}

// Optional: A constructor to make initialization cleaner
func NewHandler(store database.Store, tpl *template.Template) *Handler {
	return &Handler{
		Store: store,
		Tpl:   tpl,
	}
}
