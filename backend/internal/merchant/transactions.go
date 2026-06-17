package merchant

import (
	"log"
	"net/http"
)

// MerchantTransactionsView renders the merchant_transactions.html template
func (h *Handler) MerchantTransactionsView(w http.ResponseWriter, r *http.Request) {
	log.Println("MerchantTransactionsView running...")
	data := MerchantPageData{
		Page:     "transactions",
		Username: r.PathValue("username"),
	}
	err := h.Tpl.ExecuteTemplate(w, "merchant_transactions.html", data)
	if err != nil {
		log.Println("Error executing template:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
