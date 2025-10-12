package dto

import (
	"time"

	"github.com/google/uuid"
)

// Campaign-related DTOs

type CampaignStatus string

const (
	CAMPAIGN_STATUS_DRAFT     CampaignStatus = "draft"
	CAMPAIGN_STATUS_SCHEDULED CampaignStatus = "scheduled"
	CAMPAIGN_STATUS_SENDING   CampaignStatus = "sending"
	CAMPAIGN_STATUS_SENT      CampaignStatus = "sent"
	CAMPAIGN_STATUS_CANCELLED CampaignStatus = "cancelled"
)

type SegmentType string

const (
	SEGMENT_TYPE_CRITERIA  SegmentType = "criteria"
	SEGMENT_TYPE_CSV       SegmentType = "csv"
	SEGMENT_TYPE_ALL_USERS SegmentType = "all_users"
)

type RecipientStatus string

const (
	RECIPIENT_STATUS_PENDING   RecipientStatus = "pending"
	RECIPIENT_STATUS_SENT      RecipientStatus = "sent"
	RECIPIENT_STATUS_DELIVERED RecipientStatus = "delivered"
	RECIPIENT_STATUS_READ      RecipientStatus = "read"
	RECIPIENT_STATUS_FAILED    RecipientStatus = "failed"
)

// Campaign DTOs
type CreateCampaignRequest struct {
	Title       string                 `json:"title" validate:"required,min=1,max=255"`
	MessageType NotificationType       `json:"message_type" validate:"required,oneof=promotional kyc bonus welcome system alert payments security general"`
	Subject     string                 `json:"subject" validate:"required,min=1,max=255"`
	Content     string                 `json:"content" validate:"required,min=1,max=5000"`
	ScheduledAt *time.Time             `json:"scheduled_at,omitempty"`
	Segments    []CreateSegmentRequest `json:"segments" validate:"required,min=1"`
}

type CreateSegmentRequest struct {
	SegmentType SegmentType `json:"segment_type" validate:"required,oneof=criteria csv all_users"`
	SegmentName string      `json:"segment_name,omitempty"`
	Criteria    interface{} `json:"criteria,omitempty"`
	CSVData     string      `json:"csv_data,omitempty"`
}

type CampaignResponse struct {
	ID              uuid.UUID        `json:"id"`
	Title           string           `json:"title"`
	MessageType     NotificationType `json:"message_type"`
	Subject         string           `json:"subject"`
	Content         string           `json:"content"`
	CreatedBy       uuid.UUID        `json:"created_by"`
	Status          CampaignStatus   `json:"status"`
	ScheduledAt     *time.Time       `json:"scheduled_at,omitempty"`
	SentAt          *time.Time       `json:"sent_at,omitempty"`
	TotalRecipients int              `json:"total_recipients"`
	DeliveredCount  int              `json:"delivered_count"`
	ReadCount       int              `json:"read_count"`
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
}

type GetCampaignsRequest struct {
	Page        int               `json:"page"`
	PerPage     int               `json:"per_page"`
	Status      *CampaignStatus   `json:"status,omitempty"`
	MessageType *NotificationType `json:"message_type,omitempty"`
	CreatedBy   *uuid.UUID        `json:"created_by,omitempty"`
}

type GetCampaignsResponse struct {
	Campaigns []CampaignResponse `json:"campaigns"`
	Total     int                `json:"total"`
	Page      int                `json:"page"`
	PerPage   int                `json:"per_page"`
}

type UpdateCampaignRequest struct {
	Title       *string         `json:"title,omitempty"`
	Subject     *string         `json:"subject,omitempty"`
	Content     *string         `json:"content,omitempty"`
	ScheduledAt *time.Time      `json:"scheduled_at,omitempty"`
	Status      *CampaignStatus `json:"status,omitempty"`
}

// Segment DTOs
type SegmentResponse struct {
	ID          uuid.UUID   `json:"id"`
	CampaignID  uuid.UUID   `json:"campaign_id"`
	SegmentType SegmentType `json:"segment_type"`
	SegmentName string      `json:"segment_name,omitempty"`
	Criteria    interface{} `json:"criteria,omitempty"`
	CSVData     string      `json:"csv_data,omitempty"`
	UserCount   int         `json:"user_count"`
	CreatedAt   time.Time   `json:"created_at"`
}

// Recipient DTOs
type RecipientResponse struct {
	ID             uuid.UUID       `json:"id"`
	CampaignID     uuid.UUID       `json:"campaign_id"`
	UserID         uuid.UUID       `json:"user_id"`
	NotificationID *uuid.UUID      `json:"notification_id,omitempty"`
	Status         RecipientStatus `json:"status"`
	SentAt         *time.Time      `json:"sent_at,omitempty"`
	DeliveredAt    *time.Time      `json:"delivered_at,omitempty"`
	ReadAt         *time.Time      `json:"read_at,omitempty"`
	ErrorMessage   *string         `json:"error_message,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
}

type GetCampaignRecipientsRequest struct {
	CampaignID uuid.UUID        `json:"campaign_id"`
	Page       int              `json:"page"`
	PerPage    int              `json:"per_page"`
	Status     *RecipientStatus `json:"status,omitempty"`
}

type GetCampaignRecipientsResponse struct {
	Recipients []RecipientResponse `json:"recipients"`
	Total      int                 `json:"total"`
	Page       int                 `json:"page"`
	PerPage    int                 `json:"per_page"`
}

// Campaign Statistics
type CampaignStatsResponse struct {
	CampaignID      uuid.UUID `json:"campaign_id"`
	TotalRecipients int       `json:"total_recipients"`
	SentCount       int       `json:"sent_count"`
	DeliveredCount  int       `json:"delivered_count"`
	ReadCount       int       `json:"read_count"`
	FailedCount     int       `json:"failed_count"`
	DeliveryRate    float64   `json:"delivery_rate"`
	ReadRate        float64   `json:"read_rate"`
}

// Dashboard DTOs
type GetCampaignNotificationsDashboardRequest struct {
	Page             int        `json:"page"`
	PerPage          int        `json:"per_page"`
	Status           *string    `json:"status,omitempty"`
	NotificationType *string    `json:"type,omitempty"`
	UserID           *uuid.UUID `json:"user_id,omitempty"`
}

type CampaignNotificationDashboardItem struct {
	ID            uuid.UUID  `json:"id"`
	UserID        uuid.UUID  `json:"user_id"`
	Username      *string    `json:"username,omitempty"`
	Email         *string    `json:"email,omitempty"`
	Title         string     `json:"title"`
	Content       string     `json:"content"`
	Type          string     `json:"type"`
	Read          bool       `json:"read"`
	Delivered     bool       `json:"delivered"`
	ReadAt        *time.Time `json:"read_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	CampaignID    *uuid.UUID `json:"campaign_id,omitempty"`
	CampaignTitle *string    `json:"campaign_title,omitempty"`
}

type CampaignNotificationsDashboardResponse struct {
	Notifications []CampaignNotificationDashboardItem `json:"notifications"`
	Total         int                                 `json:"total"`
	Page          int                                 `json:"page"`
	PerPage       int                                 `json:"per_page"`
	Stats         CampaignNotificationStats           `json:"stats"`
}

type CampaignNotificationStats struct {
	TotalNotifications     int     `json:"total_notifications"`
	UnreadNotifications    int     `json:"unread_notifications"`
	DeliveredNotifications int     `json:"delivered_notifications"`
	ReadNotifications      int     `json:"read_notifications"`
	DeliveryRate           float64 `json:"delivery_rate"`
	ReadRate               float64 `json:"read_rate"`
}

// User Segmentation Criteria
type SegmentationCriteria struct {
	DaysSinceLastActivity *int       `json:"days_since_last_activity,omitempty"`
	UserLevel             *string    `json:"user_level,omitempty"`
	MinBalance            *float64   `json:"min_balance,omitempty"`
	MaxBalance            *float64   `json:"max_balance,omitempty"`
	RegistrationDateFrom  *time.Time `json:"registration_date_from,omitempty"`
	RegistrationDateTo    *time.Time `json:"registration_date_to,omitempty"`
	KYCStatus             *string    `json:"kyc_status,omitempty"`
	Country               *string    `json:"country,omitempty"`
	Currency              *string    `json:"currency,omitempty"`
}

// CSV Upload DTOs
type CSVUploadRequest struct {
	CampaignID uuid.UUID `json:"campaign_id"`
	CSVData    string    `json:"csv_data"`
}

type CSVUploadResponse struct {
	ProcessedCount int      `json:"processed_count"`
	ErrorCount     int      `json:"error_count"`
	Errors         []string `json:"errors,omitempty"`
}
