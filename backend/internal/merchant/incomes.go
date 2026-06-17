package merchant

import (
	"context"
	"log"
	"net/http"
	jsonwrite "unicard-go/backend/internal/pkg/handler"

	"github.com/shopspring/decimal"
)

type IncomeHistory struct {
	Date            string          `json:"date" db:"created_at"`
	Description     *string         `json:"description" db:"description"`
	TransactionID   string          `json:"transaction_id" db:"transaction_id"`
	CardNumber      string          `json:"card_number" db:"card_number"`
	TransactionType string          `json:"transaction_type" db:"transaction_type"`
	Amount          decimal.Decimal `json:"amount" db:"amount"`
	NetIncome       decimal.Decimal `json:"net_income" db:"net_merchant_payout"`
	ServiceFee      decimal.Decimal `json:"service_fee" db:"service_fee"`
	ProcessedBy     *string         `json:"processed_by" db:"processed_by"`
	TerminalID      *string         `json:"terminal_id" db:"terminal_id"`
}

// IncomeStat represents the merchant's income statistics
// TotalCollected  → Total money collected from customers (gross, before Unicard fee)
// UniCardFee      → Unicard's platform cut deducted per transaction
// TotalEarned     → What the merchant receives after Unicard's fee
// TotalRefunded   → Money returned back to customers
// ActualIncome    → What the merchant actually keeps after refunds (real bottom line)
// IncomeStat represents the merchant's income statistics
type IncomeStat struct {
	NetRevenue       decimal.Decimal `json:"net_revenue"`        // What the merchant gets after platform fee
	GrossRevenue     decimal.Decimal `json:"gross_revenue"`      // SUM(amount) payments
	PlatformFee      decimal.Decimal `json:"platform_fee"`       // SUM(service_fee) payments
	TotalRefunds     decimal.Decimal `json:"total_refunds"`      // SUM(amount) refunds all-time
	MonthlyNetIncome decimal.Decimal `json:"monthly_net_income"` // Net income for the current month
	MonthlyRefunds   decimal.Decimal `json:"monthly_refunds"`    // Refunds for the current month
	TotalWithdrawn   decimal.Decimal `json:"total_withdrawn"`    // Total amount withdrawn by the merchant
	AvailableBalance decimal.Decimal `json:"available_balance"`  // What the merchant can withdraw (NetRevenue - TotalWithdrawn)
	MonthlyWithdrawn decimal.Decimal `json:"monthly_withdrawn"`  // Withdrawals for the current month (if needed in the future)
}

type IncomeResponse struct {
	Stats   IncomeStat      `json:"stats"`
	History []IncomeHistory `json:"history"`
}

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

func (h *Handler) GetMerchantIncomeStats(ctx context.Context, merchantID string) (IncomeStat, error) {
	log.Println("GetMerchantIncomeStats running...")

	var stats IncomeStat
	var totalCollected, unicardFee, totalEarned, totalRefunded, earnedThisMonth, 
	refundedThisMonth, totalWithdrawn, withdrawnThisMonth decimal.Decimal

	err := h.DB.QueryRowContext(ctx, `
    SELECT 
        COALESCE(SUM(CASE WHEN transaction_type = 'payment' THEN amount ELSE 0 END), 0),
        COALESCE(SUM(CASE WHEN transaction_type = 'payment' THEN service_fee ELSE 0 END), 0),
        COALESCE(SUM(CASE WHEN transaction_type = 'payment' THEN net_merchant_payout ELSE 0 END), 0),
        COALESCE(SUM(CASE WHEN transaction_type = 'refund' THEN amount ELSE 0 END), 0),
        
        -- Earned this month
        COALESCE(SUM(CASE WHEN transaction_type = 'payment' 
            AND MONTH(created_at) = MONTH(NOW()) 
            AND YEAR(created_at) = YEAR(NOW()) 
            THEN net_merchant_payout ELSE 0 END), 0),
            
        -- Refunded this month
        COALESCE(SUM(CASE WHEN transaction_type = 'refund' 
            AND MONTH(created_at) = MONTH(NOW()) 
            AND YEAR(created_at) = YEAR(NOW()) 
            THEN amount ELSE 0 END), 0),

		-- Calculate all-time withdrawals
            COALESCE(SUM(CASE WHEN transaction_type = 'withdrawal' THEN amount ELSE 0 END), 0),

		-- Monthly withdrawals (if needed in the future)
		 	COALESCE(SUM(CASE WHEN transaction_type = 'withdrawal'
			AND MONTH(created_at) = MONTH(NOW()) 
			AND YEAR(created_at) = YEAR(NOW()) 
			THEN amount ELSE 0 END), 0)
    FROM transactions
    WHERE merchant_id = ? AND status = 'completed'
`, merchantID).Scan(
		&totalCollected,
		&unicardFee,
		&totalEarned,
		&totalRefunded,
		&earnedThisMonth,
		&refundedThisMonth,
		&totalWithdrawn,
		&withdrawnThisMonth,
	)
	if err != nil {
		log.Println("Error fetching income stats:", err)
		return IncomeStat{}, err
	}

	stats.GrossRevenue = totalCollected
	stats.PlatformFee = unicardFee
	stats.NetRevenue = totalEarned.Sub(totalRefunded)
	stats.TotalRefunds = totalRefunded
	stats.MonthlyNetIncome = earnedThisMonth.Sub(refundedThisMonth)
	stats.MonthlyRefunds = refundedThisMonth
	// NEW MATH: The Ledger Balance
	stats.TotalWithdrawn = totalWithdrawn
	stats.AvailableBalance = stats.NetRevenue.Sub(totalWithdrawn)
	stats.MonthlyWithdrawn = withdrawnThisMonth

	return stats, nil
}

func (h *Handler) GetMerchantIncomeHistory(ctx context.Context, merchantID string) ([]IncomeHistory, error) {
	log.Println("GetMerchantIncomeHistory running...")

	rows, err := h.DB.QueryContext(ctx, `
		SELECT 
			created_at, description,
			transaction_id, card_number,
			transaction_type, amount,
			net_merchant_payout, service_fee,
			processed_by,terminal_id
		FROM transactions
		WHERE merchant_id = ? AND status = 'completed'
		ORDER BY created_at DESC LIMIT 15
	`, merchantID)
	if err != nil {
		log.Println("Error fetching income history:", err)
		return nil, err
	}
	defer rows.Close()

	var history []IncomeHistory
	for rows.Next() {
		var row IncomeHistory
		if err := rows.Scan(
			&row.Date,
			&row.Description,
			&row.TransactionID,
			&row.CardNumber,
			&row.TransactionType,
			&row.Amount,
			&row.NetIncome,
			&row.ServiceFee,
			&row.ProcessedBy,
			&row.TerminalID,
		); err != nil {
			log.Println("Error scanning income history row:", err)
			continue
		}
		history = append(history, row)
	}
	return history, nil
}

func (h *Handler) IncomeHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("IncomeHandler running...")

	ctx := r.Context()
	username := r.PathValue("username")
	if username == "" {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "username is required",
		})
		return
	}

	// Resolve merchant_id from username
	var merchantID string
	err := h.DB.QueryRowContext(ctx, `
		SELECT m.merchant_id 
		FROM merchants m
		JOIN users u ON m.user_id = u.user_id
		WHERE u.username = ?
		LIMIT 1
	`, username).Scan(&merchantID)
	if err != nil {
		log.Println("Error resolving merchant ID:", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Error fetching merchant data",
		})
		return
	}

	// Get income stats
	stats, err := h.GetMerchantIncomeStats(ctx, merchantID)
	if err != nil {
		log.Println("Error getting income stats:", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Error fetching income stats",
		})
		return
	}

	// Get income history
	history, err := h.GetMerchantIncomeHistory(ctx, merchantID)
	if err != nil {
		log.Println("Error getting income history:", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Error fetching income history",
		})
		return
	}

	jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{
		Success: true,
		Message: "Income data retrieved successfully",
		Data: IncomeResponse{
			Stats:   stats,
			History: history,
		},
	})
}
