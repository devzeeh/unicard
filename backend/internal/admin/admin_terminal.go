package admin

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"
	jsonwrite "unicard-go/backend/internal/pkg/handler"
	structs "unicard-go/backend/internal/pkg/structs"
)

func (h *Handler) TerminalRegistryView(w http.ResponseWriter, r *http.Request) {
	log.Println("TerminalRegistryView running...")
	data := AdminPageData{
		Page:     "terminals",
		Username: r.PathValue("username"),
	}
	h.Tpl.ExecuteTemplate(w, "admin_terminal.html", data)
}

func (h *Handler) TerminalRegistryDataHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("TerminalRegistryDataHandler running...")

	// Parse pagination parameters
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")
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

	// Build query
	baseQuery := `FROM terminals t LEFT JOIN merchants m ON t.merchant_id = m.user_id`
	var args []interface{}
	var conditions []string

	if search != "" {
		conditions = append(conditions, `(t.terminal_id LIKE ? OR t.terminal_sn LIKE ? OR m.business_name LIKE ?)`)
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
		log.Println("Error counting terminals:", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Error counting terminals",
		})
		return
	}

	// Get paginated data
	orderClause := " ORDER BY t.created_at DESC"
	query := `SELECT t.terminal_id, t.terminal_sn, COALESCE(m.business_name, 'Unassigned / Inventory'), t.device_name, COALESCE(t.location_details, 'Not Set'), t.status ` +
		baseQuery + whereClause + orderClause + ` LIMIT ? OFFSET ?`

	args = append(args, limit, offset)

	rows, err := h.DB.Query(query, args...)
	if err != nil {
		log.Println("Error querying terminals:", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Error querying terminals",
		})
		return
	}
	defer rows.Close()

	var terminals []structs.Terminal
	for rows.Next() {
		var t structs.Terminal
		if err := rows.Scan(&t.TerminalID, &t.TerminalSN, &t.AssignedMerch, &t.DeviceName, &t.LocationDetails, &t.Status); err != nil {
			log.Println("Error scanning terminal:", err)
			continue
		}
		terminals = append(terminals, t)
	}
	if terminals == nil {
		terminals = []structs.Terminal{}
	}

	type PaginatedTerminalResponse struct {
		Terminals  []structs.Terminal `json:"terminals"`
		TotalItems int                `json:"totalItems"`
		Page       int                `json:"page"`
		Limit      int                `json:"limit"`
	}

	terminalData := PaginatedTerminalResponse{
		Terminals:  terminals,
		TotalItems: totalItems,
		Page:       page,
		Limit:      limit,
	}

	jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{
		Success: true,
		Message: "Terminals retrieved successfully",
		Data:    terminalData,
	})
	log.Println("TerminalRegistryDataHandler finished")
}

// AddTerminalRequest payload
type AddTerminalRequest struct {
	TerminalSN string `json:"terminalSn" validate:"required"`
	DeviceName string `json:"deviceName" validate:"required"`
}

// AddTerminalHandler registers a new standalone terminal
func (h *Handler) AddTerminalHandler(w http.ResponseWriter, r *http.Request) {
	var req AddTerminalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "Invalid JSON payload",
		})
		return
	}

	if req.TerminalSN == "" || req.DeviceName == "" {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "Terminal SN and Device Name are required",
		})
		return
	}

	// Generate a unique terminal ID
	timestamp := time.Now().Format("01020605") // MMDDYYss
	nTerminal, _ := rand.Int(rand.Reader, big.NewInt(10000)) // max 9999
	terminalID := fmt.Sprintf("TRM-%s%04d", timestamp, nTerminal.Int64())

	// Insert into DB with NULL merchant_id
	query := `INSERT INTO terminals (terminal_id, terminal_sn, merchant_id, device_name, status) VALUES (?, ?, NULL, ?, 'offline')`
	_, err := h.DB.Exec(query, terminalID, req.TerminalSN, req.DeviceName)
	if err != nil {
		log.Printf("Error inserting standalone terminal: %v", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Failed to register terminal (SN might already exist)",
		})
		return
	}

	jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{
		Success: true,
		Message: "Terminal registered to inventory successfully",
	})
}

type UnassignedTerminalData struct {
	TerminalSN string `json:"terminal_sn"`
	DeviceName string `json:"device_name"`
}

func (h *Handler) GetUnassignedTerminalsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query(`
		SELECT terminal_sn, device_name 
		FROM terminals 
		WHERE merchant_id IS NULL AND status = 'active'
	`)
	if err != nil {
		log.Printf("Error fetching unassigned terminals: %v", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Database error",
		})
		return
	}
	defer rows.Close()

	var terminals []UnassignedTerminalData
	for rows.Next() {
		var t UnassignedTerminalData
		if err := rows.Scan(&t.TerminalSN, &t.DeviceName); err != nil {
			log.Printf("Row scan error: %v", err)
			continue
		}
		terminals = append(terminals, t)
	}

	jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{
		Success: true,
		Data:    terminals,
	})
}
