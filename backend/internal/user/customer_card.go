package user

import (
	"fmt"
	"net/http"
)

func (h *Handler) CardView(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Card view is running...")

	username := r.PathValue("username")
	data := struct {
		Username string
	}{
		Username: username,
	}

	h.Tpl.ExecuteTemplate(w, "card.html", data)
}
