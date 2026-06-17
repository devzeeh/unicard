package merchant

import (
	"log"
	"net/http"
)

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
