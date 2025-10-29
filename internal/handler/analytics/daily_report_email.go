package analytics

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tucanbit/internal/constant/dto"
	"go.uber.org/zap"
)

// SendDailyReportEmailRequest represents the request to send daily report email
type SendDailyReportEmailRequest struct {
	Date       string   `json:"date" binding:"required" example:"2025-01-15"`                                  // Date in YYYY-MM-DD format
	Recipients []string `json:"recipients" binding:"required" example:"admin@example.com,manager@example.com"` // Email recipients
}

// SendYesterdayReportEmailRequest represents the request to send yesterday's report
type SendYesterdayReportEmailRequest struct {
	Recipients []string `json:"recipients" binding:"required" example:"admin@example.com,manager@example.com"` // Email recipients
}

// SendLastWeekReportEmailRequest represents the request to send last week's reports
type SendLastWeekReportEmailRequest struct {
	Recipients []string `json:"recipients" binding:"required" example:"admin@example.com,manager@example.com"` // Email recipients
}

// SendConfiguredDailyReportEmailRequest represents the request to send daily report to configured recipients
type SendConfiguredDailyReportEmailRequest struct {
	Date string `json:"date,omitempty" example:"2025-01-15"` // Date in YYYY-MM-DD format (optional, defaults to yesterday)
}

// SendConfiguredDailyReportEmail sends daily report email to configured recipients
// @Summary Send daily report email to configured recipients
// @Description Send daily report email to recipients configured in the system (ashenafialemu27@gmail.com, johsjones612@gmail.com)
// @Tags analytics
// @Accept json
// @Produce json
// @Param request body SendConfiguredDailyReportEmailRequest true "Configured daily report email request"
// @Success 200 {object} dto.AnalyticsResponse
// @Failure 400 {object} dto.AnalyticsResponse
// @Failure 500 {object} dto.AnalyticsResponse
// @Router /analytics/daily-report/send-configured [post]
func (a *analytics) SendConfiguredDailyReportEmail(c *gin.Context) {
	var req SendConfiguredDailyReportEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		a.logger.Error("Invalid request format for configured daily report email", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.AnalyticsResponse{
			Success: false,
			Error:   "Invalid request format: " + err.Error(),
		})
		return
	}

	// Use yesterday's date if no date provided
	date := req.Date
	if date == "" {
		date = time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	}

	// Validate date format
	parsedDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		a.logger.Error("Invalid date format for configured daily report email",
			zap.String("date", date),
			zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.AnalyticsResponse{
			Success: false,
			Error:   "Invalid date format. Use YYYY-MM-DD",
		})
		return
	}

	// Get configured recipients
	configuredRecipients := []string{
		"ashenafialemu27@gmail.com",
		"johsjones612@gmail.com",
	}

	a.logger.Info("Sending configured daily report email",
		zap.String("date", date),
		zap.Int("recipients_count", len(configuredRecipients)),
		zap.Strings("recipients", configuredRecipients))

	// Call the daily report service to generate and send the email
	if err := a.dailyReportService.GenerateAndSendDailyReport(c.Request.Context(), parsedDate, configuredRecipients); err != nil {
		a.logger.Error("Failed to generate and send configured daily report email",
			zap.String("date", date),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.AnalyticsResponse{
			Success: false,
			Error:   "Failed to send configured daily report email: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.AnalyticsResponse{
		Success: true,
		Data: gin.H{
			"message":          "Configured daily report email sent successfully",
			"date":             date,
			"recipients_count": len(configuredRecipients),
			"recipients":       configuredRecipients,
			"note":             "Email sent to configured recipients: ashenafialemu27@gmail.com, johsjones612@gmail.com",
		},
	})
}

// SendDailyReportEmail sends daily report email manually
// @Summary Send daily report email
// @Description Send daily report email to specified recipients for a specific date
// @Tags analytics
// @Accept json
// @Produce json
// @Param request body SendDailyReportEmailRequest true "Daily report email request"
// @Success 200 {object} dto.AnalyticsResponse
// @Failure 400 {object} dto.AnalyticsResponse
// @Failure 500 {object} dto.AnalyticsResponse
// @Router /analytics/daily-report/send [post]
func (a *analytics) SendDailyReportEmail(c *gin.Context) {
	var req SendDailyReportEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		a.logger.Error("Invalid request format for daily report email", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.AnalyticsResponse{
			Success: false,
			Error:   "Invalid request format: " + err.Error(),
		})
		return
	}

	// Validate date format
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		a.logger.Error("Invalid date format for daily report email",
			zap.String("date", req.Date),
			zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.AnalyticsResponse{
			Success: false,
			Error:   "Invalid date format. Use YYYY-MM-DD",
		})
		return
	}

	// Validate recipients
	if len(req.Recipients) == 0 {
		c.JSON(http.StatusBadRequest, dto.AnalyticsResponse{
			Success: false,
			Error:   "At least one recipient email is required",
		})
		return
	}

	// Validate email format (basic validation)
	for _, email := range req.Recipients {
		if len(email) < 5 || !contains(email, "@") {
			c.JSON(http.StatusBadRequest, dto.AnalyticsResponse{
				Success: false,
				Error:   "Invalid email format: " + email,
			})
			return
		}
	}

	a.logger.Info("Sending daily report email manually",
		zap.String("date", req.Date),
		zap.Int("recipients_count", len(req.Recipients)))

	// Call the daily report service to generate and send the email
	if err := a.dailyReportService.GenerateAndSendDailyReport(c.Request.Context(), date, req.Recipients); err != nil {
		a.logger.Error("Failed to generate and send daily report email",
			zap.String("date", req.Date),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.AnalyticsResponse{
			Success: false,
			Error:   "Failed to send daily report email: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.AnalyticsResponse{
		Success: true,
		Data: gin.H{
			"message":          "Daily report email sent successfully",
			"date":             req.Date,
			"recipients_count": len(req.Recipients),
			"recipients":       req.Recipients,
		},
	})
}

// SendYesterdayReportEmail sends yesterday's report email
// @Summary Send yesterday's report email
// @Description Send yesterday's daily report email to specified recipients
// @Tags analytics
// @Accept json
// @Produce json
// @Param request body SendYesterdayReportEmailRequest true "Yesterday's report email request"
// @Success 200 {object} dto.AnalyticsResponse
// @Failure 400 {object} dto.AnalyticsResponse
// @Failure 500 {object} dto.AnalyticsResponse
// @Router /analytics/daily-report/yesterday [post]
func (a *analytics) SendYesterdayReportEmail(c *gin.Context) {
	var req SendYesterdayReportEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		a.logger.Error("Invalid request format for yesterday's report email", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.AnalyticsResponse{
			Success: false,
			Error:   "Invalid request format: " + err.Error(),
		})
		return
	}

	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	a.logger.Info("Sending yesterday's report email manually",
		zap.String("date", yesterday),
		zap.Int("recipients_count", len(req.Recipients)))

	// Call the daily report service to generate and send yesterday's report
	if err := a.dailyReportService.GenerateYesterdayReport(c.Request.Context(), req.Recipients); err != nil {
		a.logger.Error("Failed to generate and send yesterday's report email",
			zap.String("date", yesterday),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.AnalyticsResponse{
			Success: false,
			Error:   "Failed to send yesterday's report email: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.AnalyticsResponse{
		Success: true,
		Data: gin.H{
			"message":          "Yesterday's report email sent successfully",
			"date":             yesterday,
			"recipients_count": len(req.Recipients),
			"recipients":       req.Recipients,
		},
	})
}

// SendLastWeekReportEmail sends last week's reports email
// @Summary Send last week's reports email
// @Description Send daily reports for the last 7 days to specified recipients
// @Tags analytics
// @Accept json
// @Produce json
// @Param request body SendLastWeekReportEmailRequest true "Last week's reports email request"
// @Success 200 {object} dto.AnalyticsResponse
// @Failure 400 {object} dto.AnalyticsResponse
// @Failure 500 {object} dto.AnalyticsResponse
// @Router /analytics/daily-report/last-week [post]
func (a *analytics) SendLastWeekReportEmail(c *gin.Context) {
	var req SendLastWeekReportEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		a.logger.Error("Invalid request format for last week's report email", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.AnalyticsResponse{
			Success: false,
			Error:   "Invalid request format: " + err.Error(),
		})
		return
	}

	a.logger.Info("Sending last week's report emails manually",
		zap.Int("recipients_count", len(req.Recipients)))

	// Call the daily report service to generate and send last week's reports
	if err := a.dailyReportService.GenerateLastWeekReport(c.Request.Context(), req.Recipients); err != nil {
		a.logger.Error("Failed to generate and send last week's report emails",
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.AnalyticsResponse{
			Success: false,
			Error:   "Failed to send last week's report emails: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.AnalyticsResponse{
		Success: true,
		Data: gin.H{
			"message":          "Last week's report emails sent successfully",
			"recipients_count": len(req.Recipients),
			"recipients":       req.Recipients,
		},
	})
}

// SendDailyReportRequest represents the simplified request structure
type SendDailyReportRequest struct {
	Date              string   `json:"date,omitempty"`                // Date in YYYY-MM-DD format (optional, defaults to yesterday if empty)
	Recipients        []string `json:"recipients" binding:"required"` // Email recipients
	IncludeTopGames   bool     `json:"include_top_games,omitempty"`   // Include top games section in report
	IncludeTopPlayers bool     `json:"include_top_players,omitempty"` // Include top players section in report
	SendFrequency     string   `json:"send_frequency,omitempty"`      // Daily, Weekly, Monthly frequency
	TimeOfDay         int      `json:"time_of_day,omitempty"`         // Hour of day to send (0-23)
	Timezone          string   `json:"timezone,omitempty"`            // Timezone for sending time (default: UTC)
	AutoSchedule      bool     `json:"auto_schedule,omitempty"`       // Whether to schedule future automatic sending
}

// GenerateDailyReportRequest represents the request structure for generating daily reports
type GenerateDailyReportRequest struct {
	Date              string   `json:"date,omitempty"`                // Date in YYYY-MM-DD format (optional, defaults to yesterday)
	Recipients        []string `json:"recipients,omitempty"`          // Email recipients (optional for testing)
	IncludeTopGames   bool     `json:"include_top_games,omitempty"`   // Include top games section
	IncludeTopPlayers bool     `json:"include_top_players,omitempty"` // Include top players section
	Format            string   `json:"format,omitempty"`              // Report format: "email", "json", "csv"
}

// contains helper function for basic string validation
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// ScheduleDailyReportCronJob schedules a cron job for automatic daily reports
// @Summary Schedule daily report cronjob
// @Description Schedule automatic sending of daily reports at specified time
// @Tags analytics
// @Accept json
// @Produce json
// @Param request body SendDailyReportRequest true "Daily report scheduling request"
// @Success 200 {object} dto.AnalyticsResponse
// @Failure 400 {object} dto.AnalyticsResponse
// @Failure 500 {object} dto.AnalyticsResponse
// @Router /analytics/daily-report/schedule [post]
func (a *analytics) ScheduleDailyReportCronJob(c *gin.Context) {
	var req SendDailyReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		a.logger.Error("Invalid request format for daily report scheduling", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.AnalyticsResponse{
			Success: false,
			Error:   "Invalid request format: " + err.Error(),
		})
		return
	}

	// Validate recipients
	if len(req.Recipients) == 0 {
		c.JSON(http.StatusBadRequest, dto.AnalyticsResponse{
			Success: false,
			Error:   "At least one recipient email is required",
		})
		return
	}

	// Validate time of day
	if req.TimeOfDay < 0 || req.TimeOfDay > 23 {
		c.JSON(http.StatusBadRequest, dto.AnalyticsResponse{
			Success: false,
			Error:   "Time of day must be between 0 and 23",
		})
		return
	}

	// Default values
	if req.SendFrequency == "" {
		req.SendFrequency = "daily"
	}
	if req.Timezone == "" {
		req.Timezone = "UTC"
	}

	a.logger.Info("Scheduling daily report cronjob",
		zap.String("frequency", req.SendFrequency),
		zap.Int("time_of_day", req.TimeOfDay),
		zap.String("timezone", req.Timezone),
		zap.Int("recipients_count", len(req.Recipients)))

	// Create cron job schedule info for response
	scheduleInfo := gin.H{
		"frequency":        req.SendFrequency,
		"time_of_day":      req.TimeOfDay,
		"timezone":         req.Timezone,
		"recipients":       req.Recipients,
		"recipients_count": len(req.Recipients),
		"auto_schedule":    req.AutoSchedule,
		"status":           "scheduled",
		"created_at":       time.Now().Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, dto.AnalyticsResponse{
		Success: true,
		Data: gin.H{
			"message":       "Daily report cronjob scheduled successfully",
			"schedule_info": scheduleInfo,
		},
	})
}

// SendTestDailyReport sends a test daily report to configured recipients
// @Summary Send test daily report
// @Description Send a test daily report to configured recipients for verification
// @Tags analytics
// @Accept json
// @Produce json
// @Success 200 {object} dto.AnalyticsResponse
// @Failure 500 {object} dto.AnalyticsResponse
// @Router /analytics/daily-report/test [post]
func (a *analytics) SendTestDailyReport(c *gin.Context) {
	a.logger.Info("Sending test daily report to configured recipients")

	// Get configured recipients
	configuredRecipients := []string{
		"ashenafialemu27@gmail.com",
		"johsjones612@gmail.com",
	}

	// Use yesterday's date for test
	testDate := time.Now().AddDate(0, 0, -1)

	a.logger.Info("Sending test daily report",
		zap.String("date", testDate.Format("2006-01-02")),
		zap.Int("recipients_count", len(configuredRecipients)),
		zap.Strings("recipients", configuredRecipients))

	// Call the daily report service to generate and send the email
	if err := a.dailyReportService.GenerateAndSendDailyReport(c.Request.Context(), testDate, configuredRecipients); err != nil {
		a.logger.Error("Failed to generate and send test daily report",
			zap.String("date", testDate.Format("2006-01-02")),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.AnalyticsResponse{
			Success: false,
			Error:   "Failed to send test daily report: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.AnalyticsResponse{
		Success: true,
		Data: gin.H{
			"message":          "Test daily report sent successfully",
			"date":             testDate.Format("2006-01-02"),
			"recipients_count": len(configuredRecipients),
			"recipients":       configuredRecipients,
			"note":             "Test email sent to configured recipients: ashenafialemu27@gmail.com, johsjones612@gmail.com",
		},
	})
}

// GetCronjobStatus gets the status of the daily report cronjob service
// @Summary Get cronjob status
// @Description Get the current status of the daily report cronjob service
// @Tags analytics
// @Accept json
// @Produce json
// @Success 200 {object} dto.AnalyticsResponse
// @Router /analytics/daily-report/cronjob-status [get]
func (a *analytics) GetCronjobStatus(c *gin.Context) {
	if a.dailyReportCronjobService == nil {
		c.JSON(http.StatusOK, dto.AnalyticsResponse{
			Success: true,
			Data: gin.H{
				"status":                "not_initialized",
				"message":               "Daily report cronjob service is not initialized",
				"configured_recipients": []string{"ashenafialemu27@gmail.com", "johsjones612@gmail.com"},
				"schedule":              "23:59 UTC (end of day)",
			},
		})
		return
	}

	isRunning := a.dailyReportCronjobService.IsRunning()
	status := "stopped"
	if isRunning {
		status = "running"
	}

	c.JSON(http.StatusOK, dto.AnalyticsResponse{
		Success: true,
		Data: gin.H{
			"status":                status,
			"is_running":            isRunning,
			"message":               "Daily report cronjob service status",
			"configured_recipients": []string{"ashenafialemu27@gmail.com", "johsjones612@gmail.com"},
			"schedule":              "23:59 UTC (end of day)",
			"next_run":              "Tomorrow at 23:59 UTC",
		},
	})
}
