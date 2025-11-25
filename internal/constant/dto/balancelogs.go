package dto

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type BalanceLogs struct {
	ID                  uuid.UUID        `json:"id"`
	UserID              uuid.UUID        `json:"user_id"`
	Component           string           `json:"component"`
	Currency            string           `json:"currency"`
	Description         string           `json:"description"`
	ChangeAmount        decimal.Decimal  `json:"change_amount"`
	OperationalGroupID  uuid.UUID        `json:"operational_group_id"`
	OperationalTypeID   uuid.UUID        `json:"operational_type_id"`
	OperationalTypeName string           `json:"operational_type_name"`
	Timestamp           *time.Time       `json:"timestamp"`
	Type                string           `json:"type"`
	BalanceAfterUpdate  *decimal.Decimal `json:"balance_after_update"`
	TransactionID       *string          `json:"transaction_id"`
	Total               int              `json:"total,omitempty"`
	Status              string           `json:"status"`
	BrandID             *uuid.UUID       `json:"brand_id,omitempty"` // brand_id from users table
}

type BalanceLogsRes struct {
	ID                 uuid.UUID        `json:"id"`
	UserID             uuid.UUID        `json:"user_id"`
	Component          string           `json:"component"`
	Currency           string           `json:"currency"`
	Description        string           `json:"description"`
	ChangeAmount       decimal.Decimal  `json:"change_amount"`
	OperationalGroupID uuid.UUID        `json:"operational_group_id"`
	OperationalType    OperationalType  `json:"operational_type"`
	Timestamp          *time.Time       `json:"timestamp"`
	Type               string           `json:"type"`
	BalanceAfterUpdate *decimal.Decimal `json:"balance_after_update"`
	TransactionID      *string          `json:"transaction_id"`
	Total              int              `json:"total,omitempty"`
	Status             string           `json:"status"`
}

type OperationalType struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type SaveBalanceLogsReq struct {
	UpdateReq            UpdateBalanceReq
	UpdateRes            UpdateBalanceRes
	OperationalGroupID   uuid.UUID
	OperationalGroupType uuid.UUID
}

type GetBalanceLogReq struct {
	PerPage            int32            `form:"per_page" `
	Page               int32            `form:"page"`
	UserID             uuid.UUID        `form:"user_id"`
	Offset             int32            `form:"offset"`
	Component          string           `form:"component"`
	OperationalGroupID uuid.UUID        `form:"operation_group_id"`
	OperationTypeID    uuid.UUID        `form:"operation_type_id"`
	StartDate          *time.Time       `json:"start_date" form:"start_date"`
	EndDate            *time.Time       `json:"end_date" form:"end_date"`
	StartAmount        *decimal.Decimal `json:"start_amount" form:"start_amount"`
	EndAmount          *decimal.Decimal `json:"end_amount" form:"end_amount"`
}

type GetBalanceResData struct {
	Page       int              `json:"page"`
	TotalPages int              `json:"total_pages"`
	Logs       []BalanceLogsRes `json:"logs"`
}

type GetBalanceLogRes struct {
	Status string            `json:"status"`
	Data   GetBalanceResData `json:"data"`
}
type BalanceLogSort struct {
	Amount string `json:"amount" form:"sort_amount"`
	Date   string `json:"date" form:"sort_date"`
}
type BalanceLogFilter struct {
	StartDate       *time.Time       `json:"start_date" form:"filter_start_date" time_format:"2006-01-02"`
	EndDate         *time.Time       `json:"end_date" form:"filter_end_date" time_format:"2006-01-02"`
	TransactionType *string          `json:"transaction_type" form:"filter_transaction_type"`
	StartAmount     *decimal.Decimal `json:"start_amount" form:"filter_start_amount"`
	EndAmount       *decimal.Decimal `json:"end_amount" form:"filter_end_amount"`
	PlayerUsername  *string          `json:"player_username" form:"filter_username"`
	Status          *string          `json:"status" form:"filter_status"`
}

type AdminGetBalanceLogsReq struct {
	Filter  BalanceLogFilter `json:"filter"`
	Sort    BalanceLogSort   `json:"sort"`
	Page    int              `json:"page" form:"page"`
	PerPage int              `json:"per_page" form:"per_page"`
}

type GetAdminBalanceResData struct {
	ID                 uuid.UUID       `json:"id"`
	Component          string          `json:"component"`
	TransactionID      string          `json:"transaction_id"`
	Currency           string          `json:"currency"`
	ChangeAmount       decimal.Decimal `json:"change_amount"`
	BalanceAfterUpdate decimal.Decimal `json:"balance_after_update"`
	Timestamp          time.Time       `json:"timestamp"`
	OperationType      string          `json:"operation_type"`
	FirstName          string          `json:"first_name"`
	LastName           string          `json:"last_name"`
	Email              string          `json:"email"`
	Username           string          `json:"username"`
	PhoneNumber        string          `json:"phone_number"`
	TotalPages         int             `json:"total_pages"`
	Description        string          `json:"description"`
	Status             string          `json:"status"`
	TransactionType    string          `json:"transaction_type"`
}

type AdminGetBalanceLogsRes struct {
	Message    string                   `json:"message"`
	Data       []GetAdminBalanceResData `json:"data"`
	TotalPages int                      `json:"total_pages"`
}

type DbAdminBalanceResData struct {
	ID                 uuid.UUID
	Component          string
	TransactionID      string
	Currency           string
	ChangeAmount       decimal.Decimal
	BalanceAfterUpdate decimal.Decimal
	Timestamp          time.Time
	OperationType      string
	FirstName          sql.NullString
	LastName           sql.NullString
	Email              sql.NullString
	Username           sql.NullString
	PhoneNumber        sql.NullString
	Description        sql.NullString
	TransactionType    string
	Status             sql.NullString
}

func NullToString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

func NullToTime(nt sql.NullTime) *time.Time {
	if nt.Valid {
		return &nt.Time
	}
	return nil
}

func NullToUUID(nu uuid.NullUUID) uuid.UUID {
	if nu.Valid {
		return nu.UUID
	}
	return uuid.Nil
}
