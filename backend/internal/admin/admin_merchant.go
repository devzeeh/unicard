package admin

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	jsonwrite "unicard-go/backend/internal/pkg/handler"
	smtp "unicard-go/backend/internal/pkg/smtpbody"
	structs "unicard-go/backend/internal/pkg/structs"

	"gopkg.in/gomail.v2"
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
	var args []any
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
	if err := h.Store.QueryRow(countQuery, args...).Scan(&totalItems); err != nil {
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
	} else if strings.ToLower(sortOrder) == "name_asc" {
		orderClause = " ORDER BY business_name ASC"
	} else if strings.ToLower(sortOrder) == "name_desc" {
		orderClause = " ORDER BY business_name DESC"
	}

	query := `SELECT merchant_id, business_name, business_type, owner_name, business_email, business_phone, status, created_at ` +
		baseQuery + whereClause + orderClause + ` LIMIT ? OFFSET ?`

	args = append(args, limit, offset)

	rows, err := h.Store.Query(query, args...)
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
		termArgs := make([]any, len(merchantIDs))
		for i, id := range merchantIDs {
			placeholders[i] = "?"
			termArgs[i] = id
		}

		termQuery := fmt.Sprintf(`
			SELECT m.merchant_id, t.terminal_id, t.terminal_sn, t.device_name, t.status 
			FROM terminals t 
			JOIN merchants m ON t.merchant_id = m.merchant_id 
			WHERE m.merchant_id IN (%s)`, strings.Join(placeholders, ","))

		termRows, err := h.Store.Query(termQuery, termArgs...)
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
	CommissionRate string `json:"commissionRate" validate:"required"`
	TerminalSn     string `json:"terminalSn" validate:"required"`
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
	tx, err := h.Store.Begin()
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

	// Get merchant user_id, email, owner_name, and business_name for the notification email and welcome tx
	var merchantUserID, merchantEmail, ownerName, businessName string
	err = tx.QueryRow("SELECT user_id, business_email, owner_name, business_name FROM merchants WHERE merchant_id = ?", merchantID).Scan(&merchantUserID, &merchantEmail, &ownerName, &businessName)
	if err != nil {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{Success: false, Message: "Merchant not found"})
		return
	}

	// Update merchants table
	// Force commission_rate to 2.00 as per requirements
	_, err = tx.Exec(`
		UPDATE merchants 
		SET status = 'active',
			document_status = 'approved',
			message = 'Congratulations! Your UniCard Merchant Account is now fully active.',
			commission_rate = 2.00,
			approved_by = ?,
			approved_at = CURRENT_TIMESTAMP
		WHERE merchant_id = ?`,
		adminUserID, merchantID)

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

	// Determine which terminal to assign: admin-provided > pending merchant request > automatic unassigned terminal
	assignTerminalSN := req.TerminalSn
	var requestID string
	if assignTerminalSN == "" {
		// Check for pending terminal request
		err = tx.QueryRow("SELECT request_id, terminal_sn FROM terminal_requests WHERE merchant_id = ? AND status = 'pending' ORDER BY requested_at DESC LIMIT 1", merchantID).Scan(&requestID, &assignTerminalSN)
		if err != nil && err != sql.ErrNoRows {
			jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{Success: false, Message: "Failed to check terminal requests"})
			return
		}
		// If still empty, try to pick an unassigned terminal
		if assignTerminalSN == "" {
			// find any unassigned inactive terminal
			err = tx.QueryRow("SELECT terminal_sn FROM terminals WHERE merchant_id IS NULL AND status = 'inactive' LIMIT 1").Scan(&assignTerminalSN)
			if err != nil && err != sql.ErrNoRows {
				jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{Success: false, Message: "Failed to find available terminal"})
				return
			}
		}
	}

	if assignTerminalSN == "" {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{Success: false, Message: "No terminal selected or available"})
		return
	}

	var terminalDeviceName string
	err = tx.QueryRow("SELECT device_name FROM terminals WHERE terminal_sn = ?", assignTerminalSN).Scan(&terminalDeviceName)
	if err != nil {
		if err == sql.ErrNoRows {
			jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{Success: false, Message: "Selected terminal was not found"})
			return
		}
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{Success: false, Message: "Failed to read terminal details"})
		return
	}

	_, err = tx.Exec("UPDATE terminals SET merchant_id = ?, device_name = ?, location_details = ?, status = 'active' WHERE terminal_sn = ?", merchantID, terminalDeviceName, businessAddress, assignTerminalSN)
	if err != nil {
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{Success: false, Message: "Failed to assign terminal"})
		return
	}

	// If we used a pending request, mark it as approved
	if requestID != "" {
		_, _ = tx.Exec("UPDATE terminal_requests SET status = 'approved', handled_by = ?, handled_at = CURRENT_TIMESTAMP WHERE request_id = ?", adminUserID, requestID)
	}

	// Format business name safely for transaction ID (remove spaces, uppercase)
	/*safeBusinessName := strings.ToUpper(strings.ReplaceAll(businessName, " ", ""))
	if len(safeBusinessName) > 15 {
		safeBusinessName = safeBusinessName[:15]
	}*/

	// Insert welcome transaction
	_, err = tx.Exec(`
		INSERT INTO user_activity_logs (user_id, activity_type, channel, status, description)
		VALUES (?, 'onboarding', 'in_app', 'completed', 'Welcome to UniCard! Your merchant account is now approved and ready to accept transactions.')`,
		merchantUserID)
	if err != nil {
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{Success: false, Message: "Failed to create welcome transaction"})
		return
	}

	if err := tx.Commit(); err != nil {
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{Success: false, Message: "Failed to finalize approval"})
		return
	}

	// Send approval email to merchant
	go func(email, name string) {
		smtpHost := os.Getenv("SMTP_HOST")
		smtpPort := 587
		smtpEmail := os.Getenv("SMTP_EMAIL")
		smtpSender := os.Getenv("SMTP_SENDER")
		smtpPass := os.Getenv("SMTP_PASSWORD")
		if smtpHost == "" || smtpEmail == "" {
			log.Println("SMTP credentials not configured, skipping approval email")
			return
		}

		m := gomail.NewMessage()
		m.SetHeader("From", smtpSender+" <"+smtpEmail+">")
		m.SetHeader("To", email)
		m.SetHeader("Subject", "Unicard Application Approved")

		loginURL := "http://0.0.0.0:3000" // Adjust if there's an env var for frontend URL
		htmlBody := fmt.Sprintf(smtp.MerchantApprovedEmail(), name, loginURL)
		m.SetBody("text/html", htmlBody)

		d := gomail.NewDialer(smtpHost, smtpPort, smtpEmail, smtpPass)
		if err := d.DialAndSend(m); err != nil {
			log.Printf("Failed to send approval email to %s: %v", email, err)
		} else {
			log.Printf("Approval email sent successfully to %s", email)
		}
	}(merchantEmail, ownerName)

	jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{Success: true, Message: "Merchant approved successfully"})
}

func (h *Handler) RejectMerchantHandler(w http.ResponseWriter, r *http.Request) {
	merchantID := r.PathValue("id")
	if merchantID == "" {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{Success: false, Message: "Merchant ID is required"})
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{Success: false, Message: "Invalid request body"})
		return
	}
	if req.Reason == "" {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{Success: false, Message: "Rejection reason is required"})
		return
	}

	tx, err := h.Store.Begin()
	if err != nil {
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{Success: false, Message: "Database error"})
		return
	}
	defer tx.Rollback()

	var merchantUserID, merchantEmail, ownerName string
	err = tx.QueryRow("SELECT user_id, business_email, owner_name FROM merchants WHERE merchant_id = ?", merchantID).Scan(&merchantUserID, &merchantEmail, &ownerName)
	if err != nil {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{Success: false, Message: "Merchant not found"})
		return
	}

	_, err = tx.Exec("UPDATE merchants SET status = 'rejected', document_status = 'rejected', message = ? WHERE merchant_id = ?", req.Reason, merchantID)
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

	// Send rejection email to merchant
	go func(email, name, reason string) {
		smtpHost := os.Getenv("SMTP_HOST")
		smtpPort := 587
		smtpEmail := os.Getenv("SMTP_EMAIL")
		smtpSender := os.Getenv("SMTP_SENDER")
		smtpPass := os.Getenv("SMTP_PASSWORD")
		if smtpHost == "" || smtpEmail == "" {
			log.Println("SMTP credentials not configured, skipping rejection email")
			return
		}

		m := gomail.NewMessage()
		m.SetHeader("From", smtpSender+" <"+smtpEmail+">")
		m.SetHeader("To", email)
		m.SetHeader("Subject", "Unicard Application Update")

		htmlBody := fmt.Sprintf(smtp.MerchantRejectedEmail(), name, reason)
		m.SetBody("text/html", htmlBody)

		d := gomail.NewDialer(smtpHost, smtpPort, smtpEmail, smtpPass)
		if err := d.DialAndSend(m); err != nil {
			log.Printf("Failed to send rejection email to %s: %v", email, err)
		} else {
			log.Printf("Rejection email sent successfully to %s", email)
		}
	}(merchantEmail, ownerName, req.Reason)

	jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{Success: true, Message: "Merchant rejected successfully"})
}

func (h *Handler) SuspendMerchantHandler(w http.ResponseWriter, r *http.Request) {
	merchantID := r.PathValue("id")
	if merchantID == "" {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{Success: false, Message: "Merchant ID is required"})
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{Success: false, Message: "Invalid request body"})
		return
	}
	if req.Reason == "" {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{Success: false, Message: "Suspension reason is required"})
		return
	}

	tx, err := h.Store.Begin()
	if err != nil {
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{Success: false, Message: "Database error"})
		return
	}
	defer tx.Rollback()

	var merchantUserID, merchantEmail, ownerName string
	err = tx.QueryRow("SELECT user_id, business_email, owner_name FROM merchants WHERE merchant_id = ?", merchantID).Scan(&merchantUserID, &merchantEmail, &ownerName)
	if err != nil {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{Success: false, Message: "Merchant not found"})
		return
	}

	_, err = tx.Exec("UPDATE merchants SET status = 'suspended', message = ? WHERE merchant_id = ?", req.Reason, merchantID)
	if err != nil {
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{Success: false, Message: "Failed to suspend merchant"})
		return
	}

	_, err = tx.Exec("UPDATE users SET status = 'inactive' WHERE user_id = ?", merchantUserID)
	if err != nil {
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{Success: false, Message: "Failed to deactivate user"})
		return
	}

	if err := tx.Commit(); err != nil {
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{Success: false, Message: "Failed to finalize suspension"})
		return
	}

	// Send suspension email to merchant
	go func(email, name, reason string) {
		smtpHost := os.Getenv("SMTP_HOST")
		smtpPort := 587
		smtpEmail := os.Getenv("SMTP_EMAIL")
		smtpSender := os.Getenv("SMTP_SENDER")
		smtpPass := os.Getenv("SMTP_PASSWORD")
		if smtpHost == "" || smtpEmail == "" {
			log.Println("SMTP credentials not configured, skipping suspension email")
			return
		}

		m := gomail.NewMessage()
		m.SetHeader("From", smtpSender+" <"+smtpEmail+">")
		m.SetHeader("To", email)
		m.SetHeader("Subject", "Unicard Account Suspended")

		htmlBody := fmt.Sprintf(smtp.MerchantSuspendedEmail(), name, reason)
		m.SetBody("text/html", htmlBody)

		d := gomail.NewDialer(smtpHost, smtpPort, smtpEmail, smtpPass)
		if err := d.DialAndSend(m); err != nil {
			log.Printf("Failed to send suspension email to %s: %v", email, err)
		} else {
			log.Printf("Suspension email sent successfully to %s", email)
		}
	}(merchantEmail, ownerName, req.Reason)

	jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{
		Success: true,
		Message: "Merchant suspended successfully",
	})
}

// ApproveMerchantDocumentsHandler approves newly uploaded documents for an already active merchant
func (h *Handler) ApproveMerchantDocumentsHandler(w http.ResponseWriter, r *http.Request) {
	merchantID := r.PathValue("id")
	if merchantID == "" {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{Success: false, Message: "Merchant ID is required"})
		return
	}

	_, err := h.Store.Exec("UPDATE merchants SET document_status = 'approved', message = 'Your newly uploaded documents have been approved.' WHERE merchant_id = ?", merchantID)
	if err != nil {
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{Success: false, Message: "Database error"})
		return
	}

	jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{Success: true, Message: "Documents approved successfully"})
}

func (h *Handler) DeleteMerchantHandler(w http.ResponseWriter, r *http.Request) {
	merchantID := r.PathValue("id")
	if merchantID == "" {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{Success: false, Message: "Merchant ID is required"})
		return
	}

	tx, err := h.Store.Begin()
	if err != nil {
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{Success: false, Message: "Database error"})
		return
	}
	defer tx.Rollback()

	var merchantUserID, ownerName, merchantEmail string
	err = tx.QueryRow("SELECT user_id, owner_name, business_email FROM merchants WHERE merchant_id = ?", merchantID).Scan(&merchantUserID, &ownerName, &merchantEmail)
	if err != nil {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{Success: false, Message: "Merchant not found"})
		return
	}

	// Update terminals assigned to this merchant
	_, err = tx.Exec("UPDATE terminals SET merchant_id = NULL, location_details = '', status = 'inactive' WHERE merchant_id = ?", merchantID)
	if err != nil {
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{Success: false, Message: "Failed to reset terminals"})
		return
	}

	// Soft delete from merchants: Set status to 'deleted' and update message
	_, err = tx.Exec("UPDATE merchants SET status = 'deleted', message = 'Your merchant account has been permanently deleted.' WHERE merchant_id = ?", merchantID)
	if err != nil {
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{Success: false, Message: "Failed to delete merchant"})
		return
	}

	// Soft delete from users: Set status to 'inactive'
	_, err = tx.Exec("UPDATE users SET status = 'inactive' WHERE user_id = ?", merchantUserID)
	if err != nil {
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{Success: false, Message: "Failed to delete user"})
		return
	}

	if err := tx.Commit(); err != nil {
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{Success: false, Message: "Failed to finalize deletion"})
		return
	}

	// Send deletion email to merchant
	go func(email, name string) {
		smtpHost := os.Getenv("SMTP_HOST")
		smtpPort := 587
		smtpEmail := os.Getenv("SMTP_EMAIL")
		smtpSender := os.Getenv("SMTP_SENDER")
		smtpPass := os.Getenv("SMTP_PASSWORD")
		if smtpHost == "" || smtpEmail == "" {
			log.Println("SMTP credentials not configured, skipping deletion email")
			return
		}

		m := gomail.NewMessage()
		m.SetHeader("From", fmt.Sprintf("%s <%s>", smtpSender, smtpEmail))
		m.SetHeader("To", email)
		m.SetHeader("Subject", "Unicard Account Deleted")

		htmlBody := fmt.Sprintf(smtp.MerchantDeletedEmail(), name)
		m.SetBody("text/html", htmlBody)

		d := gomail.NewDialer(smtpHost, smtpPort, smtpEmail, smtpPass)
		if err := d.DialAndSend(m); err != nil {
			log.Printf("Failed to send deletion email to %s: %v", email, err)
		} else {
			log.Printf("Deletion email sent successfully to %s", email)
		}
	}(merchantEmail, ownerName)

	jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{Success: true, Message: "Merchant deleted successfully"})
}

type MerchantDetailsData struct {
	MerchantID       string
	UserID           string
	BusinessName     string
	BusinessType     string
	RegistrationNum  string
	BusinessAddress  string
	City             string
	PostalCode       string
	OwnerName        string
	BusinessEmail    string
	BusinessPhone    string
	Status           string
	CommissionRate   float64
	SettlementBank   string
	SettlementName   string
	SettlementAcct   string
	CreatedAt        string
	BusinessDocument string
	BirDocument      string
	ValidId          string
	DocumentStatus   string
	Terminals        []structs.Terminal
}

type MerchantInfoResponse struct {
	Merchant MerchantDetailsData `json:"merchant"`
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

	// ADDED: Nullable types for address, phone, and type to prevent NULL crashes
	var setBank, setName, setAcct, regNum, dtiDoc, birDoc, otherDoc, city, postal, docStatus, busAddress, busPhone, busType sql.NullString

	err := h.Store.QueryRow(`
		SELECT merchant_id, user_id, business_name, business_type, business_registration_number, 
		       business_address, city, postal_code, owner_name, business_email, business_phone, status, 
		       commission_rate, settlement_bank_name, settlement_account_name, 
		       settlement_account_number, created_at,
		       business_document, bir_document, valid_id, document_status
		FROM merchants WHERE merchant_id = ?`, merchantID).Scan(
		&m.MerchantID, &m.UserID, &m.BusinessName, &busType, &regNum,
		&busAddress, &city, &postal, &m.OwnerName, &m.BusinessEmail, &busPhone, &m.Status,
		&commRate, &setBank, &setName, &setAcct, &m.CreatedAt,
		&dtiDoc, &birDoc, &otherDoc, &docStatus,
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

	// Safely map the nullable strings back to the main struct
	if busType.Valid {
		m.BusinessType = busType.String
	}
	if busAddress.Valid {
		m.BusinessAddress = busAddress.String
	}
	if busPhone.Valid {
		m.BusinessPhone = busPhone.String
	}
	if regNum.Valid {
		m.RegistrationNum = regNum.String
	}
	if city.Valid {
		m.City = city.String
	}
	if postal.Valid {
		m.PostalCode = postal.String
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
		m.BusinessDocument = dtiDoc.String
	}
	if birDoc.Valid {
		m.BirDocument = birDoc.String
	}
	if otherDoc.Valid {
		m.ValidId = otherDoc.String
	}
	if docStatus.Valid {
		m.DocumentStatus = docStatus.String
	}

	termQuery := `
		SELECT terminal_id, terminal_sn, device_name, status
		FROM terminals
		WHERE merchant_id = ?
	`

	termRows, err := h.Store.Query(termQuery, merchantID)
	if err != nil {
		log.Println("Error querying merchant terminals:", err)
		m.Terminals = []structs.Terminal{}
	} else {
		defer termRows.Close()
		var terminals []structs.Terminal
		for termRows.Next() {
			var t structs.Terminal
			if err := termRows.Scan(&t.TerminalID, &t.TerminalSN, &t.DeviceName, &t.Status); err != nil {
				log.Println("Error scanning merchant terminal:", err)
				continue
			}
			terminals = append(terminals, t)
		}
		m.Terminals = terminals
	}

	jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{
		Success: true,
		Data:    MerchantInfoResponse{Merchant: m},
	})
}
