package kafka

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/constant/types"
	"github.com/tucanbit/platform/logger"

	"github.com/IBM/sarama"
	"go.uber.org/zap"
)

type KafkaClient interface {
	RegisterKafkaEventHandler(EventType string, handler types.EventHandler)
	Route(ctx context.Context, message *sarama.ConsumerMessage)
	StartConsumer(ctx context.Context)
	WriteMessage(ctx context.Context, topic, key string, value interface{}) error
	Close() error
}

type kafkaController struct {
	BootstrapServer  string
	ClusterAPIKey    string
	ClusterAPISecret string
	SecurityProtocol string
	Mechanisms       string
	Acks             string
	Logger           logger.Logger
	topics           []string
	eventHandlers    map[string]types.EventHandler
	kafkaProducer    sarama.AsyncProducer
	kafkaConsumer    sarama.ConsumerGroup
}

func NewKafkaClient(bootStrapServer,
	ClusterAPIKey,
	ClusterAPISecret,
	SecurityProtocol,
	Mechanisms,
	acks string,
	logger *logger.Logger,
	topics []string,
) KafkaClient {
	var err error
	kf := &kafkaController{
		BootstrapServer:  bootStrapServer,
		ClusterAPIKey:    ClusterAPIKey,
		ClusterAPISecret: ClusterAPISecret,
		SecurityProtocol: SecurityProtocol,
		Mechanisms:       Mechanisms,
		Acks:             acks,
		Logger:           *logger,
		topics:           topics,
		eventHandlers:    make(map[string]types.EventHandler),
	}

	// Create Sarama configuration
	config := sarama.NewConfig()
	config.Version = sarama.V2_8_0_0

	// Configure producer
	config.Producer.RequiredAcks = func() sarama.RequiredAcks {
		switch acks {
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
	config.Net.MaxOpenRequests = 1 // Required for idempotent producer

	// Configure consumer group
	config.Consumer.Group.Session.Timeout = 20 * time.Second
	config.Consumer.Group.Heartbeat.Interval = 3 * time.Second
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin

	// Configure SASL if credentials are provided
	if ClusterAPIKey != "" && ClusterAPISecret != "" {
		config.Net.SASL.Enable = true
		config.Net.SASL.Mechanism = func() sarama.SASLMechanism {
			switch Mechanisms {
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
		config.Net.SASL.User = ClusterAPIKey
		config.Net.SASL.Password = ClusterAPISecret
	}

	// Initialize producer
	kf.kafkaProducer, err = sarama.NewAsyncProducer([]string{bootStrapServer}, config)
	if err != nil {
		kf.Logger.Fatal(context.Background(), "Failed to create Kafka producer", zap.Error(err))
	}

	// Handle producer errors
	go func() {
		for err := range kf.kafkaProducer.Errors() {
			kf.Logger.Error(context.Background(), "Kafka producer error", zap.Error(err))
		}
	}()

	// Initialize consumer group
	kf.kafkaConsumer, err = sarama.NewConsumerGroup([]string{bootStrapServer}, "tucanbit-group", config)
	if err != nil {
		kf.Logger.Fatal(context.Background(), "Failed to create Kafka consumer group", zap.Error(err))
	}

	go kf.StartConsumer(context.Background())
	return kf
}

func (k *kafkaController) Close() error {
	var err error
	if k.kafkaProducer != nil {
		k.kafkaProducer.Close()
	}
	if k.kafkaConsumer != nil {
		err = k.kafkaConsumer.Close()
	}
	return err
}

func (k *kafkaController) RegisterKafkaEventHandler(topic string, handler types.EventHandler) {
	if _, exists := k.eventHandlers[topic]; exists {
		k.Logger.Warn(context.Background(), "Event handler already registered for event type", zap.String("event_type", topic))
		return
	}
	k.eventHandlers[topic] = handler
	k.Logger.Info(context.Background(), "Registered Kafka event handler", zap.String("event_type", topic))
}

func (k *kafkaController) WriteMessage(ctx context.Context, topic, key string, value interface{}) error {
	if k.kafkaProducer == nil {
		return errors.ErrKafkaEventNotSupported.New("Kafka producer is not initialized")
	}
	message, err := json.Marshal(value)
	if err != nil {
		return errors.ErrInvalidUserInput.Wrap(err, "Failed to marshal message value to JSON")
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(message),
		Headers: []sarama.RecordHeader{
			{
				Key:   []byte("timestamp"),
				Value: []byte(time.Now().UTC().Format(time.RFC3339)),
			},
		},
	}

	k.kafkaProducer.Input() <- msg
	k.Logger.Info(ctx, "Message sent to Kafka", zap.String("topic", topic), zap.String("key", key))
	return nil
}

func (k *kafkaController) Route(ctx context.Context, message *sarama.ConsumerMessage) {
	eventType := message.Topic
	if handler, exists := k.eventHandlers[eventType]; exists {
		handled, err := handler(ctx, message.Value)
		if err != nil {
			k.Logger.Error(ctx, "Error handling Kafka message", zap.Error(err))
		} else if handled {
			k.Logger.Info(ctx, "Message handled successfully", zap.String("event_type", eventType))
		}
	} else {
		k.Logger.Warn(ctx, "No handler registered for event type", zap.String("event_type", eventType))
	}
}

func (k *kafkaController) StartConsumer(ctx context.Context) {
	if k.kafkaConsumer == nil {
		k.Logger.Error(ctx, "Kafka consumer is not initialized")
		return
	}

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	// Start consuming messages
	go func() {
		for {
			topics := k.topics
			if len(topics) == 0 {
				topics = []string{"events"} // Default topic
			}

			err := k.kafkaConsumer.Consume(ctx, topics, k)
			if err != nil {
				k.Logger.Error(ctx, "Error from consumer", zap.Error(err))
			}

			// Check if context is cancelled
			if ctx.Err() != nil {
				return
			}
		}
	}()

	k.Logger.Info(ctx, "Kafka consumer started successfully", zap.Strings("topics", k.topics))
}

// ConsumeClaim implements sarama.ConsumerGroupHandler
func (k *kafkaController) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		k.Logger.Info(context.Background(), "Received Kafka message",
			zap.String("topic", message.Topic),
			zap.Int32("partition", message.Partition),
			zap.Int64("offset", message.Offset),
			zap.String("key", string(message.Key)),
			zap.String("value", string(message.Value)))

		// Process the message based on registered handlers
		k.Route(context.Background(), message)
		session.MarkMessage(message, "")
	}

	return nil
}

// Setup implements sarama.ConsumerGroupHandler
func (k *kafkaController) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup implements sarama.ConsumerGroupHandler
func (k *kafkaController) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}
