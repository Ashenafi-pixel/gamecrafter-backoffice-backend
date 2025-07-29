package kafka

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/joshjones612/egyptkingcrash/internal/constant/errors"
	"github.com/joshjones612/egyptkingcrash/platform"
	"github.com/joshjones612/egyptkingcrash/platform/logger"
	"go.uber.org/zap"
)

type KafkaClient interface {
	RegisterKafkaEventHandler(EventType string, handler platform.EventHandler)
	Close() error
	WriteMessage(ctx context.Context, value interface{}) error
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
	eventHandlers    map[string]platform.EventHandler
	kafkaProducer    *kafka.Producer
}

func NewKafkaClient(bootStrapServer,
	ClusterAPIKey,
	ClusterAPISecret,
	SecurityProtocol,
	Mechanisms,
	acks string,
	logger *logger.Logger,
	topics []string,
) platform.Kafka {
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
		eventHandlers:    make(map[string]platform.EventHandler),
	}

	kf.kafkaProducer, err = kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": kf.BootstrapServer,
		"sasl.username":     kf.ClusterAPIKey,
		"sasl.password":     kf.ClusterAPISecret,
		"security.protocol": kf.SecurityProtocol,
		"sasl.mechanisms":   kf.Mechanisms,
		"acks":              kf.Acks,
	})
	
	if err != nil {
		kf.Logger.Fatal(context.Background(), "Failed to create Kafka producer", zap.Error(err))
	}

	go kf.StartConsumer(context.Background())
	return kf
}

func (k *kafkaController) RegisterKafkaEventHandler(topic string, handler platform.EventHandler) {
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
	deliveryChan := make(chan kafka.Event, 1)
	err = k.kafkaProducer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Key:            []byte(key),
		Value:          message,
	}, deliveryChan)
	if err != nil {
		return errors.ErrKafkaEventNotSupported.Wrap(err, "Failed to produce message to Kafka")
	}
	e := <-deliveryChan
	if msg, ok := e.(*kafka.Message); ok {
		if msg.TopicPartition.Error != nil {
			return errors.ErrKafkaEventNotSupported.Wrap(err, "Failed to deliver message to Kafka")

		}
		k.Logger.Info(ctx, "Message delivered to Kafka", zap.String("topic", topic), zap.String("key", key))
	} else {
		return errors.ErrKafkaEventNotSupported.Wrap(err, "Unexpected event type received from Kafka delivery channel")
	}
	return nil
}

func (k *kafkaController) Route(ctx context.Context, message kafka.Message) {
	eventType := string(*message.TopicPartition.Topic)
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
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": k.BootstrapServer,
		"sasl.username":     k.ClusterAPIKey,
		"sasl.password":     k.ClusterAPISecret,
		"security.protocol": k.SecurityProtocol,
		"sasl.mechanisms":   k.Mechanisms,
		"group.id":          "lottery.draw.sync",
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		k.Logger.Fatal(ctx, "Failed to create consumer", zap.Error(err))
	}
	defer c.Close()
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	err = c.SubscribeTopics(k.topics, nil)
	if err != nil {
		k.Logger.Fatal(ctx, "Failed to subscribe to topics", zap.Error(err))
	}

	for {
		select {
		case sig := <-sigchan:
			k.Logger.Info(ctx, "Received signal, shutting down consumer", zap.String("signal", sig.String()))
			return
		case <-ctx.Done():
			k.Logger.Info(ctx, "Context canceled, shutting down consumer")
			return
		default:
			ev, err := c.ReadMessage(time.Second)
			if err != nil {
				if err.(kafka.Error).Code() == kafka.ErrTimedOut {
					continue
				}
				k.Logger.Error(ctx, "Error reading message", zap.Error(err))
				continue
			}
			k.Logger.Info(ctx, "Consumed event from topic", zap.String("topic", *ev.TopicPartition.Topic))
			k.Route(ctx, *ev)
		}
	}
}
