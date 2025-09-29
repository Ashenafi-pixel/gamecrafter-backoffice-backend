package analytics

import (
	"context"
	"fmt"
	"time"

	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/module/email"
	"github.com/tucanbit/internal/storage"
	"go.uber.org/zap"
)

// DailyReportService interface for daily report operations including email notifications
type DailyReportService interface {
	GenerateAndSendDailyReport(ctx context.Context, date time.Time, recipients []string) error
	GenerateDailyReportForDate(ctx context.Context, date time.Time) (*dto.DailyReport, error)
}

// DailyReportServiceImpl implementation of daily report service
type DailyReportServiceImpl struct {
	analyticsStorage storage.Analytics
	dailyReportEmail email.DailyReportEmailService
	logger           *zap.Logger
}

// NewDailyReportService creates a new daily report service
func NewDailyReportService(
	analyticsStorage storage.Analytics,
	dailyReportEmail email.DailyReportEmailService,
	logger *zap.Logger,
) DailyReportService {
	return &DailyReportServiceImpl{
		analyticsStorage: analyticsStorage,
		dailyReportEmail: dailyReportEmail,
		logger:           logger,
	}
}

// GenerateDailyReportForDate generates a daily report for a specific date
func (d *DailyReportServiceImpl) GenerateDailyReportForDate(ctx context.Context, date time.Time) (*dto.DailyReport, error) {
	d.logger.Info("Generating daily report",
		zap.String("date", date.Format("2006-01-02")))

	// Get daily report from analytics storage
	report, err := d.analyticsStorage.GetDailyReport(ctx, date)
	if err != nil {
		d.logger.Error("Failed to get daily report from analytics storage",
			zap.String("date", date.Format("2006-01-02")),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get daily report from analytics storage: %w", err)
	}

	d.logger.Info("Daily report generated successfully",
		zap.String("date", date.Format("2006-01-02")),
		zap.Uint32("total_transactions", report.TotalTransactions),
		zap.Uint32("active_users", report.ActiveUsers),
		zap.Uint32("new_users", report.NewUsers),
		zap.Uint32("active_games", report.ActiveGames),
		zap.String("net_revenue", report.NetRevenue.String()))

	return report, nil
}

// GenerateAndSendDailyReport generates and sends a daily report email
func (d *DailyReportServiceImpl) GenerateAndSendDailyReport(ctx context.Context, date time.Time, recipients []string) error {
	d.logger.Info("Starting daily report generation and email sending",
		zap.String("date", date.Format("2006-01-02")),
		zap.Int("recipients_count", len(recipients)))

	// Validate recipients
	if len(recipients) == 0 {
		return fmt.Errorf("no recipients specified for daily report")
	}

	// Generate daily report
	report, err := d.GenerateDailyReportForDate(ctx, date)
	if err != nil {
		return fmt.Errorf("failed to generate daily report: %w", err)
	}

	// Send email
	if err := d.dailyReportEmail.SendDailyReportEmail(report, recipients); err != nil {
		d.logger.Error("Failed to send daily report email",
			zap.String("date", date.Format("2006-01-02")),
			zap.Error(err))
		return fmt.Errorf("failed to send daily report email: %w", err)
	}

	d.logger.Info("Daily report email sent successfully",
		zap.String("date", date.Format("2006-01-02")),
		zap.Int("recipients_count", len(recipients)))

	return nil
}

// GenerateYesterdayReport generates and sends yesterday's report
func (d *DailyReportServiceImpl) GenerateYesterdayReport(ctx context.Context, recipients []string) error {
	yesterday := time.Now().AddDate(0, 0, -1)
	return d.GenerateAndSendDailyReport(ctx, yesterday, recipients)
}

// GenerateLastWeekReport generates and sends last week's daily reports (last 7 days)
func (d *DailyReportServiceImpl) GenerateLastWeekReport(ctx context.Context, recipients []string) error {
	d.logger.Info("Generating last week's daily reports",
		zap.Int("recipients_count", len(recipients)))

	for i := 1; i <= 7; i++ {
		date := time.Now().AddDate(0, 0, -i)
		d.logger.Info("Generating daily report for date",
			zap.String("date", date.Format("2006-01-02")))

		if err := d.GenerateAndSendDailyReport(ctx, date, recipients); err != nil {
			d.logger.Error("Failed to generate daily report for date",
				zap.String("date", date.Format("2006-01-02")),
				zap.Error(err))
			// Continue with other dates even if one fails
			continue
		}
	}

	d.logger.Info("Last week's daily reports generation completed")
	return nil
}