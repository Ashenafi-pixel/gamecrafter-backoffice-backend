package platform

import (
	"context"
	"encoding/json"
	"time"

	"github.com/tucanbit/platform/kafka"

	"github.com/tucanbit/internal/constant/types"
	"github.com/tucanbit/internal/contracts"
	"github.com/tucanbit/platform/logger"
	"github.com/tucanbit/platform/pisi"
	"github.com/tucanbit/platform/redis"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Kafka interface {
	WriteMessage(ctx context.Context, topic, key string, value interface{}) error
	RegisterKafkaEventHandler(eventType string, handler types.EventHandler)
	StartConsumer(ctx context.Context)
}

type EventHandler func(ctx context.Context, event json.RawMessage) (bool, error)

type Event struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type Redis interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
}

type Pisi interface {
	SendSMS(ctx context.Context, phoneNumber, message string) error
	SendBulkSMS(ctx context.Context, req pisi.SendBulkSMSRequest) (string, error)
	Login(ctx context.Context) (*pisi.LoginResponse, error)
}

type Platform struct {
	Kafka Kafka
	Redis contracts.Redis
	Pisi  Pisi
}

func InitPlatform(ctx context.Context, lgr logger.Logger) *Platform {
	// Initialize Kafka i removed getKafkaConfig and inlined its logic
	topic := viper.GetString("kafka.topic")
	if topic == "" {
		lgr.Fatal(ctx, "kafka.topic config is empty")
	}
	kafkaClient := kafka.NewKafkaClient(
		viper.GetString("kafka.bootstrap_servers"),
		viper.GetString("kafka.cluster_api_key"),
		viper.GetString("kafka.cluster_api_secret"),
		viper.GetString("kafka.security_protocol"),
		viper.GetString("kafka.mechanisms"),
		viper.GetString("kafka.acks"),
		&lgr,
		[]string{topic},
	)

	// Initialize Redis
	redisClient, err := redis.NewRedisOTP(
		viper.GetString("redis.addr"),
		viper.GetString("redis.password"),
		viper.GetInt("redis.db"),
		viper.GetString("redis.key_prefix"),
		viper.GetDuration("redis.ttl"),
		viper.GetInt("auth.otp_attempt_limit"),
		lgr,
	)
	if err != nil {
		lgr.Fatal(ctx, "Failed to initialize Redis client", zap.Error(err))
	}

	// Initialize Pisi
	pisiClient := pisi.NewPisiClient(
		viper.GetString("pisi.base_url"),
		viper.GetString("pisi.password"),
		viper.GetString("pisi.vaspid"),
		viper.GetDuration("pisi.timeout"),
		viper.GetInt("pisi.retry_count"),
		viper.GetDuration("pisi.retry_delay"),
		viper.GetString("pisi.sender_id"),
		lgr,
	)

	return &Platform{
		Kafka: kafkaClient,
		Redis: redisClient,
		Pisi:  pisiClient,
	}
}
