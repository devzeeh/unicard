package user

import (
	"fmt"
	"net/http"
)

func (h *Handler) TopUpView(w http.ResponseWriter, r *http.Request) {
	fmt.Println("TopUp view is running...")

	username := r.PathValue("username")
	data := struct {
		Username string
	}{
		Username: username,
	}

	h.Tpl.ExecuteTemplate(w, "customer_topup.html", data)
}
