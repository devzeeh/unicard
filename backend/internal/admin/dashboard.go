package admin

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
)

// AdminCard represents a card entry in the admin database
type AdminCard struct {
	ID            int     `json:"id"`
	CardUID       string  `json:"card_uid"`
	CardNumber    string  `json:"card_number"`
	CardType      string  `json:"card_type"`
	InitialAmount float64 `json:"initial_amount"`
	ExpiryDate    string  `json:"expiry_date"`
	Status        string  `json:"status"`
	CreatedAt     string  `json:"created_at"`
	CardHolder    string  `json:"card_holder"`
	UserID        string  `json:"user_id"`
}

// AdminStats contains statistics about cards
type AdminStats struct {
	Total    int `json:"total"`
	Active   int `json:"active"`
	Inactive int `json:"inactive"`
	Blocked  int `json:"blocked"`
	Lost     int `json:"lost"`
}

// DashboardView handles rendering the static admin dashboard view
func (h *Handler) DashboardView(w http.ResponseWriter, r *http.Request) {
	fmt.Println("DashboardView running...")

	// Validate admin session
	cookie, err := r.Cookie("session_admin_username")
	if err != nil || cookie.Value == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	err = h.Tpl.ExecuteTemplate(w, "admin_dashboard.html", nil)
	if err != nil {
		fmt.Printf("Template execution error: %v\n", err)
	}
}

// DashboardDataHandler returns the stats and cards list as JSON
func (h *Handler) DashboardDataHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("DashboardDataHandler running...")

	//time.Sleep(5 * time.Second)

	cookie, err := r.Cookie("session_admin_username")
	if err != nil || cookie.Value == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
		return
	}

	// 1. Fetch Stats
	var stats AdminStats
	h.DB.QueryRow("SELECT COUNT(*) FROM cards").Scan(&stats.Total)
	h.DB.QueryRow("SELECT COUNT(*) FROM cards WHERE status = 'Active'").Scan(&stats.Active)
	h.DB.QueryRow("SELECT COUNT(*) FROM cards WHERE status = 'Inactive'").Scan(&stats.Inactive)
	h.DB.QueryRow("SELECT COUNT(*) FROM cards WHERE status = 'Blocked'").Scan(&stats.Blocked)
	h.DB.QueryRow("SELECT COUNT(*) FROM cards WHERE status = 'Lost'").Scan(&stats.Lost)

	// 2. Fetch Cards
	rows, err := h.DB.Query(`
		SELECT id, card_uid, card_number, card_type, initial_amount, expiry_date, status, created_at, card_holder, user_id
		FROM cards
		ORDER BY created_at DESC
	`)
	if err != nil {
		fmt.Printf("Error fetching cards: %v\n", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to fetch cards"})
		return
	}
	defer rows.Close()

	var cards []AdminCard
	for rows.Next() {
		var c AdminCard
		var cardHolderNull, userIDNull sql.NullString
		var createdAtNull sql.NullString
		err := rows.Scan(
			&c.ID,
			&c.CardUID,
			&c.CardNumber,
			&c.CardType,
			&c.InitialAmount,
			&c.ExpiryDate,
			&c.Status,
			&createdAtNull,
			&cardHolderNull,
			&userIDNull,
		)
		if err != nil {
			fmt.Printf("Error scanning card row: %v\n", err)
			continue
		}

		if createdAtNull.Valid {
			c.CreatedAt = createdAtNull.String
		} else {
			c.CreatedAt = ""
		}

		if cardHolderNull.Valid {
			c.CardHolder = cardHolderNull.String
		} else {
			c.CardHolder = "Unlinked"
		}

		if userIDNull.Valid {
			c.UserID = userIDNull.String
		} else {
			c.UserID = "None"
		}

		cards = append(cards, c)
	}

	resp := struct {
		Stats AdminStats  `json:"stats"`
		Cards []AdminCard `json:"cards"`
	}{
		Stats: stats,
		Cards: cards,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
