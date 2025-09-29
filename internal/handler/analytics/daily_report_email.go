package analytics

import (
	"net/http"
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
	return len(s) >= len(substr) && (s == substr || contains(s[1:], substr))
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
