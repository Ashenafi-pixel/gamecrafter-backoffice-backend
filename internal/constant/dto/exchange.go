package dto

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type ExchangeRes struct {
	ID           uuid.UUID       `json:"id"`
	CurrencyFrom string          `json:"currency_from"`
	CurrencyTo   string          `json:"currency_to"`
	Rate         decimal.Decimal `json:"rate"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

type ExchangeReq struct {
	CurrencyFrom string `json:"currency_from" validate:"required,max=3"`
	CurrencyTo   string `json:"currency_to" validate:"required,max=3"`
}
type ExchangeBalanceReq struct {
	UserID       uuid.UUID       `json:"user_id" swaggerignore:"true"`
	CurrencyFrom string          `json:"currency_from" validate:"required,max=3"`
	CurrencyTo   string          `json:"currency_to" validate:"required,max=3"`
	Amount       decimal.Decimal `json:"amount"`
}
type NewBalance struct {
	Currency string          `json:"currency"`
	Balance  decimal.Decimal `json:"balance"`
}
type ExchangeBalanceResData struct {
	NewFromBalance NewBalance `json:"new_from_balance"`
	NewToBalance   NewBalance `json:"new_to_balance"`
}
type ExchangeBalanceRes struct {
	Status  string                 `json:"status"`
	Message string                 `json:"message"`
	Date    ExchangeBalanceResData `json:"data"`
}

func ValidateExchangeRequest(ex ExchangeReq) error {
	validate := validator.New()
	return validate.Struct(ex)
}

type Currency struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

type GetCurrencyReqData struct {
	TotalPages int        `json:"total_pages"`
	Currencies []Currency `json:"currencies"`
}

type GetCurrencyReq struct {
	Message string             `json:"message"`
	Data    GetCurrencyReqData `json:"data"`
}
