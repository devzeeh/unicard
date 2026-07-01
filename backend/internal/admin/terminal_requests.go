package admin

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	jsonwrite "unicard-go/backend/internal/pkg/handler"
)

type TerminalRequest struct {
	ID           int        `json:"id"`
	RequestID    string     `json:"request_id"`
	MerchantID   string     `json:"merchant_id"`
	TerminalSN   *string    `json:"terminal_sn"`
	Status       string     `json:"status"`
	RequestedAt  time.Time  `json:"requested_at"`
	HandledBy    *string    `json:"handled_by"`
	HandledAt    *time.Time `json:"handled_at"`
	Notes        *string    `json:"notes"`
	BusinessName string     `json:"business_name"`
	OwnerName    string     `json:"owner_name"`
	DeviceName   *string    `json:"device_name"`
}

type TerminalRequestsResponse struct {
	Success     bool        `json:"success"`
	Message     string      `json:"message"`
	Data        interface{} `json:"data"`
	TotalItems  int         `json:"total_items,omitempty"`
	CurrentPage int         `json:"current_page,omitempty"`
	TotalPages  int         `json:"total_pages,omitempty"`
}

type ApproveTerminalRequestPayload struct {
	AssignTerminalSN string `json:"assign_terminal_sn"`
	Notes            string `json:"notes"`
}

type RejectTerminalRequestPayload struct {
	Reason string `json:"reason"`
}

func (h *Handler) TerminalRequestsView(w http.ResponseWriter, r *http.Request) {
	log.Println("TerminalRequestsView running...")
	data := AdminPageData{
		Page:     "terminal_requests",
		Username: r.PathValue("username"),
	}
	h.Tpl.ExecuteTemplate(w, "terminal_requests.html", data)
}

func (h *Handler) TerminalRequestsDataHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("TerminalRequestsDataHandler running...")

	// Parse pagination parameters
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")
	status := r.URL.Query().Get("status")
	search := r.URL.Query().Get("search")

	page := 1
	limit := 10
	if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
		page = p
	}
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		limit = l
	}

	offset := (page - 1) * limit

	// Check if table exists first
	checkTableQuery := `SELECT 1 FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = 'unicard' AND TABLE_NAME = 'terminal_requests' LIMIT 1`
	var tableExists int
	err := h.Store.QueryRow(checkTableQuery).Scan(&tableExists)
	if err != nil {
		log.Printf("Table check error: %v", err)
		// Return empty list if table doesn't exist
		response := TerminalRequestsResponse{
			Success:     true,
			Message:     "No terminal requests yet",
			Data:        []TerminalRequest{},
			TotalItems:  0,
			CurrentPage: page,
			TotalPages:  1,
		}
		jsonwrite.WriteJSON(w, http.StatusOK, response)
		return
	}

	// Build query
	baseQuery := `FROM terminal_requests tr
		JOIN merchants m ON tr.merchant_id = m.merchant_id
		LEFT JOIN terminals t ON tr.terminal_sn = t.terminal_sn`

	var args []interface{}
	var conditions []string

	if status != "" {
		conditions = append(conditions, `tr.status = ?`)
		args = append(args, status)
	}

	if search != "" {
		conditions = append(conditions, `(tr.request_id LIKE ? OR tr.merchant_id LIKE ? OR m.business_name LIKE ? OR tr.terminal_sn LIKE ?)`)
		searchPattern := "%" + search + "%"
		args = append(args, searchPattern, searchPattern, searchPattern, searchPattern)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total items
	countQuery := `SELECT COUNT(*) ` + baseQuery + whereClause
	var totalItems int
	if err := h.Store.QueryRow(countQuery, args...).Scan(&totalItems); err != nil {
		log.Printf("Count query error: %v | Query: %s | Args: %v", err, countQuery, args)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Failed to count terminal requests: %v", err),
		})
		return
	}

	// Build data query with pagination
	dataQuery := `SELECT 
		tr.id, tr.request_id, tr.merchant_id, tr.terminal_sn, tr.status, 
		tr.requested_at, tr.handled_by, tr.handled_at, tr.notes,
		m.business_name, m.owner_name, t.device_name
		` + baseQuery + whereClause + `
		ORDER BY tr.requested_at DESC
		LIMIT ? OFFSET ?`

	args = append(args, limit, offset)

	rows, err := h.Store.Query(dataQuery, args...)
	if err != nil {
		log.Printf("Data query error: %v | Query: %s | Args: %v", err, dataQuery, args)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Failed to retrieve terminal requests: %v", err),
		})
		return
	}
	defer rows.Close()

	var requests []TerminalRequest
	for rows.Next() {
		var tr TerminalRequest
		// when parseTime is not enabled in DSN, TIMESTAMP/DATETIME come as []byte
		var requestedRaw []byte
		var handledByNull sql.NullString
		var handledAtRaw []byte
		var notesNull sql.NullString

		if err := rows.Scan(
			&tr.ID, &tr.RequestID, &tr.MerchantID, &tr.TerminalSN, &tr.Status,
			&requestedRaw, &handledByNull, &handledAtRaw, &notesNull,
			&tr.BusinessName, &tr.OwnerName, &tr.DeviceName,
		); err != nil {
			log.Printf("Scan error: %v", err)
			jsonwrite.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": fmt.Sprintf("Failed to parse terminal request data: %v", err),
			})
			return
		}

		// parse requested_at
		if len(requestedRaw) > 0 {
			if t, err := time.Parse(time.RFC3339, string(requestedRaw)); err == nil {
				tr.RequestedAt = t
			} else if t, err := time.ParseInLocation("2006-01-02 15:04:05", string(requestedRaw), time.Local); err == nil {
				tr.RequestedAt = t
			} else {
				log.Printf("time parse requested_at error: %v raw=%s", err, string(requestedRaw))
			}
		}

		// handled_by
		if handledByNull.Valid {
			s := handledByNull.String
			tr.HandledBy = &s
		} else {
			tr.HandledBy = nil
		}

		// parse handled_at
		if len(handledAtRaw) > 0 {
			if t, err := time.Parse(time.RFC3339, string(handledAtRaw)); err == nil {
				tr.HandledAt = &t
			} else if t, err := time.ParseInLocation("2006-01-02 15:04:05", string(handledAtRaw), time.Local); err == nil {
				tr.HandledAt = &t
			} else {
				log.Printf("time parse handled_at error: %v raw=%s", err, string(handledAtRaw))
			}
		}

		// notes
		if notesNull.Valid {
			s := notesNull.String
			tr.Notes = &s
		} else {
			tr.Notes = nil
		}

		requests = append(requests, tr)
	}

	totalPages := (totalItems + limit - 1) / limit

	response := TerminalRequestsResponse{
		Success:     true,
		Message:     "Terminal requests retrieved successfully",
		Data:        requests,
		TotalItems:  totalItems,
		CurrentPage: page,
		TotalPages:  totalPages,
	}

	jsonwrite.WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) ApproveTerminalRequestHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("ApproveTerminalRequestHandler running...")
	adminUserID := r.PathValue("username")
	requestID := r.PathValue("id")

	if requestID == "" {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Request ID is required",
		})
		return
	}

	var payload ApproveTerminalRequestPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Invalid request payload",
		})
		return
	}

	// Start transaction
	tx, err := h.Store.Begin()
	if err != nil {
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": "Failed to start transaction",
		})
		return
	}
	defer tx.Rollback()

	// Get terminal request details
	var merchantID string
	var terminalSN *string
	var currentStatus string

	err = tx.QueryRow(
		`SELECT merchant_id, terminal_sn, status FROM terminal_requests WHERE request_id = ?`,
		requestID,
	).Scan(&merchantID, &terminalSN, &currentStatus)

	if err != nil {
		if err.Error() == "sql: no rows" {
			jsonwrite.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
				"success": false,
				"message": "Terminal request not found",
			})
			return
		}
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": "Failed to retrieve terminal request",
		})
		return
	}

	if currentStatus != "pending" {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Only pending requests can be approved",
		})
		return
	}

	// Determine terminal to assign
	assignTerminalSN := payload.AssignTerminalSN
	if assignTerminalSN == "" && terminalSN != nil {
		assignTerminalSN = *terminalSN
	}

	if assignTerminalSN == "" {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "A terminal must be assigned to approve this request",
		})
		return
	}

	// Verify terminal exists and is unassigned (unless it's being reassigned)
	if assignTerminalSN != "" {
		var existingMerchantID *string
		err := tx.QueryRow(
			`SELECT merchant_id FROM terminals WHERE terminal_sn = ?`,
			assignTerminalSN,
		).Scan(&existingMerchantID)

		if err != nil && err.Error() != "sql: no rows" {
			jsonwrite.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "Failed to verify terminal",
			})
			return
		}

		if err == nil && existingMerchantID != nil && *existingMerchantID != merchantID {
			jsonwrite.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": "Terminal is already assigned to another merchant",
			})
			return
		}
	}

	// Update terminal request status
	_, err = tx.Exec(
		`UPDATE terminal_requests SET status = 'approved', handled_by = ?, handled_at = CURRENT_TIMESTAMP WHERE request_id = ?`,
		adminUserID,
		requestID,
	)

	if err != nil {
		log.Printf("Error updating terminal request: %v", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": "Failed to approve terminal request",
		})
		return
	}

	// Assign terminal to merchant if specified
	if assignTerminalSN != "" {
		// Get merchant address for location details
		var businessAddress, city string
		err := tx.QueryRow(`SELECT business_address, city FROM merchants WHERE merchant_id = ?`, merchantID).Scan(&businessAddress, &city)

		locationDetails := ""
		if err == nil {
			if city != "" {
				locationDetails = fmt.Sprintf("%s, %s", businessAddress, city)
			} else {
				locationDetails = businessAddress
			}
		}

		_, err = tx.Exec(
			`UPDATE terminals SET merchant_id = ?, location_details = ?, status = 'active' WHERE terminal_sn = ?`,
			merchantID,
			locationDetails,
			assignTerminalSN,
		)

		if err != nil {
			log.Printf("Error assigning terminal: %v", err)
			jsonwrite.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "Failed to assign terminal",
			})
			return
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": "Failed to commit transaction",
		})
		return
	}

	jsonwrite.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Terminal request approved successfully",
		"data": map[string]interface{}{
			"request_id":  requestID,
			"terminal_sn": assignTerminalSN,
		},
	})
}

func (h *Handler) RejectTerminalRequestHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("RejectTerminalRequestHandler running...")
	adminUserID := r.PathValue("username")
	requestID := r.PathValue("id")

	if requestID == "" {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Request ID is required",
		})
		return
	}

	var payload RejectTerminalRequestPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Invalid request payload",
		})
		return
	}

	// Check current status
	var currentStatus string
	err := h.Store.QueryRow(
		`SELECT status FROM terminal_requests WHERE request_id = ?`,
		requestID,
	).Scan(&currentStatus)

	if err != nil {
		if err.Error() == "sql: no rows" {
			jsonwrite.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
				"success": false,
				"message": "Terminal request not found",
			})
			return
		}
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": "Failed to retrieve terminal request",
		})
		return
	}

	if currentStatus != "pending" {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Only pending requests can be rejected",
		})
		return
	}

	// Update status to rejected
	reason := payload.Reason
	if reason == "" {
		reason = "Rejected by admin"
	}

	_, err = h.Store.Exec(
		`UPDATE terminal_requests SET status = 'rejected', handled_by = ?, handled_at = CURRENT_TIMESTAMP, notes = ? WHERE request_id = ?`,
		adminUserID,
		reason,
		requestID,
	)

	if err != nil {
		log.Printf("Error rejecting terminal request: %v", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": "Failed to reject terminal request",
		})
		return
	}

	jsonwrite.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Terminal request rejected successfully",
		"data": map[string]interface{}{
			"request_id": requestID,
			"reason":     reason,
		},
	})
}
