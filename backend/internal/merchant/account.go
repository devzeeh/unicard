package merchant

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	jsonwrite "unicard-go/backend/internal/pkg/handler"
)

// Struct to hold data for the merchant account page
type BusinessDetails struct {
	BusinessName    string `json:"business_name"`
	BusinessType    string `json:"business_type"`
	BusinessEmail   string `json:"business_email"`
	BusinessPhone   string `json:"business_phone"`
	BusinessAddress string `json:"business_address"`
	City            string `json:"city"`
	PostalCode      string `json:"postal_code"`
}

type BusinessDocument struct {
	DocumentType      string `json:"document_type"`
	Status            string `json:"document_status"`
	Message           string `json:"message"`
	BusinessStructure string `json:"business_structure"`
	DocumentURL       string `json:"document_url"`
}

type BusinessBankDetails struct {
	AccountHolderName string `json:"account_holder_name"`
	BankName          string `json:"bank_name"`
	AccountNumber     string `json:"account_number"`
}

type AccountSummary struct {
	MerchantID          string              `json:"merchant_id"`
	AccountStatus       string              `json:"account_status"`
	DocumentStatus      string              `json:"document_status"`
	AccountMessage      string              `json:"account_message"`
	MemberSince         string              `json:"member_since"`
	BusinessDetails     BusinessDetails     `json:"business_details"`
	BusinessBankDetails BusinessBankDetails `json:"business_bank_details"`
	BusinessDocuments   []BusinessDocument  `json:"business_document"`
}

type MerchantDetails struct {
	merchantID        string `db:"merchant_id"`
	accountStatus     string `db:"status"`
	businessName      string `db:"business_name"`
	businessType      string `db:"business_type"`
	businessStructure string `db:"business_structure"`
	businessEmail     string `db:"business_email"`
	businessPhone     string `db:"business_phone"`
	businessAddress   string `db:"business_address"`
	city              string `db:"city"`
	postalCode        string `db:"postal_code"`
	accName           string `db:"settlement_account_name"`
	bankName          string `db:"settlement_bank_name"`
	accNumber         string `db:"settlement_account_number"`
	businessDoc       string `db:"business_document,bir_document"`
	birDoc            string `db:"bir_document"`
	otherDoc          string `db:"other_document"`
	docStatus         string `db:"document_status"`
	docMessage        string `db:"message"`
	createdAtStr      string `db:"created_at"`
}

// MerchantAccountView renders the merchant_account.html template
func (h *Handler) MerchantAccountView(w http.ResponseWriter, r *http.Request) {
	log.Println("MerchantAccountView running...")
	data := MerchantPageData{
		Page:     "account",
		Username: r.PathValue("username"),
	}
	err := h.Tpl.ExecuteTemplate(w, "merchant_account.html", data)
	if err != nil {
		log.Println("Error executing template:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (h *Handler) MerchantAccountDataHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("MerchantAccountDataHandler running...")

	ctx := r.Context()
	username := r.PathValue("username")
	if username == "" {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "username is required",
		})
		return
	}

	// Holds the data fetched from the database
	var merchant MerchantDetails

	// Execute the full JOIN query
	err := h.DB.QueryRowContext(ctx, `
        SELECT 
            m.merchant_id, 
            COALESCE(m.status, ''), 
            COALESCE(DATE_FORMAT(m.created_at, '%M %d, %Y'), '') as created_at,
            -- Business Info
            COALESCE(m.business_name, ''), 
            COALESCE(m.business_type, ''), 
            COALESCE(m.business_email, ''), 
            COALESCE(m.business_phone, ''), 
            COALESCE(m.business_address, ''),
            -- Location Info
            COALESCE(m.city, ''),
            COALESCE(m.postal_code, ''), 
            -- Bank Info
            COALESCE(m.settlement_account_name, ''), 
            COALESCE(m.settlement_bank_name, ''), 
            COALESCE(m.settlement_account_number, ''),
            -- Document Info
			COALESCE(m.business_document, ''),
            COALESCE(m.bir_document, ''),
            COALESCE(m.other_document, ''),
            COALESCE(m.document_status, ''),
            COALESCE(m.message, '')
        FROM merchants m
        JOIN users u ON m.user_id = u.user_id
        WHERE u.username = ?
    `, username).Scan(
		&merchant.merchantID, &merchant.accountStatus, &merchant.createdAtStr, &merchant.businessName,
		&merchant.businessType, &merchant.businessEmail, &merchant.businessPhone,
		&merchant.businessAddress, &merchant.city, &merchant.postalCode, &merchant.accName,
		&merchant.bankName, &merchant.accNumber, &merchant.businessDoc,
		&merchant.birDoc, &merchant.otherDoc, &merchant.docStatus, &merchant.docMessage,
	)

	if err != nil {
		log.Println("Error fetching merchant account data:", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Error fetching account profile",
		})
		return
	}

	// Format UI Logic
	memberSince := merchant.createdAtStr

	var maskedAccount string
	if len(merchant.accNumber) > 4 {
		maskedAccount = "**** **** **** " + merchant.accNumber[len(merchant.accNumber)-4:]
	} else if merchant.accNumber != "" {
		maskedAccount = merchant.accNumber
	}

	// Update the dynamic Document Array logic
	documents := []BusinessDocument{}

	// Business Registration (DTI or SEC)
	if merchant.businessDoc != "" {
		registrationLabel := "DTI/SEC Registration"

		documents = append(documents, BusinessDocument{
			DocumentType:      registrationLabel,
			Status:            merchant.docStatus,  // Quotes removed!
			Message:           merchant.docMessage, // Quotes removed!
			BusinessStructure: merchant.businessStructure,
			DocumentURL:       merchant.businessDoc,
		})
	}

	// Tax Registration (BIR)
	if merchant.birDoc != "" {
		documents = append(documents, BusinessDocument{
			DocumentType: "BIR Certificate",
			Status:       merchant.docStatus,
			Message:      merchant.docMessage,
			DocumentURL:  merchant.birDoc,
		})
	}

	// Other Document
	if merchant.otherDoc != "" {
		documents = append(documents, BusinessDocument{
			DocumentType: "Other Document",
			Status:       merchant.docStatus,
			Message:      merchant.docMessage,
			DocumentURL:  merchant.otherDoc,
		})
	}

	// Construct the final struct
	responseData := AccountSummary{
		MerchantID:     merchant.merchantID,
		AccountStatus:  merchant.accountStatus,
		DocumentStatus: merchant.docStatus,
		AccountMessage: merchant.docMessage,
		MemberSince:    memberSince,
		BusinessDetails: BusinessDetails{
			BusinessName:    merchant.businessName,
			BusinessType:    merchant.businessType,
			BusinessEmail:   merchant.businessEmail,
			BusinessPhone:   merchant.businessPhone,
			BusinessAddress: merchant.businessAddress,
			City:            merchant.city,
			PostalCode:      merchant.postalCode,
		},
		BusinessBankDetails: BusinessBankDetails{
			AccountHolderName: merchant.accName,
			BankName:          merchant.bankName,
			AccountNumber:     maskedAccount,
		},
		BusinessDocuments: documents,
	}

	// Send response
	jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{
		Success: true,
		Message: "Account profile retrieved successfully",
		Data:    responseData,
	})
}

func (h *Handler) UpdateBankDetails(w http.ResponseWriter, r *http.Request) {
	log.Println("UpdateBankDetails running...")
	username := r.PathValue("username")
	if username == "" {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{Success: false, Message: "Username required"})
		return
	}

	var req BusinessBankDetails
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{Success: false, Message: "Invalid request payload"})
		return
	}

	if strings.TrimSpace(req.BankName) == "" || strings.TrimSpace(req.AccountHolderName) == "" || strings.TrimSpace(req.AccountNumber) == "" {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{Success: false, Message: "All bank details fields are required"})
		return
	}

	if _, validBank := channelCodeMap[req.BankName]; !validBank {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{Success: false, Message: "Unsupported bank selected. Please choose a valid bank from the list."})
		return
	}

	var merchantID string
	err := h.DB.QueryRow("SELECT merchant_id FROM merchants WHERE user_id = (SELECT user_id FROM users WHERE username=?)", username).Scan(&merchantID)
	if err != nil {
		log.Println("Error finding merchant for update:", err)
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{Success: false, Message: "Merchant not found"})
		return
	}

	_, err = h.DB.Exec("UPDATE merchants SET settlement_bank_name=?, settlement_account_name=?, settlement_account_number=? WHERE merchant_id = ?", req.BankName, req.AccountHolderName, req.AccountNumber, merchantID)

	if err != nil {
		log.Println("Update error:", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{Success: false, Message: "Failed to update bank details"})
		return
	}

	// Insert a system transaction to log the update
	sysTxnID := fmt.Sprintf("SYS-SETTLE-%d", time.Now().UnixMilli())
	_, _ = h.DB.Exec(`
		INSERT INTO transactions 
		(transaction_id, merchant_id, transaction_type, amount, points_earned, service_fee, status, description) 
		VALUES (?, ?, 'payment', NULL, NULL, NULL, 'completed', 'Settlement bank details were updated by the merchant.')`,
		sysTxnID, merchantID)

	jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{Success: true, Message: "Bank details updated"})
}

func (h *Handler) UploadDocument(w http.ResponseWriter, r *http.Request) {
	log.Println("UploadDocument running...")
	username := r.PathValue("username")
	if username == "" {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{Success: false, Message: "Username required"})
		return
	}

	err := r.ParseMultipartForm(4 << 20) // Limit memory to 4MB
	if err != nil {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{Success: false, Message: "File too large"})
		return
	}

	docType := r.FormValue("document_type")
	file, handler, err := r.FormFile("document")
	if err != nil {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{Success: false, Message: "Failed to read file"})
		return
	}
	defer file.Close()

	if handler.Size > 4*1024*1024 {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{Success: false, Message: "File too large. Maximum size is 4MB."})
		return
	}

	ext := strings.ToLower(filepath.Ext(handler.Filename))
	validExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".pdf":  true,
	}
	if !validExts[ext] {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{Success: false, Message: "Invalid file format. Only pictures, PDF, and Word docs are allowed."})
		return
	}

	// Ensure uploads directory exists
	uploadDir := "storage/documents"
	os.MkdirAll(uploadDir, os.ModePerm)

	// Save file with auto-generated filename to eliminate raw filename
	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	filePath := filepath.Join(uploadDir, filename)
	dst, err := os.Create(filePath)
	if err != nil {
		log.Println("File create error:", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{Success: false, Message: "Internal server error"})
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		log.Println("File copy error:", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{Success: false, Message: "Internal server error"})
		return
	}

	dbPath := "/" + strings.ReplaceAll(filePath, "\\", "/")

	col := "business_document"
	switch docType {
	case "BIR Certificate":
		col = "bir_document"
	case "Other Document":
		col = "other_document"
	}

	// Remove old file if it exists
	var oldDbPath *string
	qOld := fmt.Sprintf("SELECT %s FROM merchants WHERE user_id = (SELECT user_id FROM users WHERE username=?)", col)
	if err := h.DB.QueryRow(qOld, username).Scan(&oldDbPath); err == nil && oldDbPath != nil {
		oldFile := strings.TrimPrefix(*oldDbPath, "/")
		if oldFile != "" {
			os.Remove(oldFile) // Best effort delete
		}
	}

	query := fmt.Sprintf("UPDATE merchants SET %s=?, document_status='Pending' WHERE user_id = (SELECT user_id FROM users WHERE username=?)", col)
	_, err = h.DB.Exec(query, dbPath, username)
	if err != nil {
		log.Println("DB update error:", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{Success: false, Message: "Failed to update DB"})
		return
	}

	jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{Success: true, Message: "File uploaded successfully"})
}
