package dto

import (
	"time"

	"github.com/google/uuid"
)

// AdminActivityLog represents an admin activity log entry
type AdminActivityLog struct {
	ID           uuid.UUID   `json:"id"`
	AdminUserID  uuid.UUID   `json:"admin_user_id"`
	Action       string      `json:"action"`
	ResourceType string      `json:"resource_type"`
	ResourceID   *uuid.UUID  `json:"resource_id,omitempty"`
	Description  string      `json:"description"`
	Details      interface{} `json:"details,omitempty"`
	IPAddress    string      `json:"ip_address,omitempty"`
	UserAgent    string      `json:"user_agent,omitempty"`
	SessionID    string      `json:"session_id,omitempty"`
	Severity     string      `json:"severity"`
	Category     string      `json:"category"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`

	// Additional fields for display
	AdminUsername string `json:"admin_username,omitempty"`
	AdminEmail    string `json:"admin_email,omitempty"`
	CategoryName  string `json:"category_name,omitempty"`
	CategoryColor string `json:"category_color,omitempty"`
	CategoryIcon  string `json:"category_icon,omitempty"`
}

// AdminActivityCategory represents a category for admin activities
type AdminActivityCategory struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Color       string    `json:"color"`
	Icon        string    `json:"icon"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
}

// AdminActivityAction represents a predefined admin action
type AdminActivityAction struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CategoryID  uuid.UUID `json:"category_id"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
}

// CreateAdminActivityLogReq represents a request to create an admin activity log
type CreateAdminActivityLogReq struct {
	AdminUserID  uuid.UUID   `json:"admin_user_id" validate:"required"`
	Action       string      `json:"action" validate:"required,max=100"`
	ResourceType string      `json:"resource_type" validate:"required,max=50"`
	ResourceID   *uuid.UUID  `json:"resource_id,omitempty"`
	Description  string      `json:"description" validate:"required"`
	Details      interface{} `json:"details,omitempty"`
	IPAddress    string      `json:"ip_address,omitempty"`
	UserAgent    string      `json:"user_agent,omitempty"`
	SessionID    string      `json:"session_id,omitempty"`
	Severity     string      `json:"severity" validate:"oneof=low info warning error critical"`
	Category     string      `json:"category" validate:"required,max=50"`
}

// GetAdminActivityLogsReq represents a request to get admin activity logs
type GetAdminActivityLogsReq struct {
	AdminUserID  *uuid.UUID `json:"admin_user_id,omitempty"`
	Action       string     `json:"action,omitempty"`
	ResourceType string     `json:"resource_type,omitempty"`
	ResourceID   *uuid.UUID `json:"resource_id,omitempty"`
	Category     string     `json:"category,omitempty"`
	Severity     string     `json:"severity,omitempty"`
	Search       string     `json:"search,omitempty"`
	From         *time.Time `json:"from,omitempty"`
	To           *time.Time `json:"to,omitempty"`
	Page         int        `json:"page" validate:"min=1"`
	PerPage      int        `json:"per_page" validate:"min=1,max=100"`
	SortBy       string     `json:"sort_by" validate:"oneof=created_at action category severity"`
	SortOrder    string     `json:"sort_order" validate:"oneof=asc desc"`
}

// AdminActivityLogsRes represents the response for admin activity logs
type AdminActivityLogsRes struct {
	Logs       []AdminActivityLog `json:"logs"`
	Total      int64              `json:"total"`
	Page       int                `json:"page"`
	PerPage    int                `json:"per_page"`
	TotalPages int                `json:"total_pages"`
}

// AdminActivityStats represents statistics for admin activities
type AdminActivityStats struct {
	TotalActivities      int64                `json:"total_activities"`
	ActivitiesByCategory map[string]int64     `json:"activities_by_category"`
	ActivitiesByAction   map[string]int64     `json:"activities_by_action"`
	ActivitiesBySeverity map[string]int64     `json:"activities_by_severity"`
	RecentActivities     []AdminActivityLog   `json:"recent_activities"`
	TopAdmins            []AdminActivityCount `json:"top_admins"`
}

// AdminActivityCount represents admin activity count
type AdminActivityCount struct {
	AdminUserID   uuid.UUID `json:"admin_user_id"`
	AdminUsername string    `json:"admin_username"`
	AdminEmail    string    `json:"admin_email"`
	ActivityCount int64     `json:"activity_count"`
}

// AdminActivityLogFilter represents filters for admin activity logs
type AdminActivityLogFilter struct {
	AdminUserID  *uuid.UUID
	Action       string
	ResourceType string
	ResourceID   *uuid.UUID
	Category     string
	Severity     string
	From         *time.Time
	To           *time.Time
	Search       string
}

// AdminActivityLogSort represents sorting options for admin activity logs
type AdminActivityLogSort struct {
	Field string
	Order string
}
