package redis

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"os"
	"time"

	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/contracts"
	"github.com/tucanbit/platform/logger"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func NewRedisOTP(redisAddr, redisPassword string, db int, keyPrefix string, otpTTL time.Duration, attemptLimit int, logger logger.Logger) (contracts.Redis, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		err = errors.ErrRedisClientError.Wrap(err, "failed to connect to Redis")
		logger.Error(ctx, "failed to connect to Redis", zap.Error(err))
		return nil, err
	}

	privateKey, err := loadRSAKeysFromFiles(logger)
	if err != nil {
		logger.Info(ctx, "Failed to load existing RSA keys from files, generating new ones", zap.Error(err))
		privateKey, err = rsa.GenerateKey(rand.Reader, 4096)
		if err != nil {
			err = errors.ErrRedisClientError.Wrap(err, "failed to generate RSA key pair")
			logger.Error(ctx, "failed to generate RSA key pair", zap.Error(err))
			return nil, err
		}
		if err := storeRSAKeysToFiles(privateKey, logger); err != nil {
			logger.Error(ctx, "Failed to store RSA keys to files", zap.Error(err))
		}
	} else {
		logger.Info(ctx, "Successfully loaded existing RSA keys from project files")
	}

	logger.Info(ctx, "Redis OTP client initialized successfully",
		zap.String("redis_addr", redisAddr),
		zap.String("key_prefix", keyPrefix),
		zap.Duration("otp_ttl", otpTTL),
	)

	return &RedisOTP{
		redisClient:  rdb,
		privateKey:   privateKey,
		publicKey:    &privateKey.PublicKey,
		keyPrefix:    keyPrefix,
		otpTTL:       otpTTL,
		logger:       logger,
		AttemptLimit: attemptLimit,
	}, nil
}

func NewRedisOTPFromConfig(logger logger.Logger) (contracts.Redis, error) {
	redisAddr := viper.GetString("redis.addr")
	redisPassword := viper.GetString("redis.password")
	db := viper.GetInt("redis.db")
	keyPrefix := viper.GetString("redis.key_prefix")
	otpTTL := viper.GetDuration("redis.ttl")
	if otpTTL == 0 {
		otpTTL = 5 * time.Second
	}
	attemptLimit := viper.GetInt("auth.otp_attempt_limit")
	if attemptLimit == 0 {
		attemptLimit = 4
	}
	return NewRedisOTP(redisAddr, redisPassword, db, keyPrefix, otpTTL, attemptLimit, logger)
}

func (r *RedisOTP) Get(ctx context.Context, key string) (string, error) {
	fullKey := r.otpKey(key)
	val, err := r.redisClient.Get(ctx, fullKey).Result()
	if err == redis.Nil {
		return "", errors.ErrRedisClientError.New("key not found")
	}
	if err != nil {
		r.logger.Error(ctx, "Redis GET failed",
			zap.String("key", fullKey),
			zap.Error(err))
		return "", errors.ErrRedisClientError.Wrap(err, "failed to get key")
	}
	return val, nil
}

func (r *RedisOTP) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	fullKey := r.otpKey(key)
	err := r.redisClient.Set(ctx, fullKey, value, expiration).Err()
	if err != nil {
		r.logger.Error(ctx, "Redis SET failed",
			zap.String("key", fullKey),
			zap.Error(err))
		return errors.ErrRedisClientError.Wrap(err, "failed to set key")
	}
	return nil
}

func (r *RedisOTP) Delete(ctx context.Context, key string) error {
	fullKey := r.otpKey(key)
	err := r.redisClient.Del(ctx, fullKey).Err()
	if err != nil {
		r.logger.Error(ctx, "Redis DEL failed",
			zap.String("key", fullKey),
			zap.Error(err))
		return errors.ErrRedisClientError.Wrap(err, "failed to delete key")
	}
	return nil
}

func (r *RedisOTP) SaveOTP(ctx context.Context, phone, otp string) error {
	redisKey := r.otpKey(phone)
	encryptedOTP, err := r.encryptOTP(otp)
	if err != nil {
		r.logger.Error(ctx, "encryptOTP failed", zap.Error(err))
		return err
	}

	mapping := OTPMapping{
		EncryptedOTP: encryptedOTP,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(r.otpTTL),
	}

	mappingJSON, err := json.Marshal(mapping)
	if err != nil {
		r.logger.Error(ctx, "marshal OTPMapping failed", zap.Error(err))
		return err
	}

	if err := r.Set(ctx, phone, string(mappingJSON), r.otpTTL); err != nil {
		return err
	}

	attemptKey := redisKey + ":attempts"
	if err := r.Set(ctx, attemptKey, 0, r.otpTTL); err != nil {
		return err
	}

	return nil
}

func (r *RedisOTP) VerifyAndRemoveOTP(ctx context.Context, phone, otp string) (bool, error) {
	redisKey := r.otpKey(phone)
	attempts, err := r.GetOTPAttemptCount(ctx, phone)
	if err != nil {
		return false, err
	}
	if attempts >= r.AttemptLimit {
		err = errors.ErrAcessError.New("Too many invalid attempts. Please request a new OTP.")
		r.logger.Warn(ctx, "OTP attempts exceeded", zap.String("phone", phone))
		_ = r.Delete(ctx, redisKey)
		_ = r.Delete(ctx, redisKey+":attempts")
		return false, err
	}

	mappingJSON, err := r.Get(ctx, phone)
	if err != nil {
		_, _ = r.IncrementOTPAttempt(ctx, phone)
		return false, errors.ErrAcessError.New("otp expired or invalid")
	}

	var mapping OTPMapping
	if err := json.Unmarshal([]byte(mappingJSON), &mapping); err != nil {
		_, _ = r.IncrementOTPAttempt(ctx, phone)
		return false, errors.ErrAcessError.Wrap(err, "invalid OTP format")
	}

	if time.Now().After(mapping.ExpiresAt) {
		_ = r.Delete(ctx, redisKey)
		_, _ = r.IncrementOTPAttempt(ctx, phone)
		return false, nil
	}

	decryptedOTP, err := r.decryptOTP(mapping.EncryptedOTP)
	if err != nil {
		_, _ = r.IncrementOTPAttempt(ctx, phone)
		return false, errors.ErrAcessError.Wrap(err, "failed to decrypt OTP")
	}

	if decryptedOTP != otp {
		_, _ = r.IncrementOTPAttempt(ctx, phone)
		return false, errors.ErrAcessError.New("invalid OTP")
	}

	_ = r.Delete(ctx, redisKey)
	return true, nil
}

func (r *RedisOTP) IncrementOTPAttempt(ctx context.Context, phone string) (int, error) {
	redisKey := r.otpKey(phone) + ":attempts"
	count, err := r.redisClient.Incr(ctx, redisKey).Result()
	if err != nil {
		r.logger.Error(ctx, "Redis INCR attempt count failed", zap.Error(err))
		return 0, err
	}
	if ttl, _ := r.redisClient.TTL(ctx, redisKey).Result(); ttl < 0 {
		r.redisClient.Expire(ctx, redisKey, 5*time.Minute)
	}
	return int(count), nil
}

func (r *RedisOTP) GetOTPAttemptCount(ctx context.Context, phone string) (int, error) {
	redisKey := r.otpKey(phone) + ":attempts"
	count, err := r.redisClient.Get(ctx, redisKey).Int()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		r.logger.Error(ctx, "Redis GET attempt count failed", zap.Error(err))
		return 0, err
	}
	return count, nil
}

func (r *RedisOTP) ResetOTPAttempts(ctx context.Context, phone string) error {
	redisKey := r.otpKey(phone) + ":attempts"
	return r.Set(ctx, redisKey, 0, r.otpTTL)
}

func (r *RedisOTP) encryptOTP(otp string) (string, error) {
	otpBytes := []byte(otp)
	encryptedBytes, err := rsa.EncryptPKCS1v15(rand.Reader, r.publicKey, otpBytes)
	if err != nil {
		return "", errors.ErrRedisClientError.Wrap(err, "failed to encrypt OTP")
	}
	return base64.StdEncoding.EncodeToString(encryptedBytes), nil
}

func (r *RedisOTP) decryptOTP(encryptedOTP string) (string, error) {
	encryptedBytes, err := base64.StdEncoding.DecodeString(encryptedOTP)
	if err != nil {
		return "", errors.ErrRedisClientError.Wrap(err, "failed to decode OTP")
	}
	decryptedBytes, err := rsa.DecryptPKCS1v15(rand.Reader, r.privateKey, encryptedBytes)
	if err != nil {
		return "", errors.ErrRedisClientError.Wrap(err, "failed to decrypt OTP")
	}
	return string(decryptedBytes), nil
}

func (r *RedisOTP) otpKey(phone string) string {
	return fmt.Sprintf("%s:%s", r.keyPrefix, phone)
}

// GetKeyPrefix returns the key prefix used by this Redis client
func (r *RedisOTP) GetKeyPrefix() string {
	return r.keyPrefix
}

func (r *RedisOTP) GetOTP(ctx context.Context, phone string) (string, error) {
	mappingJSON, err := r.Get(ctx, phone)
	if err != nil {
		return "", err
	}

	var mapping OTPMapping
	if err := json.Unmarshal([]byte(mappingJSON), &mapping); err != nil {
		return "", errors.ErrRedisClientError.Wrap(err, "invalid OTP format")
	}

	if time.Now().After(mapping.ExpiresAt) {
		return "", errors.ErrRedisClientError.New("OTP expired")
	}

	return r.decryptOTP(mapping.EncryptedOTP)
}

func loadRSAKeysFromFiles(logger logger.Logger) (*rsa.PrivateKey, error) {
	privateKeyPath := "config/rsa_private_key.pem"
	privateKeyPEM, err := os.ReadFile(privateKeyPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.ErrRedisClientError.New("RSA private key file not found")
		}
		return nil, errors.ErrRedisClientError.Wrap(err, "failed to read RSA private key file")
	}

	block, _ := pem.Decode(privateKeyPEM)
	if block == nil {
		return nil, errors.ErrRedisClientError.New("failed to decode private key PEM")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, errors.ErrRedisClientError.Wrap(err, "failed to parse private key")
	}

	logger.Debug(context.Background(), "Loaded RSA private key from file", zap.String("path", privateKeyPath))
	return privateKey, nil
}

func storeRSAKeysToFiles(privateKey *rsa.PrivateKey, logger logger.Logger) error {
	privateKeyPath := "config/rsa_private_key.pem"
	publicKeyPath := "config/rsa_public_key.pem"

	if err := os.MkdirAll("config", 0755); err != nil {
		return errors.ErrRedisClientError.Wrap(err, "failed to create config directory")
	}

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return errors.ErrRedisClientError.Wrap(err, "failed to marshal public key")
	}

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	if err := os.WriteFile(privateKeyPath, privateKeyPEM, 0600); err != nil {
		return errors.ErrRedisClientError.Wrap(err, "failed to write private key")
	}

	if err := os.WriteFile(publicKeyPath, publicKeyPEM, 0644); err != nil {
		return errors.ErrRedisClientError.Wrap(err, "failed to write public key")
	}

	logger.Info(context.Background(), "Stored RSA keys to files",
		zap.String("private_key", privateKeyPath),
		zap.String("public_key", publicKeyPath))
	return nil
}
