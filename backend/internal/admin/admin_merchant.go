package admin

import (
	"database/sql"
	"encoding/json"
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
	data := AdminPageData{
		Page:     "merchants",
		Username: r.PathValue("username"),
	}
	h.Tpl.ExecuteTemplate(w, "admin_merchant.html", data)
}

func (h *Handler) MerchantManagementDataHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("MerchantManagementDataHandler running...")

	// Parse pagination parameters
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")
	search := r.URL.Query().Get("search")
	sortOrder := r.URL.Query().Get("sort") // desc or asc
	category := r.URL.Query().Get("category")
	status := r.URL.Query().Get("status")

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

	if category != "" {
		conditions = append(conditions, `business_type = ?`)
		args = append(args, category)
	}

	if status != "" {
		conditions = append(conditions, `status = ?`)
		args = append(args, status)
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
			JOIN merchants m ON t.merchant_id = m.user_id 
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

type ApproveMerchantRequest struct {
	CommissionRate    string `json:"commissionRate" validate:"required"`
	SettlementBank    string `json:"settlementBank" validate:"required"`
	SettlementName    string `json:"settlementName" validate:"required"`
	SettlementAccount string `json:"settlementAccount" validate:"required"`
	TerminalSn        string `json:"terminalSn" validate:"required"`
	DeviceName        string `json:"deviceName"`
}

func (h *Handler) ApproveMerchantHandler(w http.ResponseWriter, r *http.Request) {
	merchantID := r.PathValue("id")
	adminUsername := r.PathValue("username")

	if merchantID == "" {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{Success: false, Message: "Merchant ID is required"})
		return
	}

	var req ApproveMerchantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{Success: false, Message: "Invalid request payload"})
		return
	}

	// Begin TX
	tx, err := h.DB.Begin()
	if err != nil {
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{Success: false, Message: "Database error"})
		return
	}
	defer tx.Rollback()

	// Get admin user_id
	var adminUserID string
	err = tx.QueryRow("SELECT user_id FROM users WHERE username = ?", adminUsername).Scan(&adminUserID)
	if err != nil {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{Success: false, Message: "Admin user not found"})
		return
	}

	// Get merchant user_id
	var merchantUserID string
	err = tx.QueryRow("SELECT user_id FROM merchants WHERE merchant_id = ?", merchantID).Scan(&merchantUserID)
	if err != nil {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{Success: false, Message: "Merchant not found"})
		return
	}

	// Update merchants table
	_, err = tx.Exec(`
		UPDATE merchants 
		SET status = 'active',
			commission_rate = ?,
			settlement_bank_name = ?,
			settlement_account_name = ?,
			settlement_account_number = ?,
			approved_by = ?,
			approved_at = CURRENT_TIMESTAMP
		WHERE merchant_id = ?`,
		req.CommissionRate, req.SettlementBank, req.SettlementName, req.SettlementAccount, adminUserID, merchantID)

	if err != nil {
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{Success: false, Message: "Failed to update merchant"})
		return
	}

	// Update users table
	_, err = tx.Exec("UPDATE users SET status = 'active' WHERE user_id = ?", merchantUserID)
	if err != nil {
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{Success: false, Message: "Failed to update user status"})
		return
	}

	// Update terminals table
	// First get business address for location details
	var businessAddress string
	_ = tx.QueryRow("SELECT business_address FROM merchants WHERE merchant_id = ?", merchantID).Scan(&businessAddress)

	_, err = tx.Exec("UPDATE terminals SET merchant_id = ?, location_details = ?, status = 'active' WHERE terminal_sn = ?", merchantUserID, businessAddress, req.TerminalSn)
	if err != nil {
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{Success: false, Message: "Failed to assign terminal"})
		return
	}

	if err := tx.Commit(); err != nil {
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{Success: false, Message: "Failed to finalize approval"})
		return
	}

	jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{Success: true, Message: "Merchant approved successfully"})
}

func (h *Handler) RejectMerchantHandler(w http.ResponseWriter, r *http.Request) {
	merchantID := r.PathValue("id")
	if merchantID == "" {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{Success: false, Message: "Merchant ID is required"})
		return
	}

	tx, err := h.DB.Begin()
	if err != nil {
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{Success: false, Message: "Database error"})
		return
	}
	defer tx.Rollback()

	var merchantUserID string
	err = tx.QueryRow("SELECT user_id FROM merchants WHERE merchant_id = ?", merchantID).Scan(&merchantUserID)
	if err != nil {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{Success: false, Message: "Merchant not found"})
		return
	}

	_, err = tx.Exec("UPDATE merchants SET status = 'rejected' WHERE merchant_id = ?", merchantID)
	if err != nil {
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{Success: false, Message: "Failed to reject merchant"})
		return
	}

	_, err = tx.Exec("UPDATE users SET status = 'inactive' WHERE user_id = ?", merchantUserID)
	if err != nil {
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{Success: false, Message: "Failed to deactivate user"})
		return
	}

	if err := tx.Commit(); err != nil {
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{Success: false, Message: "Failed to finalize rejection"})
		return
	}

	jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{Success: true, Message: "Merchant rejected successfully"})
}

type MerchantDetailsData struct {
	MerchantID      string
	UserID          string
	BusinessName    string
	BusinessType    string
	RegistrationNum string
	BusinessAddress string
	OwnerName       string
	BusinessEmail   string
	BusinessPhone   string
	Status          string
	CommissionRate  float64
	SettlementBank  string
	SettlementName  string
	SettlementAcct  string
	CreatedAt       string
	DtiDocument     string
	BirDocument     string
	OtherDocument   string
}

type MerchantInfoViewData struct {
	Page     string
	Username string
}

func (h *Handler) MerchantInfoView(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")

	data := MerchantInfoViewData{
		Page:     "merchants",
		Username: username,
	}

	err := h.Tpl.ExecuteTemplate(w, "merchant_info.html", data)
	if err != nil {
		fmt.Printf("Template execution error: %v\n", err)
	}
}

func (h *Handler) MerchantInfoDataHandler(w http.ResponseWriter, r *http.Request) {
	merchantID := r.PathValue("id")

	if merchantID == "" {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "Merchant ID required",
		})
		return
	}

	var m MerchantDetailsData
	var commRate sql.NullFloat64
	var setBank, setName, setAcct, regNum, dtiDoc, birDoc, otherDoc sql.NullString

	err := h.DB.QueryRow(`
		SELECT merchant_id, user_id, business_name, business_type, business_registration_number, 
		       business_address, owner_name, business_email, business_phone, status, 
		       commission_rate, settlement_bank_name, settlement_account_name, 
		       settlement_account_number, created_at,
		       dti_document, bir_document, other_document
		FROM merchants WHERE merchant_id = ?`, merchantID).Scan(
		&m.MerchantID, &m.UserID, &m.BusinessName, &m.BusinessType, &regNum,
		&m.BusinessAddress, &m.OwnerName, &m.BusinessEmail, &m.BusinessPhone, &m.Status,
		&commRate, &setBank, &setName, &setAcct, &m.CreatedAt,
		&dtiDoc, &birDoc, &otherDoc,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			jsonwrite.WriteJSON(w, http.StatusNotFound, jsonwrite.APIResponse{
				Success: false,
				Message: "Merchant not found",
			})
			return
		}
		log.Println("Error querying merchant details:", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Database error",
		})
		return
	}

	if regNum.Valid {
		m.RegistrationNum = regNum.String
	}

	if commRate.Valid {
		m.CommissionRate = commRate.Float64
	}
	if setBank.Valid {
		m.SettlementBank = setBank.String
	}
	if setName.Valid {
		m.SettlementName = setName.String
	}
	if setAcct.Valid {
		m.SettlementAcct = setAcct.String
	}
	if dtiDoc.Valid {
		m.DtiDocument = dtiDoc.String
	}
	if birDoc.Valid {
		m.BirDocument = birDoc.String
	}
	if otherDoc.Valid {
		m.OtherDocument = otherDoc.String
	}

	jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{
		Success: true,
		Data:    m,
	})
}
