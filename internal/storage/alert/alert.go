package alert

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/persistencedb"
	"go.uber.org/zap"
)

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
            currency_code, email_notifications, webhook_url, created_by
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9
        ) RETURNING *
    `

	var config dto.AlertConfiguration
	err := s.db.GetPool().QueryRow(ctx, query,
		req.Name, req.Description, req.AlertType, req.ThresholdAmount,
		req.TimeWindowMinutes, req.CurrencyCode, req.EmailNotifications,
		req.WebhookURL, createdBy,
	).Scan(
		&config.ID, &config.Name, &config.Description, &config.AlertType,
		&config.Status, &config.ThresholdAmount, &config.TimeWindowMinutes,
		&config.CurrencyCode, &config.EmailNotifications, &config.WebhookURL,
		&config.CreatedBy, &config.CreatedAt, &config.UpdatedAt, &config.UpdatedBy,
	)

	if err != nil {
		s.log.Error("Failed to create alert configuration", zap.Error(err))
		return nil, err
	}

	return &config, nil
}

// GetAlertConfigurationByID gets an alert configuration by ID
func (s *alertStorage) GetAlertConfigurationByID(ctx context.Context, id uuid.UUID) (*dto.AlertConfiguration, error) {
	query := `SELECT * FROM alert_configurations WHERE id = $1`

	var config dto.AlertConfiguration
	err := s.db.GetPool().QueryRow(ctx, query, id).Scan(
		&config.ID, &config.Name, &config.Description, &config.AlertType,
		&config.Status, &config.ThresholdAmount, &config.TimeWindowMinutes,
		&config.CurrencyCode, &config.EmailNotifications, &config.WebhookURL,
		&config.CreatedBy, &config.CreatedAt, &config.UpdatedAt, &config.UpdatedBy,
	)

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
		SELECT * FROM alert_configurations 
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
		err := rows.Scan(
			&config.ID, &config.Name, &config.Description, &config.AlertType,
			&config.Status, &config.ThresholdAmount, &config.TimeWindowMinutes,
			&config.CurrencyCode, &config.EmailNotifications, &config.WebhookURL,
			&config.CreatedBy, &config.CreatedAt, &config.UpdatedAt, &config.UpdatedBy,
		)
		if err != nil {
			s.log.Error("Failed to scan alert configuration", zap.Error(err))
			return nil, 0, err
		}
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

	// Always update updated_at and updated_by
	setClauses = append(setClauses, fmt.Sprintf("updated_at = NOW()"))
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
        RETURNING *
    `, strings.Join(setClauses, ", "), argIdx)

	args = append(args, id)

	var config dto.AlertConfiguration
	err := s.db.GetPool().QueryRow(ctx, query, args...).Scan(
		&config.ID, &config.Name, &config.Description, &config.AlertType,
		&config.Status, &config.ThresholdAmount, &config.TimeWindowMinutes,
		&config.CurrencyCode, &config.EmailNotifications, &config.WebhookURL,
		&config.CreatedBy, &config.CreatedAt, &config.UpdatedAt, &config.UpdatedBy,
	)
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
	// Implementation for checking deposit alerts
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
