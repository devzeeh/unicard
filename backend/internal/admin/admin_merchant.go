package admin

import (
	"log"
	"net/http"
	jsonwrite "unicard-go/backend/internal/pkg/handler"
	structs "unicard-go/backend/internal/pkg/structs"
)

func (h *Handler) MerchantManagementView(w http.ResponseWriter, r *http.Request) {
	log.Println("MerchantManagementView running...")
	h.Tpl.ExecuteTemplate(w, "admin_merchant.html", nil)
}

func (h *Handler) MerchantManagementDataHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("MerchantManagementDataHandler running...")

	//var dashboardData structure.AdminDashboardData

	query := `SELECT merchant_id, business_name, business_type, owner_name, business_email, business_phone, status, created_at FROM merchants`
	rows, err := h.DB.Query(query)
	if err != nil {
		log.Println("Error querying merchants:", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Error querying merchants",
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

	merchantData := structs.AdminDashboardData{
		Merchants: merchants,
	}

	jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{
		Success: true,
		Message: "Merchants retrieved successfully",
		Data:    merchantData,
	})
	log.Println("MerchantManagementDataHandler finished")
}
