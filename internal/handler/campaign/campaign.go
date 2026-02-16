package campaign

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/model/response"
	"github.com/tucanbit/internal/module"
	"go.uber.org/zap"
)

type Campaign struct {
	campaignModule module.Campaign
	log            *zap.Logger
}

func Init(campaignModule module.Campaign, log *zap.Logger) *Campaign {
	return &Campaign{
		campaignModule: campaignModule,
		log:            log,
	}
}

// CreateCampaign creates a new message campaign
// @Summary Create Message Campaign
// @Description Create a new message campaign with segments and recipients
// @Tags campaigns
// @Accept json
// @Produce json
// @Param request body dto.CreateCampaignRequest true "Campaign creation request"
// @Success 201 {object} dto.SuccessResponse{data=dto.CampaignResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /campaigns [post]
func (c *Campaign) CreateCampaign(ctx *gin.Context) {
	var req dto.CreateCampaignRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.log.Error("failed to bind create campaign request", zap.Error(err), zap.Any("error_details", err.Error()))
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid request format: " + err.Error(),
		})
		return
	}

	// Get user ID from context (assuming it's set by auth middleware)
	userID := ctx.GetString("user-id")
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Code:    http.StatusUnauthorized,
			Message: "User not authenticated",
		})
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Code:    http.StatusUnauthorized,
			Message: "Invalid user ID",
		})
		return
	}

	campaign, err := c.campaignModule.CreateCampaign(ctx.Request.Context(), req, userUUID)
	if err != nil {
		c.log.Error("failed to create campaign", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to create campaign",
		})
		return
	}

	response.SendSuccessResponse(ctx, http.StatusCreated, campaign)
}

// GetCampaignNotificationsDashboard retrieves campaign notifications dashboard with pagination
// @Summary Get Campaign Notifications Dashboard
// @Description Get paginated campaign notifications for dashboard view
// @Tags campaigns
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(10)
// @Param status query string false "Notification status filter"
// @Param type query string false "Notification type filter"
// @Param user_id query string false "User ID filter"
// @Success 200 {object} dto.SuccessResponse{data=dto.CampaignNotificationsDashboardResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /campaign-notifications-dashboard [get]
func (c *Campaign) GetCampaignNotificationsDashboard(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(ctx.DefaultQuery("per_page", "10"))

	var status *string
	if statusStr := ctx.Query("status"); statusStr != "" {
		status = &statusStr
	}

	var notificationType *string
	if typeStr := ctx.Query("type"); typeStr != "" {
		notificationType = &typeStr
	}

	var userID *uuid.UUID
	if userIDStr := ctx.Query("user_id"); userIDStr != "" {
		if parsedUUID, err := uuid.Parse(userIDStr); err == nil {
			userID = &parsedUUID
		}
	}

	req := dto.GetCampaignNotificationsDashboardRequest{
		Page:             page,
		PerPage:          perPage,
		Status:           status,
		NotificationType: notificationType,
		UserID:           userID,
	}

	dashboard, err := c.campaignModule.GetCampaignNotificationsDashboard(ctx.Request.Context(), req)
	if err != nil {
		c.log.Error("failed to get campaign notifications dashboard", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to get campaign notifications dashboard",
		})
		return
	}

	response.SendSuccessResponse(ctx, http.StatusOK, dashboard)
}

// GetCampaigns retrieves campaigns with pagination and filters
// @Summary Get Campaigns
// @Description Get campaigns with pagination and optional filters
// @Tags campaigns
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(10)
// @Param status query string false "Campaign status filter"
// @Param message_type query string false "Message type filter"
// @Param created_by query string false "Created by user ID filter"
// @Success 200 {object} dto.SuccessResponse{data=dto.GetCampaignsResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /campaigns [get]
func (c *Campaign) GetCampaigns(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(ctx.DefaultQuery("per_page", "10"))

	var status *dto.CampaignStatus
	if statusStr := ctx.Query("status"); statusStr != "" {
		s := dto.CampaignStatus(statusStr)
		status = &s
	}

	var messageType *dto.NotificationType
	if messageTypeStr := ctx.Query("message_type"); messageTypeStr != "" {
		mt := dto.NotificationType(messageTypeStr)
		messageType = &mt
	}

	var createdBy *uuid.UUID
	if createdByStr := ctx.Query("created_by"); createdByStr != "" {
		if parsedUUID, err := uuid.Parse(createdByStr); err == nil {
			createdBy = &parsedUUID
		}
	}

	req := dto.GetCampaignsRequest{
		Page:        page,
		PerPage:     perPage,
		Status:      status,
		MessageType: messageType,
		CreatedBy:   createdBy,
	}

	campaigns, err := c.campaignModule.GetCampaigns(ctx.Request.Context(), req)
	if err != nil {
		c.log.Error("failed to get campaigns", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to get campaigns",
		})
		return
	}

	response.SendSuccessResponse(ctx, http.StatusOK, campaigns)
}

// GetCampaignByID retrieves a specific campaign by ID
// @Summary Get Campaign by ID
// @Description Get a specific campaign by its ID
// @Tags campaigns
// @Accept json
// @Produce json
// @Param id path string true "Campaign ID"
// @Success 200 {object} dto.SuccessResponse{data=dto.CampaignResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /campaigns/{id} [get]
func (c *Campaign) GetCampaignByID(ctx *gin.Context) {
	campaignIDStr := ctx.Param("id")
	campaignID, err := uuid.Parse(campaignIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid campaign ID",
		})
		return
	}

	campaign, err := c.campaignModule.GetCampaignByID(ctx.Request.Context(), campaignID)
	if err != nil {
		c.log.Error("failed to get campaign", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to get campaign",
		})
		return
	}

	response.SendSuccessResponse(ctx, http.StatusOK, campaign)
}

// UpdateCampaign updates an existing campaign
// @Summary Update Campaign
// @Description Update an existing campaign
// @Tags campaigns
// @Accept json
// @Produce json
// @Param id path string true "Campaign ID"
// @Param request body dto.UpdateCampaignRequest true "Campaign update request"
// @Success 200 {object} dto.SuccessResponse{data=dto.CampaignResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /campaigns/{id} [put]
func (c *Campaign) UpdateCampaign(ctx *gin.Context) {
	campaignIDStr := ctx.Param("id")
	campaignID, err := uuid.Parse(campaignIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid campaign ID",
		})
		return
	}

	var req dto.UpdateCampaignRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.log.Error("failed to bind update campaign request", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid request format",
		})
		return
	}

	campaign, err := c.campaignModule.UpdateCampaign(ctx.Request.Context(), campaignID, req)
	if err != nil {
		c.log.Error("failed to update campaign", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to update campaign",
		})
		return
	}

	response.SendSuccessResponse(ctx, http.StatusOK, campaign)
}

// DeleteCampaign deletes a campaign
// @Summary Delete Campaign
// @Description Delete a campaign and all its associated data
// @Tags campaigns
// @Accept json
// @Produce json
// @Param id path string true "Campaign ID"
// @Success 200 {object} dto.SuccessResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /campaigns/{id} [delete]
func (c *Campaign) DeleteCampaign(ctx *gin.Context) {
	campaignIDStr := ctx.Param("id")
	campaignID, err := uuid.Parse(campaignIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid campaign ID",
		})
		return
	}

	err = c.campaignModule.DeleteCampaign(ctx.Request.Context(), campaignID)
	if err != nil {
		c.log.Error("failed to delete campaign", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to delete campaign",
		})
		return
	}

	response.SendSuccessResponse(ctx, http.StatusNoContent, nil)
}

// SendCampaign sends a campaign to all its recipients
// @Summary Send Campaign
// @Description Send a campaign to all its recipients
// @Tags campaigns
// @Accept json
// @Produce json
// @Param id path string true "Campaign ID"
// @Success 200 {object} dto.SuccessResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /campaigns/{id}/send [post]
func (c *Campaign) SendCampaign(ctx *gin.Context) {
	campaignIDStr := ctx.Param("id")
	campaignID, err := uuid.Parse(campaignIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid campaign ID",
		})
		return
	}

	err = c.campaignModule.SendCampaign(ctx.Request.Context(), campaignID)
	if err != nil {
		c.log.Error("failed to send campaign", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to send campaign",
		})
		return
	}

	response.SendSuccessResponse(ctx, http.StatusOK, map[string]string{"message": "Campaign sent successfully"})
}

// GetCampaignRecipients retrieves recipients for a campaign
// @Summary Get Campaign Recipients
// @Description Get recipients for a specific campaign
// @Tags campaigns
// @Accept json
// @Produce json
// @Param id path string true "Campaign ID"
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(10)
// @Param status query string false "Recipient status filter"
// @Success 200 {object} dto.SuccessResponse{data=dto.GetCampaignRecipientsResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /campaigns/{id}/recipients [get]
func (c *Campaign) GetCampaignRecipients(ctx *gin.Context) {
	campaignIDStr := ctx.Param("id")
	campaignID, err := uuid.Parse(campaignIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid campaign ID",
		})
		return
	}

	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(ctx.DefaultQuery("per_page", "10"))

	var status *dto.RecipientStatus
	if statusStr := ctx.Query("status"); statusStr != "" {
		s := dto.RecipientStatus(statusStr)
		status = &s
	}

	req := dto.GetCampaignRecipientsRequest{
		CampaignID: campaignID,
		Page:       page,
		PerPage:    perPage,
		Status:     status,
	}

	recipients, err := c.campaignModule.GetCampaignRecipients(ctx.Request.Context(), req)
	if err != nil {
		c.log.Error("failed to get campaign recipients", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to get campaign recipients",
		})
		return
	}

	response.SendSuccessResponse(ctx, http.StatusOK, recipients)
}

// GetCampaignStats retrieves statistics for a campaign
// @Summary Get Campaign Statistics
// @Description Get delivery and read statistics for a campaign
// @Tags campaigns
// @Accept json
// @Produce json
// @Param id path string true "Campaign ID"
// @Success 200 {object} dto.SuccessResponse{data=dto.CampaignStatsResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /campaigns/{id}/stats [get]
func (c *Campaign) GetCampaignStats(ctx *gin.Context) {
	campaignIDStr := ctx.Param("id")
	campaignID, err := uuid.Parse(campaignIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid campaign ID",
		})
		return
	}

	stats, err := c.campaignModule.GetCampaignStats(ctx.Request.Context(), campaignID)
	if err != nil {
		c.log.Error("failed to get campaign stats", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to get campaign stats",
		})
		return
	}

	response.SendSuccessResponse(ctx, http.StatusOK, stats)
}
