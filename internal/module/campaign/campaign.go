package campaign

import (
	"context"

	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/storage"
	"go.uber.org/zap"
)

type Campaign struct {
	log                 *zap.Logger
	campaignStorage     storage.Campaign
	notificationStorage storage.Notification
}

func Init(campaignStorage storage.Campaign, notificationStorage storage.Notification, log *zap.Logger) *Campaign {
	return &Campaign{
		log:                 log,
		campaignStorage:     campaignStorage,
		notificationStorage: notificationStorage,
	}
}

func (c *Campaign) CreateCampaign(ctx context.Context, req dto.CreateCampaignRequest, createdBy uuid.UUID) (dto.CampaignResponse, error) {
	// Validate request
	if err := c.validateCreateCampaignRequest(req); err != nil {
		return dto.CampaignResponse{}, errors.ErrInvalidUserInput.Wrap(err, err.Error())
	}

	return c.campaignStorage.CreateCampaign(ctx, req, createdBy)
}

func (c *Campaign) GetCampaigns(ctx context.Context, req dto.GetCampaignsRequest) (dto.GetCampaignsResponse, error) {
	return c.campaignStorage.GetCampaigns(ctx, req)
}

func (c *Campaign) GetCampaignByID(ctx context.Context, campaignID uuid.UUID) (dto.CampaignResponse, error) {
	if campaignID == uuid.Nil {
		return dto.CampaignResponse{}, errors.ErrInvalidUserInput.New("invalid campaign ID")
	}

	return c.campaignStorage.GetCampaignByID(ctx, campaignID)
}

func (c *Campaign) UpdateCampaign(ctx context.Context, campaignID uuid.UUID, req dto.UpdateCampaignRequest) (dto.CampaignResponse, error) {
	if campaignID == uuid.Nil {
		return dto.CampaignResponse{}, errors.ErrInvalidUserInput.New("invalid campaign ID")
	}

	return c.campaignStorage.UpdateCampaign(ctx, campaignID, req)
}

func (c *Campaign) DeleteCampaign(ctx context.Context, campaignID uuid.UUID) error {
	if campaignID == uuid.Nil {
		return errors.ErrInvalidUserInput.New("invalid campaign ID")
	}

	return c.campaignStorage.DeleteCampaign(ctx, campaignID)
}

func (c *Campaign) GetCampaignRecipients(ctx context.Context, req dto.GetCampaignRecipientsRequest) (dto.GetCampaignRecipientsResponse, error) {
	if req.CampaignID == uuid.Nil {
		return dto.GetCampaignRecipientsResponse{}, errors.ErrInvalidUserInput.New("invalid campaign ID")
	}

	return c.campaignStorage.GetCampaignRecipients(ctx, req)
}

func (c *Campaign) GetCampaignStats(ctx context.Context, campaignID uuid.UUID) (dto.CampaignStatsResponse, error) {
	if campaignID == uuid.Nil {
		return dto.CampaignStatsResponse{}, errors.ErrInvalidUserInput.New("invalid campaign ID")
	}

	return c.campaignStorage.GetCampaignStats(ctx, campaignID)
}

func (c *Campaign) SendCampaign(ctx context.Context, campaignID uuid.UUID) error {
	c.log.Info("Starting SendCampaign", zap.String("campaign_id", campaignID.String()))

	if campaignID == uuid.Nil {
		return errors.ErrInvalidUserInput.New("invalid campaign ID")
	}

	// Get campaign details
	c.log.Info("Getting campaign details", zap.String("campaign_id", campaignID.String()))
	campaign, err := c.campaignStorage.GetCampaignByID(ctx, campaignID)
	if err != nil {
		c.log.Error("Failed to get campaign details", zap.Error(err), zap.String("campaign_id", campaignID.String()))
		return err
	}
	c.log.Info("Campaign details retrieved", zap.String("campaign_id", campaignID.String()), zap.String("status", string(campaign.Status)))

	// Check if campaign can be sent
	if campaign.Status != dto.CAMPAIGN_STATUS_DRAFT && campaign.Status != dto.CAMPAIGN_STATUS_SCHEDULED {
		return errors.ErrInvalidUserInput.New("campaign cannot be sent in current status")
	}

	// Update campaign status to sending
	c.log.Info("Updating campaign status to sending", zap.String("campaign_id", campaignID.String()))
	_, err = c.campaignStorage.UpdateCampaign(ctx, campaignID, dto.UpdateCampaignRequest{
		Status: &[]dto.CampaignStatus{dto.CAMPAIGN_STATUS_SENDING}[0],
	})
	if err != nil {
		c.log.Error("Failed to update campaign status to sending", zap.Error(err), zap.String("campaign_id", campaignID.String()))
		return err
	}

	// Get campaign recipients
	c.log.Info("Getting campaign recipients", zap.String("campaign_id", campaignID.String()))
	recipients, err := c.campaignStorage.GetCampaignRecipients(ctx, dto.GetCampaignRecipientsRequest{
		CampaignID: campaignID,
		Page:       1,
		PerPage:    1000, // Get all recipients for sending
		Status:     nil,
	})
	if err != nil {
		c.log.Error("Failed to get campaign recipients", zap.Error(err), zap.String("campaign_id", campaignID.String()))
		return err
	}
	c.log.Info("Campaign recipients retrieved", zap.String("campaign_id", campaignID.String()), zap.Int("recipient_count", len(recipients.Recipients)))

	// Send notifications to all recipients
	successCount := 0
	failedCount := 0

	c.log.Info("Starting notification creation loop", zap.String("campaign_id", campaignID.String()))
	for i, recipient := range recipients.Recipients {
		c.log.Info("Processing recipient", zap.String("campaign_id", campaignID.String()), zap.Int("recipient_index", i), zap.String("user_id", recipient.UserID.String()))

		// Create notification payload
		notificationPayload := dto.NotificationPayload{
			UserID:    recipient.UserID,
			Title:     campaign.Subject,
			Content:   campaign.Content,
			Type:      campaign.MessageType,
			CreatedBy: &campaign.CreatedBy,
		}

		// Store notification
		c.log.Info("Storing notification", zap.String("campaign_id", campaignID.String()), zap.String("user_id", recipient.UserID.String()))
		notificationResponse, err := c.notificationStorage.StoreNotification(ctx, notificationPayload, true)
		if err != nil {
			c.log.Error("failed to store notification for campaign",
				zap.Error(err),
				zap.String("campaign_id", campaignID.String()),
				zap.String("user_id", recipient.UserID.String()))
			failedCount++
			continue
		}

		// Update recipient status
		err = c.campaignStorage.UpdateRecipientStatus(ctx, recipient.ID, dto.RECIPIENT_STATUS_SENT, "")
		if err != nil {
			c.log.Error("failed to update recipient status",
				zap.Error(err),
				zap.String("recipient_id", recipient.ID.String()))
		}

		// Update recipient with notification ID
		err = c.campaignStorage.UpdateRecipientNotificationID(ctx, recipient.ID, notificationResponse.ID)
		if err != nil {
			c.log.Error("failed to update recipient notification ID",
				zap.Error(err),
				zap.String("recipient_id", recipient.ID.String()))
		}

		successCount++
	}

	// Update campaign status - mark as sent regardless of recipient count
	err = c.campaignStorage.MarkCampaignAsSent(ctx, campaignID)
	if err != nil {
		c.log.Error("failed to mark campaign as sent", zap.Error(err))
	}

	c.log.Info("campaign sent",
		zap.String("campaign_id", campaignID.String()),
		zap.Int("success_count", successCount),
		zap.Int("failed_count", failedCount))

	return nil
}

func (c *Campaign) GetCampaignNotificationsDashboard(ctx context.Context, req dto.GetCampaignNotificationsDashboardRequest) (dto.CampaignNotificationsDashboardResponse, error) {
	c.log.Info("Getting campaign notifications dashboard", zap.Int("page", req.Page), zap.Int("per_page", req.PerPage))

	// Get notifications with campaign information
	notifications, total, err := c.campaignStorage.GetCampaignNotificationsDashboard(ctx, req)
	if err != nil {
		c.log.Error("Failed to get campaign notifications dashboard", zap.Error(err))
		return dto.CampaignNotificationsDashboardResponse{}, err
	}

	// Get dashboard stats
	stats, err := c.campaignStorage.GetCampaignNotificationStats(ctx, req)
	if err != nil {
		c.log.Error("Failed to get campaign notification stats", zap.Error(err))
		return dto.CampaignNotificationsDashboardResponse{}, err
	}

	// Calculate total pages
	totalPages := int(total) / req.PerPage
	if int(total)%req.PerPage > 0 {
		totalPages++
	}

	return dto.CampaignNotificationsDashboardResponse{
		Notifications: notifications,
		Total:         int(total),
		TotalPages:    totalPages,
		Page:          req.Page,
		PerPage:       req.PerPage,
		Stats:         stats,
	}, nil
}

func (c *Campaign) GetScheduledCampaigns(ctx context.Context) ([]dto.CampaignResponse, error) {
	return c.campaignStorage.GetScheduledCampaigns(ctx)
}

func (c *Campaign) validateCreateCampaignRequest(req dto.CreateCampaignRequest) error {
	if req.Title == "" {
		return errors.ErrInvalidUserInput.New("title is required")
	}
	if req.Subject == "" {
		return errors.ErrInvalidUserInput.New("subject is required")
	}
	if req.Content == "" {
		return errors.ErrInvalidUserInput.New("content is required")
	}
	if len(req.Segments) == 0 {
		return errors.ErrInvalidUserInput.New("at least one segment is required")
	}

	// Validate segments
	for _, segment := range req.Segments {
		if segment.SegmentType == "" {
			return errors.ErrInvalidUserInput.New("segment type is required")
		}

		switch segment.SegmentType {
		case dto.SEGMENT_TYPE_CRITERIA:
			if segment.Criteria == nil {
				return errors.ErrInvalidUserInput.New("criteria is required for criteria-based segments")
			}
		case dto.SEGMENT_TYPE_CSV:
			if segment.CSVData == "" {
				return errors.ErrInvalidUserInput.New("CSV data is required for CSV-based segments")
			}
		case dto.SEGMENT_TYPE_ALL_USERS:
			// No additional validation needed
		default:
			return errors.ErrInvalidUserInput.New("invalid segment type")
		}

		// Validate segment name if provided
		if segment.SegmentName != "" && len(segment.SegmentName) > 255 {
			return errors.ErrInvalidUserInput.New("segment name too long")
		}
	}

	return nil
}
