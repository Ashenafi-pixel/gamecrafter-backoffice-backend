package interfaces

import (
	"context"
	"time"
)

// Kafka interface for message broker operations
type Kafka interface {
	WriteMessage(ctx context.Context, topic, key string, value interface{}) error
	RegisterKafkaEventHandler(eventType string, handler EventHandler)
	StartConsumer(ctx context.Context)
}

// EventHandler for Kafka event processing
type EventHandler func(ctx context.Context, message []byte) (bool, error)

// Redis interface for caching operations
type Redis interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
}

// Pisi interface for SMS operations
type Pisi interface {
	SendSMS(ctx context.Context, phoneNumber, message string) error
	SendBulkSMS(ctx context.Context, req interface{}) (string, error)
	Login(ctx context.Context) (interface{}, error)
} 