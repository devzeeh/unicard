package merchant

import (
	"context"
	"log"
	"net/http"
	jsonwrite "unicard-go/backend/internal/pkg/handler"

	"github.com/shopspring/decimal"
)

// MerchantDashboardView renders the merchant_dashboard.html template
// It serves the main dashboard page for merchants, which includes an overview of their account and recent transactions.
type MerchantTransaction struct {
	TransactionID     string          `json:"transaction_id" db:"transaction_id"`
	CardNumber        string          `json:"card_number" db:"card_number"`
	MerchantID        *string         `json:"merchant_id" db:"merchant_id"`
	TerminalID        *string         `json:"terminal_id" db:"terminal_id"`
	TransactionType   string          `json:"transaction_type" db:"transaction_type"`
	Amount            decimal.Decimal `json:"amount" db:"transaction_amount"`
	Points            decimal.Decimal `json:"points" db:"points_earned"`
	ServiceFee        decimal.Decimal `json:"service_fee" db:"service_fee"`
	NetMerchantPayout decimal.Decimal `json:"net_merchant_payout" db:"net_merchant_payout"`
	ProcessedBy       *string         `json:"processed_by" db:"processed_by"`
	Description       *string         `json:"description" db:"description"`
	Date              string          `json:"created_at" db:"created_at"`
	Status            string          `json:"status" db:"status"`
}

// MerchantSummary struct defines the structure of the data returned by the dashboard API,
// including account info and recent transactions.
type MerchantSummary struct {
	Username           string                `json:"username"`
	MerchantID         string                `json:"merchant_id"`
	AccountRole        string                `json:"role"`
	AccountStatus      string                `json:"account_status"`
	TotalTransactions  int                   `json:"total_transactions"`
	GrossRevenue       decimal.Decimal       `json:"gross_revenue"`
	TotalRefunds       decimal.Decimal       `json:"total_refunds"`
	NetRevenue         decimal.Decimal       `json:"net_revenue"`
	TotalServiceFee    decimal.Decimal       `json:"total_service_fee"`
	TotalIncome        decimal.Decimal       `json:"total_income"`
	AvailableBalance   decimal.Decimal       `json:"available_balance"`
	MonthlyNetIncome   decimal.Decimal       `json:"monthly_net_income"`
	SettlementBank     *string               `json:"settlement_bank"`
	SettlementAccount  *string               `json:"settlement_account_number"`
	SettlementName     *string               `json:"settlement_account_name"`
	RecentTransactions []MerchantTransaction `json:"recent_transactions"`
}

// MerchantDashboardView renders the merchant_dashboard.html template
func (h *Handler) MerchantDashboardView(w http.ResponseWriter, r *http.Request) {
	log.Println("MerchantDashboardView running...")
	data := MerchantPageData{
		Page:     "dashboard",
		Username: r.PathValue("username"),
	}
	err := h.Tpl.ExecuteTemplate(w, "merchant_dashboard.html", data)
	if err != nil {
		log.Println("Error executing template:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (h *Handler) GetMerchantAccountInfo(ctx context.Context, merchantID string) (string, string, error) {
	log.Println("GetMerchantAccountInfo running...")

	var accountStatus, accountRole string
	err := h.Store.QueryRowContext(ctx, `
		SELECT u.status, u.role
		FROM users u
		JOIN merchants m ON u.user_id = m.user_id
		WHERE m.merchant_id = ? LIMIT 1`,
		merchantID).Scan(&accountStatus, &accountRole)

	if err != nil {
		log.Println("Error fetching merchant account info:", err)
		return "", "", err
	}
	return accountStatus, accountRole, nil
}

func (h *Handler) GetMerchantRecentTransactions(ctx context.Context, merchantID string) ([]MerchantTransaction, error) {
	log.Println("GetMerchantRecentTransactions running...")

	rows, err := h.Store.QueryContext(ctx, `
		SELECT * FROM (
			SELECT
				transaction_id, COALESCE(card_number, '') AS card_number,
				merchant_id, COALESCE(terminal_id, '') AS terminal_id,
				COALESCE(transaction_type, '') AS transaction_type, COALESCE(amount, 0) AS amount,
				COALESCE(points_earned, 0) AS points_earned, COALESCE(service_fee, 0) AS service_fee,
				COALESCE(net_merchant_payout, 0) AS net_merchant_payout, COALESCE(processed_by, '') AS processed_by,
				COALESCE(status, '') AS status, COALESCE(description, '') AS description, COALESCE(created_at, '') AS created_at
			FROM transactions
			WHERE merchant_id = ?
			UNION ALL
			SELECT 
				CONCAT('LOG-', ual.id) AS transaction_id, '' AS card_number,
				m.merchant_id AS merchant_id, '' AS terminal_id,
				ual.activity_type AS transaction_type, 0.00 AS amount,
				0 AS points_earned, 0.00 AS service_fee,
				0.00 AS net_merchant_payout, ual.user_id AS processed_by,
				ual.status AS status, COALESCE(ual.description, '') AS description, ual.created_at AS created_at
			FROM user_activity_logs ual
			JOIN merchants m ON ual.user_id = m.user_id
			WHERE m.merchant_id = ?
		) AS combined_txns
		ORDER BY created_at DESC
		LIMIT 10`,
		merchantID, merchantID)
	if err != nil {
		log.Println("Error fetching recent transactions:", err)
		return nil, err
	}
	defer rows.Close()

	transactions := []MerchantTransaction{}
	for rows.Next() {
		var t MerchantTransaction
		if err := rows.Scan(
			&t.TransactionID,
			&t.CardNumber,
			&t.MerchantID,
			&t.TerminalID,
			&t.TransactionType,
			&t.Amount,
			&t.Points,
			&t.ServiceFee,
			&t.NetMerchantPayout,
			&t.ProcessedBy,
			&t.Status,
			&t.Description,
			&t.Date,
		); err != nil {
			log.Println("Error scanning recent transaction:", err)
			continue
		}
		transactions = append(transactions, t)
	}
	return transactions, nil
}

func (h *Handler) MerchantDashboardDataHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("MerchantDashboardDataHandler running...")

	ctx := r.Context()

	username := r.PathValue("username")
	if username == "" {
		log.Println("Username is required")
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "username is required",
		})
		return
	}

	// Resolve merchant_id from username
	var merchantID string
	var settlementBank, settlementAccount *string
	err := h.Store.QueryRowContext(ctx, `
		SELECT m.merchant_id, m.settlement_bank_name, m.settlement_account_number
		FROM merchants m
		JOIN users u ON m.user_id = u.user_id
		WHERE u.username = ?
		LIMIT 1`,
		username).Scan(&merchantID, &settlementBank, &settlementAccount)
	if err != nil {
		log.Println("Error fetching merchant ID:", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Error fetching merchant data",
		})
		return
	}

	// Get account status and role
	accountStatus, accountRole, err := h.GetMerchantAccountInfo(ctx, merchantID)
	if err != nil {
		log.Println("Error fetching merchant account info:", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Error fetching merchant account info",
		})
		return
	}

	// Count ALL transactions regardless of type
	var totalTransactions int
	err = h.Store.QueryRowContext(ctx, `
		SELECT COUNT(*) 
		FROM transactions 
		WHERE merchant_id = ?`,
		merchantID).Scan(&totalTransactions)
	if err != nil {
		log.Println("Error fetching transaction count:", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Error fetching transaction count",
		})
		return
	}

	// Revenue summary â€” payments only
	var totalRevenue, totalServiceFee, totalIncome decimal.Decimal
	err = h.Store.QueryRowContext(ctx, `
		SELECT 
			COALESCE(SUM(amount), 0), 
			COALESCE(SUM(service_fee), 0), 
			COALESCE(SUM(net_merchant_payout), 0)
		FROM transactions 
		WHERE merchant_id = ? 
		AND transaction_type = 'payment' 
		AND status = 'completed'`,
		merchantID).Scan(
		&totalRevenue,
		&totalServiceFee,
		&totalIncome,
	)
	if err != nil {
		log.Println("Error fetching merchant summary:", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Error fetching dashboard summary",
		})
		return
	}

	// Refunds total
	var totalRefunds decimal.Decimal
	err = h.Store.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(amount), 0)
		FROM transactions
		WHERE merchant_id = ?
		AND transaction_type = 'refund'
		AND status = 'completed'`,
		merchantID).Scan(&totalRefunds)
	if err != nil {
		log.Println("Error fetching refunds:", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Error fetching refund summary",
		})
		return
	}

	// Get recent 5 transactions
	recentTransactions, err := h.GetMerchantRecentTransactions(ctx, merchantID)
	if err != nil {
		log.Println("Error fetching recent transactions:", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Error fetching recent transactions",
		})
		return
	}

	stats, err := h.GetMerchantIncomeStats(ctx, merchantID)
	var availableBalance, monthlyNetIncome decimal.Decimal
	if err == nil {
		availableBalance = stats.AvailableBalance
		monthlyNetIncome = stats.MonthlyNetIncome
	} else {
		log.Println("Error fetching available balance for dashboard:", err)
	}

	summary := MerchantSummary{
		Username:           username,
		MerchantID:         merchantID,
		AccountRole:        accountRole,
		AccountStatus:      accountStatus,
		TotalTransactions:  totalTransactions,
		GrossRevenue:       totalRevenue,
		TotalRefunds:       totalRefunds,
		NetRevenue:         totalRevenue.Sub(totalRefunds),
		TotalServiceFee:    totalServiceFee,
		TotalIncome:        totalIncome,
		AvailableBalance:   availableBalance,
		MonthlyNetIncome:   monthlyNetIncome,
		SettlementBank:     settlementBank,
		SettlementAccount:  settlementAccount,
		RecentTransactions: recentTransactions,
	}

	jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{
		Success: true,
		Message: "Dashboard data retrieved successfully",
		Data:    summary,
	})
}
