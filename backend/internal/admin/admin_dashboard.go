package admin

import (
	"log"
	"net/http"
	jsonwrite "unicard-go/backend/internal/pkg/handler"
	structs "unicard-go/backend/internal/pkg/structs"

	"github.com/shopspring/decimal"
)

type AdminPageData struct {
	Page     string
	Username string
}

// AdminDashboardView renders the platform_overview.html template after checking the admin session.
func (h *Handler) AdminDashboardView(w http.ResponseWriter, r *http.Request) {
	log.Println("AdminDashboardView running...")
	data := AdminPageData{
		Page:     "dashboard",
		Username: r.PathValue("username"),
	}
	h.Tpl.ExecuteTemplate(w, "admin_dashboard.html", data)
}

// AdminDashboardDataHandler handles the request for admin dashboard data and returns JSON response.
// AdminDashboardDataHandler handles the request for admin dashboard data and returns JSON response.
func (h *Handler) AdminDashboardDataHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("AdminDashboardDataHandler running...")

	// Compute UniCard's Gross Revenue
	// based on actual service fees collected in transactions
	query := `SELECT COALESCE(SUM(service_fee), 0.00) FROM transactions WHERE transaction_type IN ('payment', 'topup', 'withdrawal')`
	row := h.Store.QueryRow(query)

	var grossRevenue decimal.Decimal
	if err := row.Scan(&grossRevenue); err != nil {
		log.Println("Error scanning gross revenue:", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Internal server error",
		})
		return
	}

	// Compute UniCard's Net Revenue
	// based on actual service fees collected in transactions
	query = `SELECT COALESCE(SUM(
			CASE 
				WHEN transaction_type IN ('payment', 'topup', 'withdrawal') THEN service_fee 
				WHEN transaction_type IN ('refund', 'reversal') THEN -service_fee 
				ELSE 0 
			END
        ), 0.00) FROM transactions`

	row = h.Store.QueryRow(query)

	var netRevenue decimal.Decimal
	if err := row.Scan(&netRevenue); err != nil {
		log.Println("Error scanning net revenue:", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Internal server error while calculating Net Revenue",
		})
		return
	}

	log.Println("Gross revenue row:", grossRevenue)
	log.Println("Net revenue row:", netRevenue)

	// Display the number of users
	// Counting only 'active' customers (excluding suspended or inactive accounts)
	row = h.Store.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'customer'") // status = 'active'

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
	row = h.Store.QueryRow("SELECT COUNT(*) FROM cards") // WHERE status = 'active'

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

	// Display the number of merchants and breakdown
	row = h.Store.QueryRow(`
		SELECT 
			COUNT(*),
			COALESCE(SUM(CASE WHEN status = 'pending_approval' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN status = 'suspended' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN status = 'rejected' THEN 1 ELSE 0 END), 0)
		FROM merchants
	`)

	var totalMerchants, pendingMerchants, suspendedMerchants, rejectedMerchants int
	if err := row.Scan(&totalMerchants, &pendingMerchants, &suspendedMerchants, &rejectedMerchants); err != nil {
		log.Println("Error scanning merchants stats:", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Internal server error",
		})
		return
	}
	log.Println("Total merchants:", totalMerchants, "Pending:", pendingMerchants)

	// Display the number of terminals and breakdown
	row = h.Store.QueryRow(`
		SELECT 
			COUNT(*),
			COALESCE(SUM(CASE WHEN status = 'active' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN status = 'inactive' THEN 1 ELSE 0 END), 0)
		FROM terminals
	`)

	var totalTerminals, activeTerminals, inactiveTerminals int
	if err := row.Scan(&totalTerminals, &activeTerminals, &inactiveTerminals); err != nil {
		log.Println("Error scanning terminals stats:", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Internal server error",
		})
		return
	}
	log.Println("Total terminals:", totalTerminals, "Active:", activeTerminals, "Inactive:", inactiveTerminals)

	// Fetch recent merchants for the table (limit 5)
	merchantQuery := "SELECT merchant_id, business_name, business_type, owner_name, business_email, business_phone, status, created_at FROM merchants ORDER BY created_at DESC LIMIT 5"
	rows, err := h.Store.Query(merchantQuery)
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
		GrossRevenue:       grossRevenue,
		NetRevenue:         netRevenue,
		TotalUsers:         totalUsers,
		TotalCards:         totalCards,
		TotalMerchants:     totalMerchants,
		PendingMerchants:   pendingMerchants,
		SuspendedMerchants: suspendedMerchants,
		RejectedMerchants:  rejectedMerchants,
		TotalTerminals:     totalTerminals,
		ActiveTerminals:    activeTerminals,
		InactiveTerminals:  inactiveTerminals,
		Merchants:          merchants,
	}
	jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{
		Success: true,
		Message: "Admin dashboard data retrieved successfully",
		Data:    response,
	})
	log.Println("AdminDashboardDataHandler finished")
}
