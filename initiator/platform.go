package initiator

import (
	"context"
	"time"

	"github.com/joshjones612/egyptkingcrash/platform"
	"github.com/joshjones612/egyptkingcrash/platform/kafka"
	"github.com/joshjones612/egyptkingcrash/platform/logger"
	"github.com/joshjones612/egyptkingcrash/platform/pisi"
	"github.com/joshjones612/egyptkingcrash/platform/redis"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Platform struct {
	Kafka platform.Kafka
	Redis *redis.RedisOTP
	Pisi  pisi.PisiClient
}

func InitPlatform(
	_ context.Context,
	logger logger.Logger,
) Platform {
	topicName := viper.GetStringSlice("kafka.topics")
	if len(topicName) == 0 {
		logger.Fatal(context.Background(), "kafka.topic config is empty")
	}
	logger.Info(context.Background(), "Initializing Kafka connection")

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
		Kafka: kafka.NewKafkaClient(
			viper.GetString("kafka.bootstrap_server"),
			viper.GetString("kafka.cluster_api_key"),
			viper.GetString("kafka.cluster_api_secret"),
			viper.GetString("kafka.security_protocol"),
			viper.GetString("kafka.mechanisms"),
			viper.GetString("kafka.acks"),
			&logger,
			topicName,
		),
		Redis: redis,
		Pisi:  pisiClient,
	}

}
