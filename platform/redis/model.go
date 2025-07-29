package redis

import (
	"crypto/rsa"
	"time"

	"github.com/joshjones612/egyptkingcrash/platform/logger"
	"github.com/redis/go-redis/v9"
)

// RedisOTP handles OTP saving and verification using Redis with RSA encryption
type RedisOTP struct {
	redisClient  *redis.Client
	privateKey   *rsa.PrivateKey
	publicKey    *rsa.PublicKey
	keyPrefix    string
	otpTTL       time.Duration
	logger       logger.Logger
	AttemptLimit int // configurable attempt limit
}

type OTPMapping struct {
	UUID         string    `json:"uuid"`
	EncryptedOTP string    `json:"encrypted_otp"`
	CreatedAt    time.Time `json:"created_at"`
	ExpiresAt    time.Time `json:"expires_at"`
}
