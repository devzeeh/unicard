package adminauth

import (
	"fmt"
	"net/http"
	"strings"
	structMessage "unicard-go/internal/pkg"
)

// This struct represents the details of a card that we want to deactivate.
// We can use it to easily pass card data around in our functions.
// It also helps to keep our code organized and makes it easier to manage card-related data when we process the deactivation form submission.
type CardDetails struct {
	CardNumber string
	CardType   string
}

// This function renders the deactivateCard.html template when the admin visits the /admin/deactivatecard page.
// It doesn't do any processing yet, it just shows the form to the admin.
// We can also pass an empty   structMessage.MessageData struct to the template, which allows us to easily display error or success messages later on when we process the form submission.
func (h *Handler) DeactivateView(w http.ResponseWriter, r *http.Request) {
	fmt.Println("DeactivateView running...")
	// Render the deactivateCard.html template
	h.Tpl.ExecuteTemplate(w, "deactivateCard.html",  structMessage.MessageData{})
}

// This function handles the form submission from the deactivateCard.html page.
func (h *Handler) DeactivateCardHanlder(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Deactivate card handler running...")

	if err := r.ParseForm(); err != nil {
		h.Tpl.ExecuteTemplate(w, "deactivateCard.html",  structMessage.MessageData{Error: "Failed to parse form"})
		return
	}

	cardNumber := strings.TrimSpace(r.PostFormValue("cardNumber"))
	cardHolder := strings.TrimSpace(r.PostFormValue("name"))
	cardType := strings.TrimSpace(r.PostFormValue("cardType"))

	if cardNumber == "" || cardHolder == "" || cardType == "" {
		fmt.Println("Missing required fields:", cardNumber, cardHolder, cardType)
		h.Tpl.ExecuteTemplate(w, "deactivateCard.html",  structMessage.MessageData{Error: "Please fill in all required fields."})
		return
	}

	// Check if the card exists and is active, then deactivate it
	ok, err := h.deactivateCardIfActive(cardNumber, cardHolder, cardType)
	if err != nil {
		fmt.Println("Error while deactivating card:", err)
		h.Tpl.ExecuteTemplate(w, "deactivateCard.html",  structMessage.MessageData{Error: "Failed to deactivate card."})
		return
	}

	if !ok {
		fmt.Printf("Card not found or already inactive: %s with Card Type: %s\n", cardNumber, cardType)
		h.Tpl.ExecuteTemplate(w, "deactivateCard.html",  structMessage.MessageData{Error: "Card not found or already inactive."})
		return
	}

	fmt.Println("Card deactivated successfully:", cardNumber, cardType)
	h.Tpl.ExecuteTemplate(w, "deactivateCard.html",  structMessage.MessageData{Success: "Card deactivated successfully!"})
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
