package admin

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// DeleteCardHandler handles deleting a card by card_number
func (h *Handler) DeleteCardHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Delete card handler running...")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/admin/dashboard?error="+url.QueryEscape("Failed to parse form"), http.StatusSeeOther)
		return
	}

	cardNumber := strings.TrimSpace(r.PostFormValue("cardNumber"))
	redirectURL := strings.TrimSpace(r.PostFormValue("redirect"))
	if redirectURL == "" {
		redirectURL = "/admin/dashboard"
	}

	if cardNumber == "" {
		http.Redirect(w, r, redirectURL+"?error="+url.QueryEscape("Card number is required for deletion."), http.StatusSeeOther)
		return
	}

	// Delete from the database
	result, err := h.DB.Exec("DELETE FROM cards WHERE card_number = ?", cardNumber)
	if err != nil {
		fmt.Println("Error deleting card from DB:", err)
		http.Redirect(w, r, redirectURL+"?error="+url.QueryEscape("Failed to delete card due to database error."), http.StatusSeeOther)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		fmt.Println("Error reading rows affected:", err)
		http.Redirect(w, r, redirectURL+"?error="+url.QueryEscape("Failed to confirm card deletion."), http.StatusSeeOther)
		return
	}

	if rowsAffected == 0 {
		fmt.Printf("Card %s not found for deletion\n", cardNumber)
		http.Redirect(w, r, redirectURL+"?error="+url.QueryEscape("Card not found."), http.StatusSeeOther)
		return
	}

	fmt.Println("Card deleted successfully:", cardNumber)
	http.Redirect(w, r, redirectURL+"?success="+url.QueryEscape("Card deleted successfully!"), http.StatusSeeOther)
}
