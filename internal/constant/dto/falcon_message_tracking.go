package dto

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// FalconMessageStatus represents the status of a message sent to Falcon Liquidity
type FalconMessageStatus string

const (
	FalconMessageStatusPending      FalconMessageStatus = "pending"
	FalconMessageStatusSent         FalconMessageStatus = "sent"
	FalconMessageStatusFailed       FalconMessageStatus = "failed"
	FalconMessageStatusAcknowledged FalconMessageStatus = "acknowledged"
)

// FalconReconciliationStatus represents the reconciliation status
type FalconReconciliationStatus string

const (
	FalconReconciliationStatusPending    FalconReconciliationStatus = "pending"
	FalconReconciliationStatusReconciled FalconReconciliationStatus = "reconciled"
	FalconReconciliationStatusDisputed   FalconReconciliationStatus = "disputed"
)

// FalconMessageType represents the type of message sent to Falcon
type FalconMessageType string

const (
	FalconMessageTypeCasino FalconMessageType = "casino"
	FalconMessageTypeSport  FalconMessageType = "sport"
)

// FalconLiquidityMessage represents a message sent to Falcon Liquidity for tracking and reconciliation
type FalconLiquidityMessage struct {
	ID                   uuid.UUID                  `json:"id"`
	MessageID            string                     `json:"message_id"`
	TransactionID        string                     `json:"transaction_id"`
	UserID               uuid.UUID                  `json:"user_id"`
	MessageType          FalconMessageType          `json:"message_type"`
	MessageData          json.RawMessage            `json:"message_data"`
	BetAmount            decimal.Decimal            `json:"bet_amount"`
	PayoutAmount         decimal.Decimal            `json:"payout_amount"`
	Currency             string                     `json:"currency"`
	GameName             string                     `json:"game_name"`
	GameID               string                     `json:"game_id"`
	HouseEdge            decimal.Decimal            `json:"house_edge"`
	FalconRoutingKey     string                     `json:"falcon_routing_key"`
	FalconExchange       string                     `json:"falcon_exchange"`
	FalconQueue          string                     `json:"falcon_queue"`
	Status               FalconMessageStatus        `json:"status"`
	RetryCount           int                        `json:"retry_count"`
	LastRetryAt          *time.Time                 `json:"last_retry_at"`
	CreatedAt            time.Time                  `json:"created_at"`
	SentAt               *time.Time                 `json:"sent_at"`
	AcknowledgedAt       *time.Time                 `json:"acknowledged_at"`
	ErrorMessage         string                     `json:"error_message"`
	ErrorCode            string                     `json:"error_code"`
	FalconResponse       json.RawMessage            `json:"falcon_response"`
	ReconciliationStatus FalconReconciliationStatus `json:"reconciliation_status"`
	ReconciliationNotes  string                     `json:"reconciliation_notes"`
}

// CreateFalconMessageRequest represents a request to create a new Falcon message record
type CreateFalconMessageRequest struct {
	MessageID        string            `json:"message_id"`
	TransactionID    string            `json:"transaction_id"`
	UserID           uuid.UUID         `json:"user_id"`
	MessageType      FalconMessageType `json:"message_type"`
	MessageData      json.RawMessage   `json:"message_data"`
	BetAmount        decimal.Decimal   `json:"bet_amount"`
	PayoutAmount     decimal.Decimal   `json:"payout_amount"`
	Currency         string            `json:"currency"`
	GameName         string            `json:"game_name"`
	GameID           string            `json:"game_id"`
	HouseEdge        decimal.Decimal   `json:"house_edge"`
	FalconRoutingKey string            `json:"falcon_routing_key"`
	FalconExchange   string            `json:"falcon_exchange"`
	FalconQueue      string            `json:"falcon_queue"`
}

// UpdateFalconMessageRequest represents a request to update a Falcon message record
type UpdateFalconMessageRequest struct {
	Status               *FalconMessageStatus        `json:"status,omitempty"`
	RetryCount           *int                        `json:"retry_count,omitempty"`
	LastRetryAt          *time.Time                  `json:"last_retry_at,omitempty"`
	SentAt               *time.Time                  `json:"sent_at,omitempty"`
	AcknowledgedAt       *time.Time                  `json:"acknowledged_at,omitempty"`
	ErrorMessage         *string                     `json:"error_message,omitempty"`
	ErrorCode            *string                     `json:"error_code,omitempty"`
	FalconResponse       json.RawMessage             `json:"falcon_response,omitempty"`
	ReconciliationStatus *FalconReconciliationStatus `json:"reconciliation_status,omitempty"`
	ReconciliationNotes  *string                     `json:"reconciliation_notes,omitempty"`
}

// FalconMessageQuery represents query parameters for searching Falcon messages
type FalconMessageQuery struct {
	UserID               *uuid.UUID                  `json:"user_id,omitempty"`
	TransactionID        *string                     `json:"transaction_id,omitempty"`
	MessageID            *string                     `json:"message_id,omitempty"`
	Status               *FalconMessageStatus        `json:"status,omitempty"`
	MessageType          *FalconMessageType          `json:"message_type,omitempty"`
	ReconciliationStatus *FalconReconciliationStatus `json:"reconciliation_status,omitempty"`
	StartDate            *time.Time                  `json:"start_date,omitempty"`
	EndDate              *time.Time                  `json:"end_date,omitempty"`
	Limit                int                         `json:"limit"`
	Offset               int                         `json:"offset"`
}

// FalconMessageSummary represents summary statistics for Falcon messages
type FalconMessageSummary struct {
	TotalMessages        int             `json:"total_messages"`
	PendingMessages      int             `json:"pending_messages"`
	SentMessages         int             `json:"sent_messages"`
	FailedMessages       int             `json:"failed_messages"`
	AcknowledgedMessages int             `json:"acknowledged_messages"`
	DisputedMessages     int             `json:"disputed_messages"`
	TotalBetAmount       decimal.Decimal `json:"total_bet_amount"`
	TotalPayoutAmount    decimal.Decimal `json:"total_payout_amount"`
	AverageHouseEdge     decimal.Decimal `json:"average_house_edge"`
	LastMessageAt        *time.Time      `json:"last_message_at"`
}
