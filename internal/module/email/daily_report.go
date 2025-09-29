package email

import (
	"bytes"
	"fmt"
	"html/template"
	"time"

	"github.com/tucanbit/internal/constant/dto"
	"go.uber.org/zap"
)

// DailyReportEmailService interface for sending daily report emails
type DailyReportEmailService interface {
	SendDailyReportEmail(report *dto.DailyReport, recipients []string) error
}

// DailyReportEmailServiceImpl implementation of daily report email service
type DailyReportEmailServiceImpl struct {
	emailService EmailService
	logger       *zap.Logger
}

// NewDailyReportEmailService creates a new daily report email service
func NewDailyReportEmailService(emailService EmailService, logger *zap.Logger) DailyReportEmailService {
	return &DailyReportEmailServiceImpl{
		emailService: emailService,
		logger:       logger,
	}
}

// SendDailyReportEmail sends daily report email to multiple recipients
func (d *DailyReportEmailServiceImpl) SendDailyReportEmail(report *dto.DailyReport, recipients []string) error {
	subject := fmt.Sprintf("TucanBIT Daily Report - %s", report.Date.Format("January 2, 2006"))

	// Generate HTML content for the daily report
	htmlBody, err := d.generateDailyReportHTML(report)
	if err != nil {
		d.logger.Error("Failed to generate daily report HTML", zap.Error(err))
		return fmt.Errorf("failed to generate daily report HTML: %w", err)
	}

	// Send email to each recipient
	for _, recipient := range recipients {
		if err := d.sendDailyReportEmailToRecipient(recipient, subject, htmlBody); err != nil {
			d.logger.Error("Failed to send daily report email",
				zap.String("recipient", recipient),
				zap.Error(err))
			// Continue with other recipients even if one fails
			continue
		}

		d.logger.Info("Daily report email sent successfully",
			zap.String("recipient", recipient),
			zap.String("date", report.Date.Format("2006-01-02")),
			zap.Uint32("total_transactions", report.TotalTransactions),
			zap.String("net_revenue", report.NetRevenue.String()))
	}

	return nil
}

// sendDailyReportEmailToRecipient sends the daily report email to a single recipient
func (d *DailyReportEmailServiceImpl) sendDailyReportEmailToRecipient(recipient, subject, htmlBody string) error {
	// Use a generic email template structure for daily reports
	tmpl := GetDailyReportEmailTemplate()
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, map[string]interface{}{
		"Email":        recipient,
		"CurrentYear":  time.Now().Year(),
		"ReportHTML":   htmlBody,
		"BrandName":    "TucanBIT",
		"SupportEmail": "support@tucanbit.com",
	}); err != nil {
		return fmt.Errorf("failed to execute daily report template: %w", err)
	}

	finalHTMLBody := buf.String()

	// Send email using the existing email service's sendEmail method
	if impl, ok := d.emailService.(*EmailServiceImpl); ok {
		return impl.sendEmail(recipient, subject, finalHTMLBody)
	}

	return fmt.Errorf("unable to access email service implementation")
}

// generateDailyReportHTML generates HTML content for the daily report
func (d *DailyReportEmailServiceImpl) generateDailyReportHTML(report *dto.DailyReport) (string, error) {
	tmpl := `
	<div class="daily-report-content">
		<h2 style="color: #2c3e50; margin-bottom: 20px;">📊 Daily Analytics Report</h2>
		<p style="margin-bottom: 15px; font-size: 16px;"><strong>Date:</strong> {{.Date.Format "Monday, January 2, 2006"}}</p>
		
		<div class="metrics-grid" style="display: grid; grid-template-columns: repeat(auto-fit', minmax(200px, 1fr)); gap: 20px; margin: 20px 0;">
			<div class="metric-card" style="background: linear-gradient(135deg, #3498db, #2980b9); color: white; padding: 20px; border-radius: 10px; text-align: center;">
				<h3 style="margin: 0 0 10px 0; font-size: 28px; font-weight: bold;">{{.TotalTransactions}}</h3>
				<p style="margin: 0; font-size: 14px;">Total Transactions</p>
			</div>
			<div class="metric-card" style="background: linear-gradient(135deg, #27ae60, #229954); color: white; padding: 20px; border-radius: 10px; text-align: center;">
				<h3 style="margin: 0 0 10px 0; font-size: 28px; font-weight: bold;">{{.ActiveUsers}}</h3>
				<p style="margin: 0; font-size: 14px;">Active Users</p>
			</div>
			<div class="metric-card" style="background: linear-gradient(135deg, #9b59b6, #8e44ad); color: white; padding: 20px; border-radius: 10px; text-align: center;">
				<h3 style="margin: 0 0 10px 0; font-size: 28px; font-weight: bold;">{{.NewUsers}}</h3>
				<p style="margin: 0; font-size: 14px;">New Users</p>
			</div>
			<div class="metric-card" style="background: linear-gradient(135deg, #e67e22, #d35400); color: white; padding: 20px; border-radius: 10px; text-align: center;">
				<h3 style="margin: 0 0 10px 0; font-size: 28px; font-weight: bold;">{{.ActiveGames}}</h3>
				<p style="margin: 0; font-size: 14px;">Active Games</p>

			</div>
		</div>

		<div class="financial-metrics" style="margin: 30px 0;">
			<h3 style="color: #2c3e50; border-bottom: 2px solid #3498db; padding-bottom: 10px;">💰 Financial Overview</h3>
			<table style="width: 100%; border-collapse: collapse; margin: 20px 0;">
				<tr style="background-color: #f8f9fa;">
					<th style="padding: 15px; text-align: left; border: 1px solid #dee2e6;">Metric</th>
					<th style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">Amount (USD)</th>
				</tr>
				<tr>
					<td style="padding: 15px; border: 1px solid #dee2e6;"><strong>Total Deposits</strong></td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6; color: #27ae60;">${{.TotalDeposits}}</td>
				</tr>
				<tr style="background-color: #f8f9fa;">
					<td style="padding: 15px; border: 1px solid #dee2e6;"><strong>Total Withdrawals</strong></td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6; color: #e74c3c;">${{.TotalWithdrawals}}</td>
				</tr>
				<tr>
					<td style="padding: 15px; border: 1px solid #dee2e6;"><strong>Total Bets</strong></td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6; color: #3498db;">${{.TotalBets}}</td>
				</tr>
				<tr style="background-color: #f8f9fa;">
					<td style="padding: 15px; border: 1px solid #dee2e6;"><strong>Total Wins</strong></td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6; color: #9b59b6;">${{.TotalWins}}</td>
				</tr>
				<tr style="background-color: #ffeb3b;">
					<td style="padding: 15px; border: 1px solid #dee2e6;"><strong>Net Revenue</strong></td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6; color: #2c3e50; font-weight: bold;">${{.NetRevenue}}</td>
				</tr>
			</table>
		</div>

		{{if .TopGames}}
		<div class="top-games" style="margin: 30px 0;">
			<h3 style="color: #2c3e50; border-bottom: 2px solid #3498db; padding-bottom: 10px;">🎮 Top Performing Games</h3>
			<table style="width: 100%; border-collapse: collapse; margin: 20px 0;">
				<tr style="background-color: #3498db; color: white;">
					<th style="padding: 15px; text-align: center; border: 1px solid #2980b9;">Rank</th>
					<th style="padding: 15px; text-align: left; border: 1px solid #2980b9;">Game</th>
					<th style="padding: 15px; text-align: right; border: 1px solid #2980b9;">Total Bets</th>
					<th style="padding: 15px; text-align: right; border: 1px solid #2980b9;">Total Wins</th>
					<th style="padding: 15px; text-align: right; border: 1px solid #2980b9;">Net Revenue</th>
					<th style="padding: 15px; text-align: center; border: 1px solid #2980b9;">Players</th>
				</tr>
				{{range .TopGames}}
				<tr style="{{if odd .Rank}}background-color: #f8f9fa;{{end}}">
					<td style="padding: 15px; text-align: center; border: 1px solid #dee2e6;">{{.Rank}}</td>
					<td style="padding: 15px; border: 1px solid #dee2e6;">{{if .GameName}}{{.GameName}}{{else}}Game {{.GameID}}{{end}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6; color: #3498db;">>${{.TotalBets}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6; color: #27ae60;">${{.TotalWins}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6; color: #e74c3c;">${{.NetRevenue}}</td>
					<td style="padding: 15px; text-align: center; border: 1px solid #dee2e6;">{{.PlayerCount}}</td>
				</tr>
				{{end}}
			</table>
		</div>
		{{end}}

		{{if .TopPlayers}}
		<div class="top-players" style="margin: 30px 0;">
			<h3 style="color: #2c3e50; border-bottom: 2px solid #3498db; padding-bottom: 10px;">👑 Top Players</h3>
			<table style="width: 100%; border-collapse: collapse; margin: 20px 0;">
				<tr style="background-color: #27ae60; color: white;">
					<th style="padding: 15px; text-align: center; border: 1px solid #229954;">Rank</th>
					<th style="padding: 15px; text-align: left; border: 1px solid #229954;">Username</th>
					<th style="padding: 15px; text-align: right; border: 1px solid #229954;">Total Bets</th>
					<th style="padding: 15px; text-align: right; border: 1px solid #229954;">Total Wins</th>
					<th style="padding: 15px; text-align: right; border: 1px solid #229954;">Net Loss</th>
					<th style="padding: 15px; text-align: center; border: 1px solid #229954;">Transactions</th>
				</tr>
				{{range .TopPlayers}}
				<tr style="{{if odd .Rank}}background-color: #f8f9fa;{{end}}">
					<td style="padding: 15px; text-align: center; border: 1px solid #dee2e6;">{{.Rank}}</td>
					<td style="padding: 15px; border: 1px solid #dee2e6;">>{{.Username}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6; color: #3498db;">${{.TotalBets}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6; color: #27ae60;">${{.TotalWins}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6; color: #e74c3c;">${{.NetLoss}}</td>
					<td style="padding: 15px; text-align: center; border: 1px solid #dee2e6;">{{.TransactionCount}}</td>
				</tr>
				{{end}}
			</table>
		</div>
		{{end}}
	</div>`

	// Create template function for checking odd numbers
	funcMap := template.FuncMap{
		"odd": func(n int) bool {
			return n%2 == 1
		},
	}

	tmplParsed := template.Must(template.New("daily_report").Funcs(funcMap).Parse(tmpl))

	var buf bytes.Buffer
	if err := tmplParsed.Execute(&buf, report); err != nil {
		return "", fmt.Errorf("failed to execute daily report template: %w", err)
	}

	return buf.String(), nil
}
