package authentication

import (
	"html/template"
	"unicard-go/backend/internal/pkg/database"
	"unicard-go/backend/internal/pkg/storage"
)

// The struct is shared across the files in this package
type Handler struct {
	Store   database.Store     // Database store
	Tpl     *template.Template // HTML templates
	Storage storage.Service    // Cloud storage service
}

// Optional: A constructor to make initialization cleaner
func NewHandler(store database.Store, tpl *template.Template, storage storage.Service) *Handler {
	return &Handler{
		Store:   store,
		Tpl:     tpl,
		Storage: storage,
	}
}
