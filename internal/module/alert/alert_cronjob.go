package alert

import (
	"context"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// AlertCronjobService handles periodic alert checking
type AlertCronjobService interface {
	StartScheduler(ctx context.Context) error
	StopScheduler()
	IsRunning() bool
}

type alertCronjob struct {
	alertService AlertService
	cron         *cron.Cron
	log          *zap.Logger
	isRunning    bool
}

// NewAlertCronjobService creates a new alert cronjob service
func NewAlertCronjobService(
	alertService AlertService,
	log *zap.Logger,
) AlertCronjobService {
	return &alertCronjob{
		alertService: alertService,
		cron:         cron.New(cron.WithSeconds()),
		log:          log,
		isRunning:    false,
	}
}

// StartScheduler starts the cronjob scheduler to check alerts every minute
func (a *alertCronjob) StartScheduler(ctx context.Context) error {
	if a.isRunning {
		a.log.Warn("Alert cronjob scheduler is already running")
		return nil
	}

	// Schedule alert checks to run every minute
	_, err := a.cron.AddFunc("@every 1m", func() {
		a.log.Info("Running scheduled alert check")
		if err := a.alertService.CheckAllAlerts(ctx); err != nil {
			a.log.Error("Error during scheduled alert check", zap.Error(err))
		}
	})

	if err != nil {
		a.log.Error("Failed to schedule alert checks", zap.Error(err))
		return err
	}

	a.cron.Start()
	a.isRunning = true
	a.log.Info("Alert cronjob scheduler started successfully - checking alerts every minute")
	return nil
}

// StopScheduler stops the cronjob scheduler
func (a *alertCronjob) StopScheduler() {
	if !a.isRunning {
		return
	}

	a.cron.Stop()
	a.isRunning = false
	a.log.Info("Alert cronjob scheduler stopped")
}

// IsRunning returns whether the scheduler is currently running
func (a *alertCronjob) IsRunning() bool {
	return a.isRunning
}

