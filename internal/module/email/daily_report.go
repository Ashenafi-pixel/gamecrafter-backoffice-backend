package email

import (
	"bytes"
	"fmt"
	"html/template"
	"time"

	"github.com/shopspring/decimal"
	"github.com/tucanbit/internal/constant/dto"
	"go.uber.org/zap"
)

// DailyReportEmailService interface for sending daily report emails
type DailyReportEmailService interface {
	SendDailyReportEmail(report *dto.EnhancedDailyReport, recipients []string) error
	SendWeeklyReportEmail(reports []*dto.EnhancedDailyReport, recipients []string) error
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
func (d *DailyReportEmailServiceImpl) SendDailyReportEmail(report *dto.EnhancedDailyReport, recipients []string) error {
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
			zap.Uint64("total_transactions", report.TotalTransactions),
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
func (d *DailyReportEmailServiceImpl) generateDailyReportHTML(report *dto.EnhancedDailyReport) (string, error) {
	tmpl := `
	<div class="daily-report-content">
		<h2 style="color: #2c3e50; margin-bottom: 20px;">Daily Analytics Report</h2>
		<p style="margin-bottom: 15px; font-size: 16px;"><strong>Date:</strong> {{.DateFormatted}}</p>
		
		<div class="comprehensive-metrics" style="margin: 30px 0;">
			<h3 style="color: #2c3e50; border-bottom: 2px solid #2c3e50; padding-bottom: 10px;">Daily Performance Metrics with Comparison</h3>
			<div style="overflow-x: auto; margin: 20px 0; border: 1px solid #dee2e6; border-radius: 8px;">
				<table style="min-width: 1000px; width: 100%; border-collapse: collapse; margin: 0;">
				<tr style="background-color: #f8f9fa;">
					<th style="padding: 15px; text-align: left; border: 1px solid #dee2e6;">Metric</th>
					<th style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">Today</th>
					<th style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">% Change vs Yesterday</th>
					<th style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">MTD</th>
					<th style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">SPLM</th>
					<th style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">% Change MTD vs SPLM</th>
				</tr>
				<tr>
					<td style="padding: 15px; border: 1px solid #dee2e6;"><strong>Number of Registrations</strong></td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.NewUsers}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.PreviousDayChange.NewUsersChange}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.MTD.NewUsers}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.SPLM.NewUsers}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.MTDvsSPLMChange.NewUsersChange}}</td>
				</tr>
				<tr style="background-color: #f8f9fa;">
					<td style="padding: 15px; border: 1px solid #dee2e6;"><strong>Number of Unique Depositors</strong></td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.UniqueDepositors}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.PreviousDayChange.UniqueDepositorsChange}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.MTD.UniqueDepositors}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.SPLM.UniqueDepositors}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.MTDvsSPLMChange.UniqueDepositorsChange}}</td>
				</tr>
				<tr>
					<td style="padding: 15px; border: 1px solid #dee2e6;"><strong>Number of Unique Withdrawers</strong></td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.UniqueWithdrawers}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.PreviousDayChange.UniqueWithdrawersChange}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.MTD.UniqueWithdrawers}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.SPLM.UniqueWithdrawers}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.MTDvsSPLMChange.UniqueWithdrawersChange}}</td>
				</tr>
				<tr style="background-color: #f8f9fa;">
					<td style="padding: 15px; border: 1px solid #dee2e6;"><strong>Number of Active Customers</strong></td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.ActiveUsers}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.PreviousDayChange.ActiveUsersChange}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.MTD.ActiveUsers}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.SPLM.ActiveUsers}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.MTDvsSPLMChange.ActiveUsersChange}}</td>
				</tr>
				<tr>
					<td style="padding: 15px; border: 1px solid #dee2e6;"><strong>Number of Bets</strong></td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.BetCount}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.PreviousDayChange.BetCountChange}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.MTD.BetCount}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.SPLM.BetCount}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.MTDvsSPLMChange.BetCountChange}}</td>
				</tr>
				<tr style="background-color: #f8f9fa;">
					<td style="padding: 15px; border: 1px solid #dee2e6;"><strong>Bet Amount (USD)</strong></td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">${{.BetAmount}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.PreviousDayChange.TotalBetsChange}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">${{.MTD.TotalBets}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">${{.SPLM.TotalBets}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.MTDvsSPLMChange.TotalBetsChange}}</td>
				</tr>
				<tr>
					<td style="padding: 15px; border: 1px solid #dee2e6;"><strong>Win Amount (USD)</strong></td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">${{.WinAmount}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.PreviousDayChange.TotalWinsChange}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">${{.MTD.TotalWins}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">${{.SPLM.TotalWins}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.MTDvsSPLMChange.TotalWinsChange}}</td>
				</tr>
				<tr style="background-color: #f8f9fa;">
					<td style="padding: 15px; border: 1px solid #dee2e6;"><strong>GGR (Bet - Wins) USD</strong></td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">${{.GGR}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.PreviousDayChange.NetRevenueChange}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">${{.MTD.NetRevenue}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">${{.SPLM.NetRevenue}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.MTDvsSPLMChange.NetRevenueChange}}</td>
				</tr>
				<tr>
					<td style="padding: 15px; border: 1px solid #dee2e6;"><strong>Cashback Earned</strong></td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">${{.CashbackEarned}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.PreviousDayChange.CashbackEarnedChange}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">${{.MTD.CashbackEarned}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">${{.SPLM.CashbackEarned}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.MTDvsSPLMChange.CashbackEarnedChange}}</td>
				</tr>
				<tr style="background-color: #f8f9fa;">
					<td style="padding: 15px; border: 1px solid #dee2e6;"><strong>Cashback Claimed</strong></td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">${{.CashbackClaimed}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.PreviousDayChange.CashbackClaimedChange}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">${{.MTD.CashbackClaimed}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">${{.SPLM.CashbackClaimed}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.MTDvsSPLMChange.CashbackClaimedChange}}</td>
				</tr>
				<tr>
					<td style="padding: 15px; border: 1px solid #dee2e6;"><strong>NGR (GGR - Cashback Claimed)</strong></td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">${{.NGR}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">-</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">-</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">-</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">-</td>
				</tr>
				<tr style="background-color: #f8f9fa;">
					<td style="padding: 15px; border: 1px solid #dee2e6;"><strong>Number of Deposits</strong></td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.DepositCount}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.PreviousDayChange.DepositCountChange}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.MTD.DepositCount}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.SPLM.DepositCount}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.MTDvsSPLMChange.DepositCountChange}}</td>
				</tr>
				<tr>
					<td style="padding: 15px; border: 1px solid #dee2e6;"><strong>Deposit Amount (USD)</strong></td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">${{.DepositAmount}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.PreviousDayChange.TotalDepositsChange}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">${{.MTD.TotalDeposits}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">${{.SPLM.TotalDeposits}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.MTDvsSPLMChange.TotalDepositsChange}}</td>
				</tr>
				<tr style="background-color: #f8f9fa;">
					<td style="padding: 15px; border: 1px solid #dee2e6;"><strong>Number of Withdrawals</strong></td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.WithdrawalCount}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.PreviousDayChange.WithdrawalCountChange}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.MTD.WithdrawalCount}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.SPLM.WithdrawalCount}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.MTDvsSPLMChange.WithdrawalCountChange}}</td>
				</tr>
				<tr>
					<td style="padding: 15px; border: 1px solid #dee2e6;"><strong>Withdrawal Amount (USD)</strong></td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">${{.WithdrawalAmount}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.PreviousDayChange.TotalWithdrawalsChange}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">${{.MTD.TotalWithdrawals}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">${{.SPLM.TotalWithdrawals}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.MTDvsSPLMChange.TotalWithdrawalsChange}}</td>
				</tr>
				<tr style="background-color: #f8f9fa;">
					<td style="padding: 15px; border: 1px solid #dee2e6;"><strong>Admin Corrections (USD)</strong></td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">${{.AdminCorrections}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.PreviousDayChange.AdminCorrectionsChange}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">${{.MTD.AdminCorrections}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">${{.SPLM.AdminCorrections}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.MTDvsSPLMChange.AdminCorrectionsChange}}</td>
				</tr>
				</table>
			</div>
		</div>

		{{if .TopGames}}
		<div class="top-games" style="margin: 30px 0;">
			<h3 style="color: #2c3e50; border-bottom: 2px solid #3498db; padding-bottom: 10px;">Top Performing Games</h3>
			<div style="overflow-x: auto; margin: 20px 0; border: 1px solid #dee2e6; border-radius: 8px;">
				<table style="min-width: 800px; width: 100%; border-collapse: collapse; margin: 0;">
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
		</div>
		{{end}}

		{{if .TopPlayers}}
		<div class="top-players" style="margin: 30px 0;">
			<h3 style="color: #2c3e50; border-bottom: 2px solid #3498db; padding-bottom: 10px;">Top Players</h3>
			<div style="overflow-x: auto; margin: 20px 0; border: 1px solid #dee2e6; border-radius: 8px;">
				<table style="min-width: 1000px; width: 100%; border-collapse: collapse; margin: 0;">
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
		DateFormatted     string
		TotalTransactions uint64
		TotalDeposits     string
		TotalWithdrawals  string
		TotalBets         string
		TotalWins         string
		NetRevenue        string
		ActiveUsers       uint64
		ActiveGames       uint64
		NewUsers          uint64
		UniqueDepositors  uint64
		UniqueWithdrawers uint64
		BetCount          uint64
		BetAmount         string
		WinAmount         string
		GGR               string
		CashbackEarned    string
		CashbackClaimed   string
		NGR               string
		DepositCount      uint64
		DepositAmount     string
		WithdrawalCount   uint64
		WithdrawalAmount  string
		AdminCorrections  string
		TopGames          []dto.GameStats
		TopPlayers        []struct {
			dto.PlayerStats
			LastActivityFormatted string
		}
		PreviousDayChange dto.DailyReportComparison
		MTD               struct {
			TotalTransactions uint64
			TotalDeposits     string
			TotalWithdrawals  string
			TotalBets         string
			TotalWins         string
			NetRevenue        string
			ActiveUsers       uint64
			ActiveGames       uint64
			NewUsers          uint64
			UniqueDepositors  uint64
			UniqueWithdrawers uint64
			DepositCount      uint64
			WithdrawalCount   uint64
			BetCount          uint64
			WinCount          uint64
			CashbackEarned    string
			CashbackClaimed   string
			AdminCorrections  string
		}
		SPLM struct {
			TotalTransactions uint64
			TotalDeposits     string
			TotalWithdrawals  string
			TotalBets         string
			TotalWins         string
			NetRevenue        string
			ActiveUsers       uint64
			ActiveGames       uint64
			NewUsers          uint64
			UniqueDepositors  uint64
			UniqueWithdrawers uint64
			DepositCount      uint64
			WithdrawalCount   uint64
			BetCount          uint64
			WinCount          uint64
			CashbackEarned    string
			CashbackClaimed   string
			AdminCorrections  string
		}
		MTDvsSPLMChange dto.DailyReportComparison
	}{
		DateFormatted:     report.Date.Format("January 02, 2006"),
		TotalTransactions: report.TotalTransactions,
		TotalDeposits:     report.TotalDeposits.StringFixed(2),
		TotalWithdrawals:  report.TotalWithdrawals.StringFixed(2),
		TotalBets:         report.TotalBets.StringFixed(2),
		TotalWins:         report.TotalWins.StringFixed(2),
		NetRevenue:        report.NetRevenue.StringFixed(2),
		ActiveUsers:       report.ActiveUsers,
		ActiveGames:       report.ActiveGames,
		NewUsers:          report.NewUsers,
		UniqueDepositors:  report.UniqueDepositors,
		UniqueWithdrawers: report.UniqueWithdrawers,
		BetCount:          report.BetCount,
		BetAmount:         report.TotalBets.StringFixed(2),
		WinAmount:         report.TotalWins.StringFixed(2),
		GGR:               report.TotalBets.Sub(report.TotalWins).StringFixed(2),
		CashbackEarned:    report.CashbackEarned.StringFixed(2),
		CashbackClaimed:   report.CashbackClaimed.StringFixed(2),
		NGR:               report.TotalBets.Sub(report.TotalWins).Sub(report.CashbackClaimed).StringFixed(2), // GGR - Cashback Claimed
		DepositCount:      report.DepositCount,
		DepositAmount:     report.TotalDeposits.StringFixed(2),
		WithdrawalCount:   report.WithdrawalCount,
		WithdrawalAmount:  report.TotalWithdrawals.StringFixed(2),
		AdminCorrections:  report.AdminCorrections.StringFixed(2),
		PreviousDayChange: report.PreviousDayChange,
		MTD: struct {
			TotalTransactions uint64
			TotalDeposits     string
			TotalWithdrawals  string
			TotalBets         string
			TotalWins         string
			NetRevenue        string
			ActiveUsers       uint64
			ActiveGames       uint64
			NewUsers          uint64
			UniqueDepositors  uint64
			UniqueWithdrawers uint64
			DepositCount      uint64
			WithdrawalCount   uint64
			BetCount          uint64
			WinCount          uint64
			CashbackEarned    string
			CashbackClaimed   string
			AdminCorrections  string
		}{
			TotalTransactions: report.MTD.TotalTransactions,
			TotalDeposits:     report.MTD.TotalDeposits.StringFixed(2),
			TotalWithdrawals:  report.MTD.TotalWithdrawals.StringFixed(2),
			TotalBets:         report.MTD.TotalBets.StringFixed(2),
			TotalWins:         report.MTD.TotalWins.StringFixed(2),
			NetRevenue:        report.MTD.NetRevenue.StringFixed(2),
			ActiveUsers:       report.MTD.ActiveUsers,
			ActiveGames:       report.MTD.ActiveGames,
			NewUsers:          report.MTD.NewUsers,
			UniqueDepositors:  report.MTD.UniqueDepositors,
			UniqueWithdrawers: report.MTD.UniqueWithdrawers,
			DepositCount:      report.MTD.DepositCount,
			WithdrawalCount:   report.MTD.WithdrawalCount,
			BetCount:          report.MTD.BetCount,
			WinCount:          report.MTD.WinCount,
			CashbackEarned:    report.MTD.CashbackEarned.StringFixed(2),
			CashbackClaimed:   report.MTD.CashbackClaimed.StringFixed(2),
			AdminCorrections:  report.MTD.AdminCorrections.StringFixed(2),
		},
		SPLM: struct {
			TotalTransactions uint64
			TotalDeposits     string
			TotalWithdrawals  string
			TotalBets         string
			TotalWins         string
			NetRevenue        string
			ActiveUsers       uint64
			ActiveGames       uint64
			NewUsers          uint64
			UniqueDepositors  uint64
			UniqueWithdrawers uint64
			DepositCount      uint64
			WithdrawalCount   uint64
			BetCount          uint64
			WinCount          uint64
			CashbackEarned    string
			CashbackClaimed   string
			AdminCorrections  string
		}{
			TotalTransactions: report.SPLM.TotalTransactions,
			TotalDeposits:     report.SPLM.TotalDeposits.StringFixed(2),
			TotalWithdrawals:  report.SPLM.TotalWithdrawals.StringFixed(2),
			TotalBets:         report.SPLM.TotalBets.StringFixed(2),
			TotalWins:         report.SPLM.TotalWins.StringFixed(2),
			NetRevenue:        report.SPLM.NetRevenue.StringFixed(2),
			ActiveUsers:       report.SPLM.ActiveUsers,
			ActiveGames:       report.SPLM.ActiveGames,
			NewUsers:          report.SPLM.NewUsers,
			UniqueDepositors:  report.SPLM.UniqueDepositors,
			UniqueWithdrawers: report.SPLM.UniqueWithdrawers,
			DepositCount:      report.SPLM.DepositCount,
			WithdrawalCount:   report.SPLM.WithdrawalCount,
			BetCount:          report.SPLM.BetCount,
			WinCount:          report.SPLM.WinCount,
			CashbackEarned:    report.SPLM.CashbackEarned.StringFixed(2),
			CashbackClaimed:   report.SPLM.CashbackClaimed.StringFixed(2),
			AdminCorrections:  report.SPLM.AdminCorrections.StringFixed(2),
		},
		MTDvsSPLMChange: report.MTDvsSPLMChange,
		TopGames:        report.TopGames,
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

// SendWeeklyReportEmail sends weekly report email with data from multiple days
func (d *DailyReportEmailServiceImpl) SendWeeklyReportEmail(reports []*dto.EnhancedDailyReport, recipients []string) error {
	if len(reports) == 0 {
		return fmt.Errorf("no reports provided for weekly email")
	}

	// Get date range for subject
	startDate := reports[len(reports)-1].Date // Oldest date
	endDate := reports[0].Date                // Newest date

	subject := fmt.Sprintf("TucanBIT Weekly Report - %s to %s",
		startDate.Format("January 2, 2006"),
		endDate.Format("January 2, 2006"))

	// Generate HTML content for the weekly report
	htmlBody, err := d.generateWeeklyReportHTML(reports)
	if err != nil {
		d.logger.Error("Failed to generate weekly report HTML", zap.Error(err))
		return fmt.Errorf("failed to generate weekly report HTML: %w", err)
	}

	// Send email to all recipients
	for _, recipient := range recipients {
		// Use the same template structure as daily reports
		tmpl := GetDailyReportEmailTemplate()
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, map[string]interface{}{
			"Email":        recipient,
			"CurrentYear":  time.Now().Year(),
			"ReportHTML":   template.HTML(htmlBody), // Use template.HTML to prevent escaping
			"BrandName":    "TucanBIT",
			"SupportEmail": "support@tucanbit.com",
		}); err != nil {
			d.logger.Error("Failed to execute weekly report template",
				zap.String("recipient", recipient),
				zap.Error(err))
			continue
		}

		finalHTMLBody := buf.String()

		// Send email using the existing email service's sendEmail method
		if impl, ok := d.emailService.(*EmailServiceImpl); ok {
			if err := impl.sendEmail(recipient, subject, finalHTMLBody); err != nil {
				d.logger.Error("Failed to send weekly report email",
					zap.String("recipient", recipient),
					zap.Error(err))
				// Continue with other recipients even if one fails
				continue
			}
			d.logger.Info("Weekly report email sent successfully",
				zap.String("recipient", recipient),
				zap.Int("days_included", len(reports)))
		} else {
			d.logger.Error("Unable to access email service implementation",
				zap.String("recipient", recipient))
		}
	}

	return nil
}

// generateWeeklyReportHTML generates HTML content for the weekly report
func (d *DailyReportEmailServiceImpl) generateWeeklyReportHTML(reports []*dto.EnhancedDailyReport) (string, error) {
	tmpl := `
	<div class="weekly-report-content">
		<h2 style="color: #2c3e50; margin-bottom: 20px;">Weekly Analytics Report</h2>
		<p style="margin-bottom: 15px; font-size: 16px;"><strong>Period:</strong> {{.StartDate}} to {{.EndDate}}</p>
		<p style="margin-bottom: 15px; font-size: 16px;"><strong>Days Included:</strong> {{.DaysCount}}</p>
		
		<div class="weekly-summary" style="margin: 30px 0;">
			<h3 style="color: #2c3e50; border-bottom: 2px solid #2c3e50; padding-bottom: 10px;">Weekly Summary</h3>
			<div style="overflow-x: auto; margin: 20px 0; border: 1px solid #dee2e6; border-radius: 8px;">
				<table style="min-width: 1200px; width: 100%; border-collapse: collapse; margin: 0;">
				<tr style="background-color: #f8f9fa;">
					<th style="padding: 15px; text-align: left; border: 1px solid #dee2e6;">Metric</th>
					<th style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">Total</th>
					<th style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">Average per Day</th>
					<th style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">Best Day</th>
					<th style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">Worst Day</th>
					<th style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">MTD</th>
					<th style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">SPLM</th>
					<th style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">% Change MTD vs SPLM</th>
				</tr>
				<tr>
					<td style="padding: 15px; border: 1px solid #dee2e6;"><strong>Total Registrations</strong></td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.Summary.TotalNewUsers}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.Summary.AvgNewUsers}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.Summary.BestDay.NewUsers}} ({{.Summary.BestDay.Date}})</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.Summary.WorstDay.NewUsers}} ({{.Summary.WorstDay.Date}})</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.Summary.MTD.NewUsers}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.Summary.SPLM.NewUsers}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.Summary.MTDvsSPLMChange.NewUsersChange}}</td>
				</tr>
				<tr style="background-color: #f8f9fa;">
					<td style="padding: 15px; border: 1px solid #dee2e6;"><strong>Total Unique Depositors</strong></td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.Summary.TotalUniqueDepositors}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.Summary.AvgUniqueDepositors}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.Summary.BestDay.UniqueDepositors}} ({{.Summary.BestDay.Date}})</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.Summary.WorstDay.UniqueDepositors}} ({{.Summary.WorstDay.Date}})</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.Summary.MTD.UniqueDepositors}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.Summary.SPLM.UniqueDepositors}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.Summary.MTDvsSPLMChange.UniqueDepositorsChange}}</td>
				</tr>
				<tr>
					<td style="padding: 15px; border: 1px solid #dee2e6;"><strong>Total Unique Withdrawers</strong></td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.Summary.TotalUniqueWithdrawers}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.Summary.AvgUniqueWithdrawers}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.Summary.BestDay.UniqueWithdrawers}} ({{.Summary.BestDay.Date}})</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.Summary.WorstDay.UniqueWithdrawers}} ({{.Summary.WorstDay.Date}})</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.Summary.MTD.UniqueWithdrawers}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.Summary.SPLM.UniqueWithdrawers}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.Summary.MTDvsSPLMChange.UniqueWithdrawersChange}}</td>
				</tr>
				<tr style="background-color: #f8f9fa;">
					<td style="padding: 15px; border: 1px solid #dee2e6;"><strong>Total Active Users</strong></td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.Summary.TotalActiveUsers}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.Summary.AvgActiveUsers}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.Summary.BestDay.ActiveUsers}} ({{.Summary.BestDay.Date}})</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.Summary.WorstDay.ActiveUsers}} ({{.Summary.WorstDay.Date}})</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.Summary.MTD.ActiveUsers}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.Summary.SPLM.ActiveUsers}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.Summary.MTDvsSPLMChange.ActiveUsersChange}}</td>
				</tr>
				<tr>
					<td style="padding: 15px; border: 1px solid #dee2e6;"><strong>Total Bets</strong></td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.Summary.TotalBets}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.Summary.AvgBets}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.Summary.BestDay.BetCount}} ({{.Summary.BestDay.Date}})</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.Summary.WorstDay.BetCount}} ({{.Summary.WorstDay.Date}})</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.Summary.MTD.BetCount}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.Summary.SPLM.BetCount}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.Summary.MTDvsSPLMChange.BetCountChange}}</td>
				</tr>
				<tr style="background-color: #f8f9fa;">
					<td style="padding: 15px; border: 1px solid #dee2e6;"><strong>Total Bet Amount (USD)</strong></td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">${{.Summary.TotalBetAmount}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">${{.Summary.AvgBetAmount}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">${{.Summary.BestDay.TotalBets}} ({{.Summary.BestDay.Date}})</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">${{.Summary.WorstDay.TotalBets}} ({{.Summary.WorstDay.Date}})</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">${{.Summary.MTD.TotalBets}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">${{.Summary.SPLM.TotalBets}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.Summary.MTDvsSPLMChange.TotalBetsChange}}</td>
				</tr>
				<tr>
					<td style="padding: 15px; border: 1px solid #dee2e6;"><strong>Total Win Amount (USD)</strong></td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">${{.Summary.TotalWinAmount}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">${{.Summary.AvgWinAmount}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">${{.Summary.BestDay.TotalWins}} ({{.Summary.BestDay.Date}})</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">${{.Summary.WorstDay.TotalWins}} ({{.Summary.WorstDay.Date}})</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">${{.Summary.MTD.TotalWins}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">${{.Summary.SPLM.TotalWins}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.Summary.MTDvsSPLMChange.TotalWinsChange}}</td>
				</tr>
				<tr style="background-color: #f8f9fa;">
					<td style="padding: 15px; border: 1px solid #dee2e6;"><strong>Total GGR (USD)</strong></td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">${{.Summary.TotalGGR}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">${{.Summary.AvgGGR}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">${{.Summary.BestDay.NetRevenue}} ({{.Summary.BestDay.Date}})</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">${{.Summary.WorstDay.NetRevenue}} ({{.Summary.WorstDay.Date}})</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">${{.Summary.MTD.NetRevenue}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">${{.Summary.SPLM.NetRevenue}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.Summary.MTDvsSPLMChange.NetRevenueChange}}</td>
				</tr>
				<tr>
					<td style="padding: 15px; border: 1px solid #dee2e6;"><strong>Total Deposits (USD)</strong></td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">${{.Summary.TotalDeposits}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">${{.Summary.AvgDeposits}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">${{.Summary.BestDay.TotalDeposits}} ({{.Summary.BestDay.Date}})</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">${{.Summary.WorstDay.TotalDeposits}} ({{.Summary.WorstDay.Date}})</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">${{.Summary.MTD.TotalDeposits}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">${{.Summary.SPLM.TotalDeposits}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.Summary.MTDvsSPLMChange.TotalDepositsChange}}</td>
				</tr>
				<tr style="background-color: #f8f9fa;">
					<td style="padding: 15px; border: 1px solid #dee2e6;"><strong>Total Withdrawals (USD)</strong></td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">${{.Summary.TotalWithdrawals}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">${{.Summary.AvgWithdrawals}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">${{.Summary.BestDay.TotalWithdrawals}} ({{.Summary.BestDay.Date}})</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">${{.Summary.WorstDay.TotalWithdrawals}} ({{.Summary.WorstDay.Date}})</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">${{.Summary.MTD.TotalWithdrawals}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">${{.Summary.SPLM.TotalWithdrawals}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.Summary.MTDvsSPLMChange.TotalWithdrawalsChange}}</td>
				</tr>
				</table>
			</div>
		</div>

		<div class="daily-breakdown" style="margin: 30px 0;">
			<h3 style="color: #2c3e50; border-bottom: 2px solid #3498db; padding-bottom: 10px;">Daily Breakdown</h3>
			<div style="overflow-x: auto; margin: 20px 0; border: 1px solid #dee2e6; border-radius: 8px;">
				<table style="min-width: 800px; width: 100%; border-collapse: collapse; margin: 0;">
				<tr style="background-color: #3498db; color: white;">
					<th style="padding: 15px; text-align: center; border: 1px solid #2980b9;">Date</th>
					<th style="padding: 15px; text-align: right; border: 1px solid #2980b9;">New Users</th>
					<th style="padding: 15px; text-align: right; border: 1px solid #2980b9;">Unique Depositors</th>
					<th style="padding: 15px; text-align: right; border: 1px solid #2980b9;">Unique Withdrawers</th>
					<th style="padding: 15px; text-align: right; border: 1px solid #2980b9;">Active Users</th>
					<th style="padding: 15px; text-align: right; border: 1px solid #2980b9;">Bets</th>
					<th style="padding: 15px; text-align: right; border: 1px solid #2980b9;">Bet Amount</th>
					<th style="padding: 15px; text-align: right; border: 1px solid #2980b9;">GGR</th>
				</tr>
				{{range .DailyReports}}
				<tr style="{{if odd .Index}}background-color: #f8f9fa;{{end}}">
					<td style="padding: 15px; text-align: center; border: 1px solid #dee2e6;">{{.Date}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.NewUsers}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.UniqueDepositors}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.UniqueWithdrawers}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.ActiveUsers}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">{{.BetCount}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">${{.TotalBets}}</td>
					<td style="padding: 15px; text-align: right; border: 1px solid #dee2e6;">${{.NetRevenue}}</td>
				</tr>
				{{end}}
				</table>
			</div>
		</div>
	</div>`

	// Create template function for checking odd numbers
	funcMap := template.FuncMap{
		"odd": func(n int) bool {
			return n%2 == 1
		},
	}

	tmplParsed := template.Must(template.New("weekly_report").Funcs(funcMap).Parse(tmpl))

	// Calculate summary data
	summary := d.calculateWeeklySummary(reports)

	// Prepare daily reports data
	var dailyReports []struct {
		Index             int
		Date              string
		NewUsers          uint64
		UniqueDepositors  uint64
		UniqueWithdrawers uint64
		ActiveUsers       uint64
		BetCount          uint64
		TotalBets         string
		NetRevenue        string
	}

	for i, report := range reports {
		dailyReports = append(dailyReports, struct {
			Index             int
			Date              string
			NewUsers          uint64
			UniqueDepositors  uint64
			UniqueWithdrawers uint64
			ActiveUsers       uint64
			BetCount          uint64
			TotalBets         string
			NetRevenue        string
		}{
			Index:             i,
			Date:              report.Date.Format("Jan 2"),
			NewUsers:          report.NewUsers,
			UniqueDepositors:  report.UniqueDepositors,
			UniqueWithdrawers: report.UniqueWithdrawers,
			ActiveUsers:       report.ActiveUsers,
			BetCount:          report.BetCount,
			TotalBets:         report.TotalBets.StringFixed(2),
			NetRevenue:        report.NetRevenue.StringFixed(2),
		})
	}

	// Prepare data for the template
	data := struct {
		StartDate    string
		EndDate      string
		DaysCount    int
		Summary      WeeklySummary
		DailyReports []struct {
			Index             int
			Date              string
			NewUsers          uint64
			UniqueDepositors  uint64
			UniqueWithdrawers uint64
			ActiveUsers       uint64
			BetCount          uint64
			TotalBets         string
			NetRevenue        string
		}
	}{
		StartDate:    reports[len(reports)-1].Date.Format("January 2, 2006"),
		EndDate:      reports[0].Date.Format("January 2, 2006"),
		DaysCount:    len(reports),
		Summary:      summary,
		DailyReports: dailyReports,
	}

	var buf bytes.Buffer
	if err := tmplParsed.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute weekly report template: %w", err)
	}

	return buf.String(), nil
}

// WeeklySummary represents summary data for weekly reports
type WeeklySummary struct {
	TotalNewUsers          uint64
	AvgNewUsers            string
	TotalUniqueDepositors  uint64
	AvgUniqueDepositors    string
	TotalUniqueWithdrawers uint64
	AvgUniqueWithdrawers   string
	TotalActiveUsers       uint64
	AvgActiveUsers         string
	TotalBets              uint64
	AvgBets                string
	TotalBetAmount         string
	AvgBetAmount           string
	TotalWinAmount         string
	AvgWinAmount           string
	TotalGGR               string
	AvgGGR                 string
	TotalDeposits          string
	AvgDeposits            string
	TotalWithdrawals       string
	AvgWithdrawals         string
	BestDay                DailyReportSummary
	WorstDay               DailyReportSummary
	MTD                    struct {
		NewUsers          uint64
		UniqueDepositors  uint64
		UniqueWithdrawers uint64
		ActiveUsers       uint64
		BetCount          uint64
		TotalBets         string
		TotalWins         string
		NetRevenue        string
		TotalDeposits     string
		TotalWithdrawals  string
	}
	SPLM struct {
		NewUsers          uint64
		UniqueDepositors  uint64
		UniqueWithdrawers uint64
		ActiveUsers       uint64
		BetCount          uint64
		TotalBets         string
		TotalWins         string
		NetRevenue        string
		TotalDeposits     string
		TotalWithdrawals  string
	}
	MTDvsSPLMChange dto.DailyReportComparison
}

// DailyReportSummary represents summary data for a single day
type DailyReportSummary struct {
	Date              string
	NewUsers          uint64
	UniqueDepositors  uint64
	UniqueWithdrawers uint64
	ActiveUsers       uint64
	BetCount          uint64
	TotalBets         string
	TotalWins         string
	NetRevenue        string
	TotalDeposits     string
	TotalWithdrawals  string
}

// calculateWeeklySummary calculates summary statistics for weekly reports
func (d *DailyReportEmailServiceImpl) calculateWeeklySummary(reports []*dto.EnhancedDailyReport) WeeklySummary {
	if len(reports) == 0 {
		return WeeklySummary{}
	}

	var totalNewUsers, totalUniqueDepositors, totalUniqueWithdrawers, totalActiveUsers, totalBets uint64
	var totalBetAmount, totalWinAmount, totalGGR, totalDeposits, totalWithdrawals decimal.Decimal

	bestDay := reports[0]
	worstDay := reports[0]

	for _, report := range reports {
		totalNewUsers += report.NewUsers
		totalUniqueDepositors += report.UniqueDepositors
		totalUniqueWithdrawers += report.UniqueWithdrawers
		totalActiveUsers += report.ActiveUsers
		totalBets += report.BetCount
		totalBetAmount = totalBetAmount.Add(report.TotalBets)
		totalWinAmount = totalWinAmount.Add(report.TotalWins)
		totalGGR = totalGGR.Add(report.NetRevenue)
		totalDeposits = totalDeposits.Add(report.TotalDeposits)
		totalWithdrawals = totalWithdrawals.Add(report.TotalWithdrawals)

		// Track best and worst days by GGR
		if report.NetRevenue.GreaterThan(bestDay.NetRevenue) {
			bestDay = report
		}
		if report.NetRevenue.LessThan(worstDay.NetRevenue) {
			worstDay = report
		}
	}

	daysCount := decimal.NewFromInt(int64(len(reports)))

	// Get MTD, SPLM, and MTDvsSPLMChange from the most recent report (first in the array)
	latestReport := reports[0]

	return WeeklySummary{
		TotalNewUsers:          totalNewUsers,
		AvgNewUsers:            decimal.NewFromInt(int64(totalNewUsers)).Div(daysCount).StringFixed(1),
		TotalUniqueDepositors:  totalUniqueDepositors,
		AvgUniqueDepositors:    decimal.NewFromInt(int64(totalUniqueDepositors)).Div(daysCount).StringFixed(1),
		TotalUniqueWithdrawers: totalUniqueWithdrawers,
		AvgUniqueWithdrawers:   decimal.NewFromInt(int64(totalUniqueWithdrawers)).Div(daysCount).StringFixed(1),
		TotalActiveUsers:       totalActiveUsers,
		AvgActiveUsers:         decimal.NewFromInt(int64(totalActiveUsers)).Div(daysCount).StringFixed(1),
		TotalBets:              totalBets,
		AvgBets:                decimal.NewFromInt(int64(totalBets)).Div(daysCount).StringFixed(1),
		TotalBetAmount:         totalBetAmount.StringFixed(2),
		AvgBetAmount:           totalBetAmount.Div(daysCount).StringFixed(2),
		TotalWinAmount:         totalWinAmount.StringFixed(2),
		AvgWinAmount:           totalWinAmount.Div(daysCount).StringFixed(2),
		TotalGGR:               totalGGR.StringFixed(2),
		AvgGGR:                 totalGGR.Div(daysCount).StringFixed(2),
		TotalDeposits:          totalDeposits.StringFixed(2),
		AvgDeposits:            totalDeposits.Div(daysCount).StringFixed(2),
		TotalWithdrawals:       totalWithdrawals.StringFixed(2),
		AvgWithdrawals:         totalWithdrawals.Div(daysCount).StringFixed(2),
		MTD: struct {
			NewUsers          uint64
			UniqueDepositors  uint64
			UniqueWithdrawers uint64
			ActiveUsers       uint64
			BetCount          uint64
			TotalBets         string
			TotalWins         string
			NetRevenue        string
			TotalDeposits     string
			TotalWithdrawals  string
		}{
			NewUsers:          latestReport.MTD.NewUsers,
			UniqueDepositors:  latestReport.MTD.UniqueDepositors,
			UniqueWithdrawers: latestReport.MTD.UniqueWithdrawers,
			ActiveUsers:       latestReport.MTD.ActiveUsers,
			BetCount:          latestReport.MTD.BetCount,
			TotalBets:         latestReport.MTD.TotalBets.StringFixed(2),
			TotalWins:         latestReport.MTD.TotalWins.StringFixed(2),
			NetRevenue:        latestReport.MTD.NetRevenue.StringFixed(2),
			TotalDeposits:     latestReport.MTD.TotalDeposits.StringFixed(2),
			TotalWithdrawals:  latestReport.MTD.TotalWithdrawals.StringFixed(2),
		},
		SPLM: struct {
			NewUsers          uint64
			UniqueDepositors  uint64
			UniqueWithdrawers uint64
			ActiveUsers       uint64
			BetCount          uint64
			TotalBets         string
			TotalWins         string
			NetRevenue        string
			TotalDeposits     string
			TotalWithdrawals  string
		}{
			NewUsers:          latestReport.SPLM.NewUsers,
			UniqueDepositors:  latestReport.SPLM.UniqueDepositors,
			UniqueWithdrawers: latestReport.SPLM.UniqueWithdrawers,
			ActiveUsers:       latestReport.SPLM.ActiveUsers,
			BetCount:          latestReport.SPLM.BetCount,
			TotalBets:         latestReport.SPLM.TotalBets.StringFixed(2),
			TotalWins:         latestReport.SPLM.TotalWins.StringFixed(2),
			NetRevenue:        latestReport.SPLM.NetRevenue.StringFixed(2),
			TotalDeposits:     latestReport.SPLM.TotalDeposits.StringFixed(2),
			TotalWithdrawals:  latestReport.SPLM.TotalWithdrawals.StringFixed(2),
		},
		MTDvsSPLMChange: latestReport.MTDvsSPLMChange,
		BestDay: DailyReportSummary{
			Date:              bestDay.Date.Format("Jan 2"),
			NewUsers:          bestDay.NewUsers,
			UniqueDepositors:  bestDay.UniqueDepositors,
			UniqueWithdrawers: bestDay.UniqueWithdrawers,
			ActiveUsers:       bestDay.ActiveUsers,
			BetCount:          bestDay.BetCount,
			TotalBets:         bestDay.TotalBets.StringFixed(2),
			TotalWins:         bestDay.TotalWins.StringFixed(2),
			NetRevenue:        bestDay.NetRevenue.StringFixed(2),
			TotalDeposits:     bestDay.TotalDeposits.StringFixed(2),
			TotalWithdrawals:  bestDay.TotalWithdrawals.StringFixed(2),
		},
		WorstDay: DailyReportSummary{
			Date:              worstDay.Date.Format("Jan 2"),
			NewUsers:          worstDay.NewUsers,
			UniqueDepositors:  worstDay.UniqueDepositors,
			UniqueWithdrawers: worstDay.UniqueWithdrawers,
			ActiveUsers:       worstDay.ActiveUsers,
			BetCount:          worstDay.BetCount,
			TotalBets:         worstDay.TotalBets.StringFixed(2),
			TotalWins:         worstDay.TotalWins.StringFixed(2),
			NetRevenue:        worstDay.NetRevenue.StringFixed(2),
			TotalDeposits:     worstDay.TotalDeposits.StringFixed(2),
			TotalWithdrawals:  worstDay.TotalWithdrawals.StringFixed(2),
		},
	}
}
