package dto

import (
	"time"

	"github.com/google/uuid"
)

// AlertType represents the type of alert
type AlertType string

const (
	AlertTypeBetsCountLess        AlertType = "bets_count_less"
	AlertTypeBetsCountMore        AlertType = "bets_count_more"
	AlertTypeBetsAmountLess       AlertType = "bets_amount_less"
	AlertTypeBetsAmountMore       AlertType = "bets_amount_more"
	AlertTypeDepositsTotalLess    AlertType = "deposits_total_less"
	AlertTypeDepositsTotalMore    AlertType = "deposits_total_more"
	AlertTypeDepositsTypeLess     AlertType = "deposits_type_less"
	AlertTypeDepositsTypeMore     AlertType = "deposits_type_more"
	AlertTypeWithdrawalsTotalLess AlertType = "withdrawals_total_less"
	AlertTypeWithdrawalsTotalMore AlertType = "withdrawals_total_more"
	AlertTypeWithdrawalsTypeLess  AlertType = "withdrawals_type_less"
	AlertTypeWithdrawalsTypeMore  AlertType = "withdrawals_type_more"
	AlertTypeGGRTotalLess         AlertType = "ggr_total_less"
	AlertTypeGGRTotalMore         AlertType = "ggr_total_more"
	AlertTypeGGRSingleMore        AlertType = "ggr_single_more"
	AlertTypeMultipleAccountsSameIP AlertType = "multiple_accounts_same_ip"
)

// AlertStatus represents the status of an alert
type AlertStatus string

const (
	AlertStatusActive    AlertStatus = "active"
	AlertStatusInactive  AlertStatus = "inactive"
	AlertStatusTriggered AlertStatus = "triggered"
)

// CurrencyType represents supported currencies
type CurrencyType string

const (
	CurrencyUSD  CurrencyType = "USD"
	CurrencyBTC  CurrencyType = "BTC"
	CurrencyETH  CurrencyType = "ETH"
	CurrencySOL  CurrencyType = "SOL"
	CurrencyUSDT CurrencyType = "USDT"
	CurrencyUSDC CurrencyType = "USDC"
)

// AlertConfiguration represents an alert configuration
type AlertConfiguration struct {
	ID                 uuid.UUID     `json:"id" db:"id"`
	Name               string        `json:"name" db:"name"`
	Description        *string       `json:"description" db:"description"`
	AlertType          AlertType     `json:"alert_type" db:"alert_type"`
	Status             AlertStatus   `json:"status" db:"status"`
	ThresholdAmount    float64       `json:"threshold_amount" db:"threshold_amount"`
	TimeWindowMinutes  int           `json:"time_window_minutes" db:"time_window_minutes"`
	CurrencyCode       *CurrencyType `json:"currency_code" db:"currency_code"`
	EmailNotifications bool          `json:"email_notifications" db:"email_notifications"`
	WebhookURL         *string       `json:"webhook_url" db:"webhook_url"`
	EmailGroupIDs      []uuid.UUID   `json:"email_group_ids" db:"email_group_ids"` // Array of email group IDs
	CreatedBy          *uuid.UUID    `json:"created_by" db:"created_by"`
	CreatedAt          time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time     `json:"updated_at" db:"updated_at"`
	UpdatedBy          *uuid.UUID    `json:"updated_by" db:"updated_by"`
}

// AlertTrigger represents a triggered alert
type AlertTrigger struct {
	ID                   uuid.UUID     `json:"id" db:"id"`
	AlertConfigurationID uuid.UUID     `json:"alert_configuration_id" db:"alert_configuration_id"`
	TriggeredAt          time.Time     `json:"triggered_at" db:"triggered_at"`
	TriggerValue         float64       `json:"trigger_value" db:"trigger_value"`
	ThresholdValue       float64       `json:"threshold_value" db:"threshold_value"`
	UserID               *uuid.UUID    `json:"user_id" db:"user_id"`
	TransactionID        *string       `json:"transaction_id" db:"transaction_id"`
	AmountUSD            *float64      `json:"amount_usd" db:"amount_usd"`
	CurrencyCode         *CurrencyType `json:"currency_code" db:"currency_code"`
	ContextData          *string       `json:"context_data" db:"context_data"`
	Acknowledged         bool          `json:"acknowledged" db:"acknowledged"`
	AcknowledgedBy       *uuid.UUID    `json:"acknowledged_by" db:"acknowledged_by"`
	AcknowledgedAt       *time.Time    `json:"acknowledged_at" db:"acknowledged_at"`
	CreatedAt            time.Time     `json:"created_at" db:"created_at"`

	// Joined fields
	AlertConfiguration *AlertConfiguration `json:"alert_configuration,omitempty"`
	Username           *string             `json:"username,omitempty"`
	UserEmail          *string             `json:"user_email,omitempty"`
}

// CreateAlertConfigurationRequest represents the request to create an alert configuration
type CreateAlertConfigurationRequest struct {
	Name               string        `json:"name" binding:"required"`
	Description        *string       `json:"description"`
	AlertType          AlertType     `json:"alert_type" binding:"required"`
	ThresholdAmount    float64       `json:"threshold_amount" binding:"required,min=0"`
	TimeWindowMinutes  int           `json:"time_window_minutes" binding:"required,min=1"`
	CurrencyCode       *CurrencyType `json:"currency_code"`
	EmailNotifications bool          `json:"email_notifications"`
	WebhookURL         *string       `json:"webhook_url"`
	EmailGroupIDs      []uuid.UUID   `json:"email_group_ids"` // Array of email group IDs
}

// UpdateAlertConfigurationRequest represents the request to update an alert configuration
type UpdateAlertConfigurationRequest struct {
	Name               *string       `json:"name"`
	Description        *string       `json:"description"`
	Status             *AlertStatus  `json:"status"`
	ThresholdAmount    *float64      `json:"threshold_amount" binding:"omitempty,min=0"`
	TimeWindowMinutes  *int          `json:"time_window_minutes" binding:"omitempty,min=1"`
	CurrencyCode       *CurrencyType `json:"currency_code"`
	EmailNotifications *bool         `json:"email_notifications"`
	WebhookURL         *string       `json:"webhook_url"`
	EmailGroupIDs      []uuid.UUID   `json:"email_group_ids"` // Array of email group IDs
}

// GetAlertConfigurationsRequest represents the request to get alert configurations
type GetAlertConfigurationsRequest struct {
	Page      int          `form:"page" binding:"omitempty,min=1"`
	PerPage   int          `form:"per_page" binding:"omitempty,min=1,max=100"`
	AlertType *AlertType   `form:"alert_type"`
	Status    *AlertStatus `form:"status"`
	Search    string       `form:"search"`
}

// GetAlertTriggersRequest represents the request to get alert triggers
type GetAlertTriggersRequest struct {
	Page                 int        `form:"page" binding:"omitempty,min=1"`
	PerPage              int        `form:"per_page" binding:"omitempty,min=1,max=100"`
	AlertConfigurationID *uuid.UUID `form:"alert_configuration_id"`
	UserID               *uuid.UUID `form:"user_id"`
	Acknowledged         *bool      `form:"acknowledged"`
	DateFrom             *time.Time `form:"date_from"`
	DateTo               *time.Time `form:"date_to"`
	Search               string     `form:"search"`
}

// AlertConfigurationResponse represents the response for alert configuration operations
type AlertConfigurationResponse struct {
	Success bool                `json:"success"`
	Message string              `json:"message"`
	Data    *AlertConfiguration `json:"data,omitempty"`
	Error   string              `json:"error,omitempty"`
}

// AlertConfigurationsResponse represents the response for multiple alert configurations
type AlertConfigurationsResponse struct {
	Success    bool                 `json:"success"`
	Message    string               `json:"message"`
	Data       []AlertConfiguration `json:"data"`
	TotalCount int64                `json:"total_count"`
	Page       int                  `json:"page"`
	PerPage    int                  `json:"per_page"`
	Error      string               `json:"error,omitempty"`
}

// AlertTriggerResponse represents the response for alert trigger operations
type AlertTriggerResponse struct {
	Success bool          `json:"success"`
	Message string        `json:"message"`
	Data    *AlertTrigger `json:"data,omitempty"`
	Error   string        `json:"error,omitempty"`
}

// AlertTriggersResponse represents the response for multiple alert triggers
type AlertTriggersResponse struct {
	Success    bool           `json:"success"`
	Message    string         `json:"message"`
	Data       []AlertTrigger `json:"data"`
	TotalCount int64          `json:"total_count"`
	Page       int            `json:"page"`
	PerPage    int            `json:"per_page"`
	Error      string         `json:"error,omitempty"`
}

// AcknowledgeAlertRequest represents the request to acknowledge an alert
type AcknowledgeAlertRequest struct {
	Acknowledged bool `json:"acknowledged" binding:"required"`
}

// AlertEmailGroup represents an email group for alerts
type AlertEmailGroup struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	Name        string     `json:"name" db:"name"`
	Description *string    `json:"description" db:"description"`
	CreatedBy   *uuid.UUID `json:"created_by" db:"created_by"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	UpdatedBy   *uuid.UUID `json:"updated_by" db:"updated_by"`
	Emails      []string   `json:"emails,omitempty"` // Populated when fetching with members
}

// AlertEmailGroupMember represents an email member in a group
type AlertEmailGroupMember struct {
	ID        uuid.UUID `json:"id" db:"id"`
	GroupID   uuid.UUID `json:"group_id" db:"group_id"`
	Email     string    `json:"email" db:"email"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// CreateAlertEmailGroupRequest represents the request to create an email group
type CreateAlertEmailGroupRequest struct {
	Name        string   `json:"name" binding:"required"`
	Description *string  `json:"description"`
	Emails      []string `json:"emails" binding:"required,min=1"`
}

// UpdateAlertEmailGroupRequest represents the request to update an email group
type UpdateAlertEmailGroupRequest struct {
	Name        *string  `json:"name"`
	Description *string  `json:"description"`
	Emails      []string `json:"emails"` // If provided, replaces all emails
}

// AlertEmailGroupResponse represents the response for email group operations
type AlertEmailGroupResponse struct {
	Success bool             `json:"success"`
	Message string           `json:"message"`
	Data    *AlertEmailGroup `json:"data,omitempty"`
	Error   string           `json:"error,omitempty"`
}

// AlertEmailGroupsResponse represents the response for multiple email groups
type AlertEmailGroupsResponse struct {
	Success    bool              `json:"success"`
	Message    string            `json:"message"`
	Data       []AlertEmailGroup `json:"data"`
	TotalCount int64             `json:"total_count"`
	Error      string            `json:"error,omitempty"`
}
