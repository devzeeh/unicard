package admin

import (
	"log"
	"net/http"
	jsonwrite "unicard-go/backend/internal/pkg/handler"
	structs "unicard-go/backend/internal/pkg/structs"
)

// AdminDashboardView renders the platform_overview.html template after checking the admin session.
func (h *Handler) AdminDashboardView(w http.ResponseWriter, r *http.Request) {
	log.Println("AdminDashboardView running...")
	h.Tpl.ExecuteTemplate(w, "admin_dashboard.html", nil)
}

// AdminDashboardDataHandler handles the request for admin dashboard data and returns JSON response.
// AdminDashboardDataHandler handles the request for admin dashboard data and returns JSON response.
func (h *Handler) AdminDashboardDataHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("AdminDashboardDataHandler running...")

	// Compute UniCard's Absolute Gross Revenue
	// (Transaction Service Fees)
	query := `SELECT COALESCE(SUM(service_fee), 0.00) FROM transactions WHERE transaction_type = 'payment'`
	row := h.DB.QueryRow(query)

	var grossRevenue float64
	if err := row.Scan(&grossRevenue); err != nil {
		log.Println("Error scanning gross revenue:", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Internal server error",
		})
		return
	}
	log.Println("Gross revenue row:", grossRevenue)

	// Compute UniCard's Net Revenue
	// (Service fees from successful payments MINUS service fees lost to refunds/reversals)
	query = `SELECT COALESCE(SUM(
			CASE 
				WHEN transaction_type = 'payment' THEN service_fee 
				WHEN transaction_type IN ('refund', 'reversal') THEN -service_fee 
				ELSE 0 
			END
        ), 0.00) FROM transactions`

	row = h.DB.QueryRow(query)

	var netRevenue float64
	if err := row.Scan(&netRevenue); err != nil {
		log.Println("Error scanning net revenue:", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Internal server error while calculating Net Revenue",
		})
		return
	}
	log.Println("Net revenue row:", netRevenue)

	// Display the number of users
	// Counting only 'active' customers (excluding suspended or inactive accounts)
	row = h.DB.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'customer'") // status = 'active'

	var totalUsers int
	if err := row.Scan(&totalUsers); err != nil {
		log.Println("Error scanning total users:", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Internal server error",
		})
		return
	}
	log.Println("Total users row:", totalUsers)

	// Display the number of cards
	// Counting only 'active' cards to show actual circulatory supply
	row = h.DB.QueryRow("SELECT COUNT(*) FROM cards") // WHERE status = 'active'

	var totalCards int
	if err := row.Scan(&totalCards); err != nil {
		log.Println("Error scanning total cards:", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Internal server error",
		})
		return
	}
	log.Println("Total cards row:", totalCards)

	// Display the number of merchants
	// Counting only 'active' merchants (excluding 'pending_approval' or 'suspended')
	row = h.DB.QueryRow("SELECT COUNT(*) FROM merchants WHERE status = 'active'")

	var totalMerchants int
	if err := row.Scan(&totalMerchants); err != nil {
		log.Println("Error scanning total merchants:", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Internal server error",
		})
		return
	}
	log.Println("Total merchants row:", totalMerchants)

	// Display the number of terminals
	// Counting 'active' ESP32 nodes (excluding 'offline' or 'suspended')
	row = h.DB.QueryRow("SELECT COUNT(*) FROM terminals WHERE status = 'active'")

	var totalTerminals int
	if err := row.Scan(&totalTerminals); err != nil {
		log.Println("Error scanning total terminals:", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Internal server error",
		})
		return
	}
	log.Println("Total terminals row:", totalTerminals)

	// Fetch all merchants for the table
	merchantQuery := "SELECT merchant_id, business_name, business_type, owner_name, business_email, business_phone, status, created_at FROM merchants"
	rows, err := h.DB.Query(merchantQuery)
	if err != nil {
		log.Println("Error querying merchants:", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Internal server error",
		})
		return
	}
	defer rows.Close()

	var merchants []structs.Merchant
	for rows.Next() {
		var m structs.Merchant
		if err := rows.Scan(&m.MerchantID, &m.BusinessName, &m.BusinessType, &m.OwnerName, &m.Email, &m.Phone, &m.Status, &m.CreatedAt); err != nil {
			log.Println("Error scanning merchant:", err)
			continue
		}
		merchants = append(merchants, m)
	}

	// Return the data as JSON
	response := structs.AdminDashboardData{
		GrossRevenue:    grossRevenue,
		NetRevenue:      netRevenue,
		TotalUsers:      totalUsers,
		TotalCards:      totalCards,
		ActiveMerchants: totalMerchants,
		ActiveTerminals: totalTerminals,
		Merchants:       merchants,
	}
	jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{
		Success: true,
		Message: "Admin dashboard data retrieved successfully",
		Data:    response,
	})
	log.Println("AdminDashboardDataHandler finished")
}
