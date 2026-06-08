package admin

import (
	"log"
	"net/http"
	jsonwrite "unicard-go/backend/internal/pkg/handler"

	"github.com/go-playground/validator/v10"
)

var Validate = validator.New()

// PlatformOverviewView serves the new Super Admin Platform Overview
func (h *Handler) PlatformOverviewView(w http.ResponseWriter, r *http.Request) {
	data := AdminPageData{Page: "dashboard", Username: r.PathValue("username")}
	err := h.Tpl.ExecuteTemplate(w, "platform_overview.html", data)
	if err != nil {
		log.Printf("Template execution error: %v", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Internal Server Error",
		})
	}
}

// MerchantManagementView serves the Merchant Management page
func (h *Handler) MerchantManagementViews(w http.ResponseWriter, r *http.Request) {
	data := AdminPageData{Page: "merchants", Username: r.PathValue("username")}
	err := h.Tpl.ExecuteTemplate(w, "merchant_management.html", data)
	if err != nil {
		log.Printf("Template execution error: %v", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Internal Server Error",
		})
	}
}

// TerminalRegistryView serves the Hardware Registry page
func (h *Handler) TerminalRegistryViews(w http.ResponseWriter, r *http.Request) {
	data := AdminPageData{Page: "terminals", Username: r.PathValue("username")}
	err := h.Tpl.ExecuteTemplate(w, "hardware_registry.html", data)
	if err != nil {
		log.Printf("Template execution error: %v", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Internal Server Error",
		})
	}
}

// SystemSettingsView serves the System Settings page
func (h *Handler) SystemSettingsView(w http.ResponseWriter, r *http.Request) {
	data := AdminPageData{Page: "settings", Username: r.PathValue("username")}
	err := h.Tpl.ExecuteTemplate(w, "system_settings.html", data)
	if err != nil {
		log.Printf("Template execution error: %v", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Internal Server Error",
		})
	}
}
