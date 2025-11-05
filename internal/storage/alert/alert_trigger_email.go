package alert

import (
	"context"
	"fmt"

	"github.com/tucanbit/internal/constant/dto"
	"go.uber.org/zap"
)

// AlertEmailSender interface for sending alert emails
type AlertEmailSender interface {
	SendAlertEmail(ctx context.Context, to []string, alertConfig *dto.AlertConfiguration, trigger *dto.AlertTrigger) error
}

// SendAlertEmailsToGroups sends alert emails to all emails in the assigned groups
// This should be called after creating an alert trigger
func SendAlertEmailsToGroups(
	ctx context.Context,
	emailGroupStorage AlertEmailGroupStorage,
	emailService AlertEmailSender,
	alertConfig *dto.AlertConfiguration,
	trigger *dto.AlertTrigger,
	log *zap.Logger,
) error {
	// Only send emails if email notifications are enabled
	if !alertConfig.EmailNotifications {
		log.Info("Email notifications disabled for alert", zap.String("alert_id", alertConfig.ID.String()))
		return nil
	}

	// Get all emails from assigned groups
	if len(alertConfig.EmailGroupIDs) == 0 {
		log.Info("No email groups assigned to alert", zap.String("alert_id", alertConfig.ID.String()))
		return nil
	}

	// Get emails from all groups
	emailsByGroup, err := emailGroupStorage.GetEmailsByGroupIDs(ctx, alertConfig.EmailGroupIDs)
	if err != nil {
		log.Error("Failed to get emails from groups", zap.Error(err), zap.Any("group_ids", alertConfig.EmailGroupIDs))
		return fmt.Errorf("failed to get emails from groups: %w", err)
	}

	// Collect all unique emails
	emailSet := make(map[string]bool)
	for _, emails := range emailsByGroup {
		for _, email := range emails {
			if email != "" {
				emailSet[email] = true
			}
		}
	}

	if len(emailSet) == 0 {
		log.Warn("No emails found in assigned groups", zap.Any("group_ids", alertConfig.EmailGroupIDs))
		return nil
	}

	// Convert set to slice
	var emails []string
	for email := range emailSet {
		emails = append(emails, email)
	}

	// Send emails
	err = emailService.SendAlertEmail(ctx, emails, alertConfig, trigger)
	if err != nil {
		log.Error("Failed to send alert emails", zap.Error(err), zap.Strings("emails", emails))
		return fmt.Errorf("failed to send alert emails: %w", err)
	}

	log.Info("Alert emails sent successfully",
		zap.String("alert_id", alertConfig.ID.String()),
		zap.Int("email_count", len(emails)),
		zap.Strings("emails", emails))

	return nil
}
