package merchant

import (
	"log"
	"net/http"
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
	Status            string `json:"status"`
	Message           string `json:"message,omitempty"`
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
	MemberSince         string              `json:"member_since"`
	BusinessDetails     BusinessDetails     `json:"business_details"`
	BusinessBankDetails BusinessBankDetails `json:"business_bank_details"`
	BusinessDocuments   []BusinessDocument  `json:"business_document"`
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

	// 1. Holds the data fetched from the database
	var (
		merchantID, accountStatus, businessName, businessType, businessStructure,
		businessEmail, businessPhone, businessAddress, city,
		postalCode, accName, bankName, accNumber, businessDoc, birDoc,
		docStatus, docMessage, createdAtStr string
	)

	// 2. Execute the full JOIN query
	err := h.DB.QueryRowContext(ctx, `
        SELECT 
            m.merchant_id, 
            COALESCE(m.status, ''), 
            COALESCE(DATE_FORMAT(m.created_at, '%M %d, %Y'), '') as created_at,
            
            -- Business Info
            COALESCE(m.business_name, ''), 
            COALESCE(m.business_type, ''), 
            COALESCE(m.business_structure, ''),
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
            COALESCE(m.business_structure, ''),
			COALESCE(m.business_document, ''),
            COALESCE(m.bir_document, ''),
            COALESCE(m.document_status, ''),
            COALESCE(m.message, '')
            
        FROM merchants m
        JOIN users u ON m.user_id = u.user_id
        WHERE u.username = ?
    `, username).Scan(
		&merchantID,
		&accountStatus,
		&createdAtStr,
		&businessName,
		&businessType,
		&businessStructure,
		&businessEmail,
		&businessPhone,
		&businessAddress,
		&city,
		&postalCode,
		&accName,
		&bankName,
		&accNumber,
		&businessStructure,
		&businessDoc,
		&birDoc,
		&docStatus,
		&docMessage,
	)

	if err != nil {
		log.Println("Error fetching merchant account data:", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Error fetching account profile",
		})
		return
	}

	// 3. Format UI Logic
	memberSince := createdAtStr

	var maskedAccount string
	if len(accNumber) > 4 {
		maskedAccount = "**** **** **** " + accNumber[len(accNumber)-4:]
	} else if accNumber != "" {
		maskedAccount = accNumber
	}

	// 4. Update the dynamic Document Array logic
	documents := []BusinessDocument{}

	// Business Registration (DTI or SEC)
	if businessDoc != "" {
		registrationLabel := "DTI Registration"

		// NOW checking the correct column!
		if businessStructure == "corporation" || businessStructure == "partnership" {
			registrationLabel = "SEC Registration"
		}

		documents = append(documents, BusinessDocument{
			DocumentType:      registrationLabel,
			Status:            docStatus,  // Quotes removed!
			Message:           docMessage, // Quotes removed!
			BusinessStructure: businessStructure,
			DocumentURL:       businessDoc,
		})
	}

	// Tax Registration (BIR)
	if birDoc != "" {
		documents = append(documents, BusinessDocument{
			DocumentType: "BIR Certificate",
			Status:       docStatus,  // Quotes removed!
			Message:      docMessage, // Quotes removed!
			DocumentURL:  birDoc,
		})
	}

	// 5. Construct the final struct
	responseData := AccountSummary{
		MerchantID:    merchantID,
		AccountStatus: accountStatus,
		MemberSince:   memberSince,

		BusinessDetails: BusinessDetails{
			BusinessName:    businessName,
			BusinessType:    businessType,
			BusinessEmail:   businessEmail,
			BusinessPhone:   businessPhone,
			BusinessAddress: businessAddress,
			City:            city,
			PostalCode:      postalCode,
		},

		BusinessBankDetails: BusinessBankDetails{
			AccountHolderName: accName,
			BankName:          bankName,
			AccountNumber:     maskedAccount,
		},

		BusinessDocuments: documents,
	}

	// 6. Send response
	jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{
		Success: true,
		Message: "Account profile retrieved successfully",
		Data:    responseData,
	})
}
