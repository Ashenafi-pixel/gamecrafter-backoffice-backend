package campaign

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/constant/persistencedb"
	"go.uber.org/zap"
)

type Campaign struct {
	db  *persistencedb.PersistenceDB
	log *zap.Logger
}

func Init(db *persistencedb.PersistenceDB, log *zap.Logger) *Campaign {
	return &Campaign{
		db:  db,
		log: log,
	}
}

func (c *Campaign) CreateCampaign(ctx context.Context, req dto.CreateCampaignRequest, createdBy uuid.UUID) (dto.CampaignResponse, error) {
	// Start transaction
	tx, err := c.db.GetPool().Begin(ctx)
	if err != nil {
		c.log.Error("failed to begin transaction", zap.Error(err))
		return dto.CampaignResponse{}, errors.ErrUnableTocreate.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback(ctx)

	// Create campaign using direct SQL
	query := `
		INSERT INTO message_campaigns (title, message_type, subject, content, created_by, scheduled_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, title, message_type, subject, content, created_by, status, scheduled_at, sent_at, total_recipients, delivered_count, read_count, created_at, updated_at
	`
	var campaignRes struct {
		ID              uuid.UUID
		Title           string
		MessageType     string
		Subject         string
		Content         string
		CreatedBy       uuid.UUID
		Status          string
		ScheduledAt     *time.Time
		SentAt          *time.Time
		TotalRecipients int32
		DeliveredCount  int32
		ReadCount       int32
		CreatedAt       time.Time
		UpdatedAt       time.Time
	}
	// Convert FlexibleTime to *time.Time if provided
	var scheduledAt *time.Time
	if req.ScheduledAt != nil {
		scheduledAt = &req.ScheduledAt.Time
	}
	err = tx.QueryRow(ctx, query, req.Title, string(req.MessageType), req.Subject, req.Content, createdBy, scheduledAt).Scan(
		&campaignRes.ID,
		&campaignRes.Title,
		&campaignRes.MessageType,
		&campaignRes.Subject,
		&campaignRes.Content,
		&campaignRes.CreatedBy,
		&campaignRes.Status,
		&campaignRes.ScheduledAt,
		&campaignRes.SentAt,
		&campaignRes.TotalRecipients,
		&campaignRes.DeliveredCount,
		&campaignRes.ReadCount,
		&campaignRes.CreatedAt,
		&campaignRes.UpdatedAt,
	)
	if err != nil {
		c.log.Error("failed to create campaign", zap.Error(err))
		return dto.CampaignResponse{}, errors.ErrUnableTocreate.Wrap(err, "failed to create campaign")
	}

	// Create segments
	totalRecipients := 0
	for _, segmentReq := range req.Segments {
		var criteriaJSON interface{}
		if segmentReq.Criteria != nil {
			criteriaJSON = segmentReq.Criteria
		}

		segmentQuery := `
			INSERT INTO message_segments (campaign_id, segment_type, segment_name, criteria, csv_data, user_count)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id, campaign_id, segment_type, segment_name, criteria, csv_data, user_count, created_at
		`
		var segmentRes struct {
			ID          uuid.UUID
			CampaignID  uuid.UUID
			SegmentType string
			SegmentName *string
			Criteria    interface{}
			CSVData     *string
			UserCount   int32
			CreatedAt   time.Time
		}
		err = tx.QueryRow(ctx, segmentQuery, campaignRes.ID, string(segmentReq.SegmentType), segmentReq.SegmentName, criteriaJSON, segmentReq.CSVData, 0).Scan(
			&segmentRes.ID,
			&segmentRes.CampaignID,
			&segmentRes.SegmentType,
			&segmentRes.SegmentName,
			&segmentRes.Criteria,
			&segmentRes.CSVData,
			&segmentRes.UserCount,
			&segmentRes.CreatedAt,
		)
		if err != nil {
			c.log.Error("failed to create segment", zap.Error(err))
			return dto.CampaignResponse{}, errors.ErrUnableTocreate.Wrap(err, "failed to create segment")
		}

		// Calculate user count for this segment
		userCount, err := c.calculateSegmentUserCount(ctx, tx, segmentReq)
		if err != nil {
			c.log.Error("failed to calculate segment user count", zap.Error(err))
			return dto.CampaignResponse{}, errors.ErrUnableTocreate.Wrap(err, "failed to calculate segment user count")
		}

		// Update segment user count
		updateQuery := `UPDATE message_segments SET user_count = $1 WHERE id = $2`
		_, err = tx.Exec(ctx, updateQuery, userCount, segmentRes.ID)
		if err != nil {
			c.log.Error("failed to update segment user count", zap.Error(err))
			return dto.CampaignResponse{}, errors.ErrUnableTocreate.Wrap(err, "failed to update segment user count")
		}

		totalRecipients += userCount

	}

	// Create recipients for all segments
	for _, segmentReq := range req.Segments {
		if segmentReq.SegmentType == dto.SEGMENT_TYPE_ALL_USERS {
			// Insert recipients directly using a simple approach
			insertQuery := `
				INSERT INTO campaign_recipients (campaign_id, user_id, status)
				SELECT $1, id, 'pending'
				FROM users 
				WHERE is_admin IS NOT TRUE AND user_type = 'PLAYER'
			`
			result, err := tx.Exec(ctx, insertQuery, campaignRes.ID)
			if err != nil {
				c.log.Error("failed to insert campaign recipients", zap.Error(err))
				return dto.CampaignResponse{}, errors.ErrUnableTocreate.Wrap(err, "failed to insert campaign recipients")
			}

			// Log how many recipients were created
			rowsAffected := result.RowsAffected()
			c.log.Info("Created campaign recipients",
				zap.String("campaign_id", campaignRes.ID.String()),
				zap.Int64("recipients_created", rowsAffected))
		}
	}

	// Update campaign total recipients
	updateCampaignQuery := `UPDATE message_campaigns SET total_recipients = $1 WHERE id = $2`
	_, err = tx.Exec(ctx, updateCampaignQuery, totalRecipients, campaignRes.ID)
	if err != nil {
		c.log.Error("failed to update campaign total recipients", zap.Error(err))
		return dto.CampaignResponse{}, errors.ErrUnableTocreate.Wrap(err, "failed to update campaign total recipients")
	}

	if err := tx.Commit(ctx); err != nil {
		c.log.Error("failed to commit transaction", zap.Error(err))
		return dto.CampaignResponse{}, errors.ErrUnableTocreate.Wrap(err, "failed to commit transaction")
	}

	return dto.CampaignResponse{
		ID:              campaignRes.ID,
		Title:           campaignRes.Title,
		MessageType:     dto.NotificationType(campaignRes.MessageType),
		Subject:         campaignRes.Subject,
		Content:         campaignRes.Content,
		CreatedBy:       campaignRes.CreatedBy,
		Status:          dto.CampaignStatus(campaignRes.Status),
		ScheduledAt:     campaignRes.ScheduledAt,
		SentAt:          campaignRes.SentAt,
		TotalRecipients: int(campaignRes.TotalRecipients),
		DeliveredCount:  int(campaignRes.DeliveredCount),
		ReadCount:       int(campaignRes.ReadCount),
		CreatedAt:       campaignRes.CreatedAt,
		UpdatedAt:       campaignRes.UpdatedAt,
	}, nil
}

func (c *Campaign) calculateSegmentUserCount(ctx context.Context, tx pgx.Tx, segmentReq dto.CreateSegmentRequest) (int, error) {
	switch segmentReq.SegmentType {
	case dto.SEGMENT_TYPE_ALL_USERS:
		query := `SELECT COUNT(*) FROM users WHERE is_admin IS NOT TRUE AND user_type = 'PLAYER'`
		var count int
		err := tx.QueryRow(ctx, query).Scan(&count)
		return count, err

	case dto.SEGMENT_TYPE_CRITERIA:
		if segmentReq.Criteria == nil {
			return 0, nil
		}
		// Build dynamic query based on criteria
		query, args := c.buildCriteriaQuery(segmentReq.Criteria)
		var count int
		err := tx.QueryRow(ctx, query, args...).Scan(&count)
		return count, err

	case dto.SEGMENT_TYPE_CSV:
		if segmentReq.CSVData == "" {
			return 0, nil
		}
		usernames := strings.Split(segmentReq.CSVData, "\n")
		query := `SELECT COUNT(*) FROM users WHERE username = ANY($1) AND is_admin IS NOT TRUE AND user_type = 'PLAYER'`
		var count int
		err := tx.QueryRow(ctx, query, usernames).Scan(&count)
		return count, err

	default:
		return 0, fmt.Errorf("unknown segment type: %s", segmentReq.SegmentType)
	}
}

func (c *Campaign) buildCriteriaQuery(criteria interface{}) (string, []interface{}) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	// Type assert to map for easier access
	criteriaMap, ok := criteria.(map[string]interface{})
	if !ok {
		return "SELECT COUNT(*) FROM users", args
	}

	if daysSinceLastActivity, exists := criteriaMap["days_since_last_activity"]; exists && daysSinceLastActivity != nil {
		if days, ok := daysSinceLastActivity.(float64); ok {
			conditions = append(conditions, fmt.Sprintf("last_login_at <= NOW() - INTERVAL '%d days'", int(days)))
		}
	}

	if userLevel, exists := criteriaMap["user_level"]; exists && userLevel != nil {
		if level, ok := userLevel.(string); ok {
			conditions = append(conditions, fmt.Sprintf("level = $%d", argIndex))
			args = append(args, level)
			argIndex++
		}
	}

	if kycStatus, exists := criteriaMap["kyc_status"]; exists && kycStatus != nil {
		if status, ok := kycStatus.(string); ok {
			conditions = append(conditions, fmt.Sprintf("kyc_status = $%d", argIndex))
			args = append(args, status)
			argIndex++
		}
	}

	if country, exists := criteriaMap["country"]; exists && country != nil {
		if c, ok := country.(string); ok {
			conditions = append(conditions, fmt.Sprintf("country = $%d", argIndex))
			args = append(args, c)
			argIndex++
		}
	}

	if currency, exists := criteriaMap["currency"]; exists && currency != nil {
		if curr, ok := currency.(string); ok {
			conditions = append(conditions, fmt.Sprintf("default_currency = $%d", argIndex))
			args = append(args, curr)
			argIndex++
		}
	}

	if minBalance, exists := criteriaMap["min_balance"]; exists && minBalance != nil {
		if balance, ok := minBalance.(float64); ok {
			conditions = append(conditions, fmt.Sprintf("EXISTS (SELECT 1 FROM wallets w WHERE w.user_id = users.id AND w.balance >= $%d)", argIndex))
			args = append(args, balance)
			argIndex++
		}
	}

	if maxBalance, exists := criteriaMap["max_balance"]; exists && maxBalance != nil {
		if balance, ok := maxBalance.(float64); ok {
			conditions = append(conditions, fmt.Sprintf("EXISTS (SELECT 1 FROM wallets w WHERE w.user_id = users.id AND w.balance <= $%d)", argIndex))
			args = append(args, balance)
			argIndex++
		}
	}

	query := "SELECT COUNT(*) FROM users"
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	return query, args
}

func (c *Campaign) GetCampaigns(ctx context.Context, req dto.GetCampaignsRequest) (dto.GetCampaignsResponse, error) {
	query := `
		SELECT 
			id, title, message_type, subject, content, created_by, status, scheduled_at, sent_at, 
			total_recipients, delivered_count, read_count, created_at, updated_at,
			COUNT(*) OVER() AS total
		FROM message_campaigns
		WHERE 
			($1::uuid IS NULL OR created_by = $1) AND
			($2::text IS NULL OR status = $2) AND
			($3::text IS NULL OR message_type::text = $3)
		ORDER BY created_at DESC
		LIMIT $4 OFFSET $5
	`

	rows, err := c.db.GetPool().Query(ctx, query, req.CreatedBy, req.Status, req.MessageType, req.PerPage, (req.Page-1)*req.PerPage)
	if err != nil {
		c.log.Error("failed to get campaigns", zap.Error(err))
		return dto.GetCampaignsResponse{}, errors.ErrUnableTocreate.Wrap(err, "failed to get campaigns")
	}
	defer rows.Close()

	var campaigns []dto.CampaignResponse
	var total int64

	for rows.Next() {
		var campaign dto.CampaignResponse
		err := rows.Scan(
			&campaign.ID,
			&campaign.Title,
			&campaign.MessageType,
			&campaign.Subject,
			&campaign.Content,
			&campaign.CreatedBy,
			&campaign.Status,
			&campaign.ScheduledAt,
			&campaign.SentAt,
			&campaign.TotalRecipients,
			&campaign.DeliveredCount,
			&campaign.ReadCount,
			&campaign.CreatedAt,
			&campaign.UpdatedAt,
			&total,
		)
		if err != nil {
			c.log.Error("failed to scan campaign", zap.Error(err))
			return dto.GetCampaignsResponse{}, errors.ErrUnableTocreate.Wrap(err, "failed to scan campaign")
		}
		campaigns = append(campaigns, campaign)
	}

	return dto.GetCampaignsResponse{
		Campaigns: campaigns,
		Total:     int(total),
	}, nil
}

func (c *Campaign) GetCampaignByID(ctx context.Context, campaignID uuid.UUID) (dto.CampaignResponse, error) {
	query := `
		SELECT id, title, message_type, subject, content, created_by, status, scheduled_at, sent_at, 
		       total_recipients, delivered_count, read_count, created_at, updated_at
		FROM message_campaigns
		WHERE id = $1
	`

	var campaign dto.CampaignResponse
	err := c.db.GetPool().QueryRow(ctx, query, campaignID).Scan(
		&campaign.ID,
		&campaign.Title,
		&campaign.MessageType,
		&campaign.Subject,
		&campaign.Content,
		&campaign.CreatedBy,
		&campaign.Status,
		&campaign.ScheduledAt,
		&campaign.SentAt,
		&campaign.TotalRecipients,
		&campaign.DeliveredCount,
		&campaign.ReadCount,
		&campaign.CreatedAt,
		&campaign.UpdatedAt,
	)
	if err != nil {
		c.log.Error("failed to get campaign by ID", zap.Error(err))
		return dto.CampaignResponse{}, errors.ErrUnableTocreate.Wrap(err, "failed to get campaign by ID")
	}

	return campaign, nil
}

func (c *Campaign) UpdateCampaign(ctx context.Context, campaignID uuid.UUID, req dto.UpdateCampaignRequest) (dto.CampaignResponse, error) {
	query := `
		UPDATE message_campaigns
		SET 
			title = COALESCE($2, title),
			subject = COALESCE($3, subject),
			content = COALESCE($4, content),
			scheduled_at = COALESCE($5, scheduled_at),
			status = COALESCE($6, status),
			updated_at = NOW()
		WHERE id = $1
		RETURNING id, title, message_type, subject, content, created_by, status, scheduled_at, sent_at, 
		          total_recipients, delivered_count, read_count, created_at, updated_at
	`

	var campaign dto.CampaignResponse
	err := c.db.GetPool().QueryRow(ctx, query, campaignID, req.Title, req.Subject, req.Content, req.ScheduledAt, req.Status).Scan(
		&campaign.ID,
		&campaign.Title,
		&campaign.MessageType,
		&campaign.Subject,
		&campaign.Content,
		&campaign.CreatedBy,
		&campaign.Status,
		&campaign.ScheduledAt,
		&campaign.SentAt,
		&campaign.TotalRecipients,
		&campaign.DeliveredCount,
		&campaign.ReadCount,
		&campaign.CreatedAt,
		&campaign.UpdatedAt,
	)
	if err != nil {
		c.log.Error("failed to update campaign", zap.Error(err))
		return dto.CampaignResponse{}, errors.ErrUnableTocreate.Wrap(err, "failed to update campaign")
	}

	return campaign, nil
}

func (c *Campaign) DeleteCampaign(ctx context.Context, campaignID uuid.UUID) error {
	query := `DELETE FROM message_campaigns WHERE id = $1`
	_, err := c.db.GetPool().Exec(ctx, query, campaignID)
	if err != nil {
		c.log.Error("failed to delete campaign", zap.Error(err))
		return errors.ErrUnableTocreate.Wrap(err, "failed to delete campaign")
	}
	return nil
}

func (c *Campaign) GetCampaignRecipients(ctx context.Context, req dto.GetCampaignRecipientsRequest) (dto.GetCampaignRecipientsResponse, error) {
	c.log.Info("GetCampaignRecipients called", zap.String("campaign_id", req.CampaignID.String()), zap.Int("page", req.Page), zap.Int("per_page", req.PerPage))

	query := `
		SELECT 
			cr.id, cr.campaign_id, cr.user_id, cr.notification_id, cr.status, 
			cr.sent_at, cr.delivered_at, cr.read_at, cr.error_message, cr.created_at,
			COUNT(*) OVER() AS total
		FROM campaign_recipients cr
		WHERE 
			cr.campaign_id = $1 AND
			($2::text IS NULL OR cr.status = $2)
		ORDER BY cr.created_at DESC
		LIMIT $3 OFFSET $4
	`

	c.log.Info("Executing query", zap.String("query", query), zap.String("campaign_id", req.CampaignID.String()))
	rows, err := c.db.GetPool().Query(ctx, query, req.CampaignID, req.Status, req.PerPage, (req.Page-1)*req.PerPage)
	if err != nil {
		c.log.Error("failed to get campaign recipients", zap.Error(err))
		return dto.GetCampaignRecipientsResponse{}, errors.ErrUnableTocreate.Wrap(err, "failed to get campaign recipients")
	}
	defer rows.Close()

	c.log.Info("Query executed successfully, processing rows")
	var recipients []dto.RecipientResponse
	var total int64

	for rows.Next() {
		var recipient dto.RecipientResponse
		c.log.Info("Scanning row")
		err := rows.Scan(
			&recipient.ID,
			&recipient.CampaignID,
			&recipient.UserID,
			&recipient.NotificationID,
			&recipient.Status,
			&recipient.SentAt,
			&recipient.DeliveredAt,
			&recipient.ReadAt,
			&recipient.ErrorMessage,
			&recipient.CreatedAt,
			&total,
		)
		if err != nil {
			c.log.Error("failed to scan recipient", zap.Error(err))
			return dto.GetCampaignRecipientsResponse{}, errors.ErrUnableTocreate.Wrap(err, "failed to scan recipient")
		}
		c.log.Info("Row scanned successfully", zap.String("recipient_id", recipient.ID.String()))
		recipients = append(recipients, recipient)
	}

	c.log.Info("Finished processing rows", zap.Int("recipient_count", len(recipients)), zap.Int64("total", total))

	return dto.GetCampaignRecipientsResponse{
		Recipients: recipients,
		Total:      int(total),
	}, nil
}

func (c *Campaign) GetCampaignStats(ctx context.Context, campaignID uuid.UUID) (dto.CampaignStatsResponse, error) {
	query := `
		SELECT 
			campaign_id,
			COUNT(*) as total_recipients,
			COUNT(CASE WHEN status = 'sent' THEN 1 END) as sent_count,
			COUNT(CASE WHEN status = 'delivered' THEN 1 END) as delivered_count,
			COUNT(CASE WHEN status = 'read' THEN 1 END) as read_count,
			COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed_count
		FROM campaign_recipients
		WHERE campaign_id = $1
		GROUP BY campaign_id
	`

	var stats dto.CampaignStatsResponse
	err := c.db.GetPool().QueryRow(ctx, query, campaignID).Scan(
		&stats.CampaignID,
		&stats.TotalRecipients,
		&stats.SentCount,
		&stats.DeliveredCount,
		&stats.ReadCount,
		&stats.FailedCount,
	)
	if err != nil {
		c.log.Error("failed to get campaign stats", zap.Error(err))
		return dto.CampaignStatsResponse{}, errors.ErrUnableTocreate.Wrap(err, "failed to get campaign stats")
	}

	return stats, nil
}

func (c *Campaign) GetScheduledCampaigns(ctx context.Context) ([]dto.CampaignResponse, error) {
	query := `
		SELECT id, title, message_type, subject, content, created_by, status, scheduled_at, sent_at, 
		       total_recipients, delivered_count, read_count, created_at, updated_at
		FROM message_campaigns
		WHERE 
			status = 'scheduled' AND
			scheduled_at <= NOW()
		ORDER BY scheduled_at
	`

	rows, err := c.db.GetPool().Query(ctx, query)
	if err != nil {
		c.log.Error("failed to get scheduled campaigns", zap.Error(err))
		return nil, errors.ErrUnableTocreate.Wrap(err, "failed to get scheduled campaigns")
	}
	defer rows.Close()

	var campaigns []dto.CampaignResponse
	for rows.Next() {
		var campaign dto.CampaignResponse
		err := rows.Scan(
			&campaign.ID,
			&campaign.Title,
			&campaign.MessageType,
			&campaign.Subject,
			&campaign.Content,
			&campaign.CreatedBy,
			&campaign.Status,
			&campaign.ScheduledAt,
			&campaign.SentAt,
			&campaign.TotalRecipients,
			&campaign.DeliveredCount,
			&campaign.ReadCount,
			&campaign.CreatedAt,
			&campaign.UpdatedAt,
		)
		if err != nil {
			c.log.Error("failed to scan scheduled campaign", zap.Error(err))
			return nil, errors.ErrUnableTocreate.Wrap(err, "failed to scan scheduled campaign")
		}
		campaigns = append(campaigns, campaign)
	}

	return campaigns, nil
}

func (c *Campaign) UpdateRecipientStatus(ctx context.Context, recipientID uuid.UUID, status dto.RecipientStatus, errorMessage string) error {
	query := `
		UPDATE campaign_recipients
		SET 
			status = $2,
			sent_at = CASE WHEN $2 = 'sent' THEN NOW() ELSE sent_at END,
			delivered_at = CASE WHEN $2 = 'delivered' THEN NOW() ELSE delivered_at END,
			read_at = CASE WHEN $2 = 'read' THEN NOW() ELSE read_at END,
			error_message = CASE WHEN $2 = 'failed' THEN $3 ELSE error_message END
		WHERE id = $1
	`
	_, err := c.db.GetPool().Exec(ctx, query, recipientID, string(status), errorMessage)
	if err != nil {
		c.log.Error("failed to update recipient status", zap.Error(err))
		return errors.ErrUnableTocreate.Wrap(err, "failed to update recipient status")
	}
	return nil
}

func (c *Campaign) UpdateRecipientNotificationID(ctx context.Context, recipientID uuid.UUID, notificationID uuid.UUID) error {
	query := `UPDATE campaign_recipients SET notification_id = $2 WHERE id = $1`
	_, err := c.db.GetPool().Exec(ctx, query, recipientID, notificationID)
	if err != nil {
		c.log.Error("failed to update recipient notification ID", zap.Error(err))
		return errors.ErrUnableTocreate.Wrap(err, "failed to update recipient notification ID")
	}
	return nil
}

func (c *Campaign) MarkCampaignAsSent(ctx context.Context, campaignID uuid.UUID) error {
	c.log.Info("Marking campaign as sent", zap.String("campaign_id", campaignID.String()))

	query := `
		UPDATE message_campaigns
		SET 
			status = 'sent',
			sent_at = NOW(),
			updated_at = NOW()
		WHERE id = $1
	`
	result, err := c.db.GetPool().Exec(ctx, query, campaignID)
	if err != nil {
		c.log.Error("failed to mark campaign as sent", zap.Error(err))
		return errors.ErrUnableTocreate.Wrap(err, "failed to mark campaign as sent")
	}

	rowsAffected := result.RowsAffected()
	c.log.Info("Campaign marked as sent",
		zap.String("campaign_id", campaignID.String()),
		zap.Int64("rows_affected", rowsAffected))

	return nil
}

// createCampaignRecipients creates campaign recipients for a segment
func (c *Campaign) createCampaignRecipients(ctx context.Context, tx pgx.Tx, campaignID uuid.UUID, segmentReq dto.CreateSegmentRequest, userCount int) error {
	switch segmentReq.SegmentType {
	case dto.SEGMENT_TYPE_ALL_USERS:
		// Get all users
		query := `SELECT id FROM users WHERE is_admin IS NOT TRUE AND user_type = 'PLAYER'`
		rows, err := tx.Query(ctx, query)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var userID uuid.UUID
			if err := rows.Scan(&userID); err != nil {
				return err
			}

			// Create campaign recipient
			recipientQuery := `
				INSERT INTO campaign_recipients (campaign_id, user_id, status)
				VALUES ($1, $2, 'pending')
			`
			_, err = tx.Exec(ctx, recipientQuery, campaignID, userID)
			if err != nil {
				c.log.Error("failed to create campaign recipient", zap.Error(err))
				return err
			}
		}
		return nil

	case dto.SEGMENT_TYPE_CSV:
		// TODO: Implement CSV-based recipient creation
		c.log.Warn("CSV segment type not yet implemented")
		return nil

	case dto.SEGMENT_TYPE_CRITERIA:
		// TODO: Implement criteria-based recipient creation
		c.log.Warn("Criteria segment type not yet implemented")
		return nil

	default:
		return fmt.Errorf("unknown segment type: %s", segmentReq.SegmentType)
	}
}

func (c *Campaign) GetCampaignNotificationsDashboard(ctx context.Context, req dto.GetCampaignNotificationsDashboardRequest) ([]dto.CampaignNotificationDashboardItem, error) {
	c.log.Info("Getting campaign notifications dashboard", zap.Int("page", req.Page), zap.Int("per_page", req.PerPage))

	query := `
		SELECT 
			un.id, un.user_id, u.username, u.email, un.title, un.content, un.type,
			un.read, un.delivered, un.read_at, un.created_at,
			cr.campaign_id, mc.title as campaign_title
		FROM user_notifications un
		LEFT JOIN users u ON un.user_id = u.id
		LEFT JOIN campaign_recipients cr ON un.id = cr.notification_id
		LEFT JOIN message_campaigns mc ON cr.campaign_id = mc.id
		WHERE 
			($1::text IS NULL OR un.type = $1) AND
			($2::uuid IS NULL OR un.user_id = $2) AND
			($3::text IS NULL OR CASE WHEN un.read THEN 'read' ELSE 'unread' END = $3)
		ORDER BY un.created_at DESC
		LIMIT $4 OFFSET $5
	`

	rows, err := c.db.GetPool().Query(ctx, query, req.NotificationType, req.UserID, req.Status, req.PerPage, (req.Page-1)*req.PerPage)
	if err != nil {
		c.log.Error("failed to get campaign notifications dashboard", zap.Error(err))
		return nil, errors.ErrUnableTocreate.Wrap(err, "failed to get campaign notifications dashboard")
	}
	defer rows.Close()

	var notifications []dto.CampaignNotificationDashboardItem
	for rows.Next() {
		var notification dto.CampaignNotificationDashboardItem
		err := rows.Scan(
			&notification.ID,
			&notification.UserID,
			&notification.Username,
			&notification.Email,
			&notification.Title,
			&notification.Content,
			&notification.Type,
			&notification.Read,
			&notification.Delivered,
			&notification.ReadAt,
			&notification.CreatedAt,
			&notification.CampaignID,
			&notification.CampaignTitle,
		)
		if err != nil {
			c.log.Error("failed to scan campaign notification", zap.Error(err))
			return nil, errors.ErrUnableTocreate.Wrap(err, "failed to scan campaign notification")
		}
		notifications = append(notifications, notification)
	}

	return notifications, nil
}

func (c *Campaign) GetCampaignNotificationStats(ctx context.Context, req dto.GetCampaignNotificationsDashboardRequest) (dto.CampaignNotificationStats, error) {
	c.log.Info("Getting campaign notification stats")

	query := `
		SELECT 
			COUNT(*) as total_notifications,
			COUNT(*) FILTER (WHERE NOT read) as unread_notifications,
			COUNT(*) FILTER (WHERE delivered) as delivered_notifications,
			COUNT(*) FILTER (WHERE read) as read_notifications
		FROM user_notifications un
		WHERE 
			($1::text IS NULL OR un.type = $1) AND
			($2::uuid IS NULL OR un.user_id = $2) AND
			($3::text IS NULL OR CASE WHEN un.read THEN 'read' ELSE 'unread' END = $3)
	`

	var stats dto.CampaignNotificationStats
	var totalNotifications, unreadNotifications, deliveredNotifications, readNotifications int64

	err := c.db.GetPool().QueryRow(ctx, query, req.NotificationType, req.UserID, req.Status).Scan(
		&totalNotifications,
		&unreadNotifications,
		&deliveredNotifications,
		&readNotifications,
	)
	if err != nil {
		c.log.Error("failed to get campaign notification stats", zap.Error(err))
		return dto.CampaignNotificationStats{}, errors.ErrUnableTocreate.Wrap(err, "failed to get campaign notification stats")
	}

	stats.TotalNotifications = int(totalNotifications)
	stats.UnreadNotifications = int(unreadNotifications)
	stats.DeliveredNotifications = int(deliveredNotifications)
	stats.ReadNotifications = int(readNotifications)

	if totalNotifications > 0 {
		stats.DeliveryRate = float64(deliveredNotifications) / float64(totalNotifications) * 100
		stats.ReadRate = float64(readNotifications) / float64(totalNotifications) * 100
	}

	return stats, nil
}
