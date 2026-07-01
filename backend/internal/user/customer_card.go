package user

import (
	"encoding/json"
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

func (h *Handler) UpdateCardStatus(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	
	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var userID string
	err := h.Store.QueryRow("SELECT user_id FROM users WHERE username = ?", username).Scan(&userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Make sure status is valid
	if req.Status != "active" && req.Status != "inactive" && req.Status != "blocked" && req.Status != "lost" {
		http.Error(w, "Invalid status", http.StatusBadRequest)
		return
	}

	_, err = h.Store.Exec("UPDATE cards SET status = ? WHERE user_id = ?", req.Status, userID)
	if err != nil {
		http.Error(w, "Failed to update card status", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success": true}`))
}
