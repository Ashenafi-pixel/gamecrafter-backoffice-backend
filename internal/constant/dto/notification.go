package dto

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type NotificationType string

const (
	NOTIFICATION_TYPE_PROMOTIONAL NotificationType = "Promotional"
	NOTIFICATION_TYPE_KYC         NotificationType = "KYC"
	NOTIFICATION_TYPE_BONUS       NotificationType = "Bonus"
	NOTIFICATION_TYPE_WELCOME     NotificationType = "Welcome"
	NOTIFICATION_TYPE_SYSTEM      NotificationType = "System"
	NOTIFICATION_TYPE_ALERT       NotificationType = "Alert"
)

type NotificationMetadata struct {
	CTA    string `json:"cta,omitempty"`
	CTAURL string `json:"cta_url,omitempty"`
}

type NotificationPayload struct {
	UserID    uuid.UUID             `json:"user_id" validate:"required"`
	Title     string                `json:"title" validate:"required,min=1,max=200"`
	Content   string                `json:"content" validate:"required,min=1,max=1000"`
	Type      NotificationType      `json:"type" validate:"required,oneof=Promotional KYC Bonus Welcome System Alert"`
	Metadata  *NotificationMetadata `json:"metadata,omitempty"`
	CreatedBy *uuid.UUID            `json:"created_by,omitempty"`
}

type NotificationResponse struct {
	ID        uuid.UUID `json:"id,omitempty"`
	Delivered bool      `json:"delivered"`
	Timestamp time.Time `json:"timestamp"`
}

type UserNotification struct {
	ID        uuid.UUID            `json:"id"`
	UserID    uuid.UUID            `json:"user_id"`
	Title     string               `json:"title"`
	Content   string               `json:"content"`
	Type      NotificationType     `json:"type"`
	Metadata  NotificationMetadata `json:"metadata,omitempty"`
	Read      bool                 `json:"read"`
	Delivered bool                 `json:"delivered"`
	CreatedBy *uuid.UUID           `json:"created_by,omitempty"`
	CreatedAt time.Time            `json:"created_at"`
	ReadAt    *time.Time           `json:"read_at,omitempty"`
}

type GetNotificationsRequest struct {
	UserID     uuid.UUID        `json:"user_id"`
	Page       int              `json:"page" validate:"min=1"`
	PerPage    int              `json:"per_page" validate:"min=1,max=100"`
	Type       NotificationType `json:"type,omitempty"`
	UnreadOnly bool             `json:"unread_only,omitempty"`
}

type GetNotificationsResponse struct {
	Message       string             `json:"message"`
	Notifications []UserNotification `json:"notifications"`
	Total         int                `json:"total"`
	UnreadCount   int                `json:"unread_count"`
}

type MarkNotificationReadRequest struct {
	UserID         uuid.UUID `json:"user_id"`
	NotificationID uuid.UUID `json:"notification_id" validate:"required"`
}

type MarkNotificationReadResponse struct {
	Message string `json:"message"`
	Read    bool   `json:"read"`
}

type MarkAllNotificationsReadRequest struct {
	UserID uuid.UUID         `json:"user_id"`
	Type   *NotificationType `json:"type,omitempty"`
}

type MarkAllNotificationsReadResponse struct {
	Message      string `json:"message"`
	UpdatedCount int    `json:"updated_count"`
}

type DeleteNotificationRequest struct {
	UserID         uuid.UUID `json:"user_id"`
	NotificationID uuid.UUID `json:"notification_id" validate:"required"`
}

type DeleteNotificationResponse struct {
	Message string `json:"message"`
	Deleted bool   `json:"deleted"`
}

type WebSocketNotificationMessage struct {
	Type string           `json:"type"`
	Data UserNotification `json:"data"`
}

func ValidateNotificationPayload(payload NotificationPayload) error {
	validate := validator.New()
	return validate.Struct(payload)
}

func ValidateGetNotificationsRequest(req GetNotificationsRequest) error {
	validate := validator.New()
	return validate.Struct(req)
}

func ValidateMarkNotificationReadRequest(req MarkNotificationReadRequest) error {
	validate := validator.New()
	return validate.Struct(req)
}

func ValidateMarkAllNotificationsReadRequest(req MarkAllNotificationsReadRequest) error {
	validate := validator.New()
	return validate.Struct(req)
}

func ValidateDeleteNotificationRequest(req DeleteNotificationRequest) error {
	validate := validator.New()
	return validate.Struct(req)
}
