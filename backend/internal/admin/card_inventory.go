package admin

import (
	"fmt"
	"net/http"
	jsonwrite "unicard-go/backend/internal/pkg/handler"
)

// AdminCard represents a card entry in the admin database
type AdminCard struct {
	UserID     string  `json:"user_id" db:"user_id"`
	CardNumber string  `json:"card_number" db:"card_number"`
	CardType   string  `json:"card_type" db:"card_type"`
	Balance    float64 `json:"initial_amount" db:"balance"`
	ExpiryDate string  `json:"expiry_date" db:"expiry_date"`
	Status     string  `json:"status" db:"status"`
	CreatedAt  string  `json:"created_at" db:"created_at"`
}

// AdminCardInventoryStats contains statistics about cards
type AdminCardInventoryStats struct {
	Total    int `json:"total"`
	Active   int `json:"active"`
	Inactive int `json:"inactive"`
	Blocked  int `json:"blocked"`
	Lost     int `json:"lost"`
}

// CardInventoryView handles rendering the static admin dashboard view
func (h *Handler) CardInventoryView(w http.ResponseWriter, r *http.Request) {
	fmt.Println("CardInventoryView running...")
	data := AdminPageData{
		Page:     "card-inventory",
		Username: r.PathValue("username"),
	}
	err := h.Tpl.ExecuteTemplate(w, "admin_dashboard.html", data)
	if err != nil {
		fmt.Printf("Template execution error: %v\n", err)
	}
}

// CardInventoryDataHandler returns the stats and cards list as JSON
func (h *Handler) CardInventoryDataHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("CardInventoryDataHandler running...")

	// Fetch Stats
	var stats AdminCardInventoryStats
	h.DB.QueryRow("SELECT COUNT(*) FROM cards").Scan(&stats.Total)
	h.DB.QueryRow("SELECT COUNT(*) FROM cards WHERE status = 'Active'").Scan(&stats.Active)
	h.DB.QueryRow("SELECT COUNT(*) FROM cards WHERE status = 'Inactive'").Scan(&stats.Inactive)
	h.DB.QueryRow("SELECT COUNT(*) FROM cards WHERE status = 'Blocked'").Scan(&stats.Blocked)
	h.DB.QueryRow("SELECT COUNT(*) FROM cards WHERE status = 'Lost'").Scan(&stats.Lost)

	// Fetch Cards
	rows, err := h.DB.Query(`
		SELECT user_id, card_number, card_type, balance, status, expiry_date, created_at
		FROM cards
		ORDER BY created_at DESC
	`)
	if err != nil {
		fmt.Printf("Error fetching cards: %v\n", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Failed to fetch cards",
		})
		return
	}
	defer rows.Close()

	var cards []AdminCard
	for rows.Next() {
		var c AdminCard
		err := rows.Scan(
			&c.UserID,
			&c.CardNumber,
			&c.CardType,
			&c.Balance,
			&c.Status,
			&c.ExpiryDate,
			&c.CreatedAt,
		)
		if err != nil {
			fmt.Printf("Error scanning card row: %v\n", err)
			continue
		}

		cards = append(cards, c)
	}

	resp := struct {
		Stats AdminCardInventoryStats `json:"stats"`
		Cards []AdminCard             `json:"cards"`
	}{
		Stats: stats,
		Cards: cards,
	}

	jsonwrite.WriteJSON(w, http.StatusOK, resp)
}
