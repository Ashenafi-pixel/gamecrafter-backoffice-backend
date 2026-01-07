package report

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/constant/model/response"
	"github.com/tucanbit/internal/handler"
	"github.com/tucanbit/internal/module"
	"go.uber.org/zap"
)

type report struct {
	log          *zap.Logger
	reportModule module.Report
	userModule   module.User
}

func Init(reportModule module.Report, userModule module.User, log *zap.Logger) handler.Report {
	return &report{
		log:          log,
		reportModule: reportModule,
		userModule:   userModule,
	}
}

// GetDailyReport Get Daily Report.
//
//	@Summary		GetDailyReport
//	@Description	Returns daily aggregated report
//	@Tags			Report
//	@Accept			json
//	@Produce		json
//	@Param			date	query		string	true	"Report Date (YYYY-MM-DD)"
//	@Success		200		{object}	dto.DailyReportRes
//	@Failure		400		{object}	response.ErrorResponse
//	@Router			/api/admin/report/daily [get]
func (r *report) GetDailyReport(ctx *gin.Context) {
	var req dto.DailyReportReq

	if err := ctx.ShouldBindQuery(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = ctx.Error(err)
		return
	}

	reportRes, err := r.reportModule.DailyReport(ctx, req)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	response.SendSuccessResponse(ctx, http.StatusOK, reportRes)
}

// GetDuplicateIPAccounts Get Duplicate IP Accounts Report.
//
//	@Summary		GetDuplicateIPAccounts
//	@Description	Returns a report of accounts created from the same IP address (potential bot/spam accounts)
//	@Tags			Report
//	@Accept			json
//	@Produce		json
//	@Success		200		{array}	dto.DuplicateIPAccountsReport
//	@Failure		400		{object}	response.ErrorResponse
//	@Router			/api/admin/report/duplicate-ip-accounts [get]
func (r *report) GetDuplicateIPAccounts(ctx *gin.Context) {
	reportRes, err := r.reportModule.GetDuplicateIPAccounts(ctx.Request.Context())
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	response.SendSuccessResponse(ctx, http.StatusOK, reportRes)
}

// SuspendAccountsByIP suspends all accounts created from a specific IP address.
//
//	@Summary		SuspendAccountsByIP
//	@Description	Suspends all accounts that were created from the specified IP address
//	@Tags			Report
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string					true	"Bearer <token>"
//	@Param			request			body		dto.SuspendAccountsByIPReq	true	"Suspend accounts request"
//	@Success		200				{object}	dto.SuspendAccountsByIPRes
//	@Failure		400				{object}	response.ErrorResponse
//	@Failure		401				{object}	response.ErrorResponse
//	@Router			/api/admin/report/duplicate-ip-accounts/suspend [post]
func (r *report) SuspendAccountsByIP(ctx *gin.Context) {
	var req dto.SuspendAccountsByIPReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = ctx.Error(err)
		return
	}

	// Get admin user ID
	userIDStr := ctx.GetString("user-id")
	adminID, err := uuid.Parse(userIDStr)
	if err != nil {
		r.log.Warn("Failed to parse admin user ID", zap.Error(err), zap.String("userID", userIDStr))
		err = errors.ErrInvalidUserInput.Wrap(err, "invalid admin user ID")
		_ = ctx.Error(err)
		return
	}

	// Get user IDs from IP
	userIDs, err := r.reportModule.SuspendAccountsByIP(ctx.Request.Context(), req, adminID)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	if len(userIDs) == 0 {
		response.SendSuccessResponse(ctx, http.StatusOK, dto.SuspendAccountsByIPRes{
			Message:           "No accounts found for this IP address",
			IPAddress:         req.IPAddress,
			AccountsSuspended: 0,
			UserIDs:           []string{},
		})
		return
	}

	// Suspend each account
	suspendedCount := 0
	failedCount := 0
	userIDStrings := make([]string, 0, len(userIDs))

	for _, userID := range userIDs {
		// Block account with COMPLETE type and PERMANENT duration
		blockReq := dto.AccountBlockReq{
			UserID:    userID,
			BlockedBy: adminID,
			Type:      constant.BLOCK_TYPE_COMPLETE, // Block all access
			Duration:  constant.BLOCK_DURATION_PERMANENT,
			Reason: func() string {
				if req.Reason != "" {
					return req.Reason
				}
				return fmt.Sprintf("Suspended due to duplicate IP address: %s", req.IPAddress)
			}(),
			Note: req.Note,
		}

		_, err := r.userModule.BlockUser(ctx.Request.Context(), blockReq)
		if err != nil {
			r.log.Error("Failed to suspend account",
				zap.Error(err),
				zap.String("user_id", userID.String()),
				zap.String("ip_address", req.IPAddress))
			failedCount++
		} else {
			suspendedCount++
			userIDStrings = append(userIDStrings, userID.String())
		}
	}

	res := dto.SuspendAccountsByIPRes{
		Message:           fmt.Sprintf("Suspended %d accounts (failed: %d)", suspendedCount, failedCount),
		IPAddress:         req.IPAddress,
		AccountsSuspended: suspendedCount,
		UserIDs:           userIDStrings,
	}

	response.SendSuccessResponse(ctx, http.StatusOK, res)
}

// GetBigWinners Get Big Winners Report.
//
//	@Summary		GetBigWinners
//	@Description	Returns a report of players who have won large amounts over a given period
//	@Tags			Report
//	@Accept			json
//	@Produce		json
//	@Param			page				query		int		false	"Page number" default(1)
//	@Param			per_page			query		int		false	"Items per page" default(20)
//	@Param			date_from			query		string	false	"Start date (YYYY-MM-DD or YYYY-MM-DDTHH:MM:SS)"
//	@Param			date_to				query		string	false	"End date (YYYY-MM-DD or YYYY-MM-DDTHH:MM:SS)"
//	@Param			brand_id			query		string	false	"Brand ID filter"
//	@Param			game_provider		query		string	false	"Game provider filter"
//	@Param			game_id				query		string	false	"Game ID filter"
//	@Param			player_search		query		string	false	"Search by username, email, or user ID"
//	@Param			min_win_threshold	query		float	false	"Minimum win threshold" default(100)
//	@Param			bet_type			query		string	false	"Bet type: cash, bonus, both" default(both)
//	@Param			sort_by				query		string	false	"Sort by: win_amount, net_win, multiplier, date" default(date)
//	@Param			sort_order			query		string	false	"Sort order: asc, desc" default(desc)
//	@Success		200					{object}	dto.BigWinnersReportRes
//	@Failure		400					{object}	response.ErrorResponse
//	@Router			/api/admin/report/big-winners [get]
func (r *report) GetBigWinners(ctx *gin.Context) {
	var req dto.BigWinnersReportReq

	if err := ctx.ShouldBindQuery(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = ctx.Error(err)
		return
	}

	// Get user ID from context (for future brand filtering)
	_ = ctx.GetString("user-id")
	// TODO: Implement brand access control based on user's roles

	// Get user's accessible brand IDs
	// For now, we'll get the user's brand_id and if they're super admin, allow all brands
	// TODO: Implement proper brand access control based on roles
	userBrandIDs := []uuid.UUID{} // Empty means all brands (for super admin)
	// If not super admin, filter by user's brand_id
	// This should be implemented based on your brand access control logic

	reportRes, err := r.reportModule.GetBigWinners(ctx.Request.Context(), req, userBrandIDs)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	response.SendSuccessResponse(ctx, http.StatusOK, reportRes)
}

// ExportBigWinners Export Big Winners Report.
//
//	@Summary		ExportBigWinners
//	@Description	Exports big winners report to CSV or Excel
//	@Tags			Report
//	@Accept			json
//	@Produce		application/octet-stream
//	@Param			format				query		string	true	"Export format: csv or excel"
//	@Param			date_from			query		string	false	"Start date"
//	@Param			date_to				query		string	false	"End date"
//	@Param			brand_id			query		string	false	"Brand ID filter"
//	@Param			game_provider		query		string	false	"Game provider filter"
//	@Param			game_id				query		string	false	"Game ID filter"
//	@Param			player_search		query		string	false	"Search by username, email, or user ID"
//	@Param			min_win_threshold	query		float	false	"Minimum win threshold"
//	@Param			bet_type			query		string	false	"Bet type: cash, bonus, both"
//	@Success		200					file
//	@Failure		400					{object}	response.ErrorResponse
//	@Router			/api/admin/report/big-winners/export [get]
func (r *report) ExportBigWinners(ctx *gin.Context) {
	// Get format parameter
	format := ctx.Query("format")
	if format != "csv" && format != "excel" {
		err := errors.ErrInvalidUserInput.New("format must be 'csv' or 'excel'")
		_ = ctx.Error(err)
		return
	}

	// Bind request (same as GetBigWinners but with higher limit for export)
	var req dto.BigWinnersReportReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = ctx.Error(err)
		return
	}

	// Set high limit for export (max 50,000 rows)
	req.Page = 1
	req.PerPage = 50000

	// Get user ID (for future brand filtering)
	_ = ctx.GetString("user-id")
	// TODO: Implement brand access control

	userBrandIDs := []uuid.UUID{} // Empty means all brands

	reportRes, err := r.reportModule.GetBigWinners(ctx.Request.Context(), req, userBrandIDs)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	// Generate export file
	// TODO: Implement CSV/Excel generation
	// For now, return JSON (will implement proper export later)
	ctx.Header("Content-Type", "application/json")
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=big-winners-%s.json", time.Now().Format("2006-01-02")))
	response.SendSuccessResponse(ctx, http.StatusOK, reportRes)
}

// GetPlayerMetrics Get Player Metrics Report.
//
//	@Summary		GetPlayerMetrics
//	@Description	Returns player-level metrics report with key performance indicators
//	@Tags			Report
//	@Accept			json
//	@Produce		json
//	@Param			page				query		int		false	"Page number" default(1)
//	@Param			per_page			query		int		false	"Items per page" default(20)
//	@Param			player_search		query		string	false	"Search by username, email, or user ID"
//	@Param			brand_id			query		string	false	"Brand ID filter"
//	@Param			currency			query		string	false	"Currency filter"
//	@Param			country				query		string	false	"Country filter"
//	@Param			registration_from	query		string	false	"Registration date from (YYYY-MM-DD)"
//	@Param			registration_to		query		string	false	"Registration date to (YYYY-MM-DD)"
//	@Param			last_active_from	query		string	false	"Last active date from (YYYY-MM-DD)"
//	@Param			last_active_to		query		string	false	"Last active date to (YYYY-MM-DD)"
//	@Param			has_deposited		query		bool	false	"Filter by has deposited"
//	@Param			has_withdrawn		query		bool	false	"Filter by has withdrawn"
//	@Param			min_total_deposits	query		float	false	"Minimum total deposits"
//	@Param			max_total_deposits	query		float	false	"Maximum total deposits"
//	@Param			min_total_wagers	query		float	false	"Minimum total wagers"
//	@Param			max_total_wagers	query		float	false	"Maximum total wagers"
//	@Param			min_net_result		query		float	false	"Minimum net result"
//	@Param			max_net_result		query		float	false	"Maximum net result"
//	@Param			sort_by				query		string	false	"Sort by: deposits, wagering, net_loss, activity, registration" default(deposits)
//	@Param			sort_order			query		string	false	"Sort order: asc, desc" default(desc)
//	@Success		200					{object}	dto.PlayerMetricsReportRes
//	@Failure		400					{object}	response.ErrorResponse
//	@Router			/api/admin/report/player-metrics [get]
func (r *report) GetPlayerMetrics(ctx *gin.Context) {
	var req dto.PlayerMetricsReportReq

	if err := ctx.ShouldBindQuery(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = ctx.Error(err)
		return
	}

	// Get user ID (for future brand filtering)
	_ = ctx.GetString("user-id")
	// TODO: Implement brand access control
	userBrandIDs := []uuid.UUID{} // Empty means all brands

	reportRes, err := r.reportModule.GetPlayerMetrics(ctx.Request.Context(), req, userBrandIDs)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	response.SendSuccessResponse(ctx, http.StatusOK, reportRes)
}

// GetPlayerTransactions Get Player Transactions (Drill-Down).
//
//	@Summary		GetPlayerTransactions
//	@Description	Returns detailed transactions for a specific player (drill-down view)
//	@Tags			Report
//	@Accept			json
//	@Produce		json
//	@Param			player_id			path		string	true	"Player ID"
//	@Param			page				query		int		false	"Page number" default(1)
//	@Param			per_page			query		int		false	"Items per page" default(50)
//	@Param			date_from			query		string	false	"Start date"
//	@Param			date_to				query		string	false	"End date"
//	@Param			transaction_type	query		string	false	"Transaction type: deposit, withdrawal, bet, win, bonus, adjustment"
//	@Param			game_provider		query		string	false	"Game provider filter"
//	@Param			game_id				query		string	false	"Game ID filter"
//	@Param			min_amount			query		float	false	"Minimum amount"
//	@Param			max_amount			query		float	false	"Maximum amount"
//	@Param			sort_by				query		string	false	"Sort by: date, amount, net, game" default(date)
//	@Param			sort_order			query		string	false	"Sort order: asc, desc" default(desc)
//	@Success		200					{object}	dto.PlayerTransactionsRes
//	@Failure		400					{object}	response.ErrorResponse
//	@Router			/api/admin/report/player-metrics/:player_id/transactions [get]
func (r *report) GetPlayerTransactions(ctx *gin.Context) {
	var req dto.PlayerTransactionsReq

	// Get player_id from path
	playerIDStr := ctx.Param("player_id")
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, "invalid player ID")
		_ = ctx.Error(err)
		return
	}
	req.PlayerID = playerID

	if err := ctx.ShouldBindQuery(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = ctx.Error(err)
		return
	}

	transactionsRes, err := r.reportModule.GetPlayerTransactions(ctx.Request.Context(), req)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	response.SendSuccessResponse(ctx, http.StatusOK, transactionsRes)
}

// ExportPlayerMetrics Export Player Metrics Report.
//
//	@Summary		ExportPlayerMetrics
//	@Description	Exports player metrics report to CSV
//	@Tags			Report
//	@Accept			json
//	@Produce		application/octet-stream
//	@Param			format	query		string	true	"Export format: csv"
//	@Success		200		file
//	@Failure		400		{object}	response.ErrorResponse
//	@Router			/api/admin/report/player-metrics/export [get]
func (r *report) ExportPlayerMetrics(ctx *gin.Context) {
	format := ctx.Query("format")
	if format != "csv" {
		err := errors.ErrInvalidUserInput.New("format must be 'csv'")
		_ = ctx.Error(err)
		return
	}

	var req dto.PlayerMetricsReportReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = ctx.Error(err)
		return
	}

	req.Page = 1
	req.PerPage = 50000

	_ = ctx.GetString("user-id")
	userBrandIDs := []uuid.UUID{}

	reportRes, err := r.reportModule.GetPlayerMetrics(ctx.Request.Context(), req, userBrandIDs)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	// Generate CSV
	var csvBuilder strings.Builder

	// Write CSV header
	csvBuilder.WriteString("Player ID,Username,Email,Brand ID,Brand Name,Country,Registration Date,Last Activity,Main Balance,Currency,Total Deposits,Total Withdrawals,Net Deposits,Total Wagered,Total Won,Rakeback Earned,Rakeback Claimed,Net Gaming Result,Number of Sessions,Number of Bets,Account Status,Device Type,KYC Status,First Deposit Date,Last Deposit Date\n")

	// Write data rows
	for _, metric := range reportRes.Data {
		// Helper function to safely format CSV field
		formatCSVField := func(value interface{}) string {
			if value == nil {
				return ""
			}
			str := fmt.Sprintf("%v", value)
			// Escape quotes and wrap in quotes if contains comma, newline, or quote
			if strings.Contains(str, ",") || strings.Contains(str, "\n") || strings.Contains(str, `"`) {
				str = strings.ReplaceAll(str, `"`, `""`)
				return `"` + str + `"`
			}
			return str
		}

		// Format nullable fields
		email := ""
		if metric.Email != nil {
			email = *metric.Email
		}
		brandID := ""
		if metric.BrandID != nil {
			brandID = metric.BrandID.String()
		}
		brandName := ""
		if metric.BrandName != nil {
			brandName = *metric.BrandName
		}
		country := ""
		if metric.Country != nil {
			country = *metric.Country
		}
		lastActivity := ""
		if metric.LastActivity != nil {
			lastActivity = metric.LastActivity.Format("2006-01-02 15:04:05")
		}
		deviceType := ""
		if metric.DeviceType != nil {
			deviceType = *metric.DeviceType
		}
		kycStatus := ""
		if metric.KYCStatus != nil {
			kycStatus = *metric.KYCStatus
		}
		firstDepositDate := ""
		if metric.FirstDepositDate != nil {
			firstDepositDate = metric.FirstDepositDate.Format("2006-01-02 15:04:05")
		}
		lastDepositDate := ""
		if metric.LastDepositDate != nil {
			lastDepositDate = metric.LastDepositDate.Format("2006-01-02 15:04:05")
		}

		// Write row
		csvBuilder.WriteString(fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%d,%d,%s,%s,%s,%s,%s\n",
			formatCSVField(metric.PlayerID.String()),
			formatCSVField(metric.Username),
			formatCSVField(email),
			formatCSVField(brandID),
			formatCSVField(brandName),
			formatCSVField(country),
			formatCSVField(metric.RegistrationDate.Format("2006-01-02 15:04:05")),
			formatCSVField(lastActivity),
			formatCSVField(metric.MainBalance.String()),
			formatCSVField(metric.Currency),
			formatCSVField(metric.TotalDeposits.String()),
			formatCSVField(metric.TotalWithdrawals.String()),
			formatCSVField(metric.NetDeposits.String()),
			formatCSVField(metric.TotalWagered.String()),
			formatCSVField(metric.TotalWon.String()),
			formatCSVField(metric.RakebackEarned.String()),
			formatCSVField(metric.RakebackClaimed.String()),
			formatCSVField(metric.NetGamingResult.String()),
			metric.NumberOfSessions,
			metric.NumberOfBets,
			formatCSVField(metric.AccountStatus),
			formatCSVField(deviceType),
			formatCSVField(kycStatus),
			formatCSVField(firstDepositDate),
			formatCSVField(lastDepositDate),
		))
	}

	// Set headers for CSV download
	ctx.Header("Content-Type", "text/csv; charset=utf-8")
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=player-metrics-%s.csv", time.Now().Format("2006-01-02")))
	ctx.Header("Content-Transfer-Encoding", "binary")

	// Write CSV content
	ctx.Data(http.StatusOK, "text/csv; charset=utf-8", []byte(csvBuilder.String()))
}

// ExportPlayerTransactions Export Player Transactions.
//
//	@Summary		ExportPlayerTransactions
//	@Description	Exports player transactions to CSV
//	@Tags			Report
//	@Accept			json
//	@Produce		application/octet-stream
//	@Param			player_id	path		string	true	"Player ID"
//	@Param			format		query		string	true	"Export format: csv"
//	@Success		200			file
//	@Failure		400			{object}	response.ErrorResponse
//	@Router			/api/admin/report/player-metrics/:player_id/transactions/export [get]
func (r *report) ExportPlayerTransactions(ctx *gin.Context) {
	format := ctx.Query("format")
	if format != "csv" {
		err := errors.ErrInvalidUserInput.New("format must be 'csv'")
		_ = ctx.Error(err)
		return
	}

	var req dto.PlayerTransactionsReq
	playerIDStr := ctx.Param("player_id")
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, "invalid player ID")
		_ = ctx.Error(err)
		return
	}
	req.PlayerID = playerID

	if err := ctx.ShouldBindQuery(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = ctx.Error(err)
		return
	}

	req.Page = 1
	req.PerPage = 50000

	transactionsRes, err := r.reportModule.GetPlayerTransactions(ctx.Request.Context(), req)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.Header("Content-Type", "application/json")
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=player-transactions-%s-%s.json", playerIDStr, time.Now().Format("2006-01-02")))
	response.SendSuccessResponse(ctx, http.StatusOK, transactionsRes)
}

// GetCountryMetrics Get Country Report.
//
//	@Summary		GetCountryMetrics
//	@Description	Returns country-level aggregated metrics report
//	@Tags			Report
//	@Accept			json
//	@Produce		json
//	@Param			page					query		int		false	"Page number" default(1)
//	@Param			per_page				query		int		false	"Items per page" default(20)
//	@Param			date_from				query		string	true	"Start date (YYYY-MM-DD)"
//	@Param			date_to					query		string	true	"End date (YYYY-MM-DD)"
//	@Param			brand_id				query		string	false	"Brand ID filter"
//	@Param			currency				query		string	false	"Currency filter"
//	@Param			countries				query		[]string false	"Countries filter (multi-select)"
//	@Param			acquisition_channel		query		string	false	"Acquisition channel filter"
//	@Param			user_type				query		string	false	"User type: depositors, all, active" default(all)
//	@Param			sort_by					query		string	false	"Sort by: deposits, ngr, active_users, alphabetical" default(ngr)
//	@Param			sort_order				query		string	false	"Sort order: asc, desc" default(desc)
//	@Param			convert_to_base_currency	query		bool	false	"Convert to base currency"
//	@Success		200						{object}	dto.CountryReportRes
//	@Failure		400						{object}	response.ErrorResponse
//	@Router			/api/admin/report/country [get]
func (r *report) GetCountryMetrics(ctx *gin.Context) {
	var req dto.CountryReportReq

	if err := ctx.ShouldBindQuery(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = ctx.Error(err)
		return
	}

	// Handle countries array parameter
	countriesParam := ctx.QueryArray("countries")
	if len(countriesParam) > 0 {
		req.Countries = countriesParam
	}

	_ = ctx.GetString("user-id")
	userBrandIDs := []uuid.UUID{}

	reportRes, err := r.reportModule.GetCountryMetrics(ctx.Request.Context(), req, userBrandIDs)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	response.SendSuccessResponse(ctx, http.StatusOK, reportRes)
}

// GetCountryPlayers Get Country Players (Drill-Down).
//
//	@Summary		GetCountryPlayers
//	@Description	Returns players for a specific country (drill-down view)
//	@Tags			Report
//	@Accept			json
//	@Produce		json
//	@Param			country			path		string	true	"Country name"
//	@Param			page			query		int		false	"Page number" default(1)
//	@Param			per_page		query		int		false	"Items per page" default(50)
//	@Param			date_from		query		string	false	"Start date"
//	@Param			date_to			query		string	false	"End date"
//	@Param			min_deposits	query		float	false	"Minimum deposits"
//	@Param			max_deposits	query		float	false	"Maximum deposits"
//	@Param			activity_from	query		string	false	"Activity start date"
//	@Param			activity_to		query		string	false	"Activity end date"
//	@Param			kyc_status		query		string	false	"KYC status filter"
//	@Param			min_balance		query		float	false	"Minimum balance"
//	@Param			max_balance		query		float	false	"Maximum balance"
//	@Param			sort_by			query		string	false	"Sort by: deposits, ngr, activity, registration" default(ngr)
//	@Param			sort_order		query		string	false	"Sort order: asc, desc" default(desc)
//	@Success		200				{object}	dto.CountryPlayersRes
//	@Failure		400				{object}	response.ErrorResponse
//	@Router			/api/admin/report/country/:country/players [get]
func (r *report) GetCountryPlayers(ctx *gin.Context) {
	var req dto.CountryPlayersReq

	// Get country from path
	country := ctx.Param("country")
	if country == "" {
		err := errors.ErrInvalidUserInput.New("country parameter is required")
		_ = ctx.Error(err)
		return
	}
	req.Country = country

	if err := ctx.ShouldBindQuery(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = ctx.Error(err)
		return
	}

	_ = ctx.GetString("user-id")
	userBrandIDs := []uuid.UUID{}

	playersRes, err := r.reportModule.GetCountryPlayers(ctx.Request.Context(), req, userBrandIDs)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	response.SendSuccessResponse(ctx, http.StatusOK, playersRes)
}

// ExportCountryMetrics Export Country Report.
//
//	@Summary		ExportCountryMetrics
//	@Description	Exports country metrics report to CSV
//	@Tags			Report
//	@Accept			json
//	@Produce		application/octet-stream
//	@Param			format	query		string	true	"Export format: csv"
//	@Success		200		file
//	@Failure		400		{object}	response.ErrorResponse
//	@Router			/api/admin/report/country/export [get]
func (r *report) ExportCountryMetrics(ctx *gin.Context) {
	format := ctx.Query("format")
	if format != "csv" {
		err := errors.ErrInvalidUserInput.New("format must be 'csv'")
		_ = ctx.Error(err)
		return
	}

	var req dto.CountryReportReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = ctx.Error(err)
		return
	}

	countriesParam := ctx.QueryArray("countries")
	if len(countriesParam) > 0 {
		req.Countries = countriesParam
	}

	req.Page = 1
	req.PerPage = 50000

	_ = ctx.GetString("user-id")
	userBrandIDs := []uuid.UUID{}

	reportRes, err := r.reportModule.GetCountryMetrics(ctx.Request.Context(), req, userBrandIDs)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.Header("Content-Type", "application/json")
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=country-report-%s.json", time.Now().Format("2006-01-02")))
	response.SendSuccessResponse(ctx, http.StatusOK, reportRes)
}

// ExportCountryPlayers Export Country Players.
//
//	@Summary		ExportCountryPlayers
//	@Description	Exports country players to CSV
//	@Tags			Report
//	@Accept			json
//	@Produce		application/octet-stream
//	@Param			country	path		string	true	"Country name"
//	@Param			format	query		string	true	"Export format: csv"
//	@Success		200		file
//	@Failure		400		{object}	response.ErrorResponse
//	@Router			/api/admin/report/country/:country/players/export [get]
func (r *report) ExportCountryPlayers(ctx *gin.Context) {
	format := ctx.Query("format")
	if format != "csv" {
		err := errors.ErrInvalidUserInput.New("format must be 'csv'")
		_ = ctx.Error(err)
		return
	}

	var req dto.CountryPlayersReq
	country := ctx.Param("country")
	if country == "" {
		err := errors.ErrInvalidUserInput.New("country parameter is required")
		_ = ctx.Error(err)
		return
	}
	req.Country = country

	if err := ctx.ShouldBindQuery(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = ctx.Error(err)
		return
	}

	req.Page = 1
	req.PerPage = 50000

	_ = ctx.GetString("user-id")
	userBrandIDs := []uuid.UUID{}

	playersRes, err := r.reportModule.GetCountryPlayers(ctx.Request.Context(), req, userBrandIDs)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.Header("Content-Type", "application/json")
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=country-players-%s-%s.json", country, time.Now().Format("2006-01-02")))
	response.SendSuccessResponse(ctx, http.StatusOK, playersRes)
}

// GetGamePerformance Get Game Performance Report.
//
//	@Summary		GetGamePerformance
//	@Description	Returns game-level aggregated performance metrics report
//	@Tags			Report
//	@Accept			json
//	@Produce		json
//	@Param			page			query		int		false	"Page number" default(1)
//	@Param			per_page		query		int		false	"Items per page" default(20)
//	@Param			date_from		query		string	true	"Start date (YYYY-MM-DD)"
//	@Param			date_to			query		string	true	"End date (YYYY-MM-DD)"
//	@Param			brand_id		query		string	false	"Brand ID filter"
//	@Param			currency		query		string	false	"Currency filter"
//	@Param			game_provider	query		string	false	"Game provider filter"
//	@Param			game_id			query		string	false	"Game ID filter"
//	@Param			category		query		string	false	"Category filter"
//	@Param			sort_by			query		string	false	"Sort by: ggr, ngr, most_played, rtp, bet_volume" default(ggr)
//	@Param			sort_order		query		string	false	"Sort order: asc, desc" default(desc)
//	@Success		200				{object}	dto.GamePerformanceReportRes
//	@Failure		400				{object}	response.ErrorResponse
//	@Router			/api/admin/report/game-performance [get]
func (r *report) GetGamePerformance(ctx *gin.Context) {
	var req dto.GamePerformanceReportReq

	if err := ctx.ShouldBindQuery(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = ctx.Error(err)
		return
	}

	_ = ctx.GetString("user-id")
	userBrandIDs := []uuid.UUID{}

	reportRes, err := r.reportModule.GetGamePerformance(ctx.Request.Context(), req, userBrandIDs)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	response.SendSuccessResponse(ctx, http.StatusOK, reportRes)
}

// GetGamePlayers Get Game Players (Drill-Down).
//
//	@Summary		GetGamePlayers
//	@Description	Returns players for a specific game (drill-down view)
//	@Tags			Report
//	@Accept			json
//	@Produce		json
//	@Param			game_id			path		string	true	"Game ID"
//	@Param			page			query		int		false	"Page number" default(1)
//	@Param			per_page		query		int		false	"Items per page" default(50)
//	@Param			date_from		query		string	false	"Start date"
//	@Param			date_to			query		string	false	"End date"
//	@Param			currency		query		string	false	"Currency filter"
//	@Param			bet_type		query		string	false	"Bet type filter"
//	@Param			min_stake		query		float	false	"Minimum stake"
//	@Param			max_stake		query		float	false	"Maximum stake"
//	@Param			min_net_result	query		float	false	"Minimum net result"
//	@Param			max_net_result	query		float	false	"Maximum net result"
//	@Param			sort_by			query		string	false	"Sort by: total_stake, total_win, ngr, rounds, last_played" default(total_stake)
//	@Param			sort_order		query		string	false	"Sort order: asc, desc" default(desc)
//	@Success		200				{object}	dto.GamePlayersRes
//	@Failure		400				{object}	response.ErrorResponse
//	@Router			/api/admin/report/game-performance/:game_id/players [get]
func (r *report) GetGamePlayers(ctx *gin.Context) {
	var req dto.GamePlayersReq

	// Get game_id from path
	gameID := ctx.Param("game_id")
	if gameID == "" {
		err := errors.ErrInvalidUserInput.New("game_id parameter is required")
		_ = ctx.Error(err)
		return
	}
	req.GameID = gameID

	if err := ctx.ShouldBindQuery(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = ctx.Error(err)
		return
	}

	_ = ctx.GetString("user-id")
	userBrandIDs := []uuid.UUID{}

	playersRes, err := r.reportModule.GetGamePlayers(ctx.Request.Context(), req, userBrandIDs)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	response.SendSuccessResponse(ctx, http.StatusOK, playersRes)
}

// ExportGamePerformance Export Game Performance Report.
//
//	@Summary		ExportGamePerformance
//	@Description	Exports game performance report to CSV
//	@Tags			Report
//	@Accept			json
//	@Produce		application/octet-stream
//	@Param			format	query		string	true	"Export format: csv"
//	@Success		200		file
//	@Failure		400		{object}	response.ErrorResponse
//	@Router			/api/admin/report/game-performance/export [get]
func (r *report) ExportGamePerformance(ctx *gin.Context) {
	format := ctx.Query("format")
	if format != "csv" {
		err := errors.ErrInvalidUserInput.New("format must be 'csv'")
		_ = ctx.Error(err)
		return
	}

	var req dto.GamePerformanceReportReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = ctx.Error(err)
		return
	}

	req.Page = 1
	req.PerPage = 50000

	_ = ctx.GetString("user-id")
	userBrandIDs := []uuid.UUID{}

	reportRes, err := r.reportModule.GetGamePerformance(ctx.Request.Context(), req, userBrandIDs)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.Header("Content-Type", "application/json")
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=game-performance-%s.json", time.Now().Format("2006-01-02")))
	response.SendSuccessResponse(ctx, http.StatusOK, reportRes)
}

// ExportGamePlayers Export Game Players.
//
//	@Summary		ExportGamePlayers
//	@Description	Exports game players to CSV
//	@Tags			Report
//	@Accept			json
//	@Produce		application/octet-stream
//	@Param			game_id	path		string	true	"Game ID"
//	@Param			format	query		string	true	"Export format: csv"
//	@Success		200		file
//	@Failure		400		{object}	response.ErrorResponse
//	@Router			/api/admin/report/game-performance/:game_id/players/export [get]
func (r *report) ExportGamePlayers(ctx *gin.Context) {
	format := ctx.Query("format")
	if format != "csv" {
		err := errors.ErrInvalidUserInput.New("format must be 'csv'")
		_ = ctx.Error(err)
		return
	}

	var req dto.GamePlayersReq
	gameID := ctx.Param("game_id")
	if gameID == "" {
		err := errors.ErrInvalidUserInput.New("game_id parameter is required")
		_ = ctx.Error(err)
		return
	}
	req.GameID = gameID

	if err := ctx.ShouldBindQuery(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = ctx.Error(err)
		return
	}

	req.Page = 1
	req.PerPage = 50000

	_ = ctx.GetString("user-id")
	userBrandIDs := []uuid.UUID{}

	playersRes, err := r.reportModule.GetGamePlayers(ctx.Request.Context(), req, userBrandIDs)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.Header("Content-Type", "application/json")
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=game-players-%s-%s.json", gameID, time.Now().Format("2006-01-02")))
	response.SendSuccessResponse(ctx, http.StatusOK, playersRes)
}

// GetProviderPerformance Get Provider Performance Report.
//
//	@Summary		GetProviderPerformance
//	@Description	Returns provider-level aggregated performance metrics report
//	@Tags			Report
//	@Accept			json
//	@Produce		json
//	@Param			page			query		int		false	"Page number" default(1)
//	@Param			per_page		query		int		false	"Items per page" default(20)
//	@Param			date_from		query		string	true	"Start date (YYYY-MM-DD)"
//	@Param			date_to			query		string	true	"End date (YYYY-MM-DD)"
//	@Param			brand_id		query		string	false	"Brand ID filter"
//	@Param			currency		query		string	false	"Currency filter"
//	@Param			provider		query		string	false	"Provider filter"
//	@Param			category		query		string	false	"Category filter"
//	@Param			sort_by			query		string	false	"Sort by: ggr, ngr, most_played, rtp, bet_volume" default(ggr)
//	@Param			sort_order		query		string	false	"Sort order: asc, desc" default(desc)
//	@Success		200				{object}	dto.ProviderPerformanceReportRes
//	@Failure		400				{object}	response.ErrorResponse
//	@Router			/api/admin/report/provider-performance [get]
func (r *report) GetProviderPerformance(ctx *gin.Context) {
	var req dto.ProviderPerformanceReportReq

	if err := ctx.ShouldBindQuery(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = ctx.Error(err)
		return
	}

	_ = ctx.GetString("user-id")
	userBrandIDs := []uuid.UUID{}

	reportRes, err := r.reportModule.GetProviderPerformance(ctx.Request.Context(), req, userBrandIDs)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	response.SendSuccessResponse(ctx, http.StatusOK, reportRes)
}

// ExportProviderPerformance Export Provider Performance Report.
//
//	@Summary		ExportProviderPerformance
//	@Description	Exports provider performance report to CSV
//	@Tags			Report
//	@Accept			json
//	@Produce		application/octet-stream
//	@Param			format	query		string	true	"Export format: csv"
//	@Success		200		file
//	@Failure		400		{object}	response.ErrorResponse
//	@Router			/api/admin/report/provider-performance/export [get]
func (r *report) ExportProviderPerformance(ctx *gin.Context) {
	format := ctx.Query("format")
	if format != "csv" {
		err := errors.ErrInvalidUserInput.New("format must be 'csv'")
		_ = ctx.Error(err)
		return
	}

	var req dto.ProviderPerformanceReportReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = ctx.Error(err)
		return
	}

	req.Page = 1
	req.PerPage = 50000

	_ = ctx.GetString("user-id")
	userBrandIDs := []uuid.UUID{}

	reportRes, err := r.reportModule.GetProviderPerformance(ctx.Request.Context(), req, userBrandIDs)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.Header("Content-Type", "application/json")
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=provider-performance-%s.json", time.Now().Format("2006-01-02")))
	response.SendSuccessResponse(ctx, http.StatusOK, reportRes)
}

// GetAffiliateReport Get Affiliate Report.
//
//	@Summary		GetAffiliateReport
//	@Description	Returns affiliate report with daily metrics grouped by referral code
//	@Tags			Report
//	@Accept			json
//	@Produce		json
//	@Param			date_from		query		string	false	"Start date (YYYY-MM-DD)"
//	@Param			date_to			query		string	false	"End date (YYYY-MM-DD)"
//	@Param			referral_code	query		string	false	"Filter by referral code"
//	@Success		200				{object}	dto.AffiliateReportRes
//	@Failure		400				{object}	response.ErrorResponse
//	@Router			/api/admin/report/affiliate [get]
func (r *report) GetAffiliateReport(ctx *gin.Context) {
	var req dto.AffiliateReportReq

	if err := ctx.ShouldBindQuery(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = ctx.Error(err)
		return
	}

	// TODO: Get allowed referral codes based on admin's RBAC permissions
	// For now, pass empty slice to allow all referral codes
	// This should be populated based on the admin's role permissions
	allowedReferralCodes := []string{}

	reportRes, err := r.reportModule.GetAffiliateReport(ctx.Request.Context(), req, allowedReferralCodes)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	response.SendSuccessResponse(ctx, http.StatusOK, reportRes)
}

// GetAffiliatePlayersReport
//
//	@Summary		Get Affiliate Players Report (Drill-down)
//	@Description	Get player-level metrics for a specific referral code
//	@Tags			Reports
//	@Accept			json
//	@Produce		json
//	@Param			referral_code	query		string	true	"Referral code to drill down"
//	@Param			date_from		query		string	false	"Date from (YYYY-MM-DD)"
//	@Param			date_to			query		string	false	"Date to (YYYY-MM-DD)"
//	@Param			is_test_account	query		boolean	false	"Filter by test account"
//	@Success		200				{object}	dto.AffiliatePlayersReportRes
//	@Failure		400				{object}	errors.Error
//	@Failure		500				{object}	errors.Error
//	@Router			/api/admin/report/affiliate/players [get]
func (r *report) GetAffiliatePlayersReport(ctx *gin.Context) {
	var req dto.AffiliatePlayersReportReq

	if err := ctx.ShouldBindQuery(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = ctx.Error(err)
		return
	}

	// TODO: Get allowed referral codes based on admin's RBAC permissions
	allowedReferralCodes := []string{}

	reportRes, err := r.reportModule.GetAffiliatePlayersReport(ctx.Request.Context(), req, allowedReferralCodes)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	response.SendSuccessResponse(ctx, http.StatusOK, reportRes)
}
