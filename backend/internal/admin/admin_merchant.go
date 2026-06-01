package admin

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	jsonwrite "unicard-go/backend/internal/pkg/handler"
	structs "unicard-go/backend/internal/pkg/structs"
)

func (h *Handler) MerchantManagementView(w http.ResponseWriter, r *http.Request) {
	log.Println("MerchantManagementView running...")
	h.Tpl.ExecuteTemplate(w, "admin_merchant.html", nil)
}

func (h *Handler) MerchantManagementDataHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("MerchantManagementDataHandler running...")

	// Parse pagination parameters
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")
	search := r.URL.Query().Get("search")
	sortOrder := r.URL.Query().Get("sort") // desc or asc

	page := 1
	limit := 10
	if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
		page = p
	}
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		limit = l
	}

	offset := (page - 1) * limit

	// Build query
	baseQuery := `FROM merchants`
	var args []interface{}
	var conditions []string

	if search != "" {
		conditions = append(conditions, `(business_name LIKE ? OR owner_name LIKE ? OR merchant_id LIKE ?)`)
		searchPattern := "%" + search + "%"
		args = append(args, searchPattern, searchPattern, searchPattern)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total items
	countQuery := `SELECT COUNT(*) ` + baseQuery + whereClause
	var totalItems int
	if err := h.DB.QueryRow(countQuery, args...).Scan(&totalItems); err != nil {
		log.Println("Error counting merchants:", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Error counting merchants",
		})
		return
	}

	// Get paginated data
	orderClause := " ORDER BY created_at DESC"
	if strings.ToLower(sortOrder) == "asc" {
		orderClause = " ORDER BY created_at ASC"
	}

	query := `SELECT merchant_id, business_name, business_type, owner_name, business_email, business_phone, status, created_at ` +
		baseQuery + whereClause + orderClause + ` LIMIT ? OFFSET ?`

	args = append(args, limit, offset)

	rows, err := h.DB.Query(query, args...)
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
	if merchants == nil {
		merchants = []structs.Merchant{}
	} else {
		// Fetch terminals for these merchants
		var merchantIDs []string
		for _, m := range merchants {
			merchantIDs = append(merchantIDs, m.MerchantID)
		}

		placeholders := make([]string, len(merchantIDs))
		termArgs := make([]interface{}, len(merchantIDs))
		for i, id := range merchantIDs {
			placeholders[i] = "?"
			termArgs[i] = id
		}

		termQuery := fmt.Sprintf(`
			SELECT m.merchant_id, t.terminal_id, t.terminal_sn, t.device_name, t.status 
			FROM terminals t 
			JOIN merchants m ON t.merchant_id = m.id 
			WHERE m.merchant_id IN (%s)`, strings.Join(placeholders, ","))

		termRows, err := h.DB.Query(termQuery, termArgs...)
		if err == nil {
			defer termRows.Close()
			termMap := make(map[string][]structs.Terminal)
			for termRows.Next() {
				var mID string
				var t structs.Terminal
				if err := termRows.Scan(&mID, &t.TerminalID, &t.TerminalSN, &t.DeviceName, &t.Status); err == nil {
					termMap[mID] = append(termMap[mID], t)
				}
			}
			for i := range merchants {
				merchants[i].Terminals = termMap[merchants[i].MerchantID]
				if merchants[i].Terminals == nil {
					merchants[i].Terminals = []structs.Terminal{}
				}
			}
		} else {
			log.Println("Error fetching terminals for merchants:", err)
			for i := range merchants {
				merchants[i].Terminals = []structs.Terminal{}
			}
		}
	}

	type PaginatedMerchantResponse struct {
		Merchants  []structs.Merchant `json:"merchants"`
		TotalItems int                `json:"totalItems"`
		Page       int                `json:"page"`
		Limit      int                `json:"limit"`
	}

	merchantData := PaginatedMerchantResponse{
		Merchants:  merchants,
		TotalItems: totalItems,
		Page:       page,
		Limit:      limit,
	}

	jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{
		Success: true,
		Message: "Merchants retrieved successfully",
		Data:    merchantData,
	})
	log.Println("MerchantManagementDataHandler finished")
}
