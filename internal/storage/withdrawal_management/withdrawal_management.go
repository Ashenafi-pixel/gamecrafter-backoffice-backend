package withdrawal_management

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/constant/persistencedb"
	"go.uber.org/zap"
)

type WithdrawalManagement struct {
	db  *persistencedb.PersistenceDB
	log *zap.Logger
}

// PausedWithdrawal represents a paused withdrawal with its details
type PausedWithdrawal struct {
	ID                   uuid.UUID `json:"id"`
	UserID               uuid.UUID `json:"user_id"`
	WithdrawalID         string    `json:"withdrawal_id"`
	USDAmountCents       int64     `json:"usd_amount_cents"`
	CryptoAmount         string    `json:"crypto_amount"`
	CurrencyCode         string    `json:"currency_code"`
	Status               string    `json:"status"`
	PauseReason          string    `json:"pause_reason"`
	PausedAt             time.Time `json:"paused_at"`
	RequiresManualReview bool      `json:"requires_manual_review"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
	Username             *string   `json:"username,omitempty"`
	Email                *string   `json:"email,omitempty"`
}

// WithdrawalPauseDetails stores pause information in system_config
type WithdrawalPauseDetails struct {
	WithdrawalID   string    `json:"withdrawal_id"`
	PauseReason    string    `json:"pause_reason"`
	PausedAt       time.Time `json:"paused_at"`
	PausedBy       *string   `json:"paused_by,omitempty"`
	RequiresReview bool      `json:"requires_review"`
	ThresholdType  *string   `json:"threshold_type,omitempty"`
	ThresholdValue *float64  `json:"threshold_value,omitempty"`
	Notes          *string   `json:"notes,omitempty"`
}

// PausedWithdrawalsMap stores all paused withdrawals in system_config
type PausedWithdrawalsMap map[string]WithdrawalPauseDetails

func NewWithdrawalManagement(db *persistencedb.PersistenceDB, log *zap.Logger) *WithdrawalManagement {
	return &WithdrawalManagement{
		db:  db,
		log: log,
	}
}

// GetPausedWithdrawals retrieves all paused withdrawals
func (w *WithdrawalManagement) GetPausedWithdrawals(ctx context.Context, limit, offset int) ([]PausedWithdrawal, int64, error) {
	w.log.Info("Getting paused withdrawals")

	// Get paused withdrawals from system_config
	var configValue json.RawMessage
	err := w.db.GetPool().QueryRow(ctx, "SELECT config_value FROM system_config WHERE config_key = $1", "withdrawal_paused_transactions").Scan(&configValue)
	if err != nil {
		w.log.Error("Failed to get paused withdrawals config", zap.Error(err))
		return []PausedWithdrawal{}, 0, errors.ErrUnableToGet.Wrap(err, "failed to get paused withdrawals")
	}

	var pausedMap PausedWithdrawalsMap
	if err := json.Unmarshal(configValue, &pausedMap); err != nil {
		w.log.Error("Failed to unmarshal paused withdrawals", zap.Error(err))
		return []PausedWithdrawal{}, 0, errors.ErrUnableToGet.Wrap(err, "failed to unmarshal paused withdrawals")
	}

	// Get withdrawal IDs that are paused
	var withdrawalIDs []string
	for withdrawalID := range pausedMap {
		withdrawalIDs = append(withdrawalIDs, withdrawalID)
	}

	if len(withdrawalIDs) == 0 {
		return []PausedWithdrawal{}, 0, nil
	}

	// Query withdrawals table for these IDs using raw SQL
	query := `
		SELECT 
			w.id, w.user_id, w.withdrawal_id, w.usd_amount_cents, w.crypto_amount, 
			w.currency_code, w.status, w.created_at, w.updated_at,
			u.username, u.email
		FROM withdrawals w
		LEFT JOIN users u ON w.user_id = u.id
		WHERE w.withdrawal_id = ANY($1::text[])
	`
	rows, err := w.db.GetPool().Query(ctx, query, withdrawalIDs)
	if err != nil {
		w.log.Error("Failed to get withdrawals by IDs", zap.Error(err))
		return []PausedWithdrawal{}, 0, errors.ErrUnableToGet.Wrap(err, "failed to get withdrawals")
	}
	defer rows.Close()

	// Define a struct to match the query result
	type WithdrawalRow struct {
		ID             uuid.UUID
		UserID         uuid.UUID
		WithdrawalID   string
		USDAmountCents int64
		CryptoAmount   string
		CurrencyCode   string
		Status         string
		CreatedAt      time.Time
		UpdatedAt      time.Time
		Username       sql.NullString
		Email          sql.NullString
	}

	var withdrawalRows []WithdrawalRow
	for rows.Next() {
		var wr WithdrawalRow
		err := rows.Scan(
			&wr.ID, &wr.UserID, &wr.WithdrawalID, &wr.USDAmountCents, &wr.CryptoAmount,
			&wr.CurrencyCode, &wr.Status, &wr.CreatedAt, &wr.UpdatedAt,
			&wr.Username, &wr.Email,
		)
		if err != nil {
			w.log.Error("Failed to scan withdrawal row", zap.Error(err))
			continue
		}
		withdrawalRows = append(withdrawalRows, wr)
	}

	// Convert to PausedWithdrawal format
	var pausedWithdrawals []PausedWithdrawal
	for _, withdrawal := range withdrawalRows {
		pauseDetails, exists := pausedMap[withdrawal.WithdrawalID]
		if !exists {
			continue
		}

		pausedWithdrawal := PausedWithdrawal{
			ID:                   withdrawal.ID,
			UserID:               withdrawal.UserID,
			WithdrawalID:         withdrawal.WithdrawalID,
			USDAmountCents:       withdrawal.USDAmountCents,
			CryptoAmount:         withdrawal.CryptoAmount,
			CurrencyCode:         withdrawal.CurrencyCode,
			Status:               string(withdrawal.Status),
			PauseReason:          pauseDetails.PauseReason,
			PausedAt:             pauseDetails.PausedAt,
			RequiresManualReview: pauseDetails.RequiresReview,
			CreatedAt:            withdrawal.CreatedAt,
			UpdatedAt:            withdrawal.UpdatedAt,
		}

		// Get user details if available
		if withdrawal.Username.Valid {
			pausedWithdrawal.Username = &withdrawal.Username.String
		}
		if withdrawal.Email.Valid {
			pausedWithdrawal.Email = &withdrawal.Email.String
		}

		pausedWithdrawals = append(pausedWithdrawals, pausedWithdrawal)
	}

	// Apply pagination
	total := int64(len(pausedWithdrawals))
	start := offset
	end := offset + limit
	if start >= len(pausedWithdrawals) {
		return []PausedWithdrawal{}, total, nil
	}
	if end > len(pausedWithdrawals) {
		end = len(pausedWithdrawals)
	}

	return pausedWithdrawals[start:end], total, nil
}

// PauseWithdrawal pauses a withdrawal and stores the details in system_config
func (w *WithdrawalManagement) PauseWithdrawal(ctx context.Context, withdrawalID string, reason string, adminID *uuid.UUID, requiresReview bool, thresholdType *string, thresholdValue *float64) error {
	w.log.Info("Pausing withdrawal", zap.String("withdrawal_id", withdrawalID), zap.String("reason", reason))

	// Get current paused withdrawals
	var configValue json.RawMessage
	err := w.db.GetPool().QueryRow(ctx, "SELECT config_value FROM system_config WHERE config_key = $1", "withdrawal_paused_transactions").Scan(&configValue)
	if err != nil {
		w.log.Error("Failed to get paused withdrawals config", zap.Error(err))
		return errors.ErrUnableToGet.Wrap(err, "failed to get paused withdrawals config")
	}

	var pausedMap PausedWithdrawalsMap
	if err := json.Unmarshal(configValue, &pausedMap); err != nil {
		// If unmarshaling fails, initialize empty map
		pausedMap = make(PausedWithdrawalsMap)
	}

	// Add new paused withdrawal
	pauseDetails := WithdrawalPauseDetails{
		WithdrawalID:   withdrawalID,
		PauseReason:    reason,
		PausedAt:       time.Now(),
		PausedBy:       nil,
		RequiresReview: requiresReview,
		ThresholdType:  thresholdType,
		ThresholdValue: thresholdValue,
	}

	if adminID != nil {
		adminIDStr := adminID.String()
		pauseDetails.PausedBy = &adminIDStr
	}

	pausedMap[withdrawalID] = pauseDetails

	// Update system_config
	updatedConfig, err := json.Marshal(pausedMap)
	if err != nil {
		w.log.Error("Failed to marshal paused withdrawals", zap.Error(err))
		return errors.ErrUnableToUpdate.Wrap(err, "failed to marshal paused withdrawals")
	}

	_, err = w.db.GetPool().Exec(ctx, "UPDATE system_config SET config_value = $1, updated_by = $2, updated_at = NOW() WHERE config_key = $3", updatedConfig, adminID, "withdrawal_paused_transactions")
	if err != nil {
		w.log.Error("Failed to update paused withdrawals config", zap.Error(err))
		return errors.ErrUnableToUpdate.Wrap(err, "failed to update paused withdrawals config")
	}

	w.log.Info("Successfully paused withdrawal", zap.String("withdrawal_id", withdrawalID))
	return nil
}

// UnpauseWithdrawal removes a withdrawal from the paused list
func (w *WithdrawalManagement) UnpauseWithdrawal(ctx context.Context, withdrawalID string, adminID *uuid.UUID) error {
	w.log.Info("Unpausing withdrawal", zap.String("withdrawal_id", withdrawalID))

	// Get current paused withdrawals
	var configValue json.RawMessage
	err := w.db.GetPool().QueryRow(ctx, "SELECT config_value FROM system_config WHERE config_key = $1", "withdrawal_paused_transactions").Scan(&configValue)
	if err != nil {
		w.log.Error("Failed to get paused withdrawals config", zap.Error(err))
		return errors.ErrUnableToGet.Wrap(err, "failed to get paused withdrawals config")
	}

	var pausedMap PausedWithdrawalsMap
	if err := json.Unmarshal(configValue, &pausedMap); err != nil {
		w.log.Error("Failed to unmarshal paused withdrawals", zap.Error(err))
		return errors.ErrUnableToGet.Wrap(err, "failed to unmarshal paused withdrawals")
	}

	// Remove withdrawal from paused list
	delete(pausedMap, withdrawalID)

	// Update system_config
	updatedConfig, err := json.Marshal(pausedMap)
	if err != nil {
		w.log.Error("Failed to marshal paused withdrawals", zap.Error(err))
		return errors.ErrUnableToUpdate.Wrap(err, "failed to marshal paused withdrawals")
	}

	_, err = w.db.GetPool().Exec(ctx, "UPDATE system_config SET config_value = $1, updated_by = $2, updated_at = NOW() WHERE config_key = $3", updatedConfig, adminID, "withdrawal_paused_transactions")
	if err != nil {
		w.log.Error("Failed to update paused withdrawals config", zap.Error(err))
		return errors.ErrUnableToUpdate.Wrap(err, "failed to update paused withdrawals config")
	}

	w.log.Info("Successfully unpaused withdrawal", zap.String("withdrawal_id", withdrawalID))
	return nil
}

// ApproveWithdrawal approves a paused withdrawal and removes it from paused list
func (w *WithdrawalManagement) ApproveWithdrawal(ctx context.Context, withdrawalID string, adminID uuid.UUID, notes *string) error {
	w.log.Info("Approving withdrawal", zap.String("withdrawal_id", withdrawalID))

	// Remove from paused list
	err := w.UnpauseWithdrawal(ctx, withdrawalID, &adminID)
	if err != nil {
		return err
	}

	// Update withdrawal status to processing or completed
	// This would depend on your business logic
	// For now, we'll just log the approval
	w.log.Info("Withdrawal approved",
		zap.String("withdrawal_id", withdrawalID),
		zap.String("admin_id", adminID.String()),
		zap.String("notes", fmt.Sprintf("%v", notes)))

	return nil
}

// RejectWithdrawal rejects a paused withdrawal and removes it from paused list
func (w *WithdrawalManagement) RejectWithdrawal(ctx context.Context, withdrawalID string, adminID uuid.UUID, notes *string) error {
	w.log.Info("Rejecting withdrawal", zap.String("withdrawal_id", withdrawalID))

	// Remove from paused list
	err := w.UnpauseWithdrawal(ctx, withdrawalID, &adminID)
	if err != nil {
		return err
	}

	// Update withdrawal status to failed or cancelled
	// This would depend on your business logic
	// For now, we'll just log the rejection
	w.log.Info("Withdrawal rejected",
		zap.String("withdrawal_id", withdrawalID),
		zap.String("admin_id", adminID.String()),
		zap.String("notes", fmt.Sprintf("%v", notes)))

	return nil
}

// GetWithdrawalPauseStats retrieves statistics about paused withdrawals
func (w *WithdrawalManagement) GetWithdrawalPauseStats(ctx context.Context) (map[string]interface{}, error) {
	w.log.Info("Getting withdrawal pause stats")

	// Get paused withdrawals
	pausedWithdrawals, _, err := w.GetPausedWithdrawals(ctx, 1000, 0) // Get all for stats
	if err != nil {
		return nil, err
	}

	// Calculate stats
	stats := map[string]interface{}{
		"total_paused":        len(pausedWithdrawals),
		"pending_review":      0,
		"paused_today":        0,
		"paused_this_hour":    0,
		"total_paused_amount": int64(0),
	}

	now := time.Now()
	today := now.Truncate(24 * time.Hour)
	oneHourAgo := now.Add(-time.Hour)

	for _, withdrawal := range pausedWithdrawals {
		if withdrawal.RequiresManualReview {
			stats["pending_review"] = stats["pending_review"].(int) + 1
		}

		if withdrawal.PausedAt.After(today) {
			stats["paused_today"] = stats["paused_today"].(int) + 1
		}

		if withdrawal.PausedAt.After(oneHourAgo) {
			stats["paused_this_hour"] = stats["paused_this_hour"].(int) + 1
		}

		stats["total_paused_amount"] = stats["total_paused_amount"].(int64) + withdrawal.USDAmountCents
	}

	return stats, nil
}
