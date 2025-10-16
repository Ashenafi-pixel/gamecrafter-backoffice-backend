package dto

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Balance struct {
	ID            uuid.UUID       `json:"id" `
	UserId        uuid.UUID       `json:"user_id"`
	CurrencyCode  string          `json:"currency_code" validate:"required"`
	AmountCents   int64           `json:"amount_cents"`   // amount in cents
	AmountUnits   decimal.Decimal `json:"amount_units"`   // amount in units
	ReservedCents int64           `json:"reserved_cents"` // reserved amount in cents
	ReservedUnits decimal.Decimal `json:"reserved_units"` // reserved amount in units
	UpdateAt      time.Time       `json:"updated_at"`
}

type UpdateBalanceReq struct {
	UserID           uuid.UUID       `json:"user_id" swaggerignore:"true"`
	Currency         string          `json:"currency"`
	Component        string          `json:"component"`
	OperationGroupID uuid.UUID       `json:"operation_group_id"`
	OperationTypeID  uuid.UUID       `json:"operation_type_id"`
	Operation        string          `json:"operation"`
	Amount           decimal.Decimal `json:"amount"`
	Description      string          `json:"description"`
}

type BalanceData struct {
	UserID           uuid.UUID       `json:"user_id"`
	Currency         string          `json:"currency"`
	UpdatedComponent string          `json:"updated_component"`
	NewBalance       decimal.Decimal `json:"new_balance"`
	OperationGroup   string          `json:"operation_group"`
	OperationType    string          `json:"operation_type"`
}
type UpdateBalanceRes struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    BalanceData `json:"data"`
}

func ValidateBalance(b Balance) error {
	validate := validator.New()
	return validate.Struct(b)
}

type OperationalGroupAndType struct {
	OperationalGroupID uuid.UUID
	OperationalTypeID  uuid.UUID
}

type ManualFundReq struct {
	UserID        uuid.UUID       `json:"user_id"`
	AdminID       uuid.UUID       `json:"admin_id" swaggerignore:"true"`
	TransactionID string          `json:"transaction_id" swaggerignore:"true"`
	Type          string          `json:"type" swaggerignore:"true"`
	Amount        decimal.Decimal `json:"amount"`
	Reason        string          `json:"reason"`
	Currency      string          `json:"currency,omitempty"` // Make currency optional
	Note          string          `json:"note"`
	CreatedAt     time.Time       `json:"created_at" swaggerignore:"true"`
}

type ManualFundResData struct {
	ID            uuid.UUID       `json:"id"`
	UserID        uuid.UUID       `json:"user_id"`
	AdminID       uuid.UUID       `json:"admin_id" swaggerignore:"true"`
	TransactionID string          `json:"transaction_id" swaggerignore:"true"`
	Amount        decimal.Decimal `json:"amount"`
	Reason        string          `json:"reason"`
	Currency      string          `json:"currency"`
	Note          string          `json:"note"`
	CreatedAt     time.Time       `json:"created_at"`
}
type ManualFundRes struct {
	Message string            `json:"message"`
	Data    ManualFundResData `json:"data"`
}
type GetManualFundFiler struct {
	StartDate        *time.Time `json:"start_date" form:"filter_start_date"`
	EndDate          *time.Time `json:"end_date" form:"filter_end_date"`
	Type             *string    `json:"type" form:"filter_type"`
	CustomerUsername *string    `json:"customer_username" form:"filter_customer_username"`
	CustomerEmail    *string    `json:"customer_email" form:"filter_customer_email"`
	CustomerPhone    *string    `json:"customer_phone" form:"filter_customer_phone"`
	AdminUsername    *string    `json:"admin_username" form:"filter_admin_username"`
	AdminEmail       *string    `json:"admin_email" form:"filter_admin_email"`
	AdminPhone       *string    `json:"admin_phone" form:"filter_admin_phone"`
}
type GetManualFundSort struct {
	Date       string `json:"sort_date" form:"sort_date"`
	AdminEmail string `json:"admin_email" form:"sort_admin_email"`
	Amount     string `json:"amount" form:"sort_amount"`
}
type GetManualFundReq struct {
	Filter  GetManualFundFiler `json:"filter"`
	Sort    GetManualFundSort  `json:"sort"`
	Page    int                `json:"page" form:"page"`
	PerPage int                `json:"per_page" form:"per_page"`
}
type GetManualFundData struct {
	ManualFund ManualFundResData `json:"manual_funds"`
	User       User              `json:"user"`
	FundBy     User              `json:"fund_by"`
}
type GetManualFundRes struct {
	Message   string              `json:"message"`
	Data      []GetManualFundData `json:"data"`
	TotalPage int                 `json:"total_pages"`
}

type GetManualDBRes struct {
	ID            uuid.UUID       `json:"id"`
	AdminID       uuid.UUID       `json:"admin_id" `
	TransactionID string          `json:"transaction_id"`
	Amount        decimal.Decimal `json:"amount"`
	Reason        string          `json:"reason"`
	Currency      string          `json:"currency"`
	Note          string          `json:"note"`
	CreatedAt     time.Time       `json:"created_at"`
	Type          string          `json:"type"`

	CustomerUserID          uuid.UUID `json:"customer_user_id,omitempty" `
	CustomerUserName        string    `json:"customer_username" validate:"usernamevalidation,min=3,max=20"`
	CustomerPhoneNumber     string    `json:"customer_phone_number" validate:"required,e164,min=8"`
	CustomerFirstName       string    `json:"customer_first_name"`
	CustomerLastName        string    `json:"customer_last_name"`
	CustomerEmail           string    `json:"customer_email"`
	CustomerDefaultCurrency string    `json:"customer_default_currency"`
	CustomerProfilePicture  string    `json:"customer_profile_picture"`
	CustomerDateOfBirth     string    `json:"customer_date_of_birth"`

	AdminUserID         uuid.UUID `json:"admin_user_id,omitempty" `
	AdminUserName       string    `json:"admin_username" validate:"usernamevalidation,min=3,max=20"`
	AdminPhoneNumber    string    `json:"admin_phone_number" validate:"required,e164,min=8"`
	AdminFirstName      string    `json:"admin_first_name"`
	AdminLastName       string    `json:"admin_last_name"`
	AdminEmail          string    `json:"admin_email"`
	AdminProfilePicture string    `json:"admin_profile_picture"`
	AdminDateOfBirth    string    `json:"admin_date_of_birth"`
}

// CreditWalletReq is used for crediting a user's wallet after payment confirmation
// for POST /wallet/credit
type CreditWalletReq struct {
	UserID           uuid.UUID       `json:"user_id"`
	Amount           decimal.Decimal `json:"amount"`
	Currency         string          `json:"currency"`
	PaymentReference string          `json:"payment_reference"`
	Provider         string          `json:"provider"`
	TxType           string          `json:"tx_type"`
}

type CreditWalletRes struct {
	Success bool   `json:"success"`
	Reason  string `json:"reason,omitempty"`
}
