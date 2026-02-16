package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// WithdrawalPauseSettings represents the global pause settings
type WithdrawalPauseSettings struct {
	ID            uuid.UUID  `json:"id"`
	IsGloballyPaused bool    `json:"is_globally_paused"`
	PauseReason   *string    `json:"pause_reason,omitempty"`
	PausedBy      *uuid.UUID `json:"paused_by,omitempty"`
	PausedAt      *time.Time `json:"paused_at,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// UpdateWithdrawalPauseSettingsRequest represents the request to update pause settings
type UpdateWithdrawalPauseSettingsRequest struct {
	IsGloballyPaused *bool   `json:"is_globally_paused,omitempty"`
	PauseReason     *string `json:"pause_reason,omitempty"`
}

// WithdrawalThreshold represents a withdrawal threshold
type WithdrawalThreshold struct {
	ID            uuid.UUID       `json:"id"`
	ThresholdType string          `json:"threshold_type"`
	ThresholdValue decimal.Decimal `json:"threshold_value"`
	Currency      string          `json:"currency"`
	IsActive      bool            `json:"is_active"`
	CreatedBy     *uuid.UUID      `json:"created_by,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

// CreateWithdrawalThresholdRequest represents the request to create a threshold
type CreateWithdrawalThresholdRequest struct {
	ThresholdType  string          `json:"threshold_type" validate:"required,oneof=hourly_volume daily_volume single_transaction user_daily"`
	ThresholdValue decimal.Decimal `json:"threshold_value" validate:"required,gt=0"`
	Currency       string          `json:"currency" validate:"required"`
	IsActive       *bool           `json:"is_active,omitempty"`
}

// UpdateWithdrawalThresholdRequest represents the request to update a threshold
type UpdateWithdrawalThresholdRequest struct {
	ThresholdValue *decimal.Decimal `json:"threshold_value,omitempty"`
	IsActive       *bool            `json:"is_active,omitempty"`
}

// WithdrawalPauseLog represents a pause log entry
type WithdrawalPauseLog struct {
	ID            uuid.UUID       `json:"id"`
	WithdrawalID  uuid.UUID       `json:"withdrawal_id"`
	PauseReason   string          `json:"pause_reason"`
	ThresholdType *string         `json:"threshold_type,omitempty"`
	ThresholdValue *decimal.Decimal `json:"threshold_value,omitempty"`
	PausedAt      time.Time       `json:"paused_at"`
	ReviewedBy    *uuid.UUID      `json:"reviewed_by,omitempty"`
	ReviewedAt    *time.Time      `json:"reviewed_at,omitempty"`
	ActionTaken   *string         `json:"action_taken,omitempty"`
	Notes         *string         `json:"notes,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
}

// WithdrawalPauseActionRequest represents the request to approve/reject a paused withdrawal
type WithdrawalPauseActionRequest struct {
	Action string  `json:"action" validate:"required,oneof=approved rejected"`
	Notes  *string `json:"notes,omitempty"`
}

// GetPausedWithdrawalsRequest represents the request to get paused withdrawals
type GetPausedWithdrawalsRequest struct {
	Page       int     `form:"page" json:"page" validate:"min=1"`
	PerPage    int     `form:"per_page" json:"per_page" validate:"min=1,max=100"`
	Status     *string `form:"status" json:"status,omitempty"` // 'pending', 'approved', 'rejected'
	PauseReason *string `form:"pause_reason" json:"pause_reason,omitempty"`
	UserID     *uuid.UUID `form:"user_id" json:"user_id,omitempty"`
}

// PausedWithdrawal represents a paused withdrawal with additional info
type PausedWithdrawal struct {
	ID                    uuid.UUID       `json:"id"`
	UserID                uuid.UUID       `json:"user_id"`
	WithdrawalID          string          `json:"withdrawal_id"`
	Amount                decimal.Decimal `json:"amount"`
	Currency              string          `json:"currency"`
	Status                string          `json:"status"`
	IsPaused              bool            `json:"is_paused"`
	PauseReason           *string         `json:"pause_reason,omitempty"`
	PausedAt              *time.Time      `json:"paused_at,omitempty"`
	RequiresManualReview  bool            `json:"requires_manual_review"`
	CreatedAt             time.Time       `json:"created_at"`
	UpdatedAt             time.Time       `json:"updated_at"`
	// Additional user info
	Username              *string         `json:"username,omitempty"`
	Email                 *string         `json:"email,omitempty"`
	// Pause log info
	PauseLogs             []WithdrawalPauseLog `json:"pause_logs,omitempty"`
}

// GetPausedWithdrawalsResponse represents the response for paused withdrawals
type GetPausedWithdrawalsResponse struct {
	Withdrawals []PausedWithdrawal `json:"withdrawals"`
	Total       int                `json:"total"`
	Page        int                `json:"page"`
	PerPage     int                `json:"per_page"`
}

// WithdrawalPauseStats represents statistics for the pause system
type WithdrawalPauseStats struct {
	TotalPausedToday      int     `json:"total_paused_today"`
	TotalPausedThisHour   int     `json:"total_paused_this_hour"`
	PendingReview         int     `json:"pending_review"`
	ApprovedToday         int     `json:"approved_today"`
	RejectedToday         int     `json:"rejected_today"`
	AverageReviewTime     float64 `json:"average_review_time_minutes"`
	GlobalPauseStatus     bool    `json:"global_pause_status"`
	ActiveThresholds      int     `json:"active_thresholds"`
}

// Threshold types constants
const (
	THRESHOLD_HOURLY_VOLUME     = "hourly_volume"
	THRESHOLD_DAILY_VOLUME      = "daily_volume"
	THRESHOLD_SINGLE_TRANSACTION = "single_transaction"
	THRESHOLD_USER_DAILY        = "user_daily"
)

// Pause reasons constants
const (
	PAUSE_REASON_GLOBAL_PAUSE      = "global_pause"
	PAUSE_REASON_THRESHOLD_EXCEEDED = "threshold_exceeded"
	PAUSE_REASON_MANUAL_REVIEW      = "manual_review"
	PAUSE_REASON_USER_DAILY_LIMIT   = "user_daily_limit"
)

// Action taken constants
const (
	ACTION_PENDING  = "pending"
	ACTION_APPROVED = "approved"
	ACTION_REJECTED = "rejected"
)





