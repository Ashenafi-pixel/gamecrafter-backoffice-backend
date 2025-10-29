package admin_activity_logger

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/module/admin_activity_logs"
	"go.uber.org/zap"
)

type AdminActivityLogger struct {
	module admin_activity_logs.AdminActivityLogsModule
	log    *zap.Logger
}

func NewAdminActivityLogger(module admin_activity_logs.AdminActivityLogsModule, log *zap.Logger) *AdminActivityLogger {
	return &AdminActivityLogger{
		module: module,
		log:    log,
	}
}

// LogActivity logs an admin activity
func (a *AdminActivityLogger) LogActivity(ctx context.Context, req dto.CreateAdminActivityLogReq) error {
	return a.module.LogAdminActivity(ctx, req)
}

// LogUserManagement logs user management activities
func (a *AdminActivityLogger) LogUserManagement(ctx context.Context, adminUserID uuid.UUID, action, description string, resourceID *uuid.UUID, details interface{}, ipAddress, userAgent string) error {
	return a.LogActivity(ctx, dto.CreateAdminActivityLogReq{
		AdminUserID:  adminUserID,
		Action:       action,
		ResourceType: "user",
		ResourceID:   resourceID,
		Description:  description,
		Details:      details,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		Severity:     "info",
		Category:     "user_management",
	})
}

// LogFinancial logs financial activities
func (a *AdminActivityLogger) LogFinancial(ctx context.Context, adminUserID uuid.UUID, action, description string, resourceID *uuid.UUID, details interface{}, ipAddress, userAgent string) error {
	return a.LogActivity(ctx, dto.CreateAdminActivityLogReq{
		AdminUserID:  adminUserID,
		Action:       action,
		ResourceType: "balance",
		ResourceID:   resourceID,
		Description:  description,
		Details:      details,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		Severity:     "info",
		Category:     "financial",
	})
}

// LogWithdrawal logs withdrawal activities
func (a *AdminActivityLogger) LogWithdrawal(ctx context.Context, adminUserID uuid.UUID, action, description string, resourceID *uuid.UUID, details interface{}, ipAddress, userAgent string) error {
	return a.LogActivity(ctx, dto.CreateAdminActivityLogReq{
		AdminUserID:  adminUserID,
		Action:       action,
		ResourceType: "withdrawal",
		ResourceID:   resourceID,
		Description:  description,
		Details:      details,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		Severity:     "info",
		Category:     "withdrawal",
	})
}

// LogSecurity logs security activities
func (a *AdminActivityLogger) LogSecurity(ctx context.Context, adminUserID uuid.UUID, action, description string, resourceID *uuid.UUID, details interface{}, ipAddress, userAgent string) error {
	severity := "info"
	if action == "login_failed" || action == "unauthorized_access" {
		severity = "warning"
	}

	return a.LogActivity(ctx, dto.CreateAdminActivityLogReq{
		AdminUserID:  adminUserID,
		Action:       action,
		ResourceType: "security",
		ResourceID:   resourceID,
		Description:  description,
		Details:      details,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		Severity:     severity,
		Category:     "security",
	})
}

// LogSystem logs system activities
func (a *AdminActivityLogger) LogSystem(ctx context.Context, adminUserID uuid.UUID, action, description string, resourceID *uuid.UUID, details interface{}, ipAddress, userAgent string) error {
	return a.LogActivity(ctx, dto.CreateAdminActivityLogReq{
		AdminUserID:  adminUserID,
		Action:       action,
		ResourceType: "system",
		ResourceID:   resourceID,
		Description:  description,
		Details:      details,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		Severity:     "info",
		Category:     "system",
	})
}

// LogGameManagement logs game management activities
func (a *AdminActivityLogger) LogGameManagement(ctx context.Context, adminUserID uuid.UUID, action, description string, resourceID *uuid.UUID, details interface{}, ipAddress, userAgent string) error {
	return a.LogActivity(ctx, dto.CreateAdminActivityLogReq{
		AdminUserID:  adminUserID,
		Action:       action,
		ResourceType: "game",
		ResourceID:   resourceID,
		Description:  description,
		Details:      details,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		Severity:     "info",
		Category:     "game_management",
	})
}

// LogReports logs report activities
func (a *AdminActivityLogger) LogReports(ctx context.Context, adminUserID uuid.UUID, action, description string, resourceID *uuid.UUID, details interface{}, ipAddress, userAgent string) error {
	return a.LogActivity(ctx, dto.CreateAdminActivityLogReq{
		AdminUserID:  adminUserID,
		Action:       action,
		ResourceType: "report",
		ResourceID:   resourceID,
		Description:  description,
		Details:      details,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		Severity:     "info",
		Category:     "reports",
	})
}

// LogNotifications logs notification activities
func (a *AdminActivityLogger) LogNotifications(ctx context.Context, adminUserID uuid.UUID, action, description string, resourceID *uuid.UUID, details interface{}, ipAddress, userAgent string) error {
	return a.LogActivity(ctx, dto.CreateAdminActivityLogReq{
		AdminUserID:  adminUserID,
		Action:       action,
		ResourceType: "notification",
		ResourceID:   resourceID,
		Description:  description,
		Details:      details,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		Severity:     "info",
		Category:     "notifications",
	})
}

// Helper function to extract admin info from Gin context
func ExtractAdminInfo(c *gin.Context) (adminUserID uuid.UUID, ipAddress, userAgent string) {
	adminUserIDStr := c.GetString("user-id")
	if adminUserIDStr != "" {
		adminUserID, _ = uuid.Parse(adminUserIDStr)
	}

	ipAddress = c.GetString("ip")
	userAgent = c.GetHeader("User-Agent")

	return
}

// Helper function to create details map
func CreateDetailsMap(keyValuePairs ...interface{}) map[string]interface{} {
	details := make(map[string]interface{})
	for i := 0; i < len(keyValuePairs); i += 2 {
		if i+1 < len(keyValuePairs) {
			key, ok := keyValuePairs[i].(string)
			if ok {
				details[key] = keyValuePairs[i+1]
			}
		}
	}
	return details
}
