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

// set the max file size to 8MB
const maxUploadSize = 8 << 20

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
	validID           string `db:"other_document"` // valid_id
	bankDoc           string `db:"bank_document"`  //
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

// get merchant profile, merchant details, merchant bank details, and merchant documents
// merchantDocuments has a limit of 3, and the order depends on the order of the files uploaded
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
	err := h.Store.QueryRowContext(ctx, `
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
            COALESCE(m.valid_id, ''),
			COALESCE(m.bank_document, ''),
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
		&merchant.birDoc, &merchant.validID, &merchant.bankDoc, &merchant.docStatus, &merchant.docMessage,
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

	// mask bank account number with asterisks except the last 4 digits
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

	// Valid Government ID (PhilHealth, SSS, Pag-IBIG)
	if merchant.validID != "" {
		documents = append(documents, BusinessDocument{
			DocumentType: "Valid Government ID",
			Status:       merchant.docStatus,
			Message:      merchant.docMessage,
			DocumentURL:  merchant.validID,
		})
	}

	if merchant.bankDoc != "" {
		documents = append(documents, BusinessDocument{
			DocumentType: "Bank Document",
			Status:       merchant.docStatus,
			Message:      merchant.docMessage,
			DocumentURL:  merchant.bankDoc,
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

// UpdateBankDetails function will update the bank details of the merchant
// it will update the database and set the document status to pending and the message to "Bank details updated successfully"
// if the bank details are updated successfully
func (h *Handler) UpdateBankDetails(w http.ResponseWriter, r *http.Request) {
	log.Println("UpdateBankDetails running...")
	username := r.PathValue("username")
	if username == "" {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{Success: false, Message: "Username required"})
		return
	}

	// decode the request body into the BusinessBankDetails struct
	var req BusinessBankDetails
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{Success: false, Message: "Invalid request payload"})
		return
	}

	// check if all bank details fields are not empty
	if strings.TrimSpace(req.BankName) == "" || strings.TrimSpace(req.AccountHolderName) == "" || strings.TrimSpace(req.AccountNumber) == "" {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{Success: false, Message: "All bank details fields are required"})
		return
	}

	// check if the channel code of the bank is valid
	if _, validBank := channelCodeMap[req.BankName]; !validBank {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{Success: false, Message: "Unsupported bank selected. Please choose a valid bank from the list."})
		return
	}

	// get the merchant ID and existing account number from the database
	var merchantID string
	var existingAccNumber string
	err := h.Store.QueryRow("SELECT merchant_id, settlement_account_number FROM merchants WHERE user_id = (SELECT user_id FROM users WHERE username=?)", username).Scan(&merchantID, &existingAccNumber)
	if err != nil {
		log.Println("Error finding merchant for update:", err)
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{Success: false, Message: "Merchant not found"})
		return
	}

	// Prevent overwriting with masked account number
	if strings.Contains(req.AccountNumber, "****") {
		req.AccountNumber = existingAccNumber
	}

	_, err = h.Store.Exec("UPDATE merchants SET settlement_bank_name=?, settlement_account_name=?, settlement_account_number=? WHERE merchant_id = ?", req.BankName, req.AccountHolderName, req.AccountNumber, merchantID)

	if err != nil {
		log.Println("Update error:", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{Success: false, Message: "Failed to update bank details"})
		return
	}

	// Insert a system transaction to log the update
	sysTxnID := fmt.Sprintf("SYS-SETTLE-%d", time.Now().UnixMilli())
	_, _ = h.Store.Exec(`
		INSERT INTO transactions 
		(transaction_id, merchant_id, transaction_type, amount, points_earned, service_fee, status, description) 
		VALUES (?, ?, 'payment', NULL, NULL, NULL, 'completed', 'Settlement bank details were updated by the merchant.')`,
		sysTxnID, merchantID)

	jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{Success: true, Message: "Bank details updated"})
}

// UploadDocument function will upload the business documents to the server and update the database
// it will set the document status to pending and the message to "Document uploaded successfully"
// if the document is uploaded successfully
func (h *Handler) UploadDocument(w http.ResponseWriter, r *http.Request) {
	log.Println("UploadDocument running...")
	username := r.PathValue("username")
	if username == "" {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{Success: false, Message: "Username required"})
		return
	}

	err := r.ParseMultipartForm(maxUploadSize)
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

	if handler.Size > maxUploadSize {
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
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false, Message: "Invalid file format. Only pictures, PDF, and Word docs are allowed."})
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

	column := "business_document"
	switch docType {
	case "BIR Certificate":
		column = "bir_document"
	case "Valid ID":
		column = "valid_id"
	case "Bank Document":
		column = "bank_document"
	}

	// Remove old file if it exists
	var oldDbPath *string
	qOld := fmt.Sprintf(`
		SELECT %s FROM merchants
		WHERE user_id = (SELECT user_id FROM users WHERE username=?)`,
		column)
	if err := h.Store.QueryRow(qOld, username).Scan(&oldDbPath); err == nil && oldDbPath != nil {
		oldFile := strings.TrimPrefix(*oldDbPath, "/")
		if oldFile != "" {
			if err := os.Remove(oldFile); err != nil {
				log.Println("Error removing old file:", err)
			}
			log.Println("Old file removed:", oldFile)
		}
	}

	query := fmt.Sprintf(`
		UPDATE merchants SET %s=?, document_status='Pending'
		WHERE user_id = (SELECT user_id FROM users WHERE username=?)`,
		column)
	_, err = h.Store.Exec(query, dbPath, username)
	if err != nil {
		log.Println("DB update error:", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{Success: false, Message: "Failed to update DB"})
		return
	}

	jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{
		Success: true,
		Message: "File uploaded successfully",
	})
}
