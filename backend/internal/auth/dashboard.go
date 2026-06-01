package authentication

import (
	"fmt"
	"net/http"
	"unicard-go/backend/internal/pkg/structs"
)

func (h *Handler) DashboardView(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Dashboard view is running...")
	h.Tpl.ExecuteTemplate(w, "dashboard.html", nil)
}

func (h *Handler) DashboardHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Dashboard handler is running...")
	
	// Select user and card details from database
	query := `
		SELECT 
			u.id, 
			u.user_id, 
			u.username, 
			u.name, 
			COALESCE(c.balance, 0), 
			COALESCE(c.loyalty_points, 0), 
			COALESCE(c.card_type, 'Regular')
		FROM users u
		LEFT JOIN cards c ON u.user_id = c.user_id
		WHERE u.user_id = ?
	`
	rows, err := h.DB.Query(query, r.URL.Query().Get("user_id"))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer rows.Close()

	var user structure.DashboardUser
	if rows.Next() {
		if err := rows.Scan(&user.ID, &user.UserID, &user.Username, &user.Name, &user.Balance, &user.LoyaltyPoints, &user.AccountType); err != nil {
			fmt.Println(err)
			return
		}
	}
	
	if user.ID == 0 {
		fmt.Println("User not found")
		return
	}

	h.Tpl.ExecuteTemplate(w, "dashboard.html", user)
}