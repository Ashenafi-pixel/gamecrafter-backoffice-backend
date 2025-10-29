package cashback

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/storage/cashback"
	"github.com/tucanbit/internal/storage/groove"
	"go.uber.org/zap"
)

// RetryConfig represents configuration for retry mechanism
type RetryConfig struct {
	MaxRetries        int           `json:"max_retries"`
	InitialDelay      time.Duration `json:"initial_delay"`
	MaxDelay          time.Duration `json:"max_delay"`
	BackoffMultiplier float64       `json:"backoff_multiplier"`
	JitterEnabled     bool          `json:"jitter_enabled"`
}

// DefaultRetryConfig returns default retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:        3,
		InitialDelay:      time.Second,
		MaxDelay:          30 * time.Second,
		BackoffMultiplier: 2.0,
		JitterEnabled:     true,
	}
}

// Use the RetryableOperation from the storage package
type RetryableOperation = cashback.RetryableOperation

// RetryService handles retry logic for cashback operations
type RetryService struct {
	logger          *zap.Logger
	storage         cashback.CashbackStorage
	grooveStorage   groove.GrooveStorage
	cashbackService *CashbackService
	retryConfig     *RetryConfig
}

// NewRetryService creates a new retry service
func NewRetryService(logger *zap.Logger, storage cashback.CashbackStorage, grooveStorage groove.GrooveStorage, cashbackService *CashbackService, config *RetryConfig) *RetryService {
	if config == nil {
		config = DefaultRetryConfig()
	}

	return &RetryService{
		logger:          logger,
		storage:         storage,
		grooveStorage:   grooveStorage,
		cashbackService: cashbackService,
		retryConfig:     config,
	}
}

// RetryOperation executes an operation with retry logic
func (rs *RetryService) RetryOperation(ctx context.Context, operationType string, userID uuid.UUID, data map[string]interface{}, operation func() error) error {
	operationID := uuid.New()

	rs.logger.Info("Starting retryable operation",
		zap.String("operation_id", operationID.String()),
		zap.String("operation_type", operationType),
		zap.String("user_id", userID.String()))

	// Create retryable operation record
	retryOp := &RetryableOperation{
		ID:        operationID,
		Type:      operationType,
		UserID:    userID,
		Data:      data,
		Attempts:  0,
		Status:    "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Store the operation for tracking
	err := rs.storage.CreateRetryableOperation(ctx, *retryOp)
	if err != nil {
		rs.logger.Error("Failed to create retryable operation record", zap.Error(err))
		// Continue without tracking if storage fails
	}

	// Execute with retry logic
	return rs.executeWithRetry(ctx, retryOp, operation)
}

// executeWithRetry executes the operation with exponential backoff retry
func (rs *RetryService) executeWithRetry(ctx context.Context, retryOp *RetryableOperation, operation func() error) error {
	var lastErr error

	for attempt := 0; attempt <= rs.retryConfig.MaxRetries; attempt++ {
		retryOp.Attempts = attempt
		retryOp.Status = "retrying"
		retryOp.UpdatedAt = time.Now()

		// Update operation record
		rs.storage.UpdateRetryableOperation(ctx, *retryOp)

		rs.logger.Info("Executing operation attempt",
			zap.String("operation_id", retryOp.ID.String()),
			zap.String("operation_type", retryOp.Type),
			zap.Int("attempt", attempt),
			zap.Int("max_retries", rs.retryConfig.MaxRetries))

		// Execute the operation
		err := operation()
		if err == nil {
			// Success
			retryOp.Status = "completed"
			retryOp.UpdatedAt = time.Now()
			rs.storage.UpdateRetryableOperation(ctx, *retryOp)

			rs.logger.Info("Operation completed successfully",
				zap.String("operation_id", retryOp.ID.String()),
				zap.String("operation_type", retryOp.Type),
				zap.Int("attempts", attempt+1))

			return nil
		}

		lastErr = err
		retryOp.LastError = err.Error()

		rs.logger.Warn("Operation attempt failed",
			zap.String("operation_id", retryOp.ID.String()),
			zap.String("operation_type", retryOp.Type),
			zap.Int("attempt", attempt),
			zap.Error(err))

		// If this was the last attempt, mark as failed
		if attempt == rs.retryConfig.MaxRetries {
			retryOp.Status = "failed"
			retryOp.UpdatedAt = time.Now()
			rs.storage.UpdateRetryableOperation(ctx, *retryOp)

			rs.logger.Error("Operation failed after all retries",
				zap.String("operation_id", retryOp.ID.String()),
				zap.String("operation_type", retryOp.Type),
				zap.Int("attempts", attempt+1),
				zap.Error(err))

			return errors.ErrInternalServerError.Wrap(err, fmt.Sprintf("operation failed after %d retries", rs.retryConfig.MaxRetries))
		}

		// Calculate delay for next retry
		delay := rs.calculateDelay(attempt)
		nextRetryTime := time.Now().Add(delay)
		retryOp.NextRetryAt = &nextRetryTime
		retryOp.Status = "pending"
		retryOp.UpdatedAt = time.Now()
		rs.storage.UpdateRetryableOperation(ctx, *retryOp)

		rs.logger.Info("Waiting before retry",
			zap.String("operation_id", retryOp.ID.String()),
			zap.String("operation_type", retryOp.Type),
			zap.Duration("delay", delay))

		// Wait before next retry
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			// Continue to next attempt
		}
	}

	return lastErr
}

// calculateDelay calculates the delay for the next retry using exponential backoff
func (rs *RetryService) calculateDelay(attempt int) time.Duration {
	// Calculate exponential backoff: initialDelay * (backoffMultiplier ^ attempt)
	delay := float64(rs.retryConfig.InitialDelay) * math.Pow(rs.retryConfig.BackoffMultiplier, float64(attempt))

	// Apply maximum delay limit
	if delay > float64(rs.retryConfig.MaxDelay) {
		delay = float64(rs.retryConfig.MaxDelay)
	}

	// Add jitter to prevent thundering herd
	if rs.retryConfig.JitterEnabled {
		jitter := delay * 0.1 * (0.5 - math.Mod(float64(time.Now().UnixNano()), 1.0))
		delay += jitter
	}

	return time.Duration(delay)
}

// RetryFailedOperations retries all failed operations that are ready for retry
func (rs *RetryService) RetryFailedOperations(ctx context.Context) error {
	rs.logger.Info("Starting retry of failed operations")

	// Get failed operations ready for retry
	failedOps, err := rs.storage.GetFailedRetryableOperations(ctx)
	if err != nil {
		rs.logger.Error("Failed to get failed retryable operations", zap.Error(err))
		return errors.ErrInternalServerError.Wrap(err, "failed to get failed retryable operations")
	}

	rs.logger.Info("Found failed operations to retry", zap.Int("count", len(failedOps)))

	for _, op := range failedOps {
		rs.logger.Info("Retrying failed operation",
			zap.String("operation_id", op.ID.String()),
			zap.String("operation_type", op.Type),
			zap.String("user_id", op.UserID.String()))

		// Create operation function based on type
		operation, err := rs.createOperationFunction(ctx, op)
		if err != nil {
			rs.logger.Error("Failed to create operation function", zap.Error(err))
			continue
		}

		// Retry the operation
		err = rs.executeWithRetry(ctx, &op, operation)
		if err != nil {
			rs.logger.Error("Failed to retry operation",
				zap.String("operation_id", op.ID.String()),
				zap.Error(err))
		}
	}

	rs.logger.Info("Completed retry of failed operations")
	return nil
}

// createOperationFunction creates the appropriate operation function based on operation type
func (rs *RetryService) createOperationFunction(ctx context.Context, op RetryableOperation) (func() error, error) {
	switch op.Type {
	case "process_bet_cashback":
		return rs.createProcessBetCashbackOperation(ctx, op)
	case "claim_cashback":
		return rs.createClaimCashbackOperation(ctx, op)
	case "update_user_level":
		return rs.createUpdateUserLevelOperation(ctx, op)
	default:
		return nil, fmt.Errorf("unknown operation type: %s", op.Type)
	}
}

// createProcessBetCashbackOperation creates a function to retry bet cashback processing
func (rs *RetryService) createProcessBetCashbackOperation(ctx context.Context, op RetryableOperation) (func() error, error) {
	// Extract bet data from operation data
	betIDStr, ok := op.Data["bet_id"].(string)
	if !ok {
		return nil, fmt.Errorf("missing bet_id in operation data")
	}

	betID, err := uuid.Parse(betIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid bet_id: %w", err)
	}

	// Extract all required bet data
	roundIDStr, ok := op.Data["round_id"].(string)
	if !ok {
		return nil, fmt.Errorf("missing round_id in operation data")
	}

	roundID, err := uuid.Parse(roundIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid round_id: %w", err)
	}

	clientTransactionID, ok := op.Data["client_transaction_id"].(string)
	if !ok {
		return nil, fmt.Errorf("missing client_transaction_id in operation data")
	}

	amountStr, ok := op.Data["amount"].(string)
	if !ok {
		return nil, fmt.Errorf("missing amount in operation data")
	}

	amount, err := decimal.NewFromString(amountStr)
	if err != nil {
		return nil, fmt.Errorf("invalid amount: %w", err)
	}

	currency, ok := op.Data["currency"].(string)
	if !ok {
		return nil, fmt.Errorf("missing currency in operation data")
	}

	payoutStr, ok := op.Data["payout"].(string)
	if !ok {
		return nil, fmt.Errorf("missing payout in operation data")
	}

	payout, err := decimal.NewFromString(payoutStr)
	if err != nil {
		return nil, fmt.Errorf("invalid payout: %w", err)
	}

	cashOutMultiplierStr, ok := op.Data["cash_out_multiplier"].(string)
	if !ok {
		return nil, fmt.Errorf("missing cash_out_multiplier in operation data")
	}

	cashOutMultiplier, err := decimal.NewFromString(cashOutMultiplierStr)
	if err != nil {
		return nil, fmt.Errorf("invalid cash_out_multiplier: %w", err)
	}

	status, ok := op.Data["status"].(string)
	if !ok {
		return nil, fmt.Errorf("missing status in operation data")
	}

	// Create bet DTO
	bet := dto.Bet{
		BetID:               betID,
		RoundID:             roundID,
		UserID:              op.UserID,
		ClientTransactionID: clientTransactionID,
		Amount:              amount,
		Currency:            currency,
		Timestamp:           time.Now(),
		Payout:              payout,
		CashOutMultiplier:   cashOutMultiplier,
		Status:              status,
	}

	// Return operation function that calls the actual cashback service
	return func() error {
		rs.logger.Info("Retrying bet cashback processing",
			zap.String("bet_id", betID.String()),
			zap.String("user_id", op.UserID.String()),
			zap.String("amount", amount.String()))

		// Call the actual cashback service to process the bet
		err := rs.cashbackService.ProcessBetCashback(ctx, bet)
		if err != nil {
			rs.logger.Error("Bet cashback retry failed",
				zap.String("bet_id", betID.String()),
				zap.String("user_id", op.UserID.String()),
				zap.Error(err))
			return fmt.Errorf("bet cashback retry failed: %w", err)
		}

		rs.logger.Info("Bet cashback retry succeeded",
			zap.String("bet_id", betID.String()),
			zap.String("user_id", op.UserID.String()))

		return nil
	}, nil
}

// createClaimCashbackOperation creates a function to retry cashback claim
func (rs *RetryService) createClaimCashbackOperation(ctx context.Context, op RetryableOperation) (func() error, error) {
	// Extract claim data from operation data
	claimIDStr, ok := op.Data["claim_id"].(string)
	if !ok {
		return nil, fmt.Errorf("missing claim_id in operation data")
	}

	claimID, err := uuid.Parse(claimIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid claim_id: %w", err)
	}

	amountStr, ok := op.Data["amount"].(string)
	if !ok {
		return nil, fmt.Errorf("missing amount in operation data")
	}

	amount, err := decimal.NewFromString(amountStr)
	if err != nil {
		return nil, fmt.Errorf("invalid amount: %w", err)
	}

	// Create claim request
	claimRequest := dto.CashbackClaimRequest{
		Amount: amount,
	}

	// Return operation function that calls the actual cashback service
	return func() error {
		rs.logger.Info("Retrying cashback claim",
			zap.String("claim_id", claimID.String()),
			zap.String("user_id", op.UserID.String()),
			zap.String("amount", amount.String()))

		// Call the actual cashback service to process the claim
		_, err := rs.cashbackService.ClaimCashback(ctx, op.UserID, claimRequest)
		if err != nil {
			rs.logger.Error("Cashback claim retry failed",
				zap.String("claim_id", claimID.String()),
				zap.String("user_id", op.UserID.String()),
				zap.Error(err))
			return fmt.Errorf("cashback claim retry failed: %w", err)
		}

		rs.logger.Info("Cashback claim retry succeeded",
			zap.String("claim_id", claimID.String()),
			zap.String("user_id", op.UserID.String()))

		return nil
	}, nil
}

// createUpdateUserLevelOperation creates a function to retry user level update
func (rs *RetryService) createUpdateUserLevelOperation(ctx context.Context, op RetryableOperation) (func() error, error) {
	// Extract user level data from operation data
	userLevel := dto.UserLevel{
		UserID:           op.UserID,
		CurrentLevel:     int(op.Data["current_level"].(float64)),
		TotalExpectedGGR: decimal.RequireFromString(op.Data["total_ggr"].(string)),
		TotalBets:        decimal.RequireFromString(op.Data["total_bets"].(string)),
		TotalWins:        decimal.RequireFromString(op.Data["total_wins"].(string)),
		LevelProgress:    decimal.RequireFromString(op.Data["level_progress"].(string)),
		CurrentTierID:    uuid.MustParse(op.Data["current_tier_id"].(string)),
	}

	// Return operation function
	return func() error {
		_, err := rs.storage.UpdateUserLevel(ctx, userLevel)
		return err
	}, nil
}

// GetRetryableOperations returns retryable operations for a user
func (rs *RetryService) GetRetryableOperations(ctx context.Context, userID uuid.UUID) ([]RetryableOperation, error) {
	rs.logger.Info("Getting retryable operations", zap.String("user_id", userID.String()))

	operations, err := rs.storage.GetRetryableOperationsByUser(ctx, userID)
	if err != nil {
		rs.logger.Error("Failed to get retryable operations", zap.Error(err))
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to get retryable operations")
	}

	rs.logger.Info("Retrieved retryable operations", zap.Int("count", len(operations)))
	return operations, nil
}

// ManualRetryOperation manually retries a specific operation
func (rs *RetryService) ManualRetryOperation(ctx context.Context, operationID uuid.UUID) error {
	rs.logger.Info("Manually retrying operation", zap.String("operation_id", operationID.String()))

	// Get the operation
	operation, err := rs.storage.GetRetryableOperation(ctx, operationID)
	if err != nil {
		rs.logger.Error("Failed to get retryable operation", zap.Error(err))
		return errors.ErrInternalServerError.Wrap(err, "failed to get retryable operation")
	}

	if operation == nil {
		return fmt.Errorf("operation not found")
	}

	// Create operation function
	opFunc, err := rs.createOperationFunction(ctx, *operation)
	if err != nil {
		rs.logger.Error("Failed to create operation function", zap.Error(err))
		return errors.ErrInternalServerError.Wrap(err, "failed to create operation function")
	}

	// Reset operation status and retry
	operation.Status = "pending"
	operation.Attempts = 0
	operation.LastError = ""
	operation.NextRetryAt = nil
	operation.UpdatedAt = time.Now()

	// Update operation record
	err = rs.storage.UpdateRetryableOperation(ctx, *operation)
	if err != nil {
		rs.logger.Error("Failed to update retryable operation", zap.Error(err))
		return errors.ErrInternalServerError.Wrap(err, "failed to update retryable operation")
	}

	// Execute with retry
	return rs.executeWithRetry(ctx, operation, opFunc)
}
