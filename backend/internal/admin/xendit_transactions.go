package admin

import (
	"log"
	"net/http"
	"time"

	jsonwrite "unicard-go/backend/internal/pkg/handler"
	"unicard-go/backend/internal/pkg/xenditclient"

	"github.com/shopspring/decimal"
)

// XenditTransactionsView serves the dedicated Xendit Transactions page
func (h *Handler) XenditTransactionsView(w http.ResponseWriter, r *http.Request) {
	data := AdminPageData{Page: "xendit-transactions", Username: r.PathValue("username")}
	err := h.Tpl.ExecuteTemplate(w, "xendit_transactions.html", data)
	if err != nil {
		log.Printf("Template execution error: %v", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Internal Server Error",
		})
	}
}

// AllXenditTransactionsJSONHandler returns all transactions directly from Xendit
func (h *Handler) AllXenditTransactionsJSONHandler(w http.ResponseWriter, r *http.Request) {
	type TxnResponse struct {
		TransactionID    string          `json:"transaction_id"`
		Date             string          `json:"date"`
		Time             string          `json:"time"`
		Description      string          `json:"description"`
		Type             string          `json:"type"`
		Amount           decimal.Decimal `json:"amount"`
		Status           string          `json:"status"`
		SettlementStatus string          `json:"settlement_status"`
		SettlementTime   string          `json:"settlement_time"`
	}

	var transactions []TxnResponse

	xenditTx, err := xenditclient.GetAllTransactions()
	if err != nil {
		log.Printf("Error querying xendit transactions: %v\n", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Failed to fetch Xendit transactions",
		})
		return
	}

	// Use Asia/Manila timezone (UTC+8)
	loc := time.FixedZone("PHT", 8*60*60)

	for _, tx := range xenditTx.Data {
		updatedLoc := tx.Updated.In(loc)
		dateStr := updatedLoc.Format("2006-01-02")
		timeStr := updatedLoc.Format("15:04:05")

		desc := string(tx.ChannelCategory)
		if tx.ChannelCode.Get() != nil {
			desc += " - " + *tx.ChannelCode.Get()
		}

		settlementStatus := "N/A"
		if tx.SettlementStatus.Get() != nil {
			settlementStatus = *tx.SettlementStatus.Get()
		}

		settlementTime := "N/A"
		if tx.EstimatedSettlementTime.Get() != nil {
			estLoc := tx.EstimatedSettlementTime.Get().In(loc)
			settlementTime = estLoc.Format("2006-01-02 15:04:05")
		}

		transactions = append(transactions, TxnResponse{
			TransactionID:    tx.ReferenceId,
			Date:             dateStr,
			Time:             timeStr,
			Description:      "Xendit: " + desc,
			Type:             "External",
			Amount:           decimal.NewFromFloat(float64(tx.Amount)),
			Status:           string(tx.Status),
			SettlementStatus: settlementStatus,
			SettlementTime:   settlementTime,
		})
	}

	// Sort transactions by Date and Time descending
	for i := 0; i < len(transactions)-1; i++ {
		for j := i + 1; j < len(transactions); j++ {
			dtI := transactions[i].Date + " " + transactions[i].Time
			dtJ := transactions[j].Date + " " + transactions[j].Time
			if dtI < dtJ {
				// swap
				transactions[i], transactions[j] = transactions[j], transactions[i]
			}
		}
	}

	response := struct {
		Success      bool          `json:"success"`
		Transactions []TxnResponse `json:"transactions"`
	}{
		Success:      true,
		Transactions: transactions,
	}

	jsonwrite.WriteJSON(w, http.StatusOK, response)
}
