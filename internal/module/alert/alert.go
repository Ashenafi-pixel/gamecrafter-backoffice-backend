package alert

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/dto"
	alertStorage "github.com/tucanbit/internal/storage/alert"
	"go.uber.org/zap"
)

// AlertService handles alert checking and email sending
type AlertService interface {
	CheckAllAlerts(ctx context.Context) error
	CheckBetCountAlerts(ctx context.Context) error
	CheckBetAmountAlerts(ctx context.Context) error
	CheckDepositAlerts(ctx context.Context, skipDuplicateCheck bool) error
	CheckWithdrawalAlerts(ctx context.Context) error
	CheckGGRAlerts(ctx context.Context) error
	CheckMultipleAccountsSameIP(ctx context.Context) error
}

type alertService struct {
	alertStorage      alertStorage.AlertStorage
	emailGroupStorage alertStorage.AlertEmailGroupStorage
	emailService      alertStorage.AlertEmailSender
	log               *zap.Logger
}

// NewAlertService creates a new alert service
func NewAlertService(
	alertStorage alertStorage.AlertStorage,
	emailGroupStorage alertStorage.AlertEmailGroupStorage,
	emailService alertStorage.AlertEmailSender,
	log *zap.Logger,
) AlertService {
	return &alertService{
		alertStorage:      alertStorage,
		emailGroupStorage: emailGroupStorage,
		emailService:      emailService,
		log:               log,
	}
}

// CheckAllAlerts checks all alert types
func (s *alertService) CheckAllAlerts(ctx context.Context) error {
	s.log.Info("Starting comprehensive alert check")

	// Get all active alert configurations to determine time windows
	req := &dto.GetAlertConfigurationsRequest{
		Page:    1,
		PerPage: 1000, // Get all active alerts
		Status:  func() *dto.AlertStatus { status := dto.AlertStatusActive; return &status }(),
	}

	configs, _, err := s.alertStorage.GetAlertConfigurations(ctx, req)
	if err != nil {
		s.log.Error("Failed to get alert configurations", zap.Error(err))
		return err
	}

	if len(configs) == 0 {
		s.log.Info("No active alert configurations found")
		return nil
	}

	// Group alerts by type and find the maximum time window needed
	maxTimeWindow := time.Duration(0)
	for _, config := range configs {
		window := time.Duration(config.TimeWindowMinutes) * time.Minute
		if window > maxTimeWindow {
			maxTimeWindow = window
		}
	}

	// Check each alert type
	if err := s.CheckBetCountAlerts(ctx); err != nil {
		s.log.Error("Failed to check bet count alerts", zap.Error(err))
	}

	if err := s.CheckBetAmountAlerts(ctx); err != nil {
		s.log.Error("Failed to check bet amount alerts", zap.Error(err))
	}

	if err := s.CheckDepositAlerts(ctx, false); err != nil {
		s.log.Error("Failed to check deposit alerts", zap.Error(err))
	}

	if err := s.CheckWithdrawalAlerts(ctx); err != nil {
		s.log.Error("Failed to check withdrawal alerts", zap.Error(err))
	}

	if err := s.CheckGGRAlerts(ctx); err != nil {
		s.log.Error("Failed to check GGR alerts", zap.Error(err))
	}

	if err := s.CheckMultipleAccountsSameIP(ctx); err != nil {
		s.log.Error("Failed to check multiple accounts same IP alerts", zap.Error(err))
	}

	s.log.Info("Completed comprehensive alert check")
	return nil
}

// checkAndTriggerAlert is a helper function that checks if an alert should trigger,
// creates the trigger, and sends emails
func (s *alertService) checkAndTriggerAlert(
	ctx context.Context,
	config dto.AlertConfiguration,
	triggerValue float64,
	userID *uuid.UUID,
	transactionID *string,
	amountUSD *float64,
	currencyCode *dto.CurrencyType,
) error {
	// Check if threshold is exceeded
	shouldTrigger := false
	switch config.AlertType {
	case dto.AlertTypeBetsCountMore, dto.AlertTypeBetsAmountMore,
		dto.AlertTypeDepositsTotalMore, dto.AlertTypeDepositsTypeMore,
		dto.AlertTypeWithdrawalsTotalMore, dto.AlertTypeWithdrawalsTypeMore,
		dto.AlertTypeGGRTotalMore, dto.AlertTypeGGRSingleMore,
		dto.AlertTypeMultipleAccountsSameIP:
		shouldTrigger = triggerValue >= config.ThresholdAmount
	case dto.AlertTypeBetsCountLess, dto.AlertTypeBetsAmountLess,
		dto.AlertTypeDepositsTotalLess, dto.AlertTypeDepositsTypeLess,
		dto.AlertTypeWithdrawalsTotalLess, dto.AlertTypeWithdrawalsTypeLess,
		dto.AlertTypeGGRTotalLess:
		shouldTrigger = triggerValue <= config.ThresholdAmount
	default:
		return nil
	}

	if !shouldTrigger {
		return nil
	}

	// Check if we already triggered this alert recently (avoid duplicates)
	// This check is done in the individual check methods before calling this function

	// Create trigger
	trigger := &dto.AlertTrigger{
		AlertConfigurationID: config.ID,
		TriggeredAt:          time.Now(),
		TriggerValue:         triggerValue,
		ThresholdValue:       config.ThresholdAmount,
		UserID:               userID,
		TransactionID:        transactionID,
		AmountUSD:            amountUSD,
		CurrencyCode:         currencyCode,
	}

	err := s.alertStorage.CreateAlertTrigger(ctx, trigger)
	if err != nil {
		s.log.Error("Failed to create alert trigger", zap.Error(err), zap.String("alert_id", config.ID.String()))
		return err
	}

	s.log.Info("Alert triggered",
		zap.String("alert_id", config.ID.String()),
		zap.String("alert_type", string(config.AlertType)),
		zap.Float64("trigger_value", triggerValue),
		zap.Float64("threshold", config.ThresholdAmount))

	// Send emails if email notifications are enabled
	if config.EmailNotifications {
		err = alertStorage.SendAlertEmailsToGroups(
			ctx,
			s.emailGroupStorage,
			s.emailService,
			&config,
			trigger,
			s.log,
		)
		if err != nil {
			s.log.Error("Failed to send alert emails", zap.Error(err), zap.String("alert_id", config.ID.String()))
			// Don't return error - trigger is created, email failure is logged
		}
	}

	return nil
}

// sendEmailsForNewTriggers checks for triggers created since checkStartTime and sends emails
// This ensures emails are sent immediately after triggers are created
func (s *alertService) sendEmailsForNewTriggers(ctx context.Context, checkStartTime time.Time) error {
	// Get triggers created since check started (subtract 2 seconds to account for timing and ensure we catch all new triggers)
	adjustedStartTime := checkStartTime.Add(-2 * time.Second)
	req := &dto.GetAlertTriggersRequest{
		Page:  1,
		PerPage: 100,
		DateFrom: &adjustedStartTime,
	}

	triggers, _, err := s.alertStorage.GetAlertTriggers(ctx, req)
	if err != nil {
		s.log.Error("Failed to get new triggers for email sending", zap.Error(err))
		return err
	}

	s.log.Info("Checking for new triggers to send emails",
		zap.Int("trigger_count", len(triggers)),
		zap.Time("check_start_time", checkStartTime),
		zap.Time("adjusted_start_time", adjustedStartTime))

	// Send emails for each trigger
	for _, trigger := range triggers {
		if trigger.AlertConfiguration == nil {
			// Get the configuration
			config, err := s.alertStorage.GetAlertConfigurationByID(ctx, trigger.AlertConfigurationID)
			if err != nil {
				s.log.Error("Failed to get alert configuration for trigger", zap.Error(err), zap.String("trigger_id", trigger.ID.String()))
				continue
			}
			trigger.AlertConfiguration = config
		}

		config := trigger.AlertConfiguration
		s.log.Debug("Processing trigger for email sending",
			zap.String("trigger_id", trigger.ID.String()),
			zap.String("alert_id", config.ID.String()),
			zap.Bool("email_notifications", config.EmailNotifications),
			zap.Int("email_group_count", len(config.EmailGroupIDs)))
		
		if config.EmailNotifications && len(config.EmailGroupIDs) > 0 {
			err := alertStorage.SendAlertEmailsToGroups(
				ctx,
				s.emailGroupStorage,
				s.emailService,
				config,
				&trigger,
				s.log,
			)
			if err != nil {
				s.log.Error("Failed to send alert emails", zap.Error(err), zap.String("trigger_id", trigger.ID.String()))
				// Continue with other triggers
			} else {
				s.log.Info("Alert emails sent for trigger", 
					zap.String("trigger_id", trigger.ID.String()),
					zap.String("alert_id", config.ID.String()))
			}
		} else {
			s.log.Debug("Skipping email send for trigger",
				zap.String("trigger_id", trigger.ID.String()),
				zap.Bool("email_notifications", config.EmailNotifications),
				zap.Int("email_group_count", len(config.EmailGroupIDs)))
		}
	}

	return nil
}

// CheckBetCountAlerts checks bet count alerts and sends emails
func (s *alertService) CheckBetCountAlerts(ctx context.Context) error {
	checkStartTime := time.Now()
	maxTimeWindow := time.Hour * 24 // Default max window
	
	// Get max time window from configs
	req := &dto.GetAlertConfigurationsRequest{
		Page:    1,
		PerPage: 100,
		Status:  func() *dto.AlertStatus { status := dto.AlertStatusActive; return &status }(),
	}
	configs, _, _ := s.alertStorage.GetAlertConfigurations(ctx, req)
	for _, config := range configs {
		if config.AlertType == dto.AlertTypeBetsCountMore || config.AlertType == dto.AlertTypeBetsCountLess {
			window := time.Duration(config.TimeWindowMinutes) * time.Minute
			if window > maxTimeWindow {
				maxTimeWindow = window
			}
		}
	}

	err := s.alertStorage.CheckBetCountAlerts(ctx, maxTimeWindow)
	if err != nil {
		return err
	}

	// Send emails for any triggers created
	return s.sendEmailsForNewTriggers(ctx, checkStartTime)
}

// CheckBetAmountAlerts checks bet amount alerts and sends emails
func (s *alertService) CheckBetAmountAlerts(ctx context.Context) error {
	checkStartTime := time.Now()
	maxTimeWindow := time.Hour * 24
	
	req := &dto.GetAlertConfigurationsRequest{
		Page:    1,
		PerPage: 100,
		Status:  func() *dto.AlertStatus { status := dto.AlertStatusActive; return &status }(),
	}
	configs, _, _ := s.alertStorage.GetAlertConfigurations(ctx, req)
	for _, config := range configs {
		if config.AlertType == dto.AlertTypeBetsAmountMore || config.AlertType == dto.AlertTypeBetsAmountLess {
			window := time.Duration(config.TimeWindowMinutes) * time.Minute
			if window > maxTimeWindow {
				maxTimeWindow = window
			}
		}
	}

	err := s.alertStorage.CheckBetAmountAlerts(ctx, maxTimeWindow)
	if err != nil {
		return err
	}

	return s.sendEmailsForNewTriggers(ctx, checkStartTime)
}

// CheckDepositAlerts checks deposit alerts and sends emails
// skipDuplicateCheck: if true, will create triggers even if one was created recently (useful for manual fund additions)
func (s *alertService) CheckDepositAlerts(ctx context.Context, skipDuplicateCheck bool) error {
	checkStartTime := time.Now()
	maxTimeWindow := time.Hour * 24
	
	req := &dto.GetAlertConfigurationsRequest{
		Page:    1,
		PerPage: 100,
		Status:  func() *dto.AlertStatus { status := dto.AlertStatusActive; return &status }(),
	}
	configs, _, _ := s.alertStorage.GetAlertConfigurations(ctx, req)
	for _, config := range configs {
		if config.AlertType == dto.AlertTypeDepositsTotalMore || config.AlertType == dto.AlertTypeDepositsTotalLess ||
			config.AlertType == dto.AlertTypeDepositsTypeMore || config.AlertType == dto.AlertTypeDepositsTypeLess {
			window := time.Duration(config.TimeWindowMinutes) * time.Minute
			if window > maxTimeWindow {
				maxTimeWindow = window
			}
		}
	}

	err := s.alertStorage.CheckDepositAlerts(ctx, maxTimeWindow, skipDuplicateCheck)
	if err != nil {
		return err
	}

	return s.sendEmailsForNewTriggers(ctx, checkStartTime)
}

// CheckWithdrawalAlerts checks withdrawal alerts and sends emails
func (s *alertService) CheckWithdrawalAlerts(ctx context.Context) error {
	checkStartTime := time.Now()
	maxTimeWindow := time.Hour * 24
	
	req := &dto.GetAlertConfigurationsRequest{
		Page:    1,
		PerPage: 100,
		Status:  func() *dto.AlertStatus { status := dto.AlertStatusActive; return &status }(),
	}
	configs, _, _ := s.alertStorage.GetAlertConfigurations(ctx, req)
	for _, config := range configs {
		if config.AlertType == dto.AlertTypeWithdrawalsTotalMore || config.AlertType == dto.AlertTypeWithdrawalsTotalLess ||
			config.AlertType == dto.AlertTypeWithdrawalsTypeMore || config.AlertType == dto.AlertTypeWithdrawalsTypeLess {
			window := time.Duration(config.TimeWindowMinutes) * time.Minute
			if window > maxTimeWindow {
				maxTimeWindow = window
			}
		}
	}

	err := s.alertStorage.CheckWithdrawalAlerts(ctx, maxTimeWindow)
	if err != nil {
		return err
	}

	return s.sendEmailsForNewTriggers(ctx, checkStartTime)
}

// CheckGGRAlerts checks GGR alerts and sends emails
func (s *alertService) CheckGGRAlerts(ctx context.Context) error {
	checkStartTime := time.Now()
	maxTimeWindow := time.Hour * 24
	
	req := &dto.GetAlertConfigurationsRequest{
		Page:    1,
		PerPage: 100,
		Status:  func() *dto.AlertStatus { status := dto.AlertStatusActive; return &status }(),
	}
	configs, _, _ := s.alertStorage.GetAlertConfigurations(ctx, req)
	for _, config := range configs {
		if config.AlertType == dto.AlertTypeGGRTotalMore || config.AlertType == dto.AlertTypeGGRTotalLess ||
			config.AlertType == dto.AlertTypeGGRSingleMore {
			window := time.Duration(config.TimeWindowMinutes) * time.Minute
			if window > maxTimeWindow {
				maxTimeWindow = window
			}
		}
	}

	err := s.alertStorage.CheckGGRAlerts(ctx, maxTimeWindow)
	if err != nil {
		return err
	}

	return s.sendEmailsForNewTriggers(ctx, checkStartTime)
}

// CheckMultipleAccountsSameIP checks multiple accounts same IP alerts and sends emails
func (s *alertService) CheckMultipleAccountsSameIP(ctx context.Context) error {
	checkStartTime := time.Now()
	maxTimeWindow := time.Hour * 24
	
	req := &dto.GetAlertConfigurationsRequest{
		Page:    1,
		PerPage: 100,
		Status:  func() *dto.AlertStatus { status := dto.AlertStatusActive; return &status }(),
	}
	configs, _, _ := s.alertStorage.GetAlertConfigurations(ctx, req)
	for _, config := range configs {
		if config.AlertType == dto.AlertTypeMultipleAccountsSameIP {
			window := time.Duration(config.TimeWindowMinutes) * time.Minute
			if window > maxTimeWindow {
				maxTimeWindow = window
			}
		}
	}

	err := s.alertStorage.CheckMultipleAccountsSameIP(ctx, maxTimeWindow)
	if err != nil {
		return err
	}

	return s.sendEmailsForNewTriggers(ctx, checkStartTime)
}

