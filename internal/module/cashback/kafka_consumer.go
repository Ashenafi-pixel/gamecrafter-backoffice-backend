package cashback

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/platform"
	"go.uber.org/zap"
)

// BetEvent represents a bet event from Kafka
type BetEvent struct {
	EventType   string          `json:"event_type"` // "bet_completed", "bet_won", "bet_lost"
	BetID       uuid.UUID       `json:"bet_id"`
	UserID      uuid.UUID       `json:"user_id"`
	GameType    string          `json:"game_type"`
	GameVariant *string         `json:"game_variant,omitempty"`
	Amount      decimal.Decimal `json:"amount"`
	WinAmount   decimal.Decimal `json:"win_amount"`
	HouseEdge   decimal.Decimal `json:"house_edge"`
	Timestamp   time.Time       `json:"timestamp"`
}

// CashbackKafkaConsumer handles Kafka events for cashback processing
type CashbackKafkaConsumer struct {
	cashbackService *CashbackService
	kafkaClient     platform.Kafka
	logger          *zap.Logger
}

func NewCashbackKafkaConsumer(
	cashbackService *CashbackService,
	kafkaClient platform.Kafka,
	logger *zap.Logger,
) *CashbackKafkaConsumer {
	return &CashbackKafkaConsumer{
		cashbackService: cashbackService,
		kafkaClient:     kafkaClient,
		logger:          logger,
	}
}

// StartConsumer starts the Kafka consumer for bet events
func (c *CashbackKafkaConsumer) StartConsumer(ctx context.Context) error {
	c.logger.Info("Starting cashback Kafka consumer")

	// Register event handler for bet events
	c.kafkaClient.RegisterKafkaEventHandler("bet.completed", c.handleBetCompletedEvent)
	c.kafkaClient.RegisterKafkaEventHandler("bet.won", c.handleBetWonEvent)
	c.kafkaClient.RegisterKafkaEventHandler("bet.lost", c.handleBetLostEvent)

	// Start the consumer
	go c.kafkaClient.StartConsumer(ctx)

	c.logger.Info("Cashback Kafka consumer started successfully")
	return nil
}

// handleBetEvent processes incoming bet events
func (c *CashbackKafkaConsumer) handleBetEvent(ctx context.Context, message []byte) error {
	c.logger.Debug("Received bet event", zap.String("message", string(message)))

	var betEvent BetEvent
	if err := json.Unmarshal(message, &betEvent); err != nil {
		c.logger.Error("Failed to unmarshal bet event", zap.Error(err))
		return fmt.Errorf("failed to unmarshal bet event: %w", err)
	}

	// Only process completed bets
	if betEvent.EventType != "bet_completed" {
		c.logger.Debug("Skipping non-completed bet event", zap.String("event_type", betEvent.EventType))
		return nil
	}

	// Convert to bet DTO
	bet := dto.Bet{
		BetID:     betEvent.BetID,
		UserID:    betEvent.UserID,
		Amount:    betEvent.Amount,
		Currency:  "USD", // Default currency
		Payout:    betEvent.WinAmount,
		Timestamp: betEvent.Timestamp,
		Status:    "completed",
	}

	// Process cashback
	if err := c.cashbackService.ProcessBetCashback(ctx, bet); err != nil {
		c.logger.Error("Failed to process bet cashback",
			zap.Error(err),
			zap.String("bet_id", betEvent.BetID.String()),
			zap.String("user_id", betEvent.UserID.String()))
		return fmt.Errorf("failed to process bet cashback: %w", err)
	}

	c.logger.Info("Successfully processed bet cashback",
		zap.String("bet_id", betEvent.BetID.String()),
		zap.String("user_id", betEvent.UserID.String()),
		zap.String("amount", betEvent.Amount.String()))

	return nil
}

// PublishBetEvent publishes a bet event to Kafka (for testing or manual triggers)
func (c *CashbackKafkaConsumer) PublishBetEvent(ctx context.Context, betEvent BetEvent) error {
	c.logger.Info("Publishing bet event",
		zap.String("bet_id", betEvent.BetID.String()),
		zap.String("event_type", betEvent.EventType))

	err := c.kafkaClient.WriteMessage(ctx, "bet_events", betEvent.BetID.String(), betEvent)
	if err != nil {
		c.logger.Error("Failed to publish bet event", zap.Error(err))
		return fmt.Errorf("failed to publish bet event: %w", err)
	}

	c.logger.Info("Bet event published successfully", zap.String("bet_id", betEvent.BetID.String()))
	return nil
}

// ProcessExpiredCashbackJob runs periodically to process expired cashback earnings
func (c *CashbackKafkaConsumer) ProcessExpiredCashbackJob(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour) // Run every hour
	defer ticker.Stop()

	c.logger.Info("Starting expired cashback processing job")

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Expired cashback processing job stopped")
			return
		case <-ticker.C:
			c.logger.Info("Processing expired cashback earnings")

			if err := c.cashbackService.ProcessExpiredCashback(ctx); err != nil {
				c.logger.Error("Failed to process expired cashback", zap.Error(err))
			} else {
				c.logger.Info("Expired cashback processing completed")
			}
		}
	}
}

// CreateTestBetEvent creates a test bet event for testing purposes
func (c *CashbackKafkaConsumer) CreateTestBetEvent(
	userID uuid.UUID,
	gameType string,
	amount decimal.Decimal,
	houseEdge decimal.Decimal,
) BetEvent {
	return BetEvent{
		EventType: "bet_completed",
		BetID:     uuid.New(),
		UserID:    userID,
		GameType:  gameType,
		Amount:    amount,
		WinAmount: decimal.Zero, // Assume loss for cashback calculation
		HouseEdge: houseEdge,
		Timestamp: time.Now(),
	}
}

// handleBetCompletedEvent processes bet completed events
func (c *CashbackKafkaConsumer) handleBetCompletedEvent(ctx context.Context, event []byte) (bool, error) {
	c.logger.Info("Processing bet completed event for cashback")
	err := c.handleBetEvent(ctx, event)
	return err == nil, err
}

// handleBetWonEvent processes bet won events
func (c *CashbackKafkaConsumer) handleBetWonEvent(ctx context.Context, event []byte) (bool, error) {
	c.logger.Info("Processing bet won event for cashback")
	// For now, just log the event
	// In the future, we might want to apply different cashback rules for winning bets
	return true, nil
}

// handleBetLostEvent processes bet lost events
func (c *CashbackKafkaConsumer) handleBetLostEvent(ctx context.Context, event []byte) (bool, error) {
	c.logger.Info("Processing bet lost event for cashback")
	// For now, just log the event
	// In the future, we might want to apply different cashback rules for losing bets
	return true, nil
}
