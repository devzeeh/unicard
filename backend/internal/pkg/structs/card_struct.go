package structs

type CardData struct {
	CardUID    string  `json:"card_uid" db:"card_uid" validate:"required"`
	CardNumber string  `json:"cardNumber" db:"card_number" validate:"required"`
	CardHolder string  `json:"cardHolder" db:"user_id" validate:"required"`
	CardType   string  `json:"cardType" db:"card_type" validate:"required"`
	Balance    float64 `json:"initial_amount" db:"balance" validate:"required,min=0"`
}
