package merchant

import (
	"log"
	"net/http"
)

// MerchantIncomesView renders the merchant_incomes.html template
func (h *Handler) MerchantIncomesView(w http.ResponseWriter, r *http.Request) {
	log.Println("MerchantIncomesView running...")
	data := MerchantPageData{
		Page:     "incomes",
		Username: r.PathValue("username"),
	}
	err := h.Tpl.ExecuteTemplate(w, "merchant_incomes.html", data)
	if err != nil {
		log.Println("Error executing template:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
