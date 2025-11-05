package alert

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
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
	CheckDepositAlerts(ctx context.Context, timeWindow time.Duration) error
	CheckWithdrawalAlerts(ctx context.Context, timeWindow time.Duration) error
	CheckGGRAlerts(ctx context.Context, timeWindow time.Duration) error
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
	query := `
        INSERT INTO alert_configurations (
            name, description, alert_type, threshold_amount, time_window_minutes,
            currency_code, email_notifications, webhook_url, email_group_ids, created_by
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
        ) RETURNING 
            id, name, description, alert_type, status, threshold_amount, time_window_minutes,
            currency_code, email_notifications, webhook_url, email_group_ids,
            created_by, created_at, updated_at, updated_by
    `

	var emailGroupIDs interface{} = pq.Array([]uuid.UUID{})
	if len(req.EmailGroupIDs) > 0 {
		emailGroupIDs = pq.Array(req.EmailGroupIDs)
	}

	var config dto.AlertConfiguration
	var scannedEmailGroupIDs UUIDArray
	err := s.db.GetPool().QueryRow(ctx, query,
		req.Name, req.Description, req.AlertType, req.ThresholdAmount,
		req.TimeWindowMinutes, req.CurrencyCode, req.EmailNotifications,
		req.WebhookURL, emailGroupIDs, createdBy,
	).Scan(
		&config.ID, &config.Name, &config.Description, &config.AlertType,
		&config.Status, &config.ThresholdAmount, &config.TimeWindowMinutes,
		&config.CurrencyCode, &config.EmailNotifications, &config.WebhookURL,
		&scannedEmailGroupIDs, &config.CreatedBy, &config.CreatedAt, &config.UpdatedAt, &config.UpdatedBy,
	)
	if err == nil {
		config.EmailGroupIDs = []uuid.UUID(scannedEmailGroupIDs)
	}

	if err != nil {
		s.log.Error("Failed to create alert configuration", zap.Error(err))
		return nil, err
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
            t.user_id, t.transaction_id, t.amount_usd, t.currency_code, t.context_data,
            t.acknowledged, t.acknowledged_by, t.acknowledged_at, t.created_at,
            ac.id, ac.name, ac.description, ac.alert_type, ac.status, ac.threshold_amount,
            ac.time_window_minutes, ac.currency_code, ac.email_notifications, ac.webhook_url,
            ac.created_by, ac.created_at, ac.updated_at, ac.updated_by,
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
		if err := rows.Scan(
			&t.ID, &t.AlertConfigurationID, &t.TriggeredAt, &t.TriggerValue, &t.ThresholdValue,
			&t.UserID, &t.TransactionID, &t.AmountUSD, &t.CurrencyCode, &t.ContextData,
			&t.Acknowledged, &t.AcknowledgedBy, &t.AcknowledgedAt, &t.CreatedAt,
			&ac.ID, &ac.Name, &ac.Description, &ac.AlertType, &ac.Status, &ac.ThresholdAmount,
			&ac.TimeWindowMinutes, &ac.CurrencyCode, &ac.EmailNotifications, &ac.WebhookURL,
			&ac.CreatedBy, &ac.CreatedAt, &ac.UpdatedAt, &ac.UpdatedBy,
			&username, &email,
		); err != nil {
			s.log.Error("Failed to scan alert trigger", zap.Error(err))
			return nil, 0, err
		}
		// attach joined fields
		t.AlertConfiguration = &ac
		if username.Valid {
			t.Username = &username.String
		}
		if email.Valid {
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
	// Implementation for checking bet count alerts
	return nil
}

func (s *alertStorage) CheckBetAmountAlerts(ctx context.Context, timeWindow time.Duration) error {
	// Implementation for checking bet amount alerts
	return nil
}

func (s *alertStorage) CheckDepositAlerts(ctx context.Context, timeWindow time.Duration) error {
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
			if config.AlertType == alertType && config.EmailNotifications {
				depositConfigs = append(depositConfigs, config)
				break
			}
		}
	}

	if len(depositConfigs) == 0 {
		return nil // No deposit alerts configured
	}

	// Calculate time window start
	timeWindowStart := time.Now().Add(-timeWindow)

	// Check each deposit alert configuration
	for _, config := range depositConfigs {
		var totalAmount float64
		var query string
		var args []interface{}

		// Build query based on alert type
		if config.AlertType == dto.AlertTypeDepositsTotalMore || config.AlertType == dto.AlertTypeDepositsTotalLess {
			// Total deposits (all currencies, convert to USD)
			query = `
				SELECT COALESCE(SUM(amount_cents), 0) / 100.0
				FROM manual_funds
				WHERE type = 'add_fund'
					AND created_at >= $1
					AND created_at <= $2
			`
			args = []interface{}{timeWindowStart, time.Now()}
		} else if config.AlertType == dto.AlertTypeDepositsTypeMore || config.AlertType == dto.AlertTypeDepositsTypeLess {
			// Type-specific deposits
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
			args = []interface{}{string(*config.CurrencyCode), timeWindowStart, time.Now()}
		} else {
			continue
		}

		err := s.db.GetPool().QueryRow(ctx, query, args...).Scan(&totalAmount)
		if err != nil {
			s.log.Error("Failed to calculate deposit total", zap.Error(err), zap.String("alert_id", config.ID.String()))
			continue
		}

		// Check if threshold is exceeded
		shouldTrigger := false
		if config.AlertType == dto.AlertTypeDepositsTotalMore || config.AlertType == dto.AlertTypeDepositsTypeMore {
			shouldTrigger = totalAmount >= config.ThresholdAmount
		} else if config.AlertType == dto.AlertTypeDepositsTotalLess || config.AlertType == dto.AlertTypeDepositsTypeLess {
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
			err := s.db.GetPool().QueryRow(ctx, recentTriggerQuery, config.ID, timeWindowStart).Scan(&recentCount)
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
					s.log.Error("Failed to create deposit alert trigger", zap.Error(err), zap.String("alert_id", config.ID.String()))
				} else {
					s.log.Info("Deposit alert triggered",
						zap.String("alert_id", config.ID.String()),
						zap.String("alert_type", string(config.AlertType)),
						zap.Float64("total_amount", totalAmount),
						zap.Float64("threshold", config.ThresholdAmount))
					// Note: Email sending will be handled by a background job or webhook
					// For now, triggers are created and can be checked via the API
				}
			}
		}
	}

	return nil
}

func (s *alertStorage) CheckWithdrawalAlerts(ctx context.Context, timeWindow time.Duration) error {
	// Implementation for checking withdrawal alerts
	return nil
}

func (s *alertStorage) CheckGGRAlerts(ctx context.Context, timeWindow time.Duration) error {
	// Implementation for checking GGR alerts
	return nil
}
