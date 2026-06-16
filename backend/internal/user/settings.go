package user

import (
	"fmt"
	"net/http"
)

// SettingsView handles the display of the user's settings page
func (h *Handler) SettingsView(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Settings view is running...")

	username := r.PathValue("username")
	data := struct {
		Username string
	}{
		Username: username,
	}

	h.Tpl.ExecuteTemplate(w, "settings.html", data)
}
