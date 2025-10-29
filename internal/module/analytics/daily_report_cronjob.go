package analytics

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// DailyReportCronjobService handles automatic daily report scheduling
type DailyReportCronjobService interface {
	StartScheduler(ctx context.Context) error
	StopScheduler()
	SendConfiguredDailyReport(ctx context.Context, date time.Time) error
	IsRunning() bool
}

type dailyReportCronjob struct {
	logger               *zap.Logger
	dailyReportService   DailyReportService
	cron                 *cron.Cron
	httpClient           *http.Client
	appURL               string
	configuredRecipients []string
	scheduleTime         string
	timezone             string
	isRunning            bool
}

// NewDailyReportCronjobService creates a new daily report cronjob service
func NewDailyReportCronjobService(
	logger *zap.Logger,
	dailyReportService DailyReportService,
) DailyReportCronjobService {
	return &dailyReportCronjob{
		logger:             logger,
		dailyReportService: dailyReportService,
		cron:               cron.New(cron.WithSeconds()),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		appURL:               fmt.Sprintf("http://%s:%d", viper.GetString("app.host"), viper.GetInt("app.port")),
		configuredRecipients: []string{"ashenafialemu27@gmail.com", "joshjones612@gmail.com"},
		scheduleTime:         "0 59 23 * * *", // 23:59:00 (end of day)
		timezone:             "UTC",
		isRunning:            false,
	}
}

// StartScheduler starts the cronjob scheduler
func (d *dailyReportCronjob) StartScheduler(ctx context.Context) error {
	if d.isRunning {
		d.logger.Warn("Daily report cronjob scheduler is already running")
		return nil
	}

	// Add the daily report job
	_, err := d.cron.AddFunc(d.scheduleTime, func() {
		d.logger.Info("Starting scheduled daily report generation")

		// Get yesterday's date
		yesterday := time.Now().AddDate(0, 0, -1)

		// Send the daily report
		if err := d.SendConfiguredDailyReport(ctx, yesterday); err != nil {
			d.logger.Error("Failed to send scheduled daily report",
				zap.String("date", yesterday.Format("2006-01-02")),
				zap.Error(err))
		} else {
			d.logger.Info("Scheduled daily report sent successfully",
				zap.String("date", yesterday.Format("2006-01-02")),
				zap.Strings("recipients", d.configuredRecipients))
		}
	})

	if err != nil {
		d.logger.Error("Failed to add daily report cronjob", zap.Error(err))
		return err
	}

	// Start the cron scheduler
	d.cron.Start()
	d.isRunning = true

	d.logger.Info("Daily report cronjob scheduler started",
		zap.String("schedule", d.scheduleTime),
		zap.String("timezone", d.timezone),
		zap.Strings("recipients", d.configuredRecipients))

	return nil
}

// StopScheduler stops the cronjob scheduler
func (d *dailyReportCronjob) StopScheduler() {
	if !d.isRunning {
		return
	}

	d.cron.Stop()
	d.isRunning = false
	d.logger.Info("Daily report cronjob scheduler stopped")
}

// SendConfiguredDailyReport sends daily report to configured recipients
func (d *dailyReportCronjob) SendConfiguredDailyReport(ctx context.Context, date time.Time) error {
	d.logger.Info("Sending configured daily report",
		zap.String("date", date.Format("2006-01-02")),
		zap.Strings("recipients", d.configuredRecipients))

	// Use the daily report service directly
	err := d.dailyReportService.GenerateAndSendDailyReport(ctx, date, d.configuredRecipients)
	if err != nil {
		d.logger.Error("Failed to generate and send configured daily report",
			zap.String("date", date.Format("2006-01-02")),
			zap.Error(err))
		return err
	}

	d.logger.Info("Configured daily report sent successfully",
		zap.String("date", date.Format("2006-01-02")),
		zap.Strings("recipients", d.configuredRecipients))

	return nil
}

// IsRunning returns whether the scheduler is running
func (d *dailyReportCronjob) IsRunning() bool {
	return d.isRunning
}

// SendTestReport sends a test daily report for verification
func (d *dailyReportCronjob) SendTestReport(ctx context.Context) error {
	testDate := time.Now().AddDate(0, 0, -1) // Yesterday

	d.logger.Info("Sending test daily report",
		zap.String("date", testDate.Format("2006-01-02")),
		zap.Strings("recipients", d.configuredRecipients))

	return d.SendConfiguredDailyReport(ctx, testDate)
}
