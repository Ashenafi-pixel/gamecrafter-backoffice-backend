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
		"ReportHTML":   template.HTML(htmlBody), // Use template.HTML to prevent escaping
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
		<h2 style="color: #2c3e50; margin-bottom: 20px;">Daily Analytics Report</h2>
		<p style="margin-bottom: 15px; font-size: 16px;"><strong>Date:</strong> {{.DateFormatted}}</p>
		
		<div class="comprehensive-metrics" style="margin: 30px 0;">
			<h3 style="color: #2c3e50; border-bottom: 2px solid #2c3e50; padding-bottom: 10px;">Daily Performance Metrics</h3>
			<table style="width: 100%; border-collapse: collapse; margin: 20px 0;">
				<tr style="background-color: #f8f9fa;">
					<th style="padding: 15px; text-align: left; border: 1px solid #dee2e6;">Metric</th>
					<th style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">Value</th>
				</tr>
				<tr>
					<td style="padding: 15px; border: 1px solid #dee2e6;"><strong>Number of Registrations</strong></td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.NewUsers}}</td>
				</tr>
				<tr style="background-color: #f8f9fa;">
					<td style="padding: 15px; border: 1px solid #dee2e6;"><strong>Number of First Time Depositors</strong></td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.FirstTimeDepositors}}</td>
				</tr>
				<tr>
					<td style="padding: 15px; border: 1px solid #dee2e6;"><strong>Number of Active Customers</strong></td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.ActiveUsers}}</td>
				</tr>
				<tr style="background-color: #f8f9fa;">
					<td style="padding: 15px; border: 1px solid #dee2e6;"><strong>Number of Bets</strong></td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.BetCount}}</td>
				</tr>
				<tr>
					<td style="padding: 15px; border: 1px solid #dee2e6;"><strong>Bet Amount (USD)</strong></td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">${{.BetAmount}}</td>
				</tr>
				<tr style="background-color: #f8f9fa;">
					<td style="padding: 15px; border: 1px solid #dee2e6;"><strong>Win Amount (USD)</strong></td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">${{.WinAmount}}</td>
				</tr>
				<tr>
					<td style="padding: 15px; border: 1px solid #dee2e6;"><strong>GGR (Bet - Wins) USD</strong></td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">${{.GGR}}</td>
				</tr>
				<tr style="background-color: #f8f9fa;">
					<td style="padding: 15px; border: 1px solid #dee2e6;"><strong>Cashback Earned</strong></td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">${{.CashbackEarned}}</td>
				</tr>
				<tr>
					<td style="padding: 15px; border: 1px solid #dee2e6;"><strong>Cashback Claimed</strong></td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">${{.CashbackClaimed}}</td>
				</tr>
				<tr style="background-color: #f8f9fa;">
					<td style="padding: 15px; border: 1px solid #dee2e6;"><strong>NGR (GGR - Cashback Claimed)</strong></td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">${{.NGR}}</td>
				</tr>
				<tr>
					<td style="padding: 15px; border: 1px solid #dee2e6;"><strong>Number of Deposits</strong></td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.DepositCount}}</td>
				</tr>
				<tr style="background-color: #f8f9fa;">
					<td style="padding: 15px; border: 1px solid #dee2e6;"><strong>Deposit Amount (USD)</strong></td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">${{.DepositAmount}}</td>
				</tr>
				<tr>
					<td style="padding: 15px; border: 1px solid #dee2e6;"><strong>Number of Withdrawals</strong></td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.WithdrawalCount}}</td>
				</tr>
				<tr style="background-color: #f8f9fa;">
					<td style="padding: 15px; border: 1px solid #dee2e6;"><strong>Withdrawal Amount (USD)</strong></td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">${{.WithdrawalAmount}}</td>
				</tr>
				<tr>
					<td style="padding: 15px; border: 1px solid #dee2e6;"><strong>Admin Corrections (USD)</strong></td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">${{.AdminCorrections}}</td>
				</tr>
			</table>
		</div>

		{{if .TopGames}}
		<div class="top-games" style="margin: 30px 0;">
			<h3 style="color: #2c3e50; border-bottom: 2px solid #3498db; padding-bottom: 10px;">Top Performing Games</h3>
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
			<h3 style="color: #2c3e50; border-bottom: 2px solid #3498db; padding-bottom: 10px;">Top Players</h3>
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

	// Prepare data for the report content template
	data := struct {
		DateFormatted       string
		TotalTransactions   uint32
		TotalDeposits       string
		TotalWithdrawals    string
		TotalBets           string
		TotalWins           string
		NetRevenue          string
		ActiveUsers         uint32
		ActiveGames         uint32
		NewUsers            uint32
		FirstTimeDepositors uint32
		BetCount            uint32
		BetAmount           string
		WinAmount           string
		GGR                 string
		CashbackEarned      string
		CashbackClaimed     string
		NGR                 string
		DepositCount        uint32
		DepositAmount       string
		WithdrawalCount     uint32
		WithdrawalAmount    string
		AdminCorrections    string
		TopGames            []dto.GameStats
		TopPlayers          []struct {
			dto.PlayerStats
			LastActivityFormatted string
		}
	}{
		DateFormatted:       report.Date.Format("January 02, 2006"),
		TotalTransactions:   report.TotalTransactions,
		TotalDeposits:       report.TotalDeposits.StringFixed(2),
		TotalWithdrawals:    report.TotalWithdrawals.StringFixed(2),
		TotalBets:           report.TotalBets.StringFixed(2),
		TotalWins:           report.TotalWins.StringFixed(2),
		NetRevenue:          report.NetRevenue.StringFixed(2),
		ActiveUsers:         report.ActiveUsers,
		ActiveGames:         report.ActiveGames,
		NewUsers:            report.NewUsers,
		FirstTimeDepositors: 0,                        // TODO: Add this field to DailyReport DTO
		BetCount:            report.TotalTransactions, // Using total transactions as proxy for bet count
		BetAmount:           report.TotalBets.StringFixed(2),
		WinAmount:           report.TotalWins.StringFixed(2),
		GGR:                 report.TotalBets.Sub(report.TotalWins).StringFixed(2),
		CashbackEarned:      "0.00",                                                // TODO: Add this field to DailyReport DTO
		CashbackClaimed:     "0.00",                                                // TODO: Add this field to DailyReport DTO
		NGR:                 report.TotalBets.Sub(report.TotalWins).StringFixed(2), // GGR - Cashback Claimed
		DepositCount:        0,                                                     // TODO: Add actual deposit count field to DailyReport DTO
		DepositAmount:       report.TotalDeposits.StringFixed(2),
		WithdrawalCount:     0, // TODO: Add actual withdrawal count field to DailyReport DTO
		WithdrawalAmount:    report.TotalWithdrawals.StringFixed(2),
		AdminCorrections:    "0.00", // TODO: Add this field to DailyReport DTO
		TopGames:            report.TopGames,
	}

	// Format player stats for display
	for _, ps := range report.TopPlayers {
		data.TopPlayers = append(data.TopPlayers, struct {
			dto.PlayerStats
			LastActivityFormatted string
		}{
			PlayerStats:           ps,
			LastActivityFormatted: ps.LastActivity.Format("2006-01-02 15:04 UTC"),
		})
	}

	var buf bytes.Buffer
	if err := tmplParsed.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute daily report template: %w", err)
	}

	return buf.String(), nil
}
