package falcon_liquidity

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"
	"github.com/shopspring/decimal"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/storage/falcon_liquidity"
	"go.uber.org/zap"
)

// FalconLiquidityService handles publishing bet data to Falcon Liquidity via RabbitMQ
type FalconLiquidityService interface {
	PublishCasinoBet(ctx context.Context, betData *dto.FalconCasinoBetData) error
	PublishSportBet(ctx context.Context, betData *dto.FalconSportBetData) error
	Close() error
	IsConnected() bool
}

// falconLiquidityServiceImpl implements FalconLiquidityService
type falconLiquidityServiceImpl struct {
	logger         *zap.Logger
	config         *dto.FalconLiquidityConfig
	connection     *amqp091.Connection
	channel        *amqp091.Channel
	connected      bool
	messageStorage falcon_liquidity.FalconMessageStorage
	reconnectChan  chan struct{}
	stopChan       chan struct{}
}

// NewFalconLiquidityService creates a new Falcon Liquidity service
func NewFalconLiquidityService(logger *zap.Logger, config *dto.FalconLiquidityConfig, messageStorage falcon_liquidity.FalconMessageStorage) FalconLiquidityService {
	if !config.Enabled {
		logger.Info("Falcon Liquidity integration is disabled")
		return &falconLiquidityServiceImpl{
			logger: logger,
			config: config,
		}
	}

	service := &falconLiquidityServiceImpl{
		logger:         logger,
		config:         config,
		messageStorage: messageStorage,
		reconnectChan:  make(chan struct{}, 1),
		stopChan:       make(chan struct{}),
	}

	// Initialize connection
	if err := service.connect(); err != nil {
		logger.Error("Failed to initialize Falcon Liquidity connection", zap.Error(err))
		// Start connection monitoring even if initial connection fails
		go service.connectionMonitor()
		return service
	}

	// Start connection monitoring
	go service.connectionMonitor()

	logger.Info("Falcon Liquidity service initialized successfully",
		zap.String("host", config.Host),
		zap.Int("port", config.Port),
		zap.String("exchange", config.ExchangeName),
		zap.String("routing_key", config.RoutingKey))

	return service
}

// connect establishes connection to RabbitMQ
func (s *falconLiquidityServiceImpl) connect() error {
	if s.config == nil || !s.config.Enabled {
		return fmt.Errorf("falcon liquidity integration is disabled")
	}

	// Build connection URL
	url := fmt.Sprintf("amqps://%s:%s@%s:%d/%s",
		s.config.Username,
		s.config.Password,
		s.config.Host,
		s.config.Port,
		s.config.VirtualHost)

	s.logger.Info("Connecting to Falcon Liquidity RabbitMQ", zap.String("url", url))

	// Configure TLS to skip certificate verification for testing
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true, // Skip certificate verification for testing
	}

	// Configure connection with retry and heartbeat
	config := amqp091.Config{
		TLSClientConfig: tlsConfig,
		Heartbeat:       30 * time.Second, // 30 second heartbeat
		Locale:          "en_US",
		Vhost:           s.config.VirtualHost,
		Properties: map[string]interface{}{
			"connection_name": s.config.ClientName,
		},
	}

	// Establish connection with custom config
	conn, err := amqp091.DialConfig(url, config)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	// Create channel
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare exchange (this will create it if it doesn't exist)
	err = ch.ExchangeDeclare(
		s.config.ExchangeName, // name
		"direct",              // type
		true,                  // durable
		false,                 // auto-deleted
		false,                 // internal
		false,                 // no-wait
		nil,                   // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	s.connection = conn
	s.channel = ch
	s.connected = true

	// Set up connection close notification
	go s.handleConnectionClose()

	s.logger.Info("Successfully connected to Falcon Liquidity RabbitMQ",
		zap.String("host", s.config.Host),
		zap.Int("port", s.config.Port),
		zap.String("exchange", s.config.ExchangeName))
	return nil
}

// PublishCasinoBet publishes casino bet data to Falcon Liquidity with detailed tracking
func (s *falconLiquidityServiceImpl) PublishCasinoBet(ctx context.Context, betData *dto.FalconCasinoBetData) error {
	if !s.config.Enabled {
		s.logger.Debug("Falcon Liquidity integration is disabled, skipping casino bet publication")
		return nil
	}

	// Generate unique message ID for tracking
	messageID := fmt.Sprintf("falcon_casino_%s_%d", betData.BetID, time.Now().UnixNano())

	// Parse user ID for storage
	userID, err := uuid.Parse(betData.UserID)
	if err != nil {
		s.logger.Error("Failed to parse user ID for Falcon message tracking",
			zap.String("user_id", betData.UserID), zap.Error(err))
		userID = uuid.Nil // Use nil UUID if parsing fails
	}

	// Convert bet data to JSON for storage
	jsonData, err := json.Marshal(betData)
	if err != nil {
		s.logger.Error("Failed to marshal casino bet data for storage", zap.Error(err))
		return fmt.Errorf("failed to marshal casino bet data: %w", err)
	}

	// Create message record for tracking and reconciliation
	if s.messageStorage != nil {
		createReq := dto.CreateFalconMessageRequest{
			MessageID:        messageID,
			TransactionID:    betData.BetID,
			UserID:           userID,
			MessageType:      dto.FalconMessageTypeCasino,
			MessageData:      jsonData,
			BetAmount:        decimal.NewFromFloat(betData.Amount),
			PayoutAmount:     decimal.NewFromFloat(betData.Payout),
			Currency:         betData.Currency,
			GameName:         betData.GameName,
			GameID:           betData.GameID,
			HouseEdge:        decimal.NewFromFloat(betData.Edge),
			FalconRoutingKey: s.config.RoutingKey,
			FalconExchange:   s.config.ExchangeName,
			FalconQueue:      s.config.QueueName,
		}

		_, err := s.messageStorage.CreateFalconMessage(ctx, createReq)
		if err != nil {
			s.logger.Error("Failed to create Falcon message record", zap.Error(err))
			// Continue with publishing even if storage fails
		} else {
			s.logger.Info("Created Falcon message record for tracking",
				zap.String("message_id", messageID),
				zap.String("transaction_id", betData.BetID),
				zap.String("user_id", betData.UserID))
		}
	}

	// Check connection and trigger reconnection if needed
	if !s.IsConnected() {
		s.logger.Warn("Not connected to Falcon Liquidity, triggering reconnection")
		select {
		case s.reconnectChan <- struct{}{}:
		default:
		}

		// Wait a moment for reconnection attempt
		time.Sleep(2 * time.Second)

		if !s.IsConnected() {
			// Update message status to failed
			if s.messageStorage != nil {
				updateReq := dto.UpdateFalconMessageRequest{
					Status:       &[]dto.FalconMessageStatus{dto.FalconMessageStatusFailed}[0],
					ErrorMessage: &[]string{"Connection unavailable"}[0],
					ErrorCode:    &[]string{"CONNECTION_FAILED"}[0],
				}
				s.messageStorage.UpdateFalconMessage(ctx, messageID, updateReq)
			}
			return fmt.Errorf("falcon liquidity connection unavailable")
		}
	}

	// Log detailed message data for dispute resolution
	s.logger.Info("Message sent to Falcon: Casino Bet",
		zap.String("message_id", messageID),
		zap.String("transaction_id", betData.BetID),
		zap.String("user_id", betData.UserID),
		zap.String("username", betData.Username),
		zap.String("game_name", betData.GameName),
		zap.String("game_id", betData.GameID),
		zap.Float64("bet_amount", betData.Amount),
		zap.Float64("payout_amount", betData.Payout),
		zap.Float64("house_edge", betData.Edge),
		zap.String("currency", betData.Currency),
		zap.String("status", betData.Status),
		zap.Float64("payout_multiplier", betData.PayoutMultiplier),
		zap.String("provider", betData.Provider),
		zap.String("provider_type", betData.ProviderType),
		zap.String("falcon_exchange", s.config.ExchangeName),
		zap.String("falcon_routing_key", s.config.RoutingKey),
		zap.String("falcon_queue", s.config.QueueName),
		zap.Time("sent_at", time.Now()))

	// Publish message to RabbitMQ
	err = s.channel.PublishWithContext(
		ctx,
		s.config.ExchangeName, // exchange
		s.config.RoutingKey,   // routing key
		false,                 // mandatory
		false,                 // immediate
		amqp091.Publishing{
			ContentType:  "application/json",
			Body:         jsonData,
			DeliveryMode: amqp091.Persistent, // Make message persistent
			Timestamp:    time.Now(),
			MessageId:    messageID, // Add message ID for tracking
		},
	)

	if err != nil {
		// Update message status to failed
		if s.messageStorage != nil {
			updateReq := dto.UpdateFalconMessageRequest{
				Status:       &[]dto.FalconMessageStatus{dto.FalconMessageStatusFailed}[0],
				ErrorMessage: &[]string{fmt.Sprintf("Failed to publish: %v", err)}[0],
				ErrorCode:    &[]string{"PUBLISH_FAILED"}[0],
			}
			s.messageStorage.UpdateFalconMessage(ctx, messageID, updateReq)
		}

		s.logger.Error("Failed to publish casino bet to Falcon Liquidity",
			zap.Error(err),
			zap.String("message_id", messageID),
			zap.String("bet_id", betData.BetID),
			zap.String("user_id", betData.UserID),
			zap.String("game_name", betData.GameName))
		return fmt.Errorf("failed to publish casino bet: %w", err)
	}

	// Update message status to sent
	if s.messageStorage != nil {
		sentAt := time.Now()
		updateReq := dto.UpdateFalconMessageRequest{
			Status: &[]dto.FalconMessageStatus{dto.FalconMessageStatusSent}[0],
			SentAt: &sentAt,
		}
		s.messageStorage.UpdateFalconMessage(ctx, messageID, updateReq)
	}

	s.logger.Info("Successfully published casino bet to Falcon Liquidity",
		zap.String("message_id", messageID),
		zap.String("bet_id", betData.BetID),
		zap.String("user_id", betData.UserID),
		zap.String("game_name", betData.GameName),
		zap.Float64("amount", betData.Amount),
		zap.Float64("payout", betData.Payout),
		zap.Time("sent_at", time.Now()))

	return nil
}

// PublishSportBet publishes sport bet data to Falcon Liquidity
func (s *falconLiquidityServiceImpl) PublishSportBet(ctx context.Context, betData *dto.FalconSportBetData) error {
	if !s.config.Enabled {
		s.logger.Debug("Falcon Liquidity integration is disabled, skipping sport bet publication")
		return nil
	}

	if !s.connected {
		s.logger.Warn("Not connected to Falcon Liquidity, attempting to reconnect")
		if err := s.connect(); err != nil {
			return fmt.Errorf("failed to reconnect to Falcon Liquidity: %w", err)
		}
	}

	// Convert bet data to JSON
	jsonData, err := json.Marshal(betData)
	if err != nil {
		return fmt.Errorf("failed to marshal sport bet data: %w", err)
	}

	// Publish message
	err = s.channel.PublishWithContext(
		ctx,
		s.config.ExchangeName, // exchange
		s.config.RoutingKey,   // routing key
		false,                 // mandatory
		false,                 // immediate
		amqp091.Publishing{
			ContentType:  "application/json",
			Body:         jsonData,
			DeliveryMode: amqp091.Persistent, // Make message persistent
			Timestamp:    time.Now(),
		},
	)

	if err != nil {
		s.logger.Error("Failed to publish sport bet to Falcon Liquidity",
			zap.Error(err),
			zap.String("bet_id", betData.BetID),
			zap.String("user_id", betData.UserID))
		return fmt.Errorf("failed to publish sport bet: %w", err)
	}

	s.logger.Info("Successfully published sport bet to Falcon Liquidity",
		zap.String("bet_id", betData.BetID),
		zap.String("user_id", betData.UserID),
		zap.Float64("amount", betData.Amount),
		zap.Float64("payout", betData.Payout))

	return nil
}

// Close closes the RabbitMQ connection
func (s *falconLiquidityServiceImpl) Close() error {
	// Stop connection monitoring
	close(s.stopChan)

	if s.channel != nil {
		s.channel.Close()
	}
	if s.connection != nil {
		s.connection.Close()
	}
	s.connected = false
	s.logger.Info("Falcon Liquidity connection closed")
	return nil
}

// IsConnected returns whether the service is connected to RabbitMQ
func (s *falconLiquidityServiceImpl) IsConnected() bool {
	return s.connected && s.connection != nil && !s.connection.IsClosed()
}

// connectionMonitor monitors the connection and handles reconnection
func (s *falconLiquidityServiceImpl) connectionMonitor() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			s.logger.Info("Connection monitor stopped")
			return
		case <-s.reconnectChan:
			s.logger.Info("Reconnection requested")
			s.reconnect()
		case <-ticker.C:
			if !s.IsConnected() {
				s.logger.Warn("Connection lost, attempting to reconnect")
				s.reconnect()
			}
		}
	}
}

// handleConnectionClose handles connection close notifications
func (s *falconLiquidityServiceImpl) handleConnectionClose() {
	if s.connection == nil {
		return
	}

	closeChan := s.connection.NotifyClose(make(chan *amqp091.Error, 1))

	for err := range closeChan {
		if err != nil {
			s.logger.Error("Connection closed unexpectedly", zap.Error(err))
		} else {
			s.logger.Info("Connection closed gracefully")
		}

		s.connected = false
		s.connection = nil
		s.channel = nil

		// Trigger reconnection
		select {
		case s.reconnectChan <- struct{}{}:
		default:
		}
	}
}

// reconnect attempts to reconnect with exponential backoff
func (s *falconLiquidityServiceImpl) reconnect() {
	maxRetries := 5
	baseDelay := 1 * time.Second

	for i := 0; i < maxRetries; i++ {
		delay := baseDelay * time.Duration(1<<uint(i)) // Exponential backoff
		s.logger.Info("Attempting to reconnect",
			zap.Int("attempt", i+1),
			zap.Duration("delay", delay))

		time.Sleep(delay)

		if err := s.connect(); err != nil {
			s.logger.Error("Reconnection attempt failed",
				zap.Int("attempt", i+1),
				zap.Error(err))
			continue
		}

		s.logger.Info("Successfully reconnected to Falcon Liquidity")
		return
	}

	s.logger.Error("Failed to reconnect after maximum retries", zap.Int("max_retries", maxRetries))
}

// Helper function to convert decimal to float64 for Falcon API
func decimalToFloat64(d decimal.Decimal) float64 {
	f, _ := d.Float64()
	return f
}

// CreateCasinoBetFromGrooveTransaction creates a Falcon casino bet from GrooveTech transaction data
func CreateCasinoBetFromGrooveTransaction(transactionID, userID, username, gameName string, betAmount, payout decimal.Decimal, edge float64) *dto.FalconCasinoBetData {
	status := "settled"
	active := false
	if payout.GreaterThan(decimal.Zero) {
		active = true
	}

	return dto.NewFalconCasinoBet(
		transactionID,
		userID,
		username,
		gameName,
		"GrooveTech",
		"casino",
		"USD",
		status,
		decimalToFloat64(betAmount),
		decimalToFloat64(payout),
		edge,
		active,
	)
}
