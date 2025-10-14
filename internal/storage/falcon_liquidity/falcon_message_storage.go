package falcon_liquidity

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/persistencedb"
	"go.uber.org/zap"
)

// FalconMessageStorage interface for storing and retrieving Falcon Liquidity messages
type FalconMessageStorage interface {
	// Create a new Falcon message record
	CreateFalconMessage(ctx context.Context, req dto.CreateFalconMessageRequest) (*dto.FalconLiquidityMessage, error)

	// Update an existing Falcon message record
	UpdateFalconMessage(ctx context.Context, messageID string, req dto.UpdateFalconMessageRequest) error

	// Get a Falcon message by message ID
	GetFalconMessageByID(ctx context.Context, messageID string) (*dto.FalconLiquidityMessage, error)

	// Get Falcon messages by transaction ID
	GetFalconMessagesByTransactionID(ctx context.Context, transactionID string) ([]dto.FalconLiquidityMessage, error)

	// Get Falcon messages by user ID
	GetFalconMessagesByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]dto.FalconLiquidityMessage, error)

	// Query Falcon messages with filters
	QueryFalconMessages(ctx context.Context, query dto.FalconMessageQuery) ([]dto.FalconLiquidityMessage, error)

	// Get summary statistics for Falcon messages
	GetFalconMessageSummary(ctx context.Context, query dto.FalconMessageQuery) (*dto.FalconMessageSummary, error)

	// Get failed messages for retry
	GetFailedFalconMessages(ctx context.Context, maxRetries int) ([]dto.FalconLiquidityMessage, error)

	// Update reconciliation status
	UpdateReconciliationStatus(ctx context.Context, messageID string, status dto.FalconReconciliationStatus, notes string) error
}

// falconMessageStorageImpl implements FalconMessageStorage
type falconMessageStorageImpl struct {
	logger *zap.Logger
	db     *persistencedb.PersistenceDB
}

// NewFalconMessageStorage creates a new Falcon message storage implementation
func NewFalconMessageStorage(logger *zap.Logger, db *persistencedb.PersistenceDB) FalconMessageStorage {
	return &falconMessageStorageImpl{
		logger: logger,
		db:     db,
	}
}

// CreateFalconMessage creates a new Falcon message record
func (s *falconMessageStorageImpl) CreateFalconMessage(ctx context.Context, req dto.CreateFalconMessageRequest) (*dto.FalconLiquidityMessage, error) {
	s.logger.Info("Creating Falcon message record",
		zap.String("message_id", req.MessageID),
		zap.String("transaction_id", req.TransactionID),
		zap.String("user_id", req.UserID.String()),
		zap.String("message_type", string(req.MessageType)))

	query := `
		INSERT INTO falcon_liquidity_messages (
			message_id, transaction_id, user_id, message_type, message_data,
			bet_amount, payout_amount, currency, game_name, game_id, house_edge,
			falcon_routing_key, falcon_exchange, falcon_queue, status,
			created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
		) RETURNING id, created_at`

	var message dto.FalconLiquidityMessage
	var createdAt time.Time

	err := s.db.GetPool().QueryRow(ctx, query,
		req.MessageID, req.TransactionID, req.UserID, req.MessageType, req.MessageData,
		req.BetAmount, req.PayoutAmount, req.Currency, req.GameName, req.GameID, req.HouseEdge,
		req.FalconRoutingKey, req.FalconExchange, req.FalconQueue, dto.FalconMessageStatusPending,
		time.Now(),
	).Scan(&message.ID, &createdAt)

	if err != nil {
		s.logger.Error("Failed to create Falcon message record", zap.Error(err))
		return nil, err
	}

	message.MessageID = req.MessageID
	message.TransactionID = req.TransactionID
	message.UserID = req.UserID
	message.MessageType = req.MessageType
	message.MessageData = req.MessageData
	message.BetAmount = req.BetAmount
	message.PayoutAmount = req.PayoutAmount
	message.Currency = req.Currency
	message.GameName = req.GameName
	message.GameID = req.GameID
	message.HouseEdge = req.HouseEdge
	message.FalconRoutingKey = req.FalconRoutingKey
	message.FalconExchange = req.FalconExchange
	message.FalconQueue = req.FalconQueue
	message.Status = dto.FalconMessageStatusPending
	message.CreatedAt = createdAt
	message.ReconciliationStatus = dto.FalconReconciliationStatusPending

	s.logger.Info("Falcon message record created successfully",
		zap.String("message_id", req.MessageID),
		zap.String("id", message.ID.String()))

	return &message, nil
}

// UpdateFalconMessage updates an existing Falcon message record
func (s *falconMessageStorageImpl) UpdateFalconMessage(ctx context.Context, messageID string, req dto.UpdateFalconMessageRequest) error {
	s.logger.Info("Updating Falcon message record", zap.String("message_id", messageID))

	// Build dynamic update query
	updates := []string{}
	args := []interface{}{}
	argIndex := 1

	if req.Status != nil {
		updates = append(updates, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, *req.Status)
		argIndex++
	}

	if req.RetryCount != nil {
		updates = append(updates, fmt.Sprintf("retry_count = $%d", argIndex))
		args = append(args, *req.RetryCount)
		argIndex++
	}

	if req.LastRetryAt != nil {
		updates = append(updates, fmt.Sprintf("last_retry_at = $%d", argIndex))
		args = append(args, *req.LastRetryAt)
		argIndex++
	}

	if req.SentAt != nil {
		updates = append(updates, fmt.Sprintf("sent_at = $%d", argIndex))
		args = append(args, *req.SentAt)
		argIndex++
	}

	if req.AcknowledgedAt != nil {
		updates = append(updates, fmt.Sprintf("acknowledged_at = $%d", argIndex))
		args = append(args, *req.AcknowledgedAt)
		argIndex++
	}

	if req.ErrorMessage != nil {
		updates = append(updates, fmt.Sprintf("error_message = $%d", argIndex))
		args = append(args, *req.ErrorMessage)
		argIndex++
	}

	if req.ErrorCode != nil {
		updates = append(updates, fmt.Sprintf("error_code = $%d", argIndex))
		args = append(args, *req.ErrorCode)
		argIndex++
	}

	if req.FalconResponse != nil {
		updates = append(updates, fmt.Sprintf("falcon_response = $%d", argIndex))
		args = append(args, req.FalconResponse)
		argIndex++
	}

	if req.ReconciliationStatus != nil {
		updates = append(updates, fmt.Sprintf("reconciliation_status = $%d", argIndex))
		args = append(args, *req.ReconciliationStatus)
		argIndex++
	}

	if req.ReconciliationNotes != nil {
		updates = append(updates, fmt.Sprintf("reconciliation_notes = $%d", argIndex))
		args = append(args, *req.ReconciliationNotes)
		argIndex++
	}

	if len(updates) == 0 {
		return fmt.Errorf("no fields to update")
	}

	// Add message_id as the last argument
	args = append(args, messageID)

	query := fmt.Sprintf(`
		UPDATE falcon_liquidity_messages 
		SET %s 
		WHERE message_id = $%d`,
		strings.Join(updates, ", "), argIndex)

	_, err := s.db.GetPool().Exec(ctx, query, args...)
	if err != nil {
		s.logger.Error("Failed to update Falcon message record", zap.Error(err))
		return err
	}

	s.logger.Info("Falcon message record updated successfully", zap.String("message_id", messageID))
	return nil
}

// GetFalconMessageByID gets a Falcon message by message ID
func (s *falconMessageStorageImpl) GetFalconMessageByID(ctx context.Context, messageID string) (*dto.FalconLiquidityMessage, error) {
	s.logger.Info("Getting Falcon message by ID", zap.String("message_id", messageID))

	query := `
		SELECT id, message_id, transaction_id, user_id, message_type, message_data,
			   bet_amount, payout_amount, currency, game_name, game_id, house_edge,
			   falcon_routing_key, falcon_exchange, falcon_queue, status, retry_count,
			   last_retry_at, created_at, sent_at, acknowledged_at, error_message,
			   error_code, falcon_response, reconciliation_status, reconciliation_notes
		FROM falcon_liquidity_messages 
		WHERE message_id = $1`

	var message dto.FalconLiquidityMessage
	err := s.db.GetPool().QueryRow(ctx, query, messageID).Scan(
		&message.ID, &message.MessageID, &message.TransactionID, &message.UserID,
		&message.MessageType, &message.MessageData, &message.BetAmount, &message.PayoutAmount,
		&message.Currency, &message.GameName, &message.GameID, &message.HouseEdge,
		&message.FalconRoutingKey, &message.FalconExchange, &message.FalconQueue,
		&message.Status, &message.RetryCount, &message.LastRetryAt, &message.CreatedAt,
		&message.SentAt, &message.AcknowledgedAt, &message.ErrorMessage, &message.ErrorCode,
		&message.FalconResponse, &message.ReconciliationStatus, &message.ReconciliationNotes,
	)

	if err != nil {
		s.logger.Error("Failed to get Falcon message by ID", zap.Error(err))
		return nil, err
	}

	return &message, nil
}

// GetFalconMessagesByTransactionID gets Falcon messages by transaction ID
func (s *falconMessageStorageImpl) GetFalconMessagesByTransactionID(ctx context.Context, transactionID string) ([]dto.FalconLiquidityMessage, error) {
	s.logger.Info("Getting Falcon messages by transaction ID", zap.String("transaction_id", transactionID))

	query := `
		SELECT id, message_id, transaction_id, user_id, message_type, message_data,
			   bet_amount, payout_amount, currency, game_name, game_id, house_edge,
			   falcon_routing_key, falcon_exchange, falcon_queue, status, retry_count,
			   last_retry_at, created_at, sent_at, acknowledged_at, error_message,
			   error_code, falcon_response, reconciliation_status, reconciliation_notes
		FROM falcon_liquidity_messages 
		WHERE transaction_id = $1
		ORDER BY created_at DESC`

	rows, err := s.db.GetPool().Query(ctx, query, transactionID)
	if err != nil {
		s.logger.Error("Failed to query Falcon messages by transaction ID", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var messages []dto.FalconLiquidityMessage
	for rows.Next() {
		var message dto.FalconLiquidityMessage
		err := rows.Scan(
			&message.ID, &message.MessageID, &message.TransactionID, &message.UserID,
			&message.MessageType, &message.MessageData, &message.BetAmount, &message.PayoutAmount,
			&message.Currency, &message.GameName, &message.GameID, &message.HouseEdge,
			&message.FalconRoutingKey, &message.FalconExchange, &message.FalconQueue,
			&message.Status, &message.RetryCount, &message.LastRetryAt, &message.CreatedAt,
			&message.SentAt, &message.AcknowledgedAt, &message.ErrorMessage, &message.ErrorCode,
			&message.FalconResponse, &message.ReconciliationStatus, &message.ReconciliationNotes,
		)
		if err != nil {
			s.logger.Error("Failed to scan Falcon message", zap.Error(err))
			return nil, err
		}
		messages = append(messages, message)
	}

	s.logger.Info("Retrieved Falcon messages by transaction ID",
		zap.String("transaction_id", transactionID),
		zap.Int("count", len(messages)))

	return messages, nil
}

// GetFalconMessagesByUserID gets Falcon messages by user ID
func (s *falconMessageStorageImpl) GetFalconMessagesByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]dto.FalconLiquidityMessage, error) {
	s.logger.Info("Getting Falcon messages by user ID",
		zap.String("user_id", userID.String()),
		zap.Int("limit", limit),
		zap.Int("offset", offset))

	query := `
		SELECT id, message_id, transaction_id, user_id, message_type, message_data,
			   bet_amount, payout_amount, currency, game_name, game_id, house_edge,
			   falcon_routing_key, falcon_exchange, falcon_queue, status, retry_count,
			   last_retry_at, created_at, sent_at, acknowledged_at, error_message,
			   error_code, falcon_response, reconciliation_status, reconciliation_notes
		FROM falcon_liquidity_messages 
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := s.db.GetPool().Query(ctx, query, userID, limit, offset)
	if err != nil {
		s.logger.Error("Failed to query Falcon messages by user ID", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var messages []dto.FalconLiquidityMessage
	for rows.Next() {
		var message dto.FalconLiquidityMessage
		err := rows.Scan(
			&message.ID, &message.MessageID, &message.TransactionID, &message.UserID,
			&message.MessageType, &message.MessageData, &message.BetAmount, &message.PayoutAmount,
			&message.Currency, &message.GameName, &message.GameID, &message.HouseEdge,
			&message.FalconRoutingKey, &message.FalconExchange, &message.FalconQueue,
			&message.Status, &message.RetryCount, &message.LastRetryAt, &message.CreatedAt,
			&message.SentAt, &message.AcknowledgedAt, &message.ErrorMessage, &message.ErrorCode,
			&message.FalconResponse, &message.ReconciliationStatus, &message.ReconciliationNotes,
		)
		if err != nil {
			s.logger.Error("Failed to scan Falcon message", zap.Error(err))
			return nil, err
		}
		messages = append(messages, message)
	}

	s.logger.Info("Retrieved Falcon messages by user ID",
		zap.String("user_id", userID.String()),
		zap.Int("count", len(messages)))

	return messages, nil
}

// QueryFalconMessages queries Falcon messages with filters
func (s *falconMessageStorageImpl) QueryFalconMessages(ctx context.Context, query dto.FalconMessageQuery) ([]dto.FalconLiquidityMessage, error) {
	s.logger.Info("Querying Falcon messages with filters")

	// Build dynamic query
	whereClauses := []string{}
	args := []interface{}{}
	argIndex := 1

	if query.UserID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("user_id = $%d", argIndex))
		args = append(args, *query.UserID)
		argIndex++
	}

	if query.TransactionID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("transaction_id = $%d", argIndex))
		args = append(args, *query.TransactionID)
		argIndex++
	}

	if query.MessageID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("message_id = $%d", argIndex))
		args = append(args, *query.MessageID)
		argIndex++
	}

	if query.Status != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, *query.Status)
		argIndex++
	}

	if query.MessageType != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("message_type = $%d", argIndex))
		args = append(args, *query.MessageType)
		argIndex++
	}

	if query.ReconciliationStatus != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("reconciliation_status = $%d", argIndex))
		args = append(args, *query.ReconciliationStatus)
		argIndex++
	}

	if query.StartDate != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("created_at >= $%d", argIndex))
		args = append(args, *query.StartDate)
		argIndex++
	}

	if query.EndDate != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("created_at <= $%d", argIndex))
		args = append(args, *query.EndDate)
		argIndex++
	}

	whereClause := ""
	if len(whereClauses) > 0 {
		whereClause = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	// Add limit and offset
	if query.Limit <= 0 {
		query.Limit = 100 // Default limit
	}

	args = append(args, query.Limit, query.Offset)

	sqlQuery := fmt.Sprintf(`
		SELECT id, message_id, transaction_id, user_id, message_type, message_data,
			   bet_amount, payout_amount, currency, game_name, game_id, house_edge,
			   falcon_routing_key, falcon_exchange, falcon_queue, status, retry_count,
			   last_retry_at, created_at, sent_at, acknowledged_at, error_message,
			   error_code, falcon_response, reconciliation_status, reconciliation_notes
		FROM falcon_liquidity_messages 
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)

	rows, err := s.db.GetPool().Query(ctx, sqlQuery, args...)
	if err != nil {
		s.logger.Error("Failed to query Falcon messages", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var messages []dto.FalconLiquidityMessage
	for rows.Next() {
		var message dto.FalconLiquidityMessage
		err := rows.Scan(
			&message.ID, &message.MessageID, &message.TransactionID, &message.UserID,
			&message.MessageType, &message.MessageData, &message.BetAmount, &message.PayoutAmount,
			&message.Currency, &message.GameName, &message.GameID, &message.HouseEdge,
			&message.FalconRoutingKey, &message.FalconExchange, &message.FalconQueue,
			&message.Status, &message.RetryCount, &message.LastRetryAt, &message.CreatedAt,
			&message.SentAt, &message.AcknowledgedAt, &message.ErrorMessage, &message.ErrorCode,
			&message.FalconResponse, &message.ReconciliationStatus, &message.ReconciliationNotes,
		)
		if err != nil {
			s.logger.Error("Failed to scan Falcon message", zap.Error(err))
			return nil, err
		}
		messages = append(messages, message)
	}

	s.logger.Info("Retrieved Falcon messages with filters", zap.Int("count", len(messages)))
	return messages, nil
}

// GetFalconMessageSummary gets summary statistics for Falcon messages
func (s *falconMessageStorageImpl) GetFalconMessageSummary(ctx context.Context, query dto.FalconMessageQuery) (*dto.FalconMessageSummary, error) {
	s.logger.Info("Getting Falcon message summary")

	// Build dynamic query
	whereClauses := []string{}
	args := []interface{}{}
	argIndex := 1

	if query.UserID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("user_id = $%d", argIndex))
		args = append(args, *query.UserID)
		argIndex++
	}

	if query.StartDate != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("created_at >= $%d", argIndex))
		args = append(args, *query.StartDate)
		argIndex++
	}

	if query.EndDate != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("created_at <= $%d", argIndex))
		args = append(args, *query.EndDate)
		argIndex++
	}

	whereClause := ""
	if len(whereClauses) > 0 {
		whereClause = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	sqlQuery := fmt.Sprintf(`
		SELECT 
			COUNT(*) as total_messages,
			COUNT(CASE WHEN status = 'pending' THEN 1 END) as pending_messages,
			COUNT(CASE WHEN status = 'sent' THEN 1 END) as sent_messages,
			COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed_messages,
			COUNT(CASE WHEN status = 'acknowledged' THEN 1 END) as acknowledged_messages,
			COUNT(CASE WHEN reconciliation_status = 'disputed' THEN 1 END) as disputed_messages,
			COALESCE(SUM(bet_amount), 0) as total_bet_amount,
			COALESCE(SUM(payout_amount), 0) as total_payout_amount,
			COALESCE(AVG(house_edge), 0) as average_house_edge,
			MAX(created_at) as last_message_at
		FROM falcon_liquidity_messages 
		%s`, whereClause)

	var summary dto.FalconMessageSummary
	err := s.db.GetPool().QueryRow(ctx, sqlQuery, args...).Scan(
		&summary.TotalMessages, &summary.PendingMessages, &summary.SentMessages,
		&summary.FailedMessages, &summary.AcknowledgedMessages, &summary.DisputedMessages,
		&summary.TotalBetAmount, &summary.TotalPayoutAmount, &summary.AverageHouseEdge,
		&summary.LastMessageAt,
	)

	if err != nil {
		s.logger.Error("Failed to get Falcon message summary", zap.Error(err))
		return nil, err
	}

	s.logger.Info("Retrieved Falcon message summary",
		zap.Int("total_messages", summary.TotalMessages),
		zap.Int("pending_messages", summary.PendingMessages),
		zap.Int("sent_messages", summary.SentMessages),
		zap.Int("failed_messages", summary.FailedMessages))

	return &summary, nil
}

// GetFailedFalconMessages gets failed messages for retry
func (s *falconMessageStorageImpl) GetFailedFalconMessages(ctx context.Context, maxRetries int) ([]dto.FalconLiquidityMessage, error) {
	s.logger.Info("Getting failed Falcon messages for retry", zap.Int("max_retries", maxRetries))

	query := `
		SELECT id, message_id, transaction_id, user_id, message_type, message_data,
			   bet_amount, payout_amount, currency, game_name, game_id, house_edge,
			   falcon_routing_key, falcon_exchange, falcon_queue, status, retry_count,
			   last_retry_at, created_at, sent_at, acknowledged_at, error_message,
			   error_code, falcon_response, reconciliation_status, reconciliation_notes
		FROM falcon_liquidity_messages 
		WHERE status = 'failed' AND retry_count < $1
		ORDER BY created_at ASC`

	rows, err := s.db.GetPool().Query(ctx, query, maxRetries)
	if err != nil {
		s.logger.Error("Failed to query failed Falcon messages", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var messages []dto.FalconLiquidityMessage
	for rows.Next() {
		var message dto.FalconLiquidityMessage
		err := rows.Scan(
			&message.ID, &message.MessageID, &message.TransactionID, &message.UserID,
			&message.MessageType, &message.MessageData, &message.BetAmount, &message.PayoutAmount,
			&message.Currency, &message.GameName, &message.GameID, &message.HouseEdge,
			&message.FalconRoutingKey, &message.FalconExchange, &message.FalconQueue,
			&message.Status, &message.RetryCount, &message.LastRetryAt, &message.CreatedAt,
			&message.SentAt, &message.AcknowledgedAt, &message.ErrorMessage, &message.ErrorCode,
			&message.FalconResponse, &message.ReconciliationStatus, &message.ReconciliationNotes,
		)
		if err != nil {
			s.logger.Error("Failed to scan failed Falcon message", zap.Error(err))
			return nil, err
		}
		messages = append(messages, message)
	}

	s.logger.Info("Retrieved failed Falcon messages for retry", zap.Int("count", len(messages)))
	return messages, nil
}

// UpdateReconciliationStatus updates reconciliation status
func (s *falconMessageStorageImpl) UpdateReconciliationStatus(ctx context.Context, messageID string, status dto.FalconReconciliationStatus, notes string) error {
	s.logger.Info("Updating reconciliation status",
		zap.String("message_id", messageID),
		zap.String("status", string(status)))

	query := `
		UPDATE falcon_liquidity_messages 
		SET reconciliation_status = $1, reconciliation_notes = $2
		WHERE message_id = $3`

	_, err := s.db.GetPool().Exec(ctx, query, status, notes, messageID)
	if err != nil {
		s.logger.Error("Failed to update reconciliation status", zap.Error(err))
		return err
	}

	s.logger.Info("Reconciliation status updated successfully", zap.String("message_id", messageID))
	return nil
}
