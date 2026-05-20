package admin

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// DeactivatePageData represents the state passed to/from the deactivation view
type DeactivatePageData struct {
	CardNumber string
	CardHolder string
	CardType   string
	Error      string
	Success    string
}

// This function renders the deactivateCard.html template when the admin visits the /admin/deactivatecard page.
func (h *Handler) DeactivateView(w http.ResponseWriter, r *http.Request) {
	fmt.Println("DeactivateView running...")
	cardNumber := r.URL.Query().Get("cardNumber")
	name := r.URL.Query().Get("name")
	cardType := r.URL.Query().Get("cardType")

	// Render the deactivateCard.html template
	h.Tpl.ExecuteTemplate(w, "deactivateCard.html", DeactivatePageData{
		CardNumber: cardNumber,
		CardHolder: name,
		CardType:   cardType,
	})
}

// This function handles the form submission from the deactivateCard.html page.
func (h *Handler) DeactivateCardHanlder(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Deactivate card handler running...")

	if err := r.ParseForm(); err != nil {
		h.Tpl.ExecuteTemplate(w, "deactivateCard.html", DeactivatePageData{Error: "Failed to parse form"})
		return
	}

	cardNumber := strings.TrimSpace(r.PostFormValue("cardNumber"))
	cardHolder := strings.TrimSpace(r.PostFormValue("name"))
	cardType := strings.TrimSpace(r.PostFormValue("cardType"))
	redirectURL := strings.TrimSpace(r.PostFormValue("redirect"))

	if cardNumber == "" || cardHolder == "" || cardType == "" {
		fmt.Println("Missing required fields:", cardNumber, cardHolder, cardType)
		errMsg := "Please fill in all required fields."
		if redirectURL != "" {
			http.Redirect(w, r, redirectURL+"?error="+url.QueryEscape(errMsg), http.StatusSeeOther)
			return
		}
		h.Tpl.ExecuteTemplate(w, "deactivateCard.html", DeactivatePageData{
			CardNumber: cardNumber,
			CardHolder: cardHolder,
			CardType:   cardType,
			Error:      errMsg,
		})
		return
	}

	// Check if the card exists and is active, then deactivate it
	ok, err := h.deactivateCardIfActive(cardNumber, cardHolder, cardType)
	if err != nil {
		fmt.Println("Error while deactivating card:", err)
		errMsg := "Failed to deactivate card."
		if redirectURL != "" {
			http.Redirect(w, r, redirectURL+"?error="+url.QueryEscape(errMsg), http.StatusSeeOther)
			return
		}
		h.Tpl.ExecuteTemplate(w, "deactivateCard.html", DeactivatePageData{
			CardNumber: cardNumber,
			CardHolder: cardHolder,
			CardType:   cardType,
			Error:      errMsg,
		})
		return
	}

	if !ok {
		fmt.Printf("Card not found or already inactive: %s with Card Type: %s\n", cardNumber, cardType)
		errMsg := "Card not found or already inactive."
		if redirectURL != "" {
			http.Redirect(w, r, redirectURL+"?error="+url.QueryEscape(errMsg), http.StatusSeeOther)
			return
		}
		h.Tpl.ExecuteTemplate(w, "deactivateCard.html", DeactivatePageData{
			CardNumber: cardNumber,
			CardHolder: cardHolder,
			CardType:   cardType,
			Error:      errMsg,
		})
		return
	}

	fmt.Println("Card deactivated successfully:", cardNumber, cardType)
	successMsg := "Card deactivated successfully!"
	if redirectURL != "" {
		http.Redirect(w, r, redirectURL+"?success="+url.QueryEscape(successMsg), http.StatusSeeOther)
		return
	}
	h.Tpl.ExecuteTemplate(w, "deactivateCard.html", DeactivatePageData{
		Success: successMsg,
	})
}

// --- Helper functions ---

// This function checks if a card with the given number and type is active,
// and if so, it deactivates it by updating its status in the database.
func (h *Handler) deactivateCardIfActive(cardNumber, cardHolder, cardType string) (bool, error) {
	result, err := h.DB.Exec(`
		UPDATE cards
		SET status = 'Blocked'
		WHERE card_number = ?
		AND card_holder = ? 
		AND card_type = ?
		AND status = 'Active'
	`, cardNumber, cardHolder, cardType)
	if err != nil {
		return false, err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return rows > 0, nil
}
