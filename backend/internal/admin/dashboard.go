package admin

import (
	"database/sql"
	"fmt"
	"net/http"
)

// AdminCard represents a card entry in the admin database
type AdminCard struct {
	ID            int
	CardUID       string
	CardNumber    string
	CardType      string
	InitialAmount float64
	ExpiryDate    string
	Status        string
	CreatedAt     string
	CardHolder    string
	UserID        string
}

// AdminStats contains statistics about cards
type AdminStats struct {
	Total    int
	Active   int
	Inactive int
	Blocked  int
	Lost     int
}

// DashboardData is the context passed to the admin dashboard template
type DashboardData struct {
	Stats   AdminStats
	Cards   []AdminCard
	Error   string
	Success string
}

// DashboardView handles rendering the admin dashboard
func (h *Handler) DashboardView(w http.ResponseWriter, r *http.Request) {
	fmt.Println("DashboardView running...")

	queryError := r.URL.Query().Get("error")
	querySuccess := r.URL.Query().Get("success")

	// 1. Fetch Stats
	var stats AdminStats
	fmt.Println("Fetching stats...")
	h.DB.QueryRow("SELECT COUNT(*) FROM cards").Scan(&stats.Total)
	h.DB.QueryRow("SELECT COUNT(*) FROM cards WHERE status = 'Active'").Scan(&stats.Active)
	h.DB.QueryRow("SELECT COUNT(*) FROM cards WHERE status = 'Inactive'").Scan(&stats.Inactive)
	h.DB.QueryRow("SELECT COUNT(*) FROM cards WHERE status = 'Blocked'").Scan(&stats.Blocked)
	h.DB.QueryRow("SELECT COUNT(*) FROM cards WHERE status = 'Lost'").Scan(&stats.Lost)
	fmt.Println("Stats fetched:", stats)

	// 2. Fetch Cards
	fmt.Println("Querying cards...")
	rows, err := h.DB.Query(`
		SELECT id, card_uid, card_number, card_type, initial_amount, expiry_date, status, created_at, card_holder, user_id
		FROM cards
		ORDER BY created_at DESC
	`)
	if err != nil {
		fmt.Printf("Error fetching cards: %v\n", err)
		h.Tpl.ExecuteTemplate(w, "admin_dashboard.html", DashboardData{
			Stats: stats,
			Error: "Failed to fetch cards database list",
		})
		return
	}
	defer rows.Close()
	fmt.Println("Cards queried, scanning rows...")

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
	fmt.Printf("Scanned %d cards\n", len(cards))

	data := DashboardData{
		Stats:   stats,
		Cards:   cards,
		Error:   queryError,
		Success: querySuccess,
	}

	fmt.Println("Executing template admin_dashboard.html...")
	err = h.Tpl.ExecuteTemplate(w, "admin_dashboard.html", data)
	if err != nil {
		fmt.Printf("Template execution error: %v\n", err)
	} else {
		fmt.Println("Template executed successfully!")
	}
}
