package structs

// AdminDashboardData struct represents the data to be displayed on the admin dashboard
type AdminDashboardData struct {
	GrossRevenue    float64 `json:"grossRevenue"`
	NetRevenue      float64 `json:"netRevenue"`
	TotalUsers      int     `json:"totalUsers"`
	TotalCards      int     `json:"totalCards"`
	ActiveMerchants int     `json:"activeMerchants"`
	ActiveTerminals int     `json:"activeTerminals"`
}

// CardData struct represents the data required to create a new card
type CardData struct {
	CardUID    string  `json:"card_uid" db:"card_uid" validate:"required"`
	CardNumber string  `json:"cardNumber" db:"card_number" validate:"required"`
	CardHolder string  `json:"cardHolder" db:"user_id" validate:"required"`
	CardType   string  `json:"cardType" db:"card_type" validate:"required"`
	Balance    float64 `json:"initial_amount" db:"balance" validate:"required,min=0"`
}