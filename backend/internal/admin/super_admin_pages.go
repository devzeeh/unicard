package admin

import (
	"log"
	"net/http"
)

// PlatformOverviewView serves the new Super Admin Platform Overview
func (h *Handler) PlatformOverviewView(w http.ResponseWriter, r *http.Request) {
	err := h.Tpl.ExecuteTemplate(w, "platform_overview.html", nil)
	if err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// MerchantManagementView serves the Merchant Management page
func (h *Handler) MerchantManagementView(w http.ResponseWriter, r *http.Request) {
	err := h.Tpl.ExecuteTemplate(w, "merchant_management.html", nil)
	if err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// TerminalRegistryView serves the Hardware Registry page
func (h *Handler) TerminalRegistryView(w http.ResponseWriter, r *http.Request) {
	err := h.Tpl.ExecuteTemplate(w, "hardware_registry.html", nil)
	if err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// SystemSettingsView serves the System Settings page
func (h *Handler) SystemSettingsView(w http.ResponseWriter, r *http.Request) {
	err := h.Tpl.ExecuteTemplate(w, "system_settings.html", nil)
	if err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
