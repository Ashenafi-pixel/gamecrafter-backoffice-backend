package initiator

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/IBM/sarama"
	"github.com/spf13/viper"
	"github.com/tucanbit/internal/contracts"
	"github.com/tucanbit/platform/logger"
	"github.com/tucanbit/platform/pisi"
	"github.com/tucanbit/platform/redis"
	"go.uber.org/zap"
)

// Local Kafka interface to avoid import cycle
type Kafka interface {
	WriteMessage(ctx context.Context, topic, key string, value interface{}) error
	RegisterKafkaEventHandler(eventType string, handler EventHandler)
	StartConsumer(ctx context.Context)
}

type EventHandler func(ctx context.Context, event json.RawMessage) (bool, error)

type Platform struct {
	Kafka Kafka
	Redis contracts.Redis
	Pisi  pisi.PisiClient
}

// Robust Kafka client that can handle configuration issues
type robustKafkaClient struct {
	logger logger.Logger
	// Store event handlers
	eventHandlers map[string]EventHandler
	// Kafka configuration
	bootstrapServer  string
	clusterAPIKey    string
	clusterAPISecret string
	securityProtocol string
	mechanisms       string
	acks             string
	topics           []string
	// Kafka producer (will be nil if Kafka is not available)
	producer sarama.AsyncProducer
	// Kafka consumer group (will be nil if Kafka is not available)
	consumerGroup sarama.ConsumerGroup
}

func newRobustKafkaClient(
	bootstrapServer, clusterAPIKey, clusterAPISecret, securityProtocol, mechanisms, acks string,
	topics []string,
	logger logger.Logger,
) Kafka {
	client := &robustKafkaClient{
		logger:           logger,
		eventHandlers:    make(map[string]EventHandler),
		bootstrapServer:  bootstrapServer,
		clusterAPIKey:    clusterAPIKey,
		clusterAPISecret: clusterAPISecret,
		securityProtocol: securityProtocol,
		mechanisms:       mechanisms,
		acks:             acks,
		topics:           topics,
	}

	// Try to initialize Kafka producer, but don't fail if it doesn't work
	client.tryInitializeKafka()
	return client
}

func (r *robustKafkaClient) tryInitializeKafka() {
	// Check if Kafka configuration is valid
	if r.bootstrapServer == "" {
		r.logger.Warn(context.Background(), "Kafka bootstrap server not configured, using mock mode")
		return
	}

	// Initialize Kafka producer
	config := sarama.NewConfig()
	config.Version = sarama.V2_8_0_0

	// Configure producer
	config.Producer.RequiredAcks = func() sarama.RequiredAcks {
		switch r.acks {
		case "0":
			return sarama.NoResponse
		case "1":
			return sarama.WaitForLocal
		case "all":
			return sarama.WaitForAll
		default:
			return sarama.WaitForAll
		}
	}()
	config.Producer.MaxMessageBytes = 1000000 // 1MB
	config.Producer.Compression = sarama.CompressionSnappy
	config.Producer.Idempotent = true

	// Configure consumer group
	config.Consumer.Group.Session.Timeout = 20 * time.Second
	config.Consumer.Group.Heartbeat.Interval = 3 * time.Second
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin

	// Configure SASL if credentials are provided
	if r.clusterAPIKey != "" && r.clusterAPISecret != "" {
		config.Net.SASL.Enable = true
		config.Net.SASL.Mechanism = func() sarama.SASLMechanism {
			switch r.mechanisms {
			case "PLAIN":
				return sarama.SASLTypePlaintext
			case "SCRAM-SHA-256":
				return sarama.SASLTypeSCRAMSHA256
			case "SCRAM-SHA-512":
				return sarama.SASLTypeSCRAMSHA512
			default:
				return sarama.SASLTypePlaintext
			}
		}()
		config.Net.SASL.User = r.clusterAPIKey
		config.Net.SASL.Password = r.clusterAPISecret
	}

	// Configure security protocol
	switch r.securityProtocol {
	case "SASL_SSL":
		config.Net.TLS.Enable = true
		config.Net.TLS.Config = &tls.Config{
			InsecureSkipVerify: false,
		}
	case "SASL_PLAINTEXT":
		config.Net.TLS.Enable = false
	case "PLAINTEXT":
		config.Net.TLS.Enable = false
	default:
		config.Net.TLS.Enable = false
	}

	// Initialize producer
	producer, err := sarama.NewAsyncProducer([]string{r.bootstrapServer}, config)
	if err != nil {
		r.logger.Error(context.Background(), "Failed to create Kafka producer", zap.Error(err))
		r.producer = nil
		return
	}

	// Handle producer errors
	go func() {
		for err := range producer.Errors() {
			r.logger.Error(context.Background(), "Kafka producer error", zap.Error(err))
		}
	}()

	r.producer = producer
	r.logger.Info(context.Background(), "Kafka producer initialized successfully")

	// Initialize consumer group
	consumerGroup, err := sarama.NewConsumerGroup([]string{r.bootstrapServer}, "tucanbit-group", config)
	if err != nil {
		r.logger.Error(context.Background(), "Failed to create Kafka consumer group", zap.Error(err))
		r.consumerGroup = nil
		return
	}

	r.consumerGroup = consumerGroup
	r.logger.Info(context.Background(), "Kafka consumer group initialized successfully")
}

func (r *robustKafkaClient) WriteMessage(ctx context.Context, topic, key string, value interface{}) error {
	if r.producer == nil {
		// Fallback to mock mode if producer is not initialized
		messageBytes, _ := json.Marshal(value)
		r.logger.Info(ctx, "Kafka message write attempted (mock mode)",
			zap.String("topic", topic),
			zap.String("key", key),
			zap.String("message", string(messageBytes)))
		return nil
	}

	// Serialize the message
	messageBytes, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Create Kafka message
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(messageBytes),
		Headers: []sarama.RecordHeader{
			{
				Key:   []byte("timestamp"),
				Value: []byte(time.Now().UTC().Format(time.RFC3339)),
			},
		},
	}

	// Send message asynchronously
	r.producer.Input() <- msg

	r.logger.Info(ctx, "Kafka message sent successfully",
		zap.String("topic", topic),
		zap.String("key", key),
		zap.Int("message_size", len(messageBytes)))

	return nil
}

func (r *robustKafkaClient) RegisterKafkaEventHandler(eventType string, handler EventHandler) {
	if _, exists := r.eventHandlers[eventType]; exists {
		r.logger.Warn(context.Background(), "Event handler already registered for event type",
			zap.String("event_type", eventType))
		return
	}
	r.eventHandlers[eventType] = handler
	r.logger.Info(context.Background(), "Registered Kafka event handler",
		zap.String("event_type", eventType))
}

func (r *robustKafkaClient) StartConsumer(ctx context.Context) {
	if r.consumerGroup == nil {
		r.logger.Info(ctx, "Kafka consumer not initialized, running in mock mode")
		return
	}

	// Start consuming messages
	go func() {
		for {
			topics := r.topics
			if len(topics) == 0 {
				topics = []string{"events"} // Default topic
			}

			err := r.consumerGroup.Consume(ctx, topics, r)
			if err != nil {
				r.logger.Error(ctx, "Error from consumer", zap.Error(err))
			}

			// Check if context is cancelled
			if ctx.Err() != nil {
				return
			}
		}
	}()

	r.logger.Info(ctx, "Kafka consumer started successfully", zap.Strings("topics", r.topics))
}

// ConsumeClaim implements sarama.ConsumerGroupHandler
func (r *robustKafkaClient) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		r.logger.Info(context.Background(), "Received Kafka message",
			zap.String("topic", message.Topic),
			zap.Int32("partition", message.Partition),
			zap.Int64("offset", message.Offset),
			zap.String("key", string(message.Key)),
			zap.String("value", string(message.Value)))

		// Process the message based on registered handlers
		if handler, exists := r.eventHandlers[message.Topic]; exists {
			success, err := handler(context.Background(), message.Value)
			if err != nil {
				r.logger.Error(context.Background(), "Error processing message",
					zap.Error(err),
					zap.String("topic", message.Topic))
			}
			if success {
				session.MarkMessage(message, "")
			}
		} else {
			r.logger.Warn(context.Background(), "No handler registered for topic",
				zap.String("topic", message.Topic))
			// Mark as processed even without handler to avoid blocking
			session.MarkMessage(message, "")
		}
	}

	return nil
}

// Setup implements sarama.ConsumerGroupHandler
func (r *robustKafkaClient) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup implements sarama.ConsumerGroupHandler
func (r *robustKafkaClient) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func InitPlatform(
	_ context.Context,
	logger logger.Logger,
) Platform {
	// Initialize Kafka with robust configuration handling
	logger.Info(context.Background(), "Initializing Kafka with robust configuration")

	// Get Kafka configuration with fallbacks
	bootstrapServer := viper.GetString("kafka.bootstrap_server")
	if bootstrapServer == "" {
		bootstrapServer = viper.GetString("kafka.bootstrap_servers") // Try alternative key
	}

	clusterAPIKey := viper.GetString("kafka.cluster_api_key")
	clusterAPISecret := viper.GetString("kafka.cluster_api_secret")
	securityProtocol := viper.GetString("kafka.security_protocol")
	if securityProtocol == "" {
		securityProtocol = "PLAINTEXT" // Default fallback
	}
	mechanisms := viper.GetString("kafka.mechanisms")
	if mechanisms == "" {
		mechanisms = "PLAIN" // Default fallback
	}
	acks := viper.GetString("kafka.acks")
	if acks == "" {
		acks = "all" // Default fallback
	}

	// Get topics with fallback
	topics := viper.GetStringSlice("kafka.topics")
	if len(topics) == 0 {
		// Try to get as a single string and split it
		topicStr := viper.GetString("kafka.topics")
		if topicStr == "" {
			// Try alternative key
			topicStr = viper.GetString("kafka.topic")
		}
		if topicStr == "" {
			// Use default topic
			topicStr = "events"
			logger.Info(context.Background(), "Using default Kafka topic: events")
		}
		// Split by comma if it's a comma-separated string
		topics = strings.Split(topicStr, ",")
		// Trim spaces from each topic
		for i, topic := range topics {
			topics[i] = strings.TrimSpace(topic)
		}
	}

	// Create robust Kafka client
	kafkaClient := newRobustKafkaClient(
		bootstrapServer,
		clusterAPIKey,
		clusterAPISecret,
		securityProtocol,
		mechanisms,
		acks,
		topics,
		logger,
	)

	logger.Info(context.Background(), "Initializing redis connection")
	redisAddr := viper.GetString("redis.addr")
	if redisAddr == "" {
		logger.Fatal(context.Background(), "redis.addr config is empty")
	}
	redisPassword := viper.GetString("redis.password")
	if redisPassword == "" {
		logger.Fatal(context.Background(), "redis.password config is empty")
	}
	redisDB := viper.GetInt("redis.db")
	if redisDB < 0 {
		logger.Fatal(context.Background(), "redis.db config is empty")
	}
	redisKeyPrefix := viper.GetString("redis.key_prefix")
	if redisKeyPrefix == "" {
		logger.Fatal(context.Background(), "redis.key_prefix config is empty")
	}
	redisTTL := viper.GetDuration("redis.ttl")
	if redisTTL == time.Duration(0) {
		logger.Fatal(context.Background(), "redis.ttl config is empty")
	}
	redisAttempts := viper.GetInt("redis.attempts")
	if redisAttempts < 0 {
		logger.Fatal(context.Background(), "redis.attempts config is empty")
	}
	redis, err := redis.NewRedisOTP(redisAddr, redisPassword, redisDB, redisKeyPrefix, redisTTL, redisAttempts, logger)
	if err != nil {
		logger.Fatal(context.Background(), "failed to initialize redis", zap.Error(err))
	}

	// Initialize Pisi client
	pisiBaseURL := viper.GetString("pisi.base_url")
	if pisiBaseURL == "" {
		logger.Fatal(context.Background(), "pisi.base_url config is empty")
	}
	pisiPassword := viper.GetString("pisi.password")
	if pisiPassword == "" {
		logger.Fatal(context.Background(), "pisi.password config is empty")
	}

	pisiVaspid := viper.GetString("pisi.vaspid")
	if pisiVaspid == "" {
		logger.Fatal(context.Background(), "pisi.vaspid config is empty")
	}
	// retry_count
	pisiRetryCount := viper.GetInt("pisi.retry_count")
	if pisiRetryCount < 0 {
		logger.Fatal(context.Background(), "pisi.retry_count config is empty")
	}

	// sender_id
	pisiSenderID := viper.GetString("pisi.sender_id")
	if pisiSenderID == "" {
		logger.Fatal(context.Background(), "pisi.sender_id config is empty")
	}

	// retry_delay
	pisiRetryDelay := viper.GetDuration("pisi.retry_delay")
	if pisiRetryDelay <= 0 {
		logger.Fatal(context.Background(), "pisi.retry_delay config is empty")
	}

	pisiClient := pisi.NewPisiClient(pisiBaseURL, pisiPassword, pisiVaspid,
		viper.GetDuration("pisi.timeout"),
		pisiRetryCount, pisiRetryDelay, pisiSenderID, logger)

	return Platform{
		Kafka: kafkaClient,
		Redis: redis,
		Pisi:  pisiClient,
	}

}
