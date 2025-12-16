package alert

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lib/pq"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/persistencedb"
	"go.uber.org/zap"
)

// UUIDArray is a helper type for scanning PostgreSQL UUID arrays
type UUIDArray []uuid.UUID

// Scan implements the sql.Scanner interface for UUIDArray
func (a *UUIDArray) Scan(src interface{}) error {
	if src == nil {
		*a = []uuid.UUID{}
		return nil
	}

	var strArray pq.StringArray
	if err := strArray.Scan(src); err != nil {
		// If scanning as string array fails, try direct byte/string parsing
		switch v := src.(type) {
		case []byte:
			return a.scanBytes(v)
		case string:
			return a.scanBytes([]byte(v))
		default:
			return fmt.Errorf("cannot scan %T into UUIDArray", src)
		}
	}

	result := make([]uuid.UUID, 0, len(strArray))
	for _, str := range strArray {
		if id, err := uuid.Parse(str); err == nil {
			result = append(result, id)
		}
	}
	*a = result
	return nil
}

func (a *UUIDArray) scanBytes(src []byte) error {
	// PostgreSQL array format: {uuid1,uuid2,uuid3}
	if len(src) < 2 || src[0] != '{' || src[len(src)-1] != '}' {
		*a = []uuid.UUID{}
		return nil
	}

	str := string(src[1 : len(src)-1]) // Remove { }
	if str == "" {
		*a = []uuid.UUID{}
		return nil
	}

	parts := strings.Split(str, ",")
	result := make([]uuid.UUID, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		// Remove quotes if present
		if len(part) > 0 && part[0] == '"' && part[len(part)-1] == '"' {
			part = part[1 : len(part)-1]
		}
		if id, err := uuid.Parse(part); err == nil {
			result = append(result, id)
		}
	}
	*a = result
	return nil
}

type AlertStorage interface {
	// Alert Configuration methods
	CreateAlertConfiguration(ctx context.Context, req *dto.CreateAlertConfigurationRequest, createdBy uuid.UUID) (*dto.AlertConfiguration, error)
	GetAlertConfigurationByID(ctx context.Context, id uuid.UUID) (*dto.AlertConfiguration, error)
	GetAlertConfigurations(ctx context.Context, req *dto.GetAlertConfigurationsRequest) ([]dto.AlertConfiguration, int64, error)
	UpdateAlertConfiguration(ctx context.Context, id uuid.UUID, req *dto.UpdateAlertConfigurationRequest, updatedBy uuid.UUID) (*dto.AlertConfiguration, error)
	DeleteAlertConfiguration(ctx context.Context, id uuid.UUID) error

	// Alert Trigger methods
	CreateAlertTrigger(ctx context.Context, trigger *dto.AlertTrigger) error
	GetAlertTriggers(ctx context.Context, req *dto.GetAlertTriggersRequest) ([]dto.AlertTrigger, int64, error)
	GetAlertTriggerByID(ctx context.Context, id uuid.UUID) (*dto.AlertTrigger, error)
	AcknowledgeAlert(ctx context.Context, id uuid.UUID, acknowledgedBy uuid.UUID) error

	// Alert checking methods (to be implemented)
	CheckBetCountAlerts(ctx context.Context, timeWindow time.Duration) error
	CheckBetAmountAlerts(ctx context.Context, timeWindow time.Duration) error
	CheckDepositAlerts(ctx context.Context, timeWindow time.Duration, skipDuplicateCheck bool) error
	CheckWithdrawalAlerts(ctx context.Context, timeWindow time.Duration) error
	CheckGGRAlerts(ctx context.Context, timeWindow time.Duration) error
	CheckMultipleAccountsSameIP(ctx context.Context, timeWindow time.Duration) error
}

type alertStorage struct {
	db  persistencedb.PersistenceDB
	log *zap.Logger
}

func NewAlertStorage(db persistencedb.PersistenceDB, log *zap.Logger) AlertStorage {
	return &alertStorage{
		db:  db,
		log: log,
	}
}

// CreateAlertConfiguration creates a new alert configuration
func (s *alertStorage) CreateAlertConfiguration(ctx context.Context, req *dto.CreateAlertConfigurationRequest, createdBy uuid.UUID) (*dto.AlertConfiguration, error) {
	// Check if an active alert configuration with the same alert_type already exists
	checkQuery := `
		SELECT id FROM alert_configurations 
		WHERE alert_type = $1 AND status = 'active'
		LIMIT 1
	`
	var existingID uuid.UUID
	err := s.db.GetPool().QueryRow(ctx, checkQuery, req.AlertType).Scan(&existingID)
	if err == nil {
		// An active alert with this type already exists
		return nil, fmt.Errorf("an active alert configuration with alert_type '%s' already exists. Only one active alert per type is allowed", req.AlertType)
	} else if err != nil && err != pgx.ErrNoRows {
		// Database error occurred
		s.log.Error("Failed to check for existing alert configuration", zap.Error(err))
		return nil, err
	}

	// CreateAlertConfigurationRequest doesn't have Status field - always create as 'active'
	// The check above already ensures no active alert of this type exists

	// Start a transaction to ensure atomicity and better error handling
	tx, err := s.db.GetPool().Begin(ctx)
	if err != nil {
		s.log.Error("Failed to begin transaction", zap.Error(err))
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Prepare email_group_ids array
	var emailGroupIDs interface{} = pq.Array([]uuid.UUID{})
	if len(req.EmailGroupIDs) > 0 {
		emailGroupIDs = pq.Array(req.EmailGroupIDs)
	}

	// Handle nullable fields - use NULL for empty strings
	var descriptionVal interface{} = nil
	if req.Description != nil && *req.Description != "" {
		descriptionVal = *req.Description
	}

	var currencyCodeVal interface{} = nil
	if req.CurrencyCode != nil {
		currencyCodeVal = string(*req.CurrencyCode)
	}

	var webhookURLVal interface{} = nil
	if req.WebhookURL != nil && *req.WebhookURL != "" {
		webhookURLVal = *req.WebhookURL
	}

	query := `
        INSERT INTO alert_configurations (
            name, description, alert_type, status, threshold_amount, time_window_minutes,
            currency_code, email_notifications, webhook_url, email_group_ids, created_by
        ) VALUES (
            $1, $2, $3, 'active'::alert_status, $4, $5, $6, $7, $8, $9, $10
        ) RETURNING 
            id, name, description, alert_type, status, threshold_amount, time_window_minutes,
            currency_code, email_notifications, webhook_url, email_group_ids,
            created_by, created_at, updated_at, updated_by
    `

	s.log.Info("Creating alert configuration",
		zap.String("name", req.Name),
		zap.String("alert_type", string(req.AlertType)),
		zap.Any("email_group_ids", req.EmailGroupIDs),
		zap.Any("description", descriptionVal),
		zap.Any("currency_code", currencyCodeVal),
		zap.Any("webhook_url", webhookURLVal),
	)

	// Execute the query and scan the result
	var config dto.AlertConfiguration
	var scannedEmailGroupIDs UUIDArray

	// Use sql.NullString for nullable text/enum fields when scanning
	var scannedDescription sql.NullString
	var scannedCurrencyCode sql.NullString
	var scannedWebhookURL sql.NullString

	err = tx.QueryRow(ctx, query,
		req.Name, descriptionVal, req.AlertType, req.ThresholdAmount,
		req.TimeWindowMinutes, currencyCodeVal, req.EmailNotifications,
		webhookURLVal, emailGroupIDs, createdBy,
	).Scan(
		&config.ID, &config.Name, &scannedDescription, &config.AlertType,
		&config.Status, &config.ThresholdAmount, &config.TimeWindowMinutes,
		&scannedCurrencyCode, &config.EmailNotifications, &scannedWebhookURL,
		&scannedEmailGroupIDs, &config.CreatedBy, &config.CreatedAt, &config.UpdatedAt, &config.UpdatedBy,
	)

	// Convert scanned nullable fields to pointers
	if scannedDescription.Valid {
		config.Description = &scannedDescription.String
	}
	if scannedCurrencyCode.Valid {
		currencyType := dto.CurrencyType(scannedCurrencyCode.String)
		config.CurrencyCode = &currencyType
	}
	if scannedWebhookURL.Valid {
		config.WebhookURL = &scannedWebhookURL.String
	}

	if err == nil {
		config.EmailGroupIDs = []uuid.UUID(scannedEmailGroupIDs)
		// Commit the transaction
		if commitErr := tx.Commit(ctx); commitErr != nil {
			s.log.Error("Failed to commit transaction", zap.Error(commitErr))
			return nil, commitErr
		}
	}

	if err != nil {
		// Rollback transaction on error
		tx.Rollback(ctx)

		// Check if it's a unique constraint violation
		if strings.Contains(err.Error(), "idx_alert_configurations_type_unique_active") ||
			strings.Contains(err.Error(), "unique constraint") {
			return nil, fmt.Errorf("an active alert configuration with alert_type '%s' already exists. Only one active alert per type is allowed", req.AlertType)
		}

		// Check for foreign key constraint violations
		if strings.Contains(err.Error(), "foreign key") || strings.Contains(err.Error(), "violates foreign key constraint") {
			return nil, fmt.Errorf("invalid reference: %w", err)
		}

		// Log detailed error information
		s.log.Error("Failed to create alert configuration",
			zap.Error(err),
			zap.String("error_type", fmt.Sprintf("%T", err)),
			zap.String("error_message", err.Error()),
			zap.String("name", req.Name),
			zap.String("alert_type", string(req.AlertType)),
			zap.Float64("threshold_amount", req.ThresholdAmount),
			zap.Int("time_window_minutes", req.TimeWindowMinutes),
			zap.Bool("email_notifications", req.EmailNotifications),
			zap.Any("description", descriptionVal),
			zap.Any("currency_code", currencyCodeVal),
			zap.Any("webhook_url", webhookURLVal),
			zap.Any("email_group_ids", emailGroupIDs),
			zap.String("created_by", createdBy.String()),
		)

		// Return a more descriptive error
		if err == pgx.ErrNoRows || err == sql.ErrNoRows || strings.Contains(err.Error(), "no rows in result set") {
			return nil, fmt.Errorf("insert query returned no rows - this may indicate a database constraint, trigger, or data type issue. Please check server logs for details: %w", err)
		}

		return nil, fmt.Errorf("failed to create alert configuration: %w", err)
	}

	return &config, nil
}

// GetAlertConfigurationByID gets an alert configuration by ID
func (s *alertStorage) GetAlertConfigurationByID(ctx context.Context, id uuid.UUID) (*dto.AlertConfiguration, error) {
	query := `SELECT 
		id, name, description, alert_type, status, threshold_amount, time_window_minutes,
		currency_code, email_notifications, webhook_url, email_group_ids,
		created_by, created_at, updated_at, updated_by
		FROM alert_configurations WHERE id = $1`

	var config dto.AlertConfiguration
	var emailGroupIDs UUIDArray
	err := s.db.GetPool().QueryRow(ctx, query, id).Scan(
		&config.ID, &config.Name, &config.Description, &config.AlertType,
		&config.Status, &config.ThresholdAmount, &config.TimeWindowMinutes,
		&config.CurrencyCode, &config.EmailNotifications, &config.WebhookURL,
		&emailGroupIDs, &config.CreatedBy, &config.CreatedAt, &config.UpdatedAt, &config.UpdatedBy,
	)
	if err == nil {
		config.EmailGroupIDs = []uuid.UUID(emailGroupIDs)
	}

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		s.log.Error("Failed to get alert configuration", zap.Error(err))
		return nil, err
	}

	return &config, nil
}

// GetAlertConfigurations gets alert configurations with filtering and pagination
func (s *alertStorage) GetAlertConfigurations(ctx context.Context, req *dto.GetAlertConfigurationsRequest) ([]dto.AlertConfiguration, int64, error) {
	// Build WHERE clause
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argIndex := 1

	if req.AlertType != nil {
		whereClause += fmt.Sprintf(" AND alert_type = $%d", argIndex)
		args = append(args, *req.AlertType)
		argIndex++
	}

	if req.Status != nil {
		whereClause += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, *req.Status)
		argIndex++
	}

	if req.Search != "" {
		whereClause += fmt.Sprintf(" AND (name ILIKE $%d OR description ILIKE $%d)", argIndex, argIndex+1)
		searchTerm := "%" + req.Search + "%"
		args = append(args, searchTerm, searchTerm)
		argIndex += 2
	}

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM alert_configurations %s", whereClause)
	var totalCount int64
	err := s.db.GetPool().QueryRow(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		s.log.Error("Failed to get alert configurations count", zap.Error(err))
		return nil, 0, err
	}

	// Set default pagination
	page := req.Page
	if page <= 0 {
		page = 1
	}
	perPage := req.PerPage
	if perPage <= 0 {
		perPage = 20
	}

	offset := (page - 1) * perPage

	// Get configurations
	query := fmt.Sprintf(`
		SELECT 
			id, name, description, alert_type, status, threshold_amount, time_window_minutes,
			currency_code, email_notifications, webhook_url, email_group_ids,
			created_by, created_at, updated_at, updated_by
		FROM alert_configurations 
		%s 
		ORDER BY created_at DESC 
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, perPage, offset)

	rows, err := s.db.GetPool().Query(ctx, query, args...)
	if err != nil {
		s.log.Error("Failed to get alert configurations", zap.Error(err))
		return nil, 0, err
	}
	defer rows.Close()

	var configs []dto.AlertConfiguration
	for rows.Next() {
		var config dto.AlertConfiguration
		var emailGroupIDs UUIDArray
		err := rows.Scan(
			&config.ID, &config.Name, &config.Description, &config.AlertType,
			&config.Status, &config.ThresholdAmount, &config.TimeWindowMinutes,
			&config.CurrencyCode, &config.EmailNotifications, &config.WebhookURL,
			&emailGroupIDs, &config.CreatedBy, &config.CreatedAt, &config.UpdatedAt, &config.UpdatedBy,
		)
		if err != nil {
			s.log.Error("Failed to scan alert configuration", zap.Error(err))
			return nil, 0, err
		}
		config.EmailGroupIDs = []uuid.UUID(emailGroupIDs)
		configs = append(configs, config)
	}

	return configs, totalCount, nil
}

// UpdateAlertConfiguration updates an alert configuration
func (s *alertStorage) UpdateAlertConfiguration(ctx context.Context, id uuid.UUID, req *dto.UpdateAlertConfigurationRequest, updatedBy uuid.UUID) (*dto.AlertConfiguration, error) {
	setClauses := []string{}
	args := []interface{}{}
	argIdx := 1

	if req.Name != nil {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", argIdx))
		args = append(args, *req.Name)
		argIdx++
	}
	if req.Description != nil {
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", argIdx))
		args = append(args, *req.Description)
		argIdx++
	}
	if req.Status != nil {
		setClauses = append(setClauses, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, *req.Status)
		argIdx++
	}
	if req.ThresholdAmount != nil {
		setClauses = append(setClauses, fmt.Sprintf("threshold_amount = $%d", argIdx))
		args = append(args, *req.ThresholdAmount)
		argIdx++
	}
	if req.TimeWindowMinutes != nil {
		setClauses = append(setClauses, fmt.Sprintf("time_window_minutes = $%d", argIdx))
		args = append(args, *req.TimeWindowMinutes)
		argIdx++
	}
	if req.CurrencyCode != nil {
		setClauses = append(setClauses, fmt.Sprintf("currency_code = $%d", argIdx))
		args = append(args, *req.CurrencyCode)
		argIdx++
	}
	if req.EmailNotifications != nil {
		setClauses = append(setClauses, fmt.Sprintf("email_notifications = $%d", argIdx))
		args = append(args, *req.EmailNotifications)
		argIdx++
	}
	if req.WebhookURL != nil {
		setClauses = append(setClauses, fmt.Sprintf("webhook_url = $%d", argIdx))
		args = append(args, *req.WebhookURL)
		argIdx++
	}
	if req.EmailGroupIDs != nil {
		setClauses = append(setClauses, fmt.Sprintf("email_group_ids = $%d", argIdx))
		var emailGroupIDs interface{} = pq.Array([]uuid.UUID{})
		if len(req.EmailGroupIDs) > 0 {
			emailGroupIDs = pq.Array(req.EmailGroupIDs)
		}
		args = append(args, emailGroupIDs)
		argIdx++
	}

	// Always update updated_at and updated_by
	setClauses = append(setClauses, "updated_at = NOW()")
	setClauses = append(setClauses, fmt.Sprintf("updated_by = $%d", argIdx))
	args = append(args, updatedBy)
	argIdx++

	if len(setClauses) == 0 {
		// Nothing to update; return current
		return s.GetAlertConfigurationByID(ctx, id)
	}

	query := fmt.Sprintf(`
        UPDATE alert_configurations
        SET %s
        WHERE id = $%d
        RETURNING 
            id, name, description, alert_type, status, threshold_amount, time_window_minutes,
            currency_code, email_notifications, webhook_url, email_group_ids,
            created_by, created_at, updated_at, updated_by
    `, strings.Join(setClauses, ", "), argIdx)

	args = append(args, id)

	var config dto.AlertConfiguration
	var emailGroupIDs UUIDArray
	err := s.db.GetPool().QueryRow(ctx, query, args...).Scan(
		&config.ID, &config.Name, &config.Description, &config.AlertType,
		&config.Status, &config.ThresholdAmount, &config.TimeWindowMinutes,
		&config.CurrencyCode, &config.EmailNotifications, &config.WebhookURL,
		&emailGroupIDs, &config.CreatedBy, &config.CreatedAt, &config.UpdatedAt, &config.UpdatedBy,
	)
	if err == nil {
		config.EmailGroupIDs = []uuid.UUID(emailGroupIDs)
	}
	if err != nil {
		s.log.Error("Failed to update alert configuration", zap.Error(err))
		return nil, err
	}

	return &config, nil
}

// DeleteAlertConfiguration deletes an alert configuration
func (s *alertStorage) DeleteAlertConfiguration(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM alert_configurations WHERE id = $1`

	result, err := s.db.GetPool().Exec(ctx, query, id)
	if err != nil {
		s.log.Error("Failed to delete alert configuration", zap.Error(err))
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("alert configuration not found")
	}

	return nil
}

// CreateAlertTrigger creates a new alert trigger
func (s *alertStorage) CreateAlertTrigger(ctx context.Context, trigger *dto.AlertTrigger) error {
	query := `
        INSERT INTO alert_triggers (
            alert_configuration_id, triggered_at, trigger_value, threshold_value,
            user_id, transaction_id, amount_usd, currency_code, context_data
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9
        ) RETURNING id
    `

	err := s.db.GetPool().QueryRow(ctx, query,
		trigger.AlertConfigurationID, trigger.TriggeredAt, trigger.TriggerValue,
		trigger.ThresholdValue, trigger.UserID, trigger.TransactionID,
		trigger.AmountUSD, trigger.CurrencyCode, trigger.ContextData,
	).Scan(&trigger.ID)

	if err != nil {
		s.log.Error("Failed to create alert trigger", zap.Error(err))
		return err
	}

	return nil
}

// GetAlertTriggers gets alert triggers with filtering and pagination
func (s *alertStorage) GetAlertTriggers(ctx context.Context, req *dto.GetAlertTriggersRequest) ([]dto.AlertTrigger, int64, error) {
	where := "WHERE 1=1"
	args := []interface{}{}
	idx := 1

	if req.AlertConfigurationID != nil {
		where += fmt.Sprintf(" AND t.alert_configuration_id = $%d", idx)
		args = append(args, *req.AlertConfigurationID)
		idx++
	}
	if req.UserID != nil {
		where += fmt.Sprintf(" AND t.user_id = $%d", idx)
		args = append(args, *req.UserID)
		idx++
	}
	if req.Acknowledged != nil {
		where += fmt.Sprintf(" AND t.acknowledged = $%d", idx)
		args = append(args, *req.Acknowledged)
		idx++
	}
	if req.DateFrom != nil {
		where += fmt.Sprintf(" AND t.triggered_at >= $%d", idx)
		args = append(args, *req.DateFrom)
		idx++
	}
	if req.DateTo != nil {
		where += fmt.Sprintf(" AND t.triggered_at <= $%d", idx)
		args = append(args, *req.DateTo)
		idx++
	}
	if req.Search != "" {
		where += fmt.Sprintf(" AND (u.username ILIKE $%d OR u.email ILIKE $%d OR t.transaction_id ILIKE $%d)", idx, idx, idx)
		args = append(args, "%"+req.Search+"%")
		idx++
	}

	// Count
	countQuery := fmt.Sprintf(`
        SELECT COUNT(*)
        FROM alert_triggers t
        LEFT JOIN users u ON u.id = t.user_id
        LEFT JOIN alert_configurations ac ON ac.id = t.alert_configuration_id
        %s
    `, where)

	var total int64
	if err := s.db.GetPool().QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		s.log.Error("Failed to count alert triggers", zap.Error(err))
		return nil, 0, err
	}

	page := req.Page
	if page <= 0 {
		page = 1
	}
	perPage := req.PerPage
	if perPage <= 0 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	// Query rows
	query := fmt.Sprintf(`
        SELECT 
            t.id, t.alert_configuration_id, t.triggered_at, t.trigger_value, t.threshold_value,
            t.user_id::text, t.transaction_id, t.amount_usd, t.currency_code, t.context_data,
            t.acknowledged, t.acknowledged_by::text, t.acknowledged_at, t.created_at,
            ac.id, ac.name, ac.description, ac.alert_type, ac.status, ac.threshold_amount,
            ac.time_window_minutes, ac.currency_code, ac.email_notifications, ac.webhook_url,
            ac.email_group_ids, ac.created_by, ac.created_at, ac.updated_at, ac.updated_by,
            u.username, u.email
        FROM alert_triggers t
        LEFT JOIN alert_configurations ac ON ac.id = t.alert_configuration_id
        LEFT JOIN users u ON u.id = t.user_id
        %s
        ORDER BY t.triggered_at DESC
        LIMIT $%d OFFSET $%d
    `, where, idx, idx+1)

	args = append(args, perPage, offset)

	rows, err := s.db.GetPool().Query(ctx, query, args...)
	if err != nil {
		s.log.Error("Failed to get alert triggers", zap.Error(err))
		return nil, 0, err
	}
	defer rows.Close()

	triggers := []dto.AlertTrigger{}
	for rows.Next() {
		var t dto.AlertTrigger
		var ac dto.AlertConfiguration
		var username sql.NullString
		var email sql.NullString
		var userIDStr sql.NullString
		var acknowledgedByStr sql.NullString
		var transactionID sql.NullString
		var amountUSD sql.NullFloat64
		var currencyCode sql.NullString
		var contextData sql.NullString
		var acknowledgedAt sql.NullTime
		var emailGroupIDs UUIDArray

		if err := rows.Scan(
			&t.ID, &t.AlertConfigurationID, &t.TriggeredAt, &t.TriggerValue, &t.ThresholdValue,
			&userIDStr, &transactionID, &amountUSD, &currencyCode, &contextData,
			&t.Acknowledged, &acknowledgedByStr, &acknowledgedAt, &t.CreatedAt,
			&ac.ID, &ac.Name, &ac.Description, &ac.AlertType, &ac.Status, &ac.ThresholdAmount,
			&ac.TimeWindowMinutes, &ac.CurrencyCode, &ac.EmailNotifications, &ac.WebhookURL,
			&emailGroupIDs, &ac.CreatedBy, &ac.CreatedAt, &ac.UpdatedAt, &ac.UpdatedBy,
			&username, &email,
		); err != nil {
			s.log.Error("Failed to scan alert trigger", zap.Error(err))
			return nil, 0, err
		}

		// Handle nullable UUID fields - scan as string and parse
		if userIDStr.Valid && userIDStr.String != "" {
			if userUUID, err := uuid.Parse(userIDStr.String); err == nil {
				t.UserID = &userUUID
			} else {
				s.log.Warn("Failed to parse user_id", zap.String("user_id_str", userIDStr.String), zap.Error(err))
			}
		}

		if acknowledgedByStr.Valid && acknowledgedByStr.String != "" {
			if ackByUUID, err := uuid.Parse(acknowledgedByStr.String); err == nil {
				t.AcknowledgedBy = &ackByUUID
			} else {
				s.log.Warn("Failed to parse acknowledged_by", zap.String("acknowledged_by_str", acknowledgedByStr.String), zap.Error(err))
			}
		}

		// Handle nullable string fields
		if transactionID.Valid && transactionID.String != "" {
			t.TransactionID = &transactionID.String
		}

		// Handle nullable float fields
		if amountUSD.Valid {
			t.AmountUSD = &amountUSD.Float64
		}

		// Handle nullable currency code
		if currencyCode.Valid && currencyCode.String != "" {
			currencyType := dto.CurrencyType(currencyCode.String)
			t.CurrencyCode = &currencyType
		}

		// Handle nullable context data
		if contextData.Valid && contextData.String != "" {
			t.ContextData = &contextData.String
		}

		// Handle nullable timestamp
		if acknowledgedAt.Valid {
			t.AcknowledgedAt = &acknowledgedAt.Time
		}

		// Populate email group IDs
		ac.EmailGroupIDs = []uuid.UUID(emailGroupIDs)

		// attach joined fields
		t.AlertConfiguration = &ac

		// Set username from users table (not user_id)
		if username.Valid && username.String != "" {
			t.Username = &username.String
		}

		// Set user email from users table
		if email.Valid && email.String != "" {
			t.UserEmail = &email.String
		}

		triggers = append(triggers, t)
	}

	return triggers, total, nil
}

// GetAlertTriggerByID gets an alert trigger by ID
func (s *alertStorage) GetAlertTriggerByID(ctx context.Context, id uuid.UUID) (*dto.AlertTrigger, error) {
	// Simplified implementation for now
	return nil, nil
}

// AcknowledgeAlert acknowledges an alert trigger
func (s *alertStorage) AcknowledgeAlert(ctx context.Context, id uuid.UUID, acknowledgedBy uuid.UUID) error {
	query := `
		UPDATE alert_triggers 
		SET acknowledged = true, acknowledged_by = $1, acknowledged_at = NOW()
		WHERE id = $2
	`

	result, err := s.db.GetPool().Exec(ctx, query, acknowledgedBy, id)
	if err != nil {
		s.log.Error("Failed to acknowledge alert", zap.Error(err))
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("alert trigger not found")
	}

	return nil
}

// Alert checking methods (to be implemented based on business logic)
func (s *alertStorage) CheckBetCountAlerts(ctx context.Context, timeWindow time.Duration) error {
	// Get all active bet count alert configurations
	req := &dto.GetAlertConfigurationsRequest{
		Page:    1,
		PerPage: 100,
		Status:  func() *dto.AlertStatus { status := dto.AlertStatusActive; return &status }(),
	}

	configs, _, err := s.GetAlertConfigurations(ctx, req)
	if err != nil {
		s.log.Error("Failed to get alert configurations", zap.Error(err))
		return err
	}

	// Filter for bet count-related alerts
	betCountAlertTypes := []dto.AlertType{
		dto.AlertTypeBetsCountMore,
		dto.AlertTypeBetsCountLess,
	}

	betCountConfigs := make([]dto.AlertConfiguration, 0)
	for _, config := range configs {
		for _, alertType := range betCountAlertTypes {
			if config.AlertType == alertType {
				betCountConfigs = append(betCountConfigs, config)
				break
			}
		}
	}

	if len(betCountConfigs) == 0 {
		return nil // No bet count alerts configured
	}

	// Check each bet count alert configuration
	for _, config := range betCountConfigs {
		// Skip inactive alerts
		if config.Status != dto.AlertStatusActive {
			continue
		}

		// Use the config's time window instead of the parameter
		configTimeWindowStart := time.Now().Add(-time.Duration(config.TimeWindowMinutes) * time.Minute)

		// Count bets from all sources: transactions (bet-type rows), groove_transactions (wager type), sport_bets, plinko
		query := `
			SELECT COUNT(*)::float
			FROM (
				-- Main bet source: transactions table by transaction_type
				SELECT id::text as id, created_at
				FROM transactions
				WHERE created_at >= $1 AND created_at <= $2
				  AND transaction_type IN ('bet', 'groove_bet')
				UNION ALL
				SELECT id::text as id, created_at FROM groove_transactions 
				WHERE type = 'wager' AND created_at >= $1 AND created_at <= $2
				UNION ALL
				SELECT id::text as id, created_at FROM sport_bets 
				WHERE created_at >= $1 AND created_at <= $2
				UNION ALL
				SELECT id::text as id, timestamp as created_at FROM plinko 
				WHERE timestamp >= $1 AND timestamp <= $2
			) all_bets
		`

		var betCount float64
		err := s.db.GetPool().QueryRow(ctx, query, configTimeWindowStart, time.Now()).Scan(&betCount)
		if err != nil {
			s.log.Error("Failed to calculate bet count", zap.Error(err), zap.String("alert_id", config.ID.String()))
			continue
		}

		// Check if threshold is exceeded
		shouldTrigger := false
		if config.AlertType == dto.AlertTypeBetsCountMore {
			shouldTrigger = betCount >= config.ThresholdAmount
		} else if config.AlertType == dto.AlertTypeBetsCountLess {
			shouldTrigger = betCount <= config.ThresholdAmount
		}

		if shouldTrigger {
			// Check if we already triggered this alert recently (avoid duplicates within the last 5 minutes)
			// Use a shorter window (5 minutes) to prevent spam while still allowing new triggers
			recentWindowStart := time.Now().Add(-5 * time.Minute)
			recentTriggerQuery := `
				SELECT COUNT(*) FROM alert_triggers
				WHERE alert_configuration_id = $1
					AND triggered_at >= $2
			`
			var recentCount int64
			err := s.db.GetPool().QueryRow(ctx, recentTriggerQuery, config.ID, recentWindowStart).Scan(&recentCount)
			if err != nil {
				s.log.Error("Failed to check for recent triggers", zap.Error(err), zap.String("alert_id", config.ID.String()))
				// Continue anyway - don't block trigger creation due to check failure
			}

			if recentCount == 0 {
				// Create trigger
				trigger := &dto.AlertTrigger{
					AlertConfigurationID: config.ID,
					TriggeredAt:          time.Now(),
					TriggerValue:         betCount,
					ThresholdValue:       config.ThresholdAmount,
				}

				err = s.CreateAlertTrigger(ctx, trigger)
				if err != nil {
					s.log.Error("Failed to create bet count alert trigger", zap.Error(err), zap.String("alert_id", config.ID.String()))
				} else {
					s.log.Info("Bet count alert triggered",
						zap.String("alert_id", config.ID.String()),
						zap.String("alert_type", string(config.AlertType)),
						zap.Float64("bet_count", betCount),
						zap.Float64("threshold", config.ThresholdAmount),
						zap.String("trigger_id", trigger.ID.String()))
					// Note: user_id is null for aggregate alerts (total bet count across all users)
					// Email sending will be handled by the module layer's sendEmailsForNewTriggers
				}
			} else {
				s.log.Debug("Skipping duplicate trigger",
					zap.String("alert_id", config.ID.String()),
					zap.Int64("recent_count", recentCount))
			}
		}
	}

	return nil
}

func (s *alertStorage) CheckBetAmountAlerts(ctx context.Context, timeWindow time.Duration) error {
	// Get all active bet amount alert configurations
	req := &dto.GetAlertConfigurationsRequest{
		Page:    1,
		PerPage: 100,
		Status:  func() *dto.AlertStatus { status := dto.AlertStatusActive; return &status }(),
	}

	configs, _, err := s.GetAlertConfigurations(ctx, req)
	if err != nil {
		s.log.Error("Failed to get alert configurations", zap.Error(err))
		return err
	}

	// Filter for bet amount-related alerts
	betAmountAlertTypes := []dto.AlertType{
		dto.AlertTypeBetsAmountMore,
		dto.AlertTypeBetsAmountLess,
	}

	betAmountConfigs := make([]dto.AlertConfiguration, 0)
	for _, config := range configs {
		for _, alertType := range betAmountAlertTypes {
			if config.AlertType == alertType {
				betAmountConfigs = append(betAmountConfigs, config)
				break
			}
		}
	}

	if len(betAmountConfigs) == 0 {
		return nil // No bet amount alerts configured
	}

	// Check each bet amount alert configuration
	for _, config := range betAmountConfigs {
		// Use the config's time window
		configTimeWindowStart := time.Now().Add(-time.Duration(config.TimeWindowMinutes) * time.Minute)
		configTimeWindowEnd := time.Now()

		// Sum bet amounts from all sources (convert to USD)
		// Note: For now, we assume all amounts are in USD or the same currency
		// TODO: Add proper currency conversion if needed
		query := `
			SELECT COALESCE(SUM(amount_usd), 0)
			FROM (
				-- Bets table: handle NULL timestamps by using created_at as fallback
				SELECT COALESCE(b.amount, 0) as amount_usd
				FROM bets b
				WHERE (COALESCE(b.timestamp, NOW()) >= $1 AND COALESCE(b.timestamp, NOW()) <= $2)
					AND b.amount > 0
				UNION ALL
				-- GrooveTech transactions: only wager type with negative amount (bets)
				SELECT COALESCE(ABS(gt.amount), 0) as amount_usd
				FROM groove_transactions gt
				WHERE gt.type = 'wager' 
					AND gt.amount < 0 
					AND gt.created_at >= $1 
					AND gt.created_at <= $2
					AND gt.status = 'completed'
				UNION ALL
				-- Sport bets: only include actual bet amounts
				SELECT COALESCE(sb.bet_amount, 0) as amount_usd
				FROM sport_bets sb
				WHERE sb.created_at >= $1 
					AND sb.created_at <= $2
					AND sb.bet_amount > 0
					AND COALESCE(sb.status, 'completed') = 'completed'
				UNION ALL
				-- Plinko bets: handle NULL timestamps
				SELECT COALESCE(p.bet_amount, 0) as amount_usd
				FROM plinko p
				WHERE (COALESCE(p.timestamp, p.created_at, NOW()) >= $1 AND COALESCE(p.timestamp, p.created_at, NOW()) <= $2)
					AND p.bet_amount > 0
			) all_bets
		`

		var totalBetAmount float64
		err := s.db.GetPool().QueryRow(ctx, query, configTimeWindowStart, configTimeWindowEnd).Scan(&totalBetAmount)
		if err != nil {
			s.log.Error("Failed to calculate bet amount total", zap.Error(err), zap.String("alert_id", config.ID.String()))
			continue
		}

		s.log.Debug("Bet amount alert check",
			zap.String("alert_id", config.ID.String()),
			zap.String("alert_type", string(config.AlertType)),
			zap.Float64("total_bet_amount", totalBetAmount),
			zap.Float64("threshold", config.ThresholdAmount),
			zap.Time("time_window_start", configTimeWindowStart),
			zap.Time("time_window_end", configTimeWindowEnd))

		// Check if threshold is exceeded
		shouldTrigger := false
		if config.AlertType == dto.AlertTypeBetsAmountMore {
			// Trigger when total is greater than or equal to threshold
			shouldTrigger = totalBetAmount >= config.ThresholdAmount
		} else if config.AlertType == dto.AlertTypeBetsAmountLess {
			// Trigger when total is strictly less than threshold (not equal)
			// This ensures we only alert when bets are actually below the threshold
			shouldTrigger = totalBetAmount < config.ThresholdAmount
		}

		if shouldTrigger {
			// Check if we already triggered this alert recently (avoid duplicates)
			recentTriggerQuery := `
				SELECT COUNT(*) FROM alert_triggers
				WHERE alert_configuration_id = $1
					AND triggered_at >= $2
			`
			var recentCount int64
			err := s.db.GetPool().QueryRow(ctx, recentTriggerQuery, config.ID, configTimeWindowStart).Scan(&recentCount)
			if err == nil && recentCount == 0 {
				// Create trigger
				trigger := &dto.AlertTrigger{
					AlertConfigurationID: config.ID,
					TriggeredAt:          time.Now(),
					TriggerValue:         totalBetAmount,
					ThresholdValue:       config.ThresholdAmount,
					AmountUSD:            &totalBetAmount,
				}

				err = s.CreateAlertTrigger(ctx, trigger)
				if err != nil {
					s.log.Error("Failed to create bet amount alert trigger", zap.Error(err), zap.String("alert_id", config.ID.String()))
				} else {
					s.log.Info("Bet amount alert triggered",
						zap.String("alert_id", config.ID.String()),
						zap.String("alert_type", string(config.AlertType)),
						zap.Float64("total_amount", totalBetAmount),
						zap.Float64("threshold", config.ThresholdAmount),
						zap.String("trigger_id", trigger.ID.String()))
					// Note: user_id is null for aggregate alerts (total bet amount across all users)
					// Email sending will be handled by the module layer's sendEmailsForNewTriggers
				}
			}
		}
	}

	return nil
}

func (s *alertStorage) CheckDepositAlerts(ctx context.Context, timeWindow time.Duration, skipDuplicateCheck bool) error {
	// Get all active deposit alert configurations
	req := &dto.GetAlertConfigurationsRequest{
		Page:    1,
		PerPage: 100,
		Status:  func() *dto.AlertStatus { status := dto.AlertStatusActive; return &status }(),
	}

	configs, _, err := s.GetAlertConfigurations(ctx, req)
	if err != nil {
		s.log.Error("Failed to get alert configurations", zap.Error(err))
		return err
	}

	// Filter for deposit-related alerts
	depositAlertTypes := []dto.AlertType{
		dto.AlertTypeDepositsTotalMore,
		dto.AlertTypeDepositsTotalLess,
		dto.AlertTypeDepositsTypeMore,
		dto.AlertTypeDepositsTypeLess,
	}

	depositConfigs := make([]dto.AlertConfiguration, 0)
	for _, config := range configs {
		for _, alertType := range depositAlertTypes {
			if config.AlertType == alertType {
				depositConfigs = append(depositConfigs, config)
				break
			}
		}
	}

	if len(depositConfigs) == 0 {
		return nil // No deposit alerts configured
	}

	// Check each deposit alert configuration
	for _, config := range depositConfigs {
		// Use the config's time window instead of the passed parameter
		configTimeWindowStart := time.Now().Add(-time.Duration(config.TimeWindowMinutes) * time.Minute)

		var totalAmount float64
		var query string
		var args []interface{}

		// Build query based on alert type - also get the most recent deposit's user_id and amount
		var mostRecentUserID *uuid.UUID
		var mostRecentAmount float64
		var userIDQuery string
		if config.AlertType == dto.AlertTypeDepositsTotalMore || config.AlertType == dto.AlertTypeDepositsTotalLess {
			// Total deposits (all currencies, convert to USD) - also get most recent user_id and amount
			query = `
				SELECT COALESCE(SUM(amount_cents), 0) / 100.0
				FROM manual_funds
				WHERE type = 'add_fund'
					AND created_at >= $1
					AND created_at <= $2
			`
			userIDQuery = `
				SELECT user_id, amount_cents / 100.0 as amount
				FROM manual_funds 
				WHERE type = 'add_fund' 
				  AND created_at >= $1 
				  AND created_at <= $2 
				ORDER BY created_at DESC LIMIT 1
			`
			args = []interface{}{configTimeWindowStart, time.Now()}
		} else if config.AlertType == dto.AlertTypeDepositsTypeMore || config.AlertType == dto.AlertTypeDepositsTypeLess {
			// Type-specific deposits - also get most recent user_id and amount
			if config.CurrencyCode == nil {
				continue // Skip if no currency specified
			}
			query = `
				SELECT COALESCE(SUM(amount_cents), 0) / 100.0
				FROM manual_funds
				WHERE type = 'add_fund'
					AND currency_code = $1
					AND created_at >= $2
					AND created_at <= $3
			`
			userIDQuery = `
				SELECT user_id, amount_cents / 100.0 as amount
				FROM manual_funds 
				WHERE type = 'add_fund' 
				  AND currency_code = $1
				  AND created_at >= $2 
				  AND created_at <= $3 
				ORDER BY created_at DESC LIMIT 1
			`
			args = []interface{}{string(*config.CurrencyCode), configTimeWindowStart, time.Now()}
		} else {
			continue
		}

		err := s.db.GetPool().QueryRow(ctx, query, args...).Scan(&totalAmount)
		if err != nil {
			s.log.Error("Failed to calculate deposit total", zap.Error(err), zap.String("alert_id", config.ID.String()))
			continue
		}

		// Get the most recent deposit's user_id and amount
		var recentUserIDStr sql.NullString
		var recentAmount sql.NullFloat64
		err = s.db.GetPool().QueryRow(ctx, userIDQuery, args...).Scan(&recentUserIDStr, &recentAmount)
		if err == nil {
			if recentUserIDStr.Valid && recentUserIDStr.String != "" {
				if uid, parseErr := uuid.Parse(recentUserIDStr.String); parseErr == nil {
					mostRecentUserID = &uid
					s.log.Debug("Found most recent deposit user_id",
						zap.String("user_id", mostRecentUserID.String()),
						zap.String("alert_id", config.ID.String()))
				}
			}
			if recentAmount.Valid {
				mostRecentAmount = recentAmount.Float64
				s.log.Debug("Found most recent deposit amount",
					zap.Float64("amount", mostRecentAmount),
					zap.String("alert_id", config.ID.String()))
			}
		}

		s.log.Debug("Checking deposit alert",
			zap.String("alert_id", config.ID.String()),
			zap.String("alert_type", string(config.AlertType)),
			zap.Float64("total_amount", totalAmount),
			zap.Float64("threshold", config.ThresholdAmount),
			zap.Int("time_window_minutes", config.TimeWindowMinutes))

		// Check if threshold is exceeded
		shouldTrigger := false
		if config.AlertType == dto.AlertTypeDepositsTotalMore || config.AlertType == dto.AlertTypeDepositsTypeMore {
			shouldTrigger = totalAmount >= config.ThresholdAmount
		} else if config.AlertType == dto.AlertTypeDepositsTotalLess || config.AlertType == dto.AlertTypeDepositsTypeLess {
			shouldTrigger = totalAmount <= config.ThresholdAmount
		}

		if shouldTrigger {
			shouldCreateTrigger := true

			// Check for duplicates only if skipDuplicateCheck is false
			if !skipDuplicateCheck {
				// Check if we already triggered this alert recently (avoid duplicates within last 5 minutes)
				// Use a shorter window (5 minutes) to prevent spam while still allowing new triggers
				// when conditions are met again after the cooldown period
				recentWindowStart := time.Now().Add(-5 * time.Minute)
				recentTriggerQuery := `
					SELECT COUNT(*) FROM alert_triggers
					WHERE alert_configuration_id = $1
						AND triggered_at >= $2
				`
				var recentCount int64
				err := s.db.GetPool().QueryRow(ctx, recentTriggerQuery, config.ID, recentWindowStart).Scan(&recentCount)
				if err != nil {
					s.log.Error("Failed to check for recent deposit triggers", zap.Error(err), zap.String("alert_id", config.ID.String()))
					// Continue anyway - don't block trigger creation due to check failure
				} else if recentCount > 0 {
					shouldCreateTrigger = false
					s.log.Debug("Skipping duplicate deposit trigger",
						zap.String("alert_id", config.ID.String()),
						zap.Int64("recent_count", recentCount),
						zap.Float64("total_amount", totalAmount),
						zap.Float64("threshold", config.ThresholdAmount))
				}
			} else {
				s.log.Debug("Skipping duplicate check for deposit alert (manual fund addition)",
					zap.String("alert_id", config.ID.String()))
			}

			if shouldCreateTrigger {
				// Create trigger - use the most recent deposit amount as trigger value (not the total)
				// This shows the specific deposit that triggered the alert
				triggerValue := totalAmount
				amountUSD := &totalAmount
				if mostRecentAmount > 0 {
					// Use the most recent deposit amount as the trigger value
					triggerValue = mostRecentAmount
					amountUSD = &mostRecentAmount
				}

				trigger := &dto.AlertTrigger{
					AlertConfigurationID: config.ID,
					TriggeredAt:          time.Now(),
					TriggerValue:         triggerValue, // Use the individual deposit amount, not total
					ThresholdValue:       config.ThresholdAmount,
					AmountUSD:            amountUSD,        // Use the individual deposit amount
					UserID:               mostRecentUserID, // Include user_id from most recent deposit
				}
				if config.CurrencyCode != nil {
					trigger.CurrencyCode = config.CurrencyCode
				}

				err = s.CreateAlertTrigger(ctx, trigger)
				if err != nil {
					s.log.Error("Failed to create deposit alert trigger", zap.Error(err), zap.String("alert_id", config.ID.String()))
				} else {
					logFields := []zap.Field{
						zap.String("alert_id", config.ID.String()),
						zap.String("alert_type", string(config.AlertType)),
						zap.Float64("total_amount", totalAmount),
						zap.Float64("threshold", config.ThresholdAmount),
						zap.String("trigger_id", trigger.ID.String()),
					}
					if mostRecentUserID != nil {
						logFields = append(logFields, zap.String("user_id", mostRecentUserID.String()))
					}
					s.log.Info("Deposit alert triggered", logFields...)
					// Email sending will be handled by the module layer's sendEmailsForNewTriggers
				}
			}
		} else {
			s.log.Debug("Deposit alert condition not met",
				zap.String("alert_id", config.ID.String()),
				zap.String("alert_type", string(config.AlertType)),
				zap.Float64("total_amount", totalAmount),
				zap.Float64("threshold", config.ThresholdAmount))
		}
	}

	return nil
}

func (s *alertStorage) CheckWithdrawalAlerts(ctx context.Context, timeWindow time.Duration) error {
	// Get all active withdrawal alert configurations
	req := &dto.GetAlertConfigurationsRequest{
		Page:    1,
		PerPage: 100,
		Status:  func() *dto.AlertStatus { status := dto.AlertStatusActive; return &status }(),
	}

	configs, _, err := s.GetAlertConfigurations(ctx, req)
	if err != nil {
		s.log.Error("Failed to get alert configurations", zap.Error(err))
		return err
	}

	// Filter for withdrawal-related alerts
	withdrawalAlertTypes := []dto.AlertType{
		dto.AlertTypeWithdrawalsTotalMore,
		dto.AlertTypeWithdrawalsTotalLess,
		dto.AlertTypeWithdrawalsTypeMore,
		dto.AlertTypeWithdrawalsTypeLess,
	}

	withdrawalConfigs := make([]dto.AlertConfiguration, 0)
	for _, config := range configs {
		for _, alertType := range withdrawalAlertTypes {
			if config.AlertType == alertType {
				withdrawalConfigs = append(withdrawalConfigs, config)
				break
			}
		}
	}

	if len(withdrawalConfigs) == 0 {
		return nil // No withdrawal alerts configured
	}

	// Check each withdrawal alert configuration
	for _, config := range withdrawalConfigs {
		configTimeWindowStart := time.Now().Add(-time.Duration(config.TimeWindowMinutes) * time.Minute)
		var totalAmount float64
		var query string
		var args []interface{}

		// Build query based on alert type
		if config.AlertType == dto.AlertTypeWithdrawalsTotalMore || config.AlertType == dto.AlertTypeWithdrawalsTotalLess {
			// Total withdrawals (all currencies, convert to USD)
			query = `
				SELECT COALESCE(SUM(usd_amount_cents), 0) / 100.0
				FROM withdrawals
				WHERE status IN ('completed', 'processing')
					AND created_at >= $1
					AND created_at <= $2
			`
			args = []interface{}{configTimeWindowStart, time.Now()}
		} else if config.AlertType == dto.AlertTypeWithdrawalsTypeMore || config.AlertType == dto.AlertTypeWithdrawalsTypeLess {
			// Type-specific withdrawals
			if config.CurrencyCode == nil {
				continue // Skip if no currency specified
			}
			query = `
				SELECT COALESCE(SUM(usd_amount_cents), 0) / 100.0
				FROM withdrawals
				WHERE status IN ('completed', 'processing')
					AND currency_code = $1
					AND created_at >= $2
					AND created_at <= $3
			`
			args = []interface{}{string(*config.CurrencyCode), configTimeWindowStart, time.Now()}
		} else {
			continue
		}

		err := s.db.GetPool().QueryRow(ctx, query, args...).Scan(&totalAmount)
		if err != nil {
			s.log.Error("Failed to calculate withdrawal total", zap.Error(err), zap.String("alert_id", config.ID.String()))
			continue
		}

		// Check if threshold is exceeded
		shouldTrigger := false
		if config.AlertType == dto.AlertTypeWithdrawalsTotalMore || config.AlertType == dto.AlertTypeWithdrawalsTypeMore {
			shouldTrigger = totalAmount >= config.ThresholdAmount
		} else if config.AlertType == dto.AlertTypeWithdrawalsTotalLess || config.AlertType == dto.AlertTypeWithdrawalsTypeLess {
			shouldTrigger = totalAmount <= config.ThresholdAmount
		}

		if shouldTrigger {
			// Check if we already triggered this alert recently (avoid duplicates)
			recentTriggerQuery := `
				SELECT COUNT(*) FROM alert_triggers
				WHERE alert_configuration_id = $1
					AND triggered_at >= $2
			`
			var recentCount int64
			err := s.db.GetPool().QueryRow(ctx, recentTriggerQuery, config.ID, configTimeWindowStart).Scan(&recentCount)
			if err == nil && recentCount == 0 {
				// Create trigger
				trigger := &dto.AlertTrigger{
					AlertConfigurationID: config.ID,
					TriggeredAt:          time.Now(),
					TriggerValue:         totalAmount,
					ThresholdValue:       config.ThresholdAmount,
					AmountUSD:            &totalAmount,
				}
				if config.CurrencyCode != nil {
					trigger.CurrencyCode = config.CurrencyCode
				}

				err = s.CreateAlertTrigger(ctx, trigger)
				if err != nil {
					s.log.Error("Failed to create withdrawal alert trigger", zap.Error(err), zap.String("alert_id", config.ID.String()))
				} else {
					s.log.Info("Withdrawal alert triggered",
						zap.String("alert_id", config.ID.String()),
						zap.String("alert_type", string(config.AlertType)),
						zap.Float64("total_amount", totalAmount),
						zap.Float64("threshold", config.ThresholdAmount),
						zap.String("trigger_id", trigger.ID.String()))
					// Note: user_id is null for aggregate alerts (total withdrawals across all users)
					// Email sending will be handled by the module layer's sendEmailsForNewTriggers
				}
			}
		}
	}

	return nil
}

func (s *alertStorage) CheckGGRAlerts(ctx context.Context, timeWindow time.Duration) error {
	// Get all active GGR alert configurations
	req := &dto.GetAlertConfigurationsRequest{
		Page:    1,
		PerPage: 100,
		Status:  func() *dto.AlertStatus { status := dto.AlertStatusActive; return &status }(),
	}

	configs, _, err := s.GetAlertConfigurations(ctx, req)
	if err != nil {
		s.log.Error("Failed to get alert configurations", zap.Error(err))
		return err
	}

	// Filter for GGR-related alerts
	ggrAlertTypes := []dto.AlertType{
		dto.AlertTypeGGRTotalMore,
		dto.AlertTypeGGRTotalLess,
		dto.AlertTypeGGRSingleMore,
	}

	ggrConfigs := make([]dto.AlertConfiguration, 0)
	for _, config := range configs {
		for _, alertType := range ggrAlertTypes {
			if config.AlertType == alertType {
				ggrConfigs = append(ggrConfigs, config)
				break
			}
		}
	}

	if len(ggrConfigs) == 0 {
		return nil // No GGR alerts configured
	}

	// Check each GGR alert configuration
	for _, config := range ggrConfigs {
		configTimeWindowStart := time.Now().Add(-time.Duration(config.TimeWindowMinutes) * time.Minute)

		if config.AlertType == dto.AlertTypeGGRSingleMore {
			// Check for single transaction GGR (bet - win) exceeding threshold
			// This checks individual transactions, not aggregates
			query := `
				SELECT 
					gt.account_id,
					gt.transaction_id,
					ABS(gt.amount) as bet_amount,
					COALESCE(gt2.amount, 0) as win_amount,
					(ABS(gt.amount) - COALESCE(gt2.amount, 0)) as ggr
				FROM groove_transactions gt
				LEFT JOIN groove_transactions gt2 ON gt2.account_id = gt.account_id 
					AND gt2.round_id = gt.round_id 
					AND gt2.type = 'result'
					AND gt2.created_at > gt.created_at
					AND gt2.created_at <= gt.created_at + INTERVAL '5 minutes'
				WHERE gt.type = 'wager'
					AND gt.amount < 0
					AND gt.created_at >= $1
					AND gt.created_at <= $2
					AND (ABS(gt.amount) - COALESCE(gt2.amount, 0)) >= $3
				ORDER BY gt.created_at DESC
				LIMIT 1
			`

			var accountID, transactionID sql.NullString
			var betAmount, winAmount, ggr float64
			err := s.db.GetPool().QueryRow(ctx, query, configTimeWindowStart, time.Now(), config.ThresholdAmount).Scan(
				&accountID, &transactionID, &betAmount, &winAmount, &ggr)
			if err == nil && ggr >= config.ThresholdAmount {
				// Check if we already triggered this alert recently
				recentTriggerQuery := `
					SELECT COUNT(*) FROM alert_triggers
					WHERE alert_configuration_id = $1
						AND triggered_at >= $2
						AND transaction_id = $3
				`
				var recentCount int64
				txID := transactionID.String
				err := s.db.GetPool().QueryRow(ctx, recentTriggerQuery, config.ID, configTimeWindowStart, txID).Scan(&recentCount)
				if err == nil && recentCount == 0 {
					trigger := &dto.AlertTrigger{
						AlertConfigurationID: config.ID,
						TriggeredAt:          time.Now(),
						TriggerValue:         ggr,
						ThresholdValue:       config.ThresholdAmount,
						AmountUSD:            &ggr,
					}
					if transactionID.Valid {
						trigger.TransactionID = &transactionID.String
					}

					err = s.CreateAlertTrigger(ctx, trigger)
					if err != nil {
						s.log.Error("Failed to create GGR single alert trigger", zap.Error(err), zap.String("alert_id", config.ID.String()))
					} else {
						s.log.Info("GGR single alert triggered",
							zap.String("alert_id", config.ID.String()),
							zap.Float64("ggr", ggr),
							zap.Float64("threshold", config.ThresholdAmount),
							zap.String("trigger_id", trigger.ID.String()))
						// Note: user_id may be set for single-user GGR alerts, null for aggregate
						// Email sending will be handled by the module layer's sendEmailsForNewTriggers
					}
				}
			}
		} else {
			// Total GGR alerts (GGRTotalMore, GGRTotalLess)
			// GGR = Total Bets - Total Wins
			query := `
				SELECT 
					COALESCE(SUM(bet_amount), 0) - COALESCE(SUM(win_amount), 0) as total_ggr
				FROM (
					SELECT 
						ABS(gt.amount) as bet_amount,
						0 as win_amount,
						gt.created_at
					FROM groove_transactions gt
					WHERE gt.type = 'wager' AND gt.amount < 0
						AND gt.created_at >= $1 AND gt.created_at <= $2
					UNION ALL
					SELECT 
						0 as bet_amount,
						gt.amount as win_amount,
						gt.created_at
					FROM groove_transactions gt
					WHERE gt.type = 'result' AND gt.amount > 0
						AND gt.created_at >= $1 AND gt.created_at <= $2
					UNION ALL
					SELECT 
						b.amount as bet_amount,
						COALESCE(b.payout, 0) as win_amount,
						COALESCE(b.timestamp, NOW()) as created_at
					FROM bets b
					WHERE COALESCE(b.timestamp, NOW()) >= $1 AND COALESCE(b.timestamp, NOW()) <= $2
					UNION ALL
					SELECT 
						sb.bet_amount as bet_amount,
						COALESCE(sb.actual_win, 0) as win_amount,
						sb.created_at
					FROM sport_bets sb
					WHERE sb.created_at >= $1 AND sb.created_at <= $2
					UNION ALL
					SELECT 
						p.bet_amount as bet_amount,
						COALESCE(p.win_amount, 0) as win_amount,
						p.timestamp as created_at
					FROM plinko p
					WHERE p.timestamp >= $1 AND p.timestamp <= $2
				) all_transactions
			`

			var totalGGR float64
			err := s.db.GetPool().QueryRow(ctx, query, configTimeWindowStart, time.Now()).Scan(&totalGGR)
			if err != nil {
				s.log.Error("Failed to calculate GGR total", zap.Error(err), zap.String("alert_id", config.ID.String()))
				continue
			}

			// Check if threshold is exceeded
			shouldTrigger := false
			if config.AlertType == dto.AlertTypeGGRTotalMore {
				shouldTrigger = totalGGR >= config.ThresholdAmount
			} else if config.AlertType == dto.AlertTypeGGRTotalLess {
				shouldTrigger = totalGGR <= config.ThresholdAmount
			}

			if shouldTrigger {
				// Check if we already triggered this alert recently (avoid duplicates)
				recentTriggerQuery := `
					SELECT COUNT(*) FROM alert_triggers
					WHERE alert_configuration_id = $1
						AND triggered_at >= $2
				`
				var recentCount int64
				err := s.db.GetPool().QueryRow(ctx, recentTriggerQuery, config.ID, configTimeWindowStart).Scan(&recentCount)
				if err == nil && recentCount == 0 {
					// Create trigger
					trigger := &dto.AlertTrigger{
						AlertConfigurationID: config.ID,
						TriggeredAt:          time.Now(),
						TriggerValue:         totalGGR,
						ThresholdValue:       config.ThresholdAmount,
						AmountUSD:            &totalGGR,
					}

					err = s.CreateAlertTrigger(ctx, trigger)
					if err != nil {
						s.log.Error("Failed to create GGR alert trigger", zap.Error(err), zap.String("alert_id", config.ID.String()))
					} else {
						s.log.Info("GGR alert triggered",
							zap.String("alert_id", config.ID.String()),
							zap.String("alert_type", string(config.AlertType)),
							zap.Float64("total_ggr", totalGGR),
							zap.Float64("threshold", config.ThresholdAmount),
							zap.String("trigger_id", trigger.ID.String()))
						// Note: user_id is null for aggregate GGR alerts (total GGR across all users)
						// Email sending will be handled by the module layer's sendEmailsForNewTriggers
					}
				}
			}
		}
	}

	return nil
}

func (s *alertStorage) CheckMultipleAccountsSameIP(ctx context.Context, timeWindow time.Duration) error {
	// Get all active multiple_accounts_same_ip alert configurations
	req := &dto.GetAlertConfigurationsRequest{
		Page:    1,
		PerPage: 100,
		Status:  func() *dto.AlertStatus { status := dto.AlertStatusActive; return &status }(),
	}

	configs, _, err := s.GetAlertConfigurations(ctx, req)
	if err != nil {
		s.log.Error("Failed to get alert configurations", zap.Error(err))
		return err
	}

	// Filter for multiple_accounts_same_ip alerts
	ipAlertConfigs := make([]dto.AlertConfiguration, 0)
	for _, config := range configs {
		if config.AlertType == dto.AlertTypeMultipleAccountsSameIP {
			ipAlertConfigs = append(ipAlertConfigs, config)
		}
	}

	if len(ipAlertConfigs) == 0 {
		return nil // No multiple accounts same IP alerts configured
	}

	// Check each alert configuration
	for _, config := range ipAlertConfigs {
		configTimeWindowStart := time.Now().Add(-time.Duration(config.TimeWindowMinutes) * time.Minute)

		// Find IP addresses that have more than threshold number of accounts created within the time window
		// We use the first session IP for each user as a proxy for registration IP
		query := `
			WITH first_sessions AS (
				SELECT DISTINCT ON (us.user_id)
					us.user_id,
					us.ip_address,
					u.created_at as user_created_at
				FROM user_sessions us
				INNER JOIN users u ON u.id = us.user_id
				WHERE us.ip_address IS NOT NULL 
					AND us.ip_address != ''
					AND u.created_at >= $1
					AND u.created_at <= $2
				ORDER BY us.user_id, us.created_at ASC
			),
			ip_account_counts AS (
				SELECT 
					ip_address,
					COUNT(DISTINCT user_id) as account_count,
					ARRAY_AGG(DISTINCT user_id) as user_ids,
					MIN(user_created_at) as first_account_created,
					MAX(user_created_at) as last_account_created
				FROM first_sessions
				GROUP BY ip_address
				HAVING COUNT(DISTINCT user_id) >= $3
			)
			SELECT 
				ip_address,
				account_count,
				user_ids,
				first_account_created,
				last_account_created
			FROM ip_account_counts
			ORDER BY account_count DESC, ip_address
			LIMIT 10
		`

		rows, err := s.db.GetPool().Query(ctx, query, configTimeWindowStart, time.Now(), int(config.ThresholdAmount))
		if err != nil {
			s.log.Error("Failed to query multiple accounts same IP", zap.Error(err), zap.String("alert_id", config.ID.String()))
			continue
		}
		defer rows.Close()

		for rows.Next() {
			var ipAddress string
			var accountCount int
			var userIDs pq.StringArray
			var firstAccountCreated, lastAccountCreated time.Time

			err := rows.Scan(&ipAddress, &accountCount, &userIDs, &firstAccountCreated, &lastAccountCreated)
			if err != nil {
				s.log.Error("Failed to scan multiple accounts same IP result", zap.Error(err))
				continue
			}

			// Check if we already triggered this alert recently for this IP (avoid duplicates)
			recentTriggerQuery := `
				SELECT COUNT(*) FROM alert_triggers
				WHERE alert_configuration_id = $1
					AND triggered_at >= $2
					AND context_data LIKE $3
			`
			var recentCount int64
			ipPattern := fmt.Sprintf(`%%"ip_address":"%s"%%`, ipAddress)
			err = s.db.GetPool().QueryRow(ctx, recentTriggerQuery, config.ID, configTimeWindowStart, ipPattern).Scan(&recentCount)
			if err != nil {
				s.log.Error("Failed to check for recent triggers", zap.Error(err))
				continue
			}

			if recentCount == 0 {
				// Create context data with IP and user IDs
				contextData := fmt.Sprintf(`{"ip_address":"%s","account_count":%d,"user_ids":[`, ipAddress, accountCount)
				for i, userID := range userIDs {
					if i > 0 {
						contextData += ","
					}
					contextData += fmt.Sprintf(`"%s"`, userID)
				}
				contextData += fmt.Sprintf(`],"first_account_created":"%s","last_account_created":"%s"}`,
					firstAccountCreated.Format(time.RFC3339), lastAccountCreated.Format(time.RFC3339))

				// Create trigger
				trigger := &dto.AlertTrigger{
					AlertConfigurationID: config.ID,
					TriggeredAt:          time.Now(),
					TriggerValue:         float64(accountCount),
					ThresholdValue:       config.ThresholdAmount,
					ContextData:          &contextData,
				}

				err = s.CreateAlertTrigger(ctx, trigger)
				if err != nil {
					s.log.Error("Failed to create multiple accounts same IP alert trigger", zap.Error(err), zap.String("alert_id", config.ID.String()))
				} else {
					s.log.Info("Multiple accounts same IP alert triggered",
						zap.String("alert_id", config.ID.String()),
						zap.String("ip_address", ipAddress),
						zap.Int("account_count", accountCount),
						zap.Float64("threshold", config.ThresholdAmount),
						zap.String("trigger_id", trigger.ID.String()))
					// Email sending will be handled by the module layer's sendEmailsForNewTriggers
				}
			}
		}
	}

	return nil
}
