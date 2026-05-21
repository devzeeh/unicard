package user

import (
	"fmt"
	"net/http"

	jsonwrite "unicard-go/backend/internal/pkg/handler"
)

func (h *Handler) TransactionView(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Transaction view is running...")
	cookie, err := r.Cookie("session_user_id")
	if err != nil || cookie.Value == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	h.Tpl.ExecuteTemplate(w, "transaction.html", nil)
}

func (h *Handler) TopupView(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Topup view is running...")
	cookie, err := r.Cookie("session_user_id")
	if err != nil || cookie.Value == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	h.Tpl.ExecuteTemplate(w, "topup.html", nil)
}

func (h *Handler) ProfileView(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Profile view is running...")
	cookie, err := r.Cookie("session_user_id")
	if err != nil || cookie.Value == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	h.Tpl.ExecuteTemplate(w, "profile.html", nil)
}

func (h *Handler) SettingsView(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Settings view is running...")
	cookie, err := r.Cookie("session_user_id")
	if err != nil || cookie.Value == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	h.Tpl.ExecuteTemplate(w, "settings.html", nil)
}

func (h *Handler) CardView(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Card view is running...")
	cookie, err := r.Cookie("session_user_id")
	if err != nil || cookie.Value == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	h.Tpl.ExecuteTemplate(w, "card.html", nil)
}

func (h *Handler) TransactionsJSONHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Transactions JSON handler is running...")

	// Get session cookie
	cookie, err := r.Cookie("session_user_id")
	if err != nil || cookie.Value == "" {
		jsonwrite.WriteJSON(w, http.StatusUnauthorized, jsonwrite.APIResponse{
			Success: false,
			Message: "Unauthorized",
		})
		return
	}
	userID := cookie.Value

	// Fetch all transactions
	rows, err := h.DB.Query("SELECT created_at, description, transaction_type, amount FROM transactions WHERE user_id = ? ORDER BY created_at DESC", userID)
	var transactions []Transaction
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var t Transaction
			var createdAt string
			if err := rows.Scan(&createdAt, &t.Description, &t.Type, &t.Amount); err == nil {
				t.Date = formatDate(createdAt) // Using formatDate from dashboard.go
				transactions = append(transactions, t)
			}
		}
	} else {
		fmt.Printf("Error fetching transactions: %v\n", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Failed to load transactions",
		})
		return
	}

	// If no transactions, ensure we send an empty array, not null
	if transactions == nil {
		transactions = []Transaction{}
	}

	jsonwrite.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"success":      true,
		"transactions": transactions,
	})
}
