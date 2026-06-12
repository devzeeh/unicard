package user

import (
	"fmt"
	"net/http"
)

func (h *Handler) ProfileView(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Profile view is running...")

	username := r.PathValue("username")
	data := struct {
		Username string
	}{
		Username: username,
	}

	h.Tpl.ExecuteTemplate(w, "profile.html", data)
}
